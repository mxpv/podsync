package builder

import (
	"context"
	"strconv"
	"time"

	"github.com/mxpv/podsync/pkg/feed"
	"github.com/pkg/errors"
	soundcloudapi "github.com/zackradisic/soundcloud-api"

	"github.com/mxpv/podsync/pkg/model"
)

type SoundCloudBuilder struct {
	client *soundcloudapi.API
}

func (s *SoundCloudBuilder) Build(_ctx context.Context, cfg *feed.Config) (*model.Feed, error) {
	info, err := ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	feed := &model.Feed{
		ItemID:    info.ItemID,
		Provider:  info.Provider,
		LinkType:  info.LinkType,
		Format:    cfg.Format,
		Quality:   cfg.Quality,
		PageSize:  cfg.PageSize,
		UpdatedAt: time.Now().UTC(),
	}

	if info.LinkType == model.TypePlaylist {
		if soundcloudapi.IsPlaylistURL(cfg.URL) {
			scplaylist, err := s.client.GetPlaylistInfo(cfg.URL)
			if err != nil {
				return nil, err
			}

			feed.Title = scplaylist.Title
			feed.Description = scplaylist.Description
			feed.ItemURL = cfg.URL

			date, err := time.Parse(time.RFC3339, scplaylist.CreatedAt)
			if err == nil {
				feed.PubDate = date
			}
			feed.Author = scplaylist.User.Username
			feed.CoverArt = scplaylist.ArtworkURL

			var added = 0
			for _, track := range scplaylist.Tracks {
				pubDate, _ := time.Parse(time.RFC3339, track.CreatedAt)
				var (
					videoID   = strconv.FormatInt(track.ID, 10)
					duration  = track.DurationMS / 1000
					mediaURL  = track.PermalinkURL
					trackSize = track.DurationMS * 15 // very rough estimate
				)

				// For the moment we don't support detecting live videos on SoundCloud
				var isLive bool = false

				feed.Episodes = append(feed.Episodes, &model.Episode{
					ID:          videoID,
					Title:       track.Title,
					Description: track.Description,
					Duration:    duration,
					Size:        trackSize,
					VideoURL:    mediaURL,
					PubDate:     pubDate,
					Thumbnail:   track.ArtworkURL,
					Status:      model.EpisodeNew,
					IsLive:      isLive,
				})

				added++

				if added >= feed.PageSize {
					return feed, nil
				}
			}

			return feed, nil
		}
	}

	return nil, errors.New(("unsupported soundcloud feed type"))
}

func NewSoundcloudBuilder() (*SoundCloudBuilder, error) {
	sc, err := soundcloudapi.New(soundcloudapi.APIOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create soundcloud client")
	}

	return &SoundCloudBuilder{client: sc}, nil
}
