package builder

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BrianHicks/finch/duration"
	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/link"
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
func (yt *YouTubeBuilder) listChannels(linkType link.Type, id string, parts string) (*youtube.Channel, error) {
	req := yt.client.Channels.List(parts)

	switch linkType {
	case link.TypeChannel:
		req = req.Id(id)
	case link.TypeUser:
		req = req.ForUsername(id)
	default:
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
func (yt *YouTubeBuilder) listPlaylists(id, channelID string, parts string) (*youtube.Playlist, error) {
	req := yt.client.Playlists.List(parts)

	if id != "" {
		req = req.Id(id)
	} else {
		req = req.ChannelId(channelID)
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
func (yt *YouTubeBuilder) listPlaylistItems(itemID string, pageToken string) ([]*youtube.PlaylistItem, string, error) {
	req := yt.client.PlaylistItems.List("id,snippet").MaxResults(maxYoutubeResults).PlaylistId(itemID)
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

func (yt *YouTubeBuilder) selectThumbnail(snippet *youtube.ThumbnailDetails, quality config.Quality, videoID string) string {
	if snippet == nil {
		if videoID != "" {
			return fmt.Sprintf("https://img.youtube.com/vi/%s/default.jpg", videoID)
		}

		// TODO: use Podsync's preview image if unable to retrieve from YouTube
		return ""
	}

	// Use high resolution thumbnails for high quality mode
	// https://github.com/mxpv/Podsync/issues/14
	if quality == config.QualityHigh {
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

func (yt *YouTubeBuilder) GetVideoCount(info *link.Info) (uint64, error) {
	switch info.LinkType {
	case link.TypeChannel, link.TypeUser:
		// Cost: 3 units
		if channel, err := yt.listChannels(info.LinkType, info.ItemID, "id,statistics"); err != nil {
			return 0, err
		} else { // nolint:golint
			return channel.Statistics.VideoCount, nil
		}

	case link.TypePlaylist:
		// Cost: 3 units
		if playlist, err := yt.listPlaylists(info.ItemID, "", "id,contentDetails"); err != nil {
			return 0, err
		} else { // nolint:golint
			return uint64(playlist.ContentDetails.ItemCount), nil
		}

	default:
		return 0, errors.New("unsupported link format")
	}
}

func (yt *YouTubeBuilder) queryFeed(info *link.Info, config *config.Feed) (*Feed, error) {
	var (
		thumbnails *youtube.ThumbnailDetails
	)

	feed := Feed{
		ItemID:    info.ItemID,
		Provider:  info.Provider,
		LinkType:  info.LinkType,
		UpdatedAt: time.Now().UTC(),
	}

	switch info.LinkType {
	case link.TypeChannel, link.TypeUser:
		// Cost: 5 units for channel or user
		channel, err := yt.listChannels(info.LinkType, info.ItemID, "id,snippet,contentDetails")
		if err != nil {
			return nil, err
		}

		feed.Title = channel.Snippet.Title
		feed.Description = channel.Snippet.Description

		if channel.Kind == "youtube#channel" {
			feed.ItemURL = fmt.Sprintf("https://youtube.com/channel/%s", channel.Id)
			feed.Author = "<notfound>"
		} else {
			feed.ItemURL = fmt.Sprintf("https://youtube.com/user/%s", channel.Snippet.CustomUrl)
			feed.Author = channel.Snippet.CustomUrl
		}

		feed.ItemID = channel.ContentDetails.RelatedPlaylists.Uploads

		if date, err := yt.parseDate(channel.Snippet.PublishedAt); err != nil {
			return nil, err
		} else { // nolint:golint
			feed.PubDate = date
		}

		thumbnails = channel.Snippet.Thumbnails

	case link.TypePlaylist:
		// Cost: 3 units for playlist
		playlist, err := yt.listPlaylists(info.ItemID, "", "id,snippet")
		if err != nil {
			return nil, err
		}

		feed.Title = fmt.Sprintf("%s: %s", playlist.Snippet.ChannelTitle, playlist.Snippet.Title)
		feed.Description = playlist.Snippet.Description

		feed.ItemURL = fmt.Sprintf("https://youtube.com/playlist?list=%s", playlist.Id)
		feed.ItemID = playlist.Id

		feed.Author = "<notfound>"

		if date, err := yt.parseDate(playlist.Snippet.PublishedAt); err != nil {
			return nil, err
		} else { // nolint:golint
			feed.PubDate = date
		}

		thumbnails = playlist.Snippet.Thumbnails

	default:
		return nil, errors.New("unsupported link format")
	}

	// Apply customizations and default values

	if feed.Description == "" {
		feed.Description = fmt.Sprintf("%s (%s)", feed.Title, feed.PubDate)
	}

	if config.CoverArt != "" {
		feed.CoverArt = config.CoverArt
	} else {
		feed.CoverArt = yt.selectThumbnail(thumbnails, config.Quality, "")
	}

	return &feed, nil
}

// Video size information requires 1 additional call for each video (1 feed = 50 videos = 50 calls),
// which is too expensive, so get approximated size depending on duration and definition params
func (yt *YouTubeBuilder) getSize(duration int64, cfg *config.Feed) int64 {
	if cfg.Format == config.FormatAudio {
		if cfg.Quality == config.QualityHigh {
			return highAudioBytesPerSecond * duration
		}

		return lowAudioBytesPerSecond * duration
	}

	// Video format

	if cfg.Quality == config.QualityHigh {
		return duration * hdBytesPerSecond
	}

	return duration * ldBytesPerSecond
}

// Cost: 5 units (call: 1, snippet: 2, contentDetails: 2)
// See https://developers.google.com/youtube/v3/docs/videos/list#part
func (yt *YouTubeBuilder) queryVideoDescriptions(playlist map[string]*youtube.PlaylistItemSnippet, feed *Feed, cfg *config.Feed) error {
	// Make the list of video ids
	ids := make([]string, 0, len(playlist))
	for _, s := range playlist {
		ids = append(ids, s.ResourceId.VideoId)
	}

	req, err := yt.client.Videos.List("id,snippet,contentDetails").Id(strings.Join(ids, ",")).Do(yt.key)
	if err != nil {
		return errors.Wrap(err, "failed to query video descriptions")
	}

	for _, video := range req.Items {
		var (
			snippet  = video.Snippet
			videoID  = video.Id
			videoURL = fmt.Sprintf("https://youtube.com/watch?v=%s", video.Id)
			image    = yt.selectThumbnail(snippet.Thumbnails, cfg.Quality, videoID)
		)

		// Parse date added to playlist / publication date
		dateStr := ""
		playlistItem, ok := playlist[video.Id]
		if ok {
			dateStr = playlistItem.PublishedAt
		} else {
			dateStr = snippet.PublishedAt
		}

		pubDate, err := yt.parseDate(dateStr)
		if err != nil {
			return errors.Wrapf(err, "failed to parse video publish date: %s", dateStr)
		}

		// Sometimes YouTube retrun empty content defailt, use arbitrary one
		var seconds int64 = 1
		if video.ContentDetails != nil {
			// Parse duration
			d, err := duration.FromString(video.ContentDetails.Duration)
			if err != nil {
				return errors.Wrapf(err, "failed to parse duration %s", video.ContentDetails.Duration)
			}

			seconds = int64(d.ToDuration().Seconds())
		}

		var (
			order = strconv.FormatInt(playlistItem.Position, 10)
			size  = yt.getSize(seconds, cfg)
		)

		feed.Episodes = append(feed.Episodes, &Item{
			ID:          video.Id,
			Title:       snippet.Title,
			Description: snippet.Description,
			Thumbnail:   image,
			Duration:    seconds,
			Size:        size,
			VideoURL:    videoURL,
			PubDate:     pubDate,
			Order:       order,
		})
	}

	return nil
}

// Cost: (3 units + 5 units) * X pages = 8 units per page
func (yt *YouTubeBuilder) queryItems(feed *Feed, cfg *config.Feed) error {
	var (
		token string
		count int
	)

	for {
		items, pageToken, err := yt.listPlaylistItems(feed.ItemID, token)
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
		if err := yt.queryVideoDescriptions(snippets, feed, cfg); err != nil {
			return err
		}

		if count >= cfg.PageSize || token == "" {
			return nil
		}
	}
}

func (yt *YouTubeBuilder) Build(cfg *config.Feed) (*Feed, error) {
	info, err := link.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}

	// Query general information about feed (title, description, lang, etc)
	feed, err := yt.queryFeed(&info, cfg)
	if err != nil {
		return nil, err
	}

	if err := yt.queryItems(feed, cfg); err != nil {
		return nil, err
	}

	sort.Slice(feed.Episodes, func(i, j int) bool {
		item1, _ := strconv.Atoi(feed.Episodes[i].Order)
		item2, _ := strconv.Atoi(feed.Episodes[j].Order)
		return item1 < item2
	})

	return feed, nil
}

func NewYouTubeBuilder(key string) (*YouTubeBuilder, error) {
	yt, err := youtube.New(&http.Client{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create youtube client")
	}

	return &YouTubeBuilder{client: yt, key: apiKey(key)}, nil
}
