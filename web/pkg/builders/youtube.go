package builders

import (
	"fmt"
	"github.com/eduncan911/podcast"
	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	linkTypeChannel  = linkType(1)
	linkTypePlaylist = linkType(2)
	linkTypeUser     = linkType(3)
)

const (
	maxResults       = 50
	podsyncGenerator = "Podsync YouTube generator"
)

type linkType int
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
		err = errors.New("invalid youtube link")
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
			err = errors.New("invalid channel link")
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

func (yt *YouTubeBuilder) queryChannel(id, username string) (*youtube.Channel, error) {
	req := yt.client.Channels.List("id,snippet,contentDetails")

	if id != "" {
		req = req.Id(id)
	} else {
		req = req.ForUsername(username)
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

func (yt *YouTubeBuilder) queryPlaylist(id, channelId string) (*youtube.Playlist, error) {
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

func (yt *YouTubeBuilder) queryFeed(kind linkType, id string) (*podcast.Podcast, string, error) {
	now := time.Now()

	if kind == linkTypeChannel {
		channel, err := yt.queryChannel(id, "")
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to query channel")
		}

		feed := podcast.New(channel.Snippet.Title, "", channel.Snippet.Description, nil, &now)
		feed.PubDate = channel.Snippet.PublishedAt
		feed.Generator = podsyncGenerator

		return &feed, channel.ContentDetails.RelatedPlaylists.Uploads, nil
	}

	if kind == linkTypeUser {
		channel, err := yt.queryChannel("", id)
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to query channel")
		}

		feed := podcast.New(channel.Snippet.Title, "", channel.Snippet.Description, nil, &now)
		feed.PubDate = channel.Snippet.PublishedAt
		feed.Generator = podsyncGenerator

		return &feed, channel.ContentDetails.RelatedPlaylists.Uploads, nil
	}

	if kind == linkTypePlaylist {
		playlist, err := yt.queryPlaylist(id, "")
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to query playlist")

		}

		feed := podcast.New(playlist.Snippet.Title, "", playlist.Snippet.Description, nil, &now)
		feed.PubDate = playlist.Snippet.PublishedAt
		feed.Generator = podsyncGenerator

		return &feed, playlist.Id, nil
	}

	return nil, "", errors.New("unsupported link format")
}

func (yt *YouTubeBuilder) queryPlaylistItems(itemId string, pageToken string) ([]*youtube.PlaylistItem, string, error) {
	req := yt.client.PlaylistItems.List("id,snippet").MaxResults(maxResults).PlaylistId(itemId)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, err := req.Do(yt.key)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to query playlist items")
	}

	return resp.Items, resp.NextPageToken, nil
}

func (yt *YouTubeBuilder) queryVideoDescriptions(ids []string, feed *podcast.Podcast) error {
	req, err := yt.client.Videos.List("id,snippet,contentDetails").Id(strings.Join(ids, ",")).Do(yt.key)
	if err != nil {
		return errors.Wrap(err, "failed to query video descriptions")
	}

	for _, video := range req.Items {
		item := podcast.Item{}

		item.GUID = video.Id
		item.Link = fmt.Sprintf("https://youtube.com/watch?v=%s", video.Id)
		item.Title = video.Snippet.Title
		item.Description = video.Snippet.Description
		item.PubDateFormatted = video.Snippet.PublishedAt

		_, err := feed.AddItem(item)
		if err != nil {
			return errors.Wrapf(err, "failed to add item to feed (id '%s')", video.Id)
		}
	}

	return nil
}

func (yt *YouTubeBuilder) queryItems(itemId string, pageSize int, feed *podcast.Podcast) error {
	pageToken := ""
	count := 0

	for {
		items, pageToken, err := yt.queryPlaylistItems(itemId, pageToken)
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
		if err := yt.queryVideoDescriptions(ids, feed); err != nil {
			return err
		}

		if count >= pageSize || pageToken == "" {
			return nil
		}
	}
}

func (yt *YouTubeBuilder) Build(url string, pageSize int) (*podcast.Podcast, error) {
	kind, id, err := yt.parseUrl(url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse link: %s", url)
	}

	// Query general information about feed (title, description, lang, etc)

	feed, itemId, err := yt.queryFeed(kind, id)
	if err != nil {
		return nil, err
	}

	// Get video descriptions

	if err := yt.queryItems(itemId, pageSize, feed); err != nil {
		return nil, err
	}

	return feed, nil
}

func NewYouTubeBuilder(key string) (*YouTubeBuilder, error) {
	yt, err := youtube.New(&http.Client{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create youtube client")
	}

	return &YouTubeBuilder{client: yt, key: apiKey(key)}, nil
}
