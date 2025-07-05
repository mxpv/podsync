package builder

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BrianHicks/finch/duration"
	"github.com/mxpv/podsync/pkg/feed"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/youtube/v3"

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

// resolveHandle uses youtube.Search.List to find a channel ID for a given handle.
// Cost: 100 units for the search.
func (yt *YouTubeBuilder) resolveHandle(ctx context.Context, handle string) (string, error) {
	if !strings.HasPrefix(handle, "@") {
		return "", errors.Errorf("handle %q does not start with @", handle)
	}

	searchParts := []string{"id"}
	searchCall := yt.client.Search.List(searchParts).Q(handle).Type("channel").MaxResults(1)

	resp, err := searchCall.Context(ctx).Do(yt.key)

	if err != nil {
		return "", errors.Wrapf(err, "failed to search for handle %s", handle)
	}

	if len(resp.Items) == 0 || resp.Items[0].Id == nil || resp.Items[0].Id.ChannelId == "" {
		return "", errors.Wrapf(model.ErrNotFound, "handle %s not found or no channel ID associated", handle)
	}

	return resp.Items[0].Id.ChannelId, nil
}

// Cost: 5 units (call method: 1, snippet: 2, contentDetails: 2) if not resolving handle.
// If resolving handle, additional 100 units for search.
// See https://developers.google.com/youtube/v3/docs/channels/list#part
func (yt *YouTubeBuilder) listChannels(ctx context.Context, linkType model.Type, id string, partsStr string) (*youtube.Channel, error) {
	partsSlice := strings.Split(partsStr, ",")
	req := yt.client.Channels.List(partsSlice)
	var resolvedID = id

	if strings.HasPrefix(id, "@") {
		// Handles are always treated as channels for the purpose of fetching channel details.
		// The original linkType might have been TypeUser or TypeChannel based on URL parsing,
		// but the core identifier is the handle.
		actualChannelID, err := yt.resolveHandle(ctx, id)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to resolve handle %s", id)
		}
		resolvedID = actualChannelID
		req = req.Id(resolvedID)
	} else {
		switch linkType {
		case model.TypeChannel:
			req = req.Id(id)
		case model.TypeUser: // Legacy user URL
			req = req.ForUsername(id)
		default:
			return nil, errors.Errorf("unsupported link type for listChannels: %s", linkType)
		}
	}

	resp, err := req.Context(ctx).Do(yt.key)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to query channel")
	}

	if len(resp.Items) == 0 {
		return nil, model.ErrNotFound
	}

	item := resp.Items[0]
	return item, nil
}

// Cost: 3 units (call method: 1, snippet: 2)
// See https://developers.google.com/youtube/v3/docs/playlists/list#part
func (yt *YouTubeBuilder) listPlaylists(ctx context.Context, id, channelID string, partsStr string) (*youtube.Playlist, error) {
	partsSlice := strings.Split(partsStr, ",")
	req := yt.client.Playlists.List(partsSlice)

	if id != "" {
		req = req.Id(id)
	} else {
		req = req.ChannelId(channelID)
	}

	resp, err := req.Context(ctx).Do(yt.key)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to query playlist")
	}

	if len(resp.Items) == 0 {
		return nil, model.ErrNotFound
	}

	item := resp.Items[0]
	return item, nil
}

// Cost: 3 units (call: 1, snippet: 2)
// See https://developers.google.com/youtube/v3/docs/playlistItems/list#part
func (yt *YouTubeBuilder) listPlaylistItems(ctx context.Context, feed *model.Feed, pageToken string) ([]*youtube.PlaylistItem, string, error) {
	count := maxYoutubeResults
	if count > feed.PageSize {
		// If we need less than 50
		count = feed.PageSize
	}

	// Parts for PlaylistItems.List is hardcoded to "id,snippet" in the original code constructing `req`.
	playlistItemParts := []string{"id", "snippet"}
	req := yt.client.PlaylistItems.List(playlistItemParts).MaxResults(int64(count)).PlaylistId(feed.ItemID)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, err := req.Context(ctx).Do(yt.key)

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

func (yt *YouTubeBuilder) selectThumbnail(snippet *youtube.ThumbnailDetails, quality model.Quality, videoID string) string {
	if snippet == nil {
		if videoID != "" {
			return fmt.Sprintf("https://img.youtube.com/vi/%s/default.jpg", videoID)
		}

		// TODO: use Podsync's preview image if unable to retrieve from YouTube
		return ""
	}

	// Use high resolution thumbnails for high quality mode
	// https://github.com/mxpv/Podsync/issues/14
	if quality == model.QualityHigh {
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

func (yt *YouTubeBuilder) GetVideoCount(ctx context.Context, info *model.Info) (uint64, error) {
	switch info.LinkType {
	case model.TypeChannel, model.TypeUser:
		// Cost: 3 units
		if channel, err := yt.listChannels(ctx, info.LinkType, info.ItemID, "id,statistics"); err != nil {
			return 0, err
		} else { // nolint:golint
			return channel.Statistics.VideoCount, nil
		}

	case model.TypePlaylist:
		// Cost: 3 units
		if playlist, err := yt.listPlaylists(ctx, info.ItemID, "", "id,contentDetails"); err != nil {
			return 0, err
		} else { // nolint:golint
			return uint64(playlist.ContentDetails.ItemCount), nil
		}

	default:
		return 0, errors.New("unsupported link format")
	}
}

func (yt *YouTubeBuilder) queryFeed(ctx context.Context, feed *model.Feed, info *model.Info) error {
	var (
		thumbnails *youtube.ThumbnailDetails
	)
	originalItemID := info.ItemID // Store original item ID for later use, especially for handles

	switch info.LinkType {
	case model.TypeChannel, model.TypeUser:
		// Cost: 5 units for channel or user (plus 100 if resolving handle)
		channel, err := yt.listChannels(ctx, info.LinkType, info.ItemID, "id,snippet,contentDetails")
		if err != nil {
			return err
		}

		feed.Title = channel.Snippet.Title
		feed.Description = channel.Snippet.Description

		// If the original ItemID was a handle, use it for ItemURL and Author
		if strings.HasPrefix(originalItemID, "@") {
			feed.ItemURL = fmt.Sprintf("https://youtube.com/%s", originalItemID)
			feed.Author = originalItemID
		} else {
			// Existing logic for non-handle channel/user URLs
			if channel.Kind == "youtube#channel" {
				feed.ItemURL = fmt.Sprintf("https://youtube.com/channel/%s", channel.Id)
				// For channels not identified by a user-facing name initially,
				// it's often better to use channel title as author if CustomUrl is not available or not fitting.
				// However, CustomUrl if available and different from ID is usually what users expect.
				// If channel.Snippet.CustomUrl is available and starts with @, it's a handle.
				// If not, channel title is a good fallback.
				if channel.Snippet.CustomUrl != "" {
					feed.Author = channel.Snippet.CustomUrl
				} else {
					feed.Author = channel.Snippet.Title // Fallback to title
				}
			} else { // Legacy user (rarely distinct from channel nowadays)
				feed.ItemURL = fmt.Sprintf("https://youtube.com/user/%s", channel.Snippet.CustomUrl) // CustomUrl is the username here
				feed.Author = channel.Snippet.CustomUrl
			}
		}

		feed.ItemID = channel.ContentDetails.RelatedPlaylists.Uploads // This ItemID is now the uploads playlist ID

		if date, err := yt.parseDate(channel.Snippet.PublishedAt); err != nil {
			return err
		} else { // nolint:golint
			feed.PubDate = date
		}

		thumbnails = channel.Snippet.Thumbnails

	case model.TypePlaylist:
		originalItemID := info.ItemID // Preserve original ItemID for author/URL if it's a handle
		isHandle := strings.HasPrefix(info.ItemID, "@")

		if isHandle {
			channelId, err := yt.resolveHandle(ctx, info.ItemID)
			if err != nil {
				return errors.Wrapf(err, "failed to resolve handle for playlist query: %s", info.ItemID)
			}

			// Fetch channel details to get uploads playlist and metadata
			// Using TypeChannel here as we are operating on a channelId
			// Cost: 5 units (plus 100 for prior handle resolution)
			channel, err := yt.listChannels(ctx, model.TypeChannel, channelId, "id,snippet,contentDetails")
			if err != nil {
				return errors.Wrapf(err, "failed to list channel for handle-based playlist: %s", info.ItemID)
			}

			feed.Title = fmt.Sprintf("%s - All Uploads", channel.Snippet.Title)
			feed.Description = channel.Snippet.Description
			// Use the original handle for ItemURL as per task description
			feed.ItemURL = fmt.Sprintf("https://youtube.com/%s/playlists", originalItemID)
			feed.Author = originalItemID // The handle itself

			if date, errDate := yt.parseDate(channel.Snippet.PublishedAt); errDate != nil {
				return errDate
			} else {
				feed.PubDate = date
			}
			thumbnails = channel.Snippet.Thumbnails // Use channel thumbnails for cover art

			// This is key: set ItemID to the channel's uploads playlist ID
			if channel.ContentDetails == nil || channel.ContentDetails.RelatedPlaylists == nil || channel.ContentDetails.RelatedPlaylists.Uploads == "" {
				return errors.Errorf("could not find uploads playlist for channel: %s (resolved from handle %s)", channelId, originalItemID)
			}
			feed.ItemID = channel.ContentDetails.RelatedPlaylists.Uploads
			// The LinkType on the feed object can remain TypePlaylist.
			// The ItemID now points to a concrete playlist (the uploads playlist).
		} else { // Not a handle, so info.ItemID is assumed to be a standard playlist ID (PL...)
			// Cost: 3 units for playlist (id,snippet,status,contentDetails)
			playlist, err := yt.listPlaylists(ctx, info.ItemID, "", "id,snippet,status,contentDetails")
			if err != nil {
				return err
			}

			if playlist.Status != nil && playlist.Status.PrivacyStatus == "private" && !feed.PrivateFeed {
				return errors.Errorf("playlist %s is private. To allow, set private_feed = true in config for this feed", info.ItemID)
			}

			// Existing logic for standard playlists
			feed.Title = fmt.Sprintf("%s: %s", playlist.Snippet.ChannelTitle, playlist.Snippet.Title)
			feed.Description = playlist.Snippet.Description
			feed.ItemURL = fmt.Sprintf("https://youtube.com/playlist?list=%s", playlist.Id)
			feed.Author = playlist.Snippet.ChannelTitle

			if date, errDate := yt.parseDate(playlist.Snippet.PublishedAt); errDate != nil {
				return errDate
			} else {
				feed.PubDate = date
			}
			thumbnails = playlist.Snippet.Thumbnails
			feed.ItemID = playlist.Id // Ensure ItemID is the playlist ID
		}

	default:
		return errors.New("unsupported link format")
	}

	if feed.Description == "" {
		feed.Description = fmt.Sprintf("%s (%s)", feed.Title, feed.PubDate)
	}

	feed.CoverArt = yt.selectThumbnail(thumbnails, feed.CoverArtQuality, "")

	return nil
}

// Video size information requires 1 additional call for each video (1 feed = 50 videos = 50 calls),
// which is too expensive, so get approximated size depending on duration and definition params
func (yt *YouTubeBuilder) getSize(duration int64, feed *model.Feed) int64 {
	if feed.Format == model.FormatAudio {
		if feed.Quality == model.QualityHigh {
			return highAudioBytesPerSecond * duration
		}

		return lowAudioBytesPerSecond * duration
	}

	// Video format

	if feed.Quality == model.QualityHigh {
		return duration * hdBytesPerSecond
	}

	return duration * ldBytesPerSecond
}

// Cost: 5 units (call: 1, snippet: 2, contentDetails: 2)
// See https://developers.google.com/youtube/v3/docs/videos/list#part
func (yt *YouTubeBuilder) queryVideoDescriptions(ctx context.Context, playlist map[string]*youtube.PlaylistItemSnippet, feed *model.Feed) error {
	// Make the list of video ids
	ids := make([]string, 0, len(playlist))
	for _, s := range playlist {
		ids = append(ids, s.ResourceId.VideoId)
	}

	// Init a list that will contains the aggregated strings of videos IDs (capped at 50 IDs per API Calls)
	idsList := make([]string, 0, 1)

	// Chunk the list of IDs by slices limited to maxYoutubeResults
	for i := 0; i < len(ids); i += maxYoutubeResults {
		end := i + maxYoutubeResults
		if end > len(ids) {
			end = len(ids)
		}
		// Save each slice as comma-delimited string
		idsList = append(idsList, strings.Join(ids[i:end], ","))
	}

	// Show how many API calls will be required
	log.Debugf("Expected to make %d API calls to get the descriptions for %d episode(s).", len(idsList), len(ids))

	// Loop in each slices of 50 (or less) IDs and query their description
	for _, idsI := range idsList {
		videoPartsSlice := []string{"id", "snippet", "contentDetails"}
		req := yt.client.Videos.List(videoPartsSlice).Id(idsI)

		resp, err := req.Context(ctx).Do(yt.key)

		if err != nil {
			return errors.Wrap(err, "failed to query video descriptions")
		}

		for _, video := range resp.Items {
			var (
				snippet  = video.Snippet
				videoID  = video.Id
				videoURL = fmt.Sprintf("https://youtube.com/watch?v=%s", video.Id)
				image    = yt.selectThumbnail(snippet.Thumbnails, feed.Quality, videoID)
			)

			// Skip unreleased/airing Premiere videos
			if snippet.LiveBroadcastContent == "upcoming" || snippet.LiveBroadcastContent == "live" {
				continue
			}

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
				size  = yt.getSize(seconds, feed)
			)

			feed.Episodes = append(feed.Episodes, &model.Episode{
				ID:          video.Id,
				Title:       snippet.Title,
				Description: snippet.Description,
				Thumbnail:   image,
				Duration:    seconds,
				Size:        size,
				VideoURL:    videoURL,
				PubDate:     pubDate,
				Order:       order,
				Status:      model.EpisodeNew,
			})
		}
	}

	return nil
}

// Cost:
// ASC mode = (3 units + 5 units) * X pages = 8 units per page
// DESC mode = 3 units * (number of pages in the entire playlist) + 5 units
func (yt *YouTubeBuilder) queryItems(ctx context.Context, feed *model.Feed) error {
	var (
		token       string
		count       int
		allSnippets []*youtube.PlaylistItemSnippet
	)

	for {
		items, pageToken, err := yt.listPlaylistItems(ctx, feed, token)
		if err != nil {
			return err
		}

		token = pageToken

		if len(items) == 0 {
			break
		}

		// Extract playlist snippets
		for _, item := range items {
			allSnippets = append(allSnippets, item.Snippet)
			count++
		}

		if (feed.PlaylistSort != model.SortingDesc && count >= feed.PageSize) || token == "" {
			break
		}
	}

	if len(allSnippets) > feed.PageSize {
		if feed.PlaylistSort != model.SortingDesc {
			allSnippets = allSnippets[:feed.PageSize]
		} else {
			allSnippets = allSnippets[len(allSnippets)-feed.PageSize:]
		}
	}

	snippets := map[string]*youtube.PlaylistItemSnippet{}
	for _, snippet := range allSnippets {
		snippets[snippet.ResourceId.VideoId] = snippet
	}

	// Query video descriptions from the list of ids
	if err := yt.queryVideoDescriptions(ctx, snippets, feed); err != nil {
		return err
	}

	return nil
}

func (yt *YouTubeBuilder) Build(ctx context.Context, cfg *feed.Config) (*model.Feed, error) {
	info, err := ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	_feed := &model.Feed{
		ItemID:          info.ItemID,
		Provider:        info.Provider,
		LinkType:        info.LinkType,
		Format:          cfg.Format,
		Quality:         cfg.Quality,
		CoverArtQuality: cfg.Custom.CoverArtQuality,
		PageSize:        cfg.PageSize,
		PlaylistSort:    cfg.PlaylistSort,
		PrivateFeed:     cfg.PrivateFeed,
		UpdatedAt:       time.Now().UTC(),
	}

	if _feed.PageSize == 0 {
		_feed.PageSize = maxYoutubeResults
	}

	// Query general information about feed (title, description, lang, etc)
	if err := yt.queryFeed(ctx, _feed, &info); err != nil {
		return nil, err
	}

	if err := yt.queryItems(ctx, _feed); err != nil {
		return nil, err
	}

	// YT API client gets 50 episodes per query.
	// Round up to page size.
	if len(_feed.Episodes) > _feed.PageSize {
		_feed.Episodes = _feed.Episodes[:_feed.PageSize]
	}

	sort.Slice(_feed.Episodes, func(i, j int) bool {
		item1, _ := strconv.Atoi(_feed.Episodes[i].Order)
		item2, _ := strconv.Atoi(_feed.Episodes[j].Order)
		return item1 < item2
	})

	return _feed, nil
}

func NewYouTubeBuilder(key string) (*YouTubeBuilder, error) {
	if key == "" {
		return nil, errors.New("empty YouTube API key")
	}


	httpClient := &http.Client{} // Or any other http.Client setup you might have
	ytService, err := youtube.New(httpClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create youtube client for builder")
	}

	return &YouTubeBuilder{client: ytService, key: apiKey(key)}, nil
}
