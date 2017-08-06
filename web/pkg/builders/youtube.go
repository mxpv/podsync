package builders

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/BrianHicks/finch/duration"
	itunes "github.com/mxpv/podcast"
	"github.com/mxpv/podsync/web/pkg/storage"
	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"
)

const (
	maxYoutubeResults = 50
	hdBytesPerSecond  = 350000
	ldBytesPerSecond  = 100000
)

type apiKey string

func (key apiKey) Get() (string, string) {
	return "key", string(key)
}

type YouTubeBuilder struct {
	client *youtube.Service
	key    apiKey
}

func (yt *YouTubeBuilder) parseUrl(link string) (kind linkType, id string, err error) {
	parsed, err := url.Parse(link)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse url: %s", link)
		return
	}

	if !strings.HasSuffix(parsed.Host, "youtube.com") {
		err = errors.New("invalid youtube host")
		return
	}

	path := parsed.EscapedPath()

	// Parse
	// https://www.youtube.com/playlist?list=PLCB9F975ECF01953C
	if strings.HasPrefix(path, "/playlist") {
		kind = linkTypePlaylist

		id = parsed.Query().Get("list")
		if id != "" {
			return
		}

		err = errors.New("invalid playlist link")
		return
	}

	// Parse
	// - https://www.youtube.com/channel/UC5XPnUk8Vvv_pWslhwom6Og
	// - https://www.youtube.com/channel/UCrlakW-ewUT8sOod6Wmzyow/videos
	if strings.HasPrefix(path, "/channel") {
		kind = linkTypeChannel
		parts := strings.Split(parsed.EscapedPath(), "/")
		if len(parts) <= 2 {
			err = errors.New("invalid youtube channel link")
			return
		}

		id = parts[2]
		if id == "" {
			err = errors.New("invalid id")
		}

		return
	}

	// Parse
	// - https://www.youtube.com/user/fxigr1
	if strings.HasPrefix(path, "/user") {
		kind = linkTypeUser

		parts := strings.Split(parsed.EscapedPath(), "/")
		if len(parts) <= 2 {
			err = errors.New("invalid user link")
			return
		}

		id = parts[2]
		if id == "" {
			err = errors.New("invalid id")
		}

		return
	}

	err = errors.New("unsupported link format")
	return
}

func (yt *YouTubeBuilder) listChannels(kind linkType, id string) (*youtube.Channel, error) {
	req := yt.client.Channels.List("id,snippet,contentDetails")

	if kind == linkTypeChannel {
		req = req.Id(id)
	} else if kind == linkTypeUser {
		req = req.ForUsername(id)
	} else {
		return nil, errors.New("unsupported link type")
	}

	resp, err := req.Do(yt.key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query channel")
	}

	if len(resp.Items) == 0 {
		return nil, errors.New("channel not found")
	}

	item := resp.Items[0]
	return item, nil
}

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
		return nil, errors.New("playlist not found")
	}

	item := resp.Items[0]
	return item, nil
}

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

func (yt *YouTubeBuilder) selectThumbnail(snippet *youtube.ThumbnailDetails, quality storage.Quality) string {
	// Use high resolution thumbnails for high quality mode
	// https://github.com/mxpv/Podsync/issues/14
	if quality == storage.HighQuality {
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

func (yt *YouTubeBuilder) queryFeed(kind linkType, id string, feed *storage.Feed) (*itunes.Podcast, string, error) {
	now := time.Now()

	if kind == linkTypeChannel || kind == linkTypeUser {
		channel, err := yt.listChannels(kind, id)
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to query channel")
		}

		itemId := channel.ContentDetails.RelatedPlaylists.Uploads

		link := ""
		if kind == linkTypeChannel {
			link = fmt.Sprintf("https://youtube.com/channel/%s", itemId)
		} else {
			link = fmt.Sprintf("https://youtube.com/user/%s", itemId)
		}

		pubDate, err := yt.parseDate(channel.Snippet.PublishedAt)
		if err != nil {
			return nil, "", err
		}

		title := channel.Snippet.Title

		podcast := itunes.New(title, link, channel.Snippet.Description, &pubDate, &now)
		podcast.Generator = podsyncGenerator

		podcast.AddSubTitle(title)
		podcast.AddCategory(defaultCategory, nil)
		podcast.AddImage(yt.selectThumbnail(channel.Snippet.Thumbnails, feed.Quality))

		return &podcast, itemId, nil
	}

	if kind == linkTypePlaylist {
		playlist, err := yt.listPlaylists(id, "")
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to query playlist")
		}

		link := fmt.Sprintf("https://youtube.com/playlist?list=%s", playlist.Id)

		snippet := playlist.Snippet

		pubDate, err := yt.parseDate(snippet.PublishedAt)
		if err != nil {
			return nil, "", err
		}

		title := fmt.Sprintf("%s: %s", snippet.ChannelTitle, snippet.Title)

		podcast := itunes.New(title, link, snippet.Description, &pubDate, &now)
		podcast.Generator = podsyncGenerator

		podcast.AddSubTitle(title)
		podcast.AddCategory(defaultCategory, nil)
		podcast.AddImage(yt.selectThumbnail(snippet.Thumbnails, feed.Quality))

		return &podcast, playlist.Id, nil
	}

	return nil, "", errors.New("unsupported link format")
}

func (yt *YouTubeBuilder) getVideoSize(definition string, duration int64, fmt storage.Format) int64 {
	// Video size information requires 1 additional call for each video (1 feed = 50 videos = 50 calls),
	// which is too expensive, so get approximated size depending on duration and definition params
	var size int64 = 0

	if definition == "hd" {
		size = duration * hdBytesPerSecond
	} else {
		size = duration * ldBytesPerSecond
	}

	// Some podcasts are coming in with exactly double the actual runtime and with the second half just silence.
	// https://github.com/mxpv/Podsync/issues/6
	if fmt == storage.AudioFormat {
		size /= 2
	}

	return size
}

func (yt *YouTubeBuilder) queryVideoDescriptions(ids []string, feed *storage.Feed, podcast *itunes.Podcast) error {
	req, err := yt.client.Videos.List("id,snippet,contentDetails").Id(strings.Join(ids, ",")).Do(yt.key)
	if err != nil {
		return errors.Wrap(err, "failed to query video descriptions")
	}

	for _, video := range req.Items {
		snippet := video.Snippet

		item := itunes.Item{}

		item.GUID = video.Id
		item.Link = fmt.Sprintf("https://youtube.com/watch?v=%s", video.Id)
		item.Title = snippet.Title
		item.Description = snippet.Description
		item.ISubtitle = snippet.Title

		// Select thumbnail

		item.AddImage(yt.selectThumbnail(snippet.Thumbnails, feed.Quality))

		// Parse publication date

		pubDate, err := yt.parseDate(snippet.PublishedAt)
		if err != nil {
			return errors.Wrapf(err, "failed to parse video publish date: %s", snippet.PublishedAt)
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

		size := yt.getVideoSize(video.ContentDetails.Definition, seconds, feed.Format)
		item.AddEnclosure(makeEnclosure(feed, video.Id, size))

		_, err = podcast.AddItem(item)
		if err != nil {
			return errors.Wrapf(err, "failed to add item to podcast (id '%s')", video.Id)
		}
	}

	return nil
}

func (yt *YouTubeBuilder) queryItems(itemId string, feed *storage.Feed, podcast *itunes.Podcast) error {
	pageToken := ""
	count := 0

	for {
		items, pageToken, err := yt.listPlaylistItems(itemId, pageToken)
		if err != nil {
			return err
		}

		if len(items) == 0 {
			return nil
		}

		// Extract video ids
		ids := make([]string, len(items))
		for index, item := range items {
			ids[index] = item.Snippet.ResourceId.VideoId
			count++
		}

		// Query video descriptions from the list of ids
		if err := yt.queryVideoDescriptions(ids, feed, podcast); err != nil {
			return err
		}

		if count >= feed.PageSize || pageToken == "" {
			return nil
		}
	}
}

func (yt *YouTubeBuilder) Build(feed *storage.Feed) (*itunes.Podcast, error) {
	kind, id, err := yt.parseUrl(feed.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse link: %s", feed.URL)
	}

	// Query general information about feed (title, description, lang, etc)

	podcast, itemId, err := yt.queryFeed(kind, id, feed)
	if err != nil {
		return nil, err
	}

	// Get video descriptions

	if err := yt.queryItems(itemId, feed, podcast); err != nil {
		return nil, err
	}

	return podcast, nil
}

func NewYouTubeBuilder(key string) (*YouTubeBuilder, error) {
	yt, err := youtube.New(&http.Client{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create youtube client")
	}

	return &YouTubeBuilder{client: yt, key: apiKey(key)}, nil
}
