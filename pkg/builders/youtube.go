package builders

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BrianHicks/finch/duration"
	itunes "github.com/mxpv/podcast"
	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
)

const (
	maxYoutubeResults       = 50
	hdBytesPerSecond        = 350000
	ldBytesPerSecond        = 100000
	lowAudioBytesPerSecond  = 48000 / 8
	highAudioBytesPerSecond = 128000 / 8
)

type apiKey string

func (key apiKey) Get() (string, string) {
	return "key", string(key)
}

type YouTubeBuilder struct {
	client *youtube.Service
	key    apiKey
}

// Cost: 5 units (call method: 1, snippet: 2, contentDetails: 2)
// See https://developers.google.com/youtube/v3/docs/channels/list#part
func (yt *YouTubeBuilder) listChannels(linkType api.LinkType, id string) (*youtube.Channel, error) {
	req := yt.client.Channels.List("id,snippet,contentDetails")

	if linkType == api.LinkTypeChannel {
		req = req.Id(id)
	} else if linkType == api.LinkTypeUser {
		req = req.ForUsername(id)
	} else {
		return nil, errors.New("unsupported link type")
	}

	resp, err := req.Do(yt.key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query channel")
	}

	if len(resp.Items) == 0 {
		return nil, api.ErrNotFound
	}

	item := resp.Items[0]
	return item, nil
}

// Cost: 3 units (call method: 1, snippet: 2)
// See https://developers.google.com/youtube/v3/docs/playlists/list#part
func (yt *YouTubeBuilder) listPlaylists(id, channelId string) (*youtube.Playlist, error) {
	req := yt.client.Playlists.List("id,snippet")

	if id != "" {
		req = req.Id(id)
	} else {
		req = req.ChannelId(channelId)
	}

	resp, err := req.Do(yt.key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query playlist")
	}

	if len(resp.Items) == 0 {
		return nil, api.ErrNotFound
	}

	item := resp.Items[0]
	return item, nil
}

// Cost: 3 units (call: 1, snippet: 2)
// See https://developers.google.com/youtube/v3/docs/playlistItems/list#part
func (yt *YouTubeBuilder) listPlaylistItems(itemId string, pageToken string) ([]*youtube.PlaylistItem, string, error) {
	req := yt.client.PlaylistItems.List("id,snippet").MaxResults(maxYoutubeResults).PlaylistId(itemId)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, err := req.Do(yt.key)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to query playlist items")
	}

	return resp.Items, resp.NextPageToken, nil
}

func (yt *YouTubeBuilder) parseDate(s string) (time.Time, error) {
	date, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "failed to parse date: %s", s)
	}

	return date, nil
}

func (yt *YouTubeBuilder) selectThumbnail(snippet *youtube.ThumbnailDetails, quality api.Quality) string {
	// Use high resolution thumbnails for high quality mode
	// https://github.com/mxpv/Podsync/issues/14
	if quality == api.QualityHigh {
		if snippet.Maxres != nil {
			return snippet.Maxres.Url
		}

		if snippet.High != nil {
			return snippet.High.Url
		}

		if snippet.Medium != nil {
			return snippet.Medium.Url
		}
	}

	return snippet.Default.Url
}

// Cost:
// - 5 units for channel or user
// - 3 units for playlist
func (yt *YouTubeBuilder) queryFeed(feed *model.Feed) (*itunes.Podcast, string, error) {
	var (
		title      string
		desc       string
		link       string // URL link to YouTube resource
		itemID     string // ID of YouTube's channel, user or playlist
		pubDate    time.Time
		thumbnails *youtube.ThumbnailDetails
	)

	switch feed.LinkType {
	case api.LinkTypeChannel, api.LinkTypeUser:
		channel, err := yt.listChannels(feed.LinkType, feed.ItemID)
		if err != nil {
			return nil, "", err
		}

		title = channel.Snippet.Title
		desc = channel.Snippet.Description

		if channel.Kind == "youtube#channel" {
			link = fmt.Sprintf("https://youtube.com/channel/%s", channel.Id)
		} else {
			link = fmt.Sprintf("https://youtube.com/user/%s", channel.Snippet.CustomUrl)
		}

		itemID = channel.ContentDetails.RelatedPlaylists.Uploads

		if date, err := yt.parseDate(channel.Snippet.PublishedAt); err != nil {
			return nil, "", err
		} else {
			pubDate = date
		}

		thumbnails = channel.Snippet.Thumbnails

	case api.LinkTypePlaylist:
		playlist, err := yt.listPlaylists(feed.ItemID, "")
		if err != nil {
			return nil, "", err
		}

		title = fmt.Sprintf("%s: %s", playlist.Snippet.ChannelTitle, playlist.Snippet.Title)
		desc = playlist.Snippet.Description

		link = fmt.Sprintf("https://youtube.com/playlist?list=%s", playlist.Id)

		itemID = playlist.Id

		if date, err := yt.parseDate(playlist.Snippet.PublishedAt); err != nil {
			return nil, "", err
		} else {
			pubDate = date
		}

		thumbnails = playlist.Snippet.Thumbnails

	default:
		return nil, "", errors.New("unsupported link format")
	}

	// Apply customizations and default values

	if desc == "" {
		desc = fmt.Sprintf("%s (%s)", title, pubDate)
	}

	// Build iTunes feed
	buildTime := time.Now()
	podcast := itunes.New(title, link, desc, &pubDate, &buildTime)
	podcast.Generator = podsyncGenerator

	podcast.AddSubTitle(title)
	podcast.AddCategory(defaultCategory, nil)
	podcast.AddImage(yt.selectThumbnail(thumbnails, feed.Quality))

	podcast.IAuthor = title
	podcast.AddSummary(desc)

	return &podcast, itemID, nil
}

// Video size information requires 1 additional call for each video (1 feed = 50 videos = 50 calls),
// which is too expensive, so get approximated size depending on duration and definition params
func (yt *YouTubeBuilder) getSize(duration int64, feed *model.Feed) int64 {
	if feed.Format == api.FormatAudio {
		if feed.Quality == api.QualityHigh {
			return highAudioBytesPerSecond * duration
		} else {
			return lowAudioBytesPerSecond * duration
		}
	} else {
		if feed.Quality == api.QualityHigh {
			return duration * hdBytesPerSecond
		} else {
			return duration * ldBytesPerSecond
		}
	}
}

// Cost: 5 units (call: 1, snippet: 2, contentDetails: 2)
// See https://developers.google.com/youtube/v3/docs/videos/list#part
func (yt *YouTubeBuilder) queryVideoDescriptions(playlistItems map[string]*youtube.PlaylistItemSnippet, feed *model.Feed, podcast *itunes.Podcast) error {
	// Make the list of video ids
	ids := make([]string, 0, len(playlistItems))
	for _, s := range playlistItems {
		ids = append(ids, s.ResourceId.VideoId)
	}

	req, err := yt.client.Videos.List("id,snippet,contentDetails").Id(strings.Join(ids, ",")).Do(yt.key)
	if err != nil {
		return errors.Wrap(err, "failed to query video descriptions")
	}

	for _, video := range req.Items {
		snippet := video.Snippet

		item := itunes.Item{
			GUID: video.Id,
			Link: fmt.Sprintf("https://youtube.com/watch?v=%s", video.Id),
			Title: snippet.Title,
			Description: snippet.Description,
			ISubtitle: snippet.Title,
		}

		desc := snippet.Description
		if desc == "" {
			desc = fmt.Sprintf("%s (%s)", snippet.Title, snippet.PublishedAt)
		}

		item.AddSummary(desc)

		// Select thumbnail
		item.AddImage(yt.selectThumbnail(snippet.Thumbnails, feed.Quality))

		// Parse date added to playlist / publication date
		dateStr := ""
		playlistItem, ok := playlistItems[video.Id]
		if ok {
			dateStr = playlistItem.PublishedAt
		} else {
			dateStr = snippet.PublishedAt
		}

		pubDate, err := yt.parseDate(dateStr)
		if err != nil {
			return errors.Wrapf(err, "failed to parse video publish date: %s", dateStr)
		}

		item.AddPubDate(&pubDate)

		// Parse duration
		d, err := duration.FromString(video.ContentDetails.Duration)
		if err != nil {
			return errors.Wrapf(err, "failed to parse duration %s", video.ContentDetails.Duration)
		}

		seconds := int64(d.ToDuration().Seconds())
		item.AddDuration(seconds)

		// Add download links
		size := yt.getSize(seconds, feed)
		item.AddEnclosure(makeEnclosure(feed, video.Id, size))

		// podcast.AddItem requires description to be not empty, use workaround
		if item.Description == "" {
			item.Description = " "
		}

		item.IOrder = strconv.FormatInt(playlistItem.Position, 10)

		_, err = podcast.AddItem(item)
		if err != nil {
			return errors.Wrapf(err, "failed to add item to podcast (id '%s')", video.Id)
		}
	}

	return nil
}

// Cost: (3 units + 5 units) * X pages = 8 units per page
func (yt *YouTubeBuilder) queryItems(itemId string, feed *model.Feed, podcast *itunes.Podcast) error {
	var (
		token string
		count int
	)

	for {
		items, pageToken, err := yt.listPlaylistItems(itemId, token)
		if err != nil {
			return err
		}

		token = pageToken

		if len(items) == 0 {
			return nil
		}

		// Extract playlist snippets
		snippets := map[string]*youtube.PlaylistItemSnippet{}
		for _, item := range items {
			snippets[item.Snippet.ResourceId.VideoId] = item.Snippet
			count++
		}

		// Query video descriptions from the list of ids
		if err := yt.queryVideoDescriptions(snippets, feed, podcast); err != nil {
			return err
		}

		if count >= feed.PageSize || token == "" {
			return nil
		}
	}
}

func (yt *YouTubeBuilder) Build(feed *model.Feed) (*itunes.Podcast, error) {

	// Query general information about feed (title, description, lang, etc)
	podcast, itemId, err := yt.queryFeed(feed)
	if err != nil {
		return nil, err
	}

	// Get video descriptions
	if feed.PageSize == 0 {
		feed.PageSize = maxYoutubeResults
	}

	if err := yt.queryItems(itemId, feed, podcast); err != nil {
		return nil, err
	}

	// Sort episodes by <itunes:order> tag

	sort.Slice(podcast.Items, func(i, j int) bool {
		item1, _ := strconv.Atoi(podcast.Items[i].IOrder)
		item2, _ := strconv.Atoi(podcast.Items[j].IOrder)
		return item1 < item2
	})

	return podcast, nil
}

func NewYouTubeBuilder(key string) (*YouTubeBuilder, error) {
	yt, err := youtube.New(&http.Client{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create youtube client")
	}

	return &YouTubeBuilder{client: yt, key: apiKey(key)}, nil
}
