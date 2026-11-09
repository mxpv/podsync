package builder

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/mxpv/podsync/pkg/feed"
	"github.com/pkg/errors"
	soundcloudapi "github.com/zackradisic/soundcloud-api"

	"github.com/mxpv/podsync/pkg/model"
)

type SoundCloudBuilder struct {
	client     *soundcloudapi.API
	httpClient *http.Client
}

// Build implements Builder for SoundCloud.
//
// Supported URL formats (see url.go parsing):
//   - Playlist: https://soundcloud.com/<user>/sets/<playlist>
//   - User:     https://soundcloud.com/<user>  (or /tracks)
//
// Podsync’s downloader uses Episode.VideoURL; for SoundCloud we can safely set this
// to the public track permalink URL and let yt-dlp resolve the actual audio.
func (s *SoundCloudBuilder) Build(ctx context.Context, cfg *feed.Config) (*model.Feed, error) {
	info, err := ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	_feed := &model.Feed{
		ItemID:    info.ItemID,
		Provider:  info.Provider,
		LinkType:  info.LinkType,
		Format:    cfg.Format,
		Quality:   cfg.Quality,
		PageSize:  cfg.PageSize,
		UpdatedAt: time.Now().UTC(),
		ItemURL:   cfg.URL,
	}

	switch info.LinkType {
	case model.TypePlaylist:
		return s.buildPlaylist(ctx, cfg, _feed)
	case model.TypeUser:
		return s.buildUser(ctx, cfg, _feed)
	default:
		return nil, errors.New("unsupported soundcloud feed type")
	}
}

func (s *SoundCloudBuilder) buildPlaylist(_ctx context.Context, cfg *feed.Config, _feed *model.Feed) (*model.Feed, error) {
	if !soundcloudapi.IsPlaylistURL(cfg.URL) {
		return nil, errors.New("invalid soundcloud playlist url")
	}

	scplaylist, err := s.client.GetPlaylistInfo(cfg.URL)
	if err != nil {
		return nil, err
	}

	_feed.Title = scplaylist.Title
	_feed.Description = scplaylist.Description
	_feed.Author = scplaylist.User.Username
	_feed.CoverArt = scplaylist.ArtworkURL

	if date, err := time.Parse(time.RFC3339, scplaylist.CreatedAt); err == nil {
		_feed.PubDate = date
	}

	added := 0
	for _, track := range scplaylist.Tracks {
		_feed.Episodes = append(_feed.Episodes, trackToEpisode(track))
		added++

		// PageSize <= 0 means "no limit"
		if _feed.PageSize > 0 && added >= _feed.PageSize {
			break
		}
	}

	return _feed, nil
}

func (s *SoundCloudBuilder) buildUser(ctx context.Context, cfg *feed.Config, _feed *model.Feed) (*model.Feed, error) {
	// Resolve profile URL to numeric user ID.
	user, err := s.client.GetUser(soundcloudapi.GetUserOptions{
		ProfileURL: cfg.URL,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve soundcloud user profile")
	}

	_feed.Title = user.Username
	_feed.Author = user.Username
	_feed.Description = user.Description
	_feed.CoverArt = user.AvatarURL

	limit := cfg.PageSize
	if limit <= 0 {
		// Keep a sane default; the feed can still be "unlimited" by setting PageSize high.
		limit = 20
	}

	tracks, err := s.fetchUserTracks(ctx, user.ID, limit)
	if err != nil {
		return nil, err
	}

	for _, track := range tracks {
		_feed.Episodes = append(_feed.Episodes, trackToEpisode(track))
	}

	return _feed, nil
}

func trackToEpisode(track soundcloudapi.Track) *model.Episode {
	pubDate, _ := time.Parse(time.RFC3339, track.CreatedAt)

	videoID := strconv.FormatInt(track.ID, 10)
	duration := track.DurationMS / 1000
	mediaURL := track.PermalinkURL
	trackSize := track.DurationMS * 15 // very rough estimate

	return &model.Episode{
		ID:          videoID,
		Title:       track.Title,
		Description: track.Description,
		Duration:    duration,
		Size:        trackSize,
		VideoURL:    mediaURL,
		PubDate:     pubDate,
		Thumbnail:   track.ArtworkURL,
		Status:      model.EpisodeNew,
	}
}

// fetchUserTracks fetches the most recent public uploads for a SoundCloud user via api-v2.
//
// We keep this call isolated because SoundCloud’s private API is subject to change.
// The soundcloud-api library is used for client_id scraping and resolving profile URL -> user ID.
func (s *SoundCloudBuilder) fetchUserTracks(ctx context.Context, userID int64, limit int) ([]soundcloudapi.Track, error) {
	clientID := s.client.ClientID()
	if clientID == "" {
		cid, err := soundcloudapi.FetchClientID()
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch soundcloud client_id")
		}
		s.client.SetClientID(cid)
		clientID = cid
	}

	endpoint, err := url.Parse("https://api-v2.soundcloud.com/users/" + strconv.FormatInt(userID, 10) + "/tracks")
	if err != nil {
		return nil, errors.Wrap(err, "failed to build soundcloud api url")
	}

	q := endpoint.Query()
	q.Set("client_id", clientID)
	q.Set("limit", strconv.Itoa(limit))
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build soundcloud api request")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "soundcloud user tracks request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.Errorf("soundcloud user tracks request returned http %d", resp.StatusCode)
	}

	var tracks []soundcloudapi.Track
	if err := json.NewDecoder(resp.Body).Decode(&tracks); err != nil {
		return nil, errors.Wrap(err, "failed to decode soundcloud user tracks response")
	}

	return tracks, nil
}

// NewSoundcloudBuilder creates a SoundCloud builder.
//
// The key parameter is optional and is interpreted as a SoundCloud client_id override.
// If empty, the underlying library will still be able to scrape a client_id when needed.
func NewSoundcloudBuilder(key string) (*SoundCloudBuilder, error) {
	sc, err := soundcloudapi.New(soundcloudapi.APIOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create soundcloud client")
	}

	if key != "" {
		sc.SetClientID(key)
	}

	return &SoundCloudBuilder{
		client:     sc,
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}, nil
}
