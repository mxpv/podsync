package migrate

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/db"
	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/fs"
	"github.com/mxpv/podsync/pkg/model"
)

type Service struct {
	feeds  map[string]*feed.Config
	db     db.Storage
	fs     fs.Storage
	dryRun bool
}

type Result struct {
	Feeds       int
	Episodes    int
	Migrated    int
	AlreadyGood int
	MissingOld  int
	SkippedDueToExistingTarget int
}

func New(feeds map[string]*feed.Config, db db.Storage, storage fs.Storage, dryRun bool) *Service {
	return &Service{feeds: feeds, db: db, fs: storage, dryRun: dryRun}
}

func (s *Service) Run(ctx context.Context) (*Result, error) {
	result := &Result{Feeds: len(s.feeds)}
	var allErr *multierror.Error

	feedIDs := make([]string, 0, len(s.feeds))
	for feedID := range s.feeds {
		feedIDs = append(feedIDs, feedID)
	}
	sort.Strings(feedIDs)

	for _, feedID := range feedIDs {
		cfg := s.feeds[feedID]
		logger := log.WithField("feed_id", feedID)
		logger.Infof("starting filename migration (dry_run=%t)", s.dryRun)

		err := s.db.WalkEpisodes(ctx, feedID, func(episode *model.Episode) error {
			if episode.Status != model.EpisodeDownloaded {
				return nil
			}

			result.Episodes++
			newName := feed.EpisodeName(cfg, episode)
			newPath := fmt.Sprintf("%s/%s", feedID, newName)
			legacyName := feed.LegacyEpisodeName(cfg, episode)
			legacyPath := fmt.Sprintf("%s/%s", feedID, legacyName)

			newSize, newErr := s.fs.Size(ctx, newPath)
			if newErr == nil {
				result.AlreadyGood++
				return s.updateEpisode(feedID, episode.ID, newSize)
			}
			if !os.IsNotExist(newErr) {
				return errors.Wrapf(newErr, "failed to stat target file %q", newPath)
			}

			if _, legacyErr := s.fs.Size(ctx, legacyPath); legacyErr != nil {
				if os.IsNotExist(legacyErr) {
					result.MissingOld++
					return nil
				}
				return errors.Wrapf(legacyErr, "failed to stat legacy file %q", legacyPath)
			}

			if s.dryRun {
				result.Migrated++
				return nil
			}

			if _, existingErr := s.fs.Size(ctx, newPath); existingErr == nil {
				result.SkippedDueToExistingTarget++
				return nil
			}

			legacyFile, err := s.fs.Open(legacyPath)
			if err != nil {
				return errors.Wrapf(err, "failed to open legacy file %q", legacyPath)
			}

			size, err := s.fs.Create(ctx, newPath, legacyFile)
			closeErr := legacyFile.Close()
			if err != nil {
				return errors.Wrapf(err, "failed to create migrated file %q", newPath)
			}
			if closeErr != nil {
				return errors.Wrapf(closeErr, "failed to close legacy file %q", legacyPath)
			}

			if err := s.fs.Delete(ctx, legacyPath); err != nil && !os.IsNotExist(err) {
				return errors.Wrapf(err, "failed to delete legacy file %q", legacyPath)
			}

			if err := s.updateEpisode(feedID, episode.ID, size); err != nil {
				return err
			}

			result.Migrated++
			return nil
		})
		if err != nil {
			allErr = multierror.Append(allErr, errors.Wrapf(err, "feed %s migration failed", feedID))
		}
	}

	return result, allErr.ErrorOrNil()
}

func (s *Service) updateEpisode(feedID string, episodeID string, size int64) error {
	if s.dryRun {
		return nil
	}

	return s.db.UpdateEpisode(feedID, episodeID, func(episode *model.Episode) error {
		episode.Size = size
		episode.Status = model.EpisodeDownloaded
		return nil
	})
}
