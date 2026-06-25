package update

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/builder"
	"github.com/mxpv/podsync/pkg/db"
	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/fs"
	"github.com/mxpv/podsync/pkg/model"
	"github.com/mxpv/podsync/pkg/ytdl"
)

type Downloader interface {
	Download(ctx context.Context, feedConfig *feed.Config, episode *model.Episode) (io.ReadCloser, error)
	// PlaylistMetadata is required by the platform builders (passed through to builder.New).
	PlaylistMetadata(ctx context.Context, url string) (metadata ytdl.PlaylistMetadata, err error)
}

type TokenList []string

type Manager struct {
	hostname   string
	downloader Downloader
	db         db.Storage
	fs         fs.Storage
	feeds      map[string]*feed.Config
	keys       map[model.Provider]feed.KeyProvider
}

func NewUpdater(
	feeds map[string]*feed.Config,
	keys map[model.Provider]feed.KeyProvider,
	hostname string,
	downloader Downloader,
	db db.Storage,
	fs fs.Storage,
) (*Manager, error) {
	return &Manager{
		hostname:   hostname,
		downloader: downloader,
		db:         db,
		fs:         fs,
		feeds:      feeds,
		keys:       keys,
	}, nil
}

func (u *Manager) Update(ctx context.Context, feedConfig *feed.Config) error {
	log.WithFields(log.Fields{
		"feed_id": feedConfig.ID,
		"format":  feedConfig.Format,
		"quality": feedConfig.Quality,
	}).Infof("-> updating %s", feedConfig.URL)

	started := time.Now()

	if err := u.updateFeed(ctx, feedConfig); err != nil {
		return errors.Wrap(err, "update failed")
	}

	// Fetch episodes for download
	episodesToDownload, err := u.fetchEpisodes(ctx, feedConfig)
	if err != nil {
		return errors.Wrap(err, "fetch episodes failed")
	}

	if err := u.downloadEpisodes(ctx, feedConfig, episodesToDownload); err != nil {
		return errors.Wrap(err, "download failed")
	}

	if err := u.cleanup(ctx, feedConfig); err != nil {
		log.WithError(err).Error("cleanup failed")
	}

	if err := u.buildXML(ctx, feedConfig); err != nil {
		return errors.Wrap(err, "xml build failed")
	}

	if err := u.buildOPML(ctx); err != nil {
		return errors.Wrap(err, "opml build failed")
	}

	elapsed := time.Since(started)
	log.Infof("successfully updated feed in %s", elapsed)
	return nil
}

// updateFeed pulls API for new episodes and saves them to database
func (u *Manager) updateFeed(ctx context.Context, feedConfig *feed.Config) error {
	info, err := builder.ParseURL(feedConfig.URL)
	if err != nil {
		return errors.Wrapf(err, "failed to parse URL: %s", feedConfig.URL)
	}

	keyProvider, ok := u.keys[info.Provider]
	if !ok {
		return errors.Errorf("key provider %q not loaded", info.Provider)
	}

	// Create an updater for this feed type
	provider, err := builder.New(ctx, info.Provider, keyProvider.Get(), u.downloader)
	if err != nil {
		return err
	}

	// Decide how far back to discover. Normally this is a shallow scan (page_size); when max_age
	// has been expanded beyond what was ever scanned, request a one-time deep scan to the cutoff.
	since, watermark := u.discoveryWindow(ctx, feedConfig)
	feedConfig.DiscoverSince = since
	if !since.IsZero() {
		log.WithField("feed_id", feedConfig.ID).Infof("max_age extended beyond scanned history, performing deep discovery back to %s", since.Format("2006-01-02"))
	}

	// Query API to get episodes
	log.Debug("building feed")
	result, err := provider.Build(ctx, feedConfig)
	feedConfig.DiscoverSince = time.Time{} // consumed by the builder; don't leak into the next cycle
	if err != nil {
		return err
	}

	log.Debugf("received %d episode(s) for %q", len(result.Episodes), result.Title)

	// Index the fresh API metadata. Used to restore metadata for cleaned episodes that older
	// versions stored without a title/description when they are resurrected.
	apiEpisodes := make(map[string]*model.Episode, len(result.Episodes))
	for _, episode := range result.Episodes {
		apiEpisodes[episode.ID] = episode
	}

	var (
		episodeSet = make(map[string]struct{})
		cleaned    []*model.Episode
		downloaded []*model.Episode
		// processed holds episodes that don't need a first download attempt (downloaded, cleaned,
		// or errored). Used to decide when a deep-scan catch-up is complete.
		processed = make(map[string]struct{})
	)
	if err := u.db.WalkEpisodes(ctx, feedConfig.ID, func(episode *model.Episode) error {
		switch episode.Status {
		case model.EpisodeDownloaded:
			downloaded = append(downloaded, episode)
			processed[episode.ID] = struct{}{}
		case model.EpisodeCleaned:
			cleaned = append(cleaned, episode)
			processed[episode.ID] = struct{}{}
		default:
			episodeSet[episode.ID] = struct{}{}
			if episode.Status == model.EpisodeError {
				processed[episode.ID] = struct{}{}
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Carry forward the discovery high-water mark so the deep scan only repeats while catching up.
	result.ScannedThrough = nextScannedThrough(since, watermark, result, processed, &feedConfig.Filters)

	if err := u.db.AddFeed(ctx, feedConfig.ID, result); err != nil {
		return err
	}

	for _, episode := range result.Episodes {
		delete(episodeSet, episode.ID)
	}

	// removing episodes that are no longer available in the feed and not downloaded or cleaned
	for id := range episodeSet {
		log.Infof("removing episode %q", id)
		err := u.db.DeleteEpisode(feedConfig.ID, id)
		if err != nil {
			return err
		}
	}

	if err := u.resurrectEpisodes(feedConfig, cleaned, downloaded, apiEpisodes); err != nil {
		return err
	}

	log.Debug("successfully saved updates to storage")
	return nil
}

// discoveryWindow decides whether the next build should perform a deep (max_age-driven) discovery
// pass. It returns the cutoff date to page back to (zero means the normal shallow discovery) and
// the current scan high-water mark to carry forward. A deep scan is requested only when max_age
// now reaches further back than anything previously scanned, so it happens once per expansion
// rather than every cycle.
func (u *Manager) discoveryWindow(ctx context.Context, feedConfig *feed.Config) (since, watermark time.Time) {
	maxAge := feedConfig.Filters.MaxAge
	if maxAge <= 0 {
		return time.Time{}, time.Time{}
	}

	existing, err := u.db.GetFeed(ctx, feedConfig.ID)
	if err != nil {
		if err == model.ErrNotFound {
			// Brand new feed: nothing scanned yet. Stay shallow to avoid a surprise back-catalog download.
			return time.Time{}, time.Time{}
		}
		log.WithError(err).WithField("feed_id", feedConfig.ID).Warn("failed to load existing feed for discovery-window calculation; staying shallow")
		return time.Time{}, time.Time{}
	}

	watermark = existing.ScannedThrough
	if watermark.IsZero() {
		// Upgrade path (feed predates this field): derive the mark from the oldest episode we have
		// already processed, so an expansion past it still triggers a one-time catch-up.
		watermark = oldestProcessedPubDate(existing.Episodes)
	}
	if watermark.IsZero() {
		return time.Time{}, watermark
	}

	cutoff := time.Now().AddDate(0, 0, -maxAge)
	if cutoff.Before(watermark) {
		return cutoff, watermark
	}
	return time.Time{}, watermark
}

// nextScannedThrough computes the discovery high-water mark to persist after a build. On a shallow
// scan it preserves the existing mark (or establishes a baseline for a new feed). On a deep scan it
// advances the mark to the cutoff only once every matching in-range episode is accounted for
// (downloaded, cleaned, or errored); until then it keeps the old mark so the next cycle scans deep
// again and the freshly discovered episodes are not pruned before they finish downloading.
func nextScannedThrough(since, watermark time.Time, result *model.Feed, processed map[string]struct{}, filters *feed.Filters) time.Time {
	if since.IsZero() {
		if !watermark.IsZero() {
			return watermark
		}
		return oldestPubDate(result.Episodes)
	}

	for _, episode := range result.Episodes {
		if !matchFilters(episode, filters) {
			continue
		}
		if _, ok := processed[episode.ID]; !ok {
			return watermark
		}
	}
	return since
}

func oldestProcessedPubDate(episodes []*model.Episode) time.Time {
	var oldest time.Time
	for _, episode := range episodes {
		if episode.Status != model.EpisodeDownloaded && episode.Status != model.EpisodeCleaned {
			continue
		}
		if episode.PubDate.IsZero() {
			continue
		}
		if oldest.IsZero() || episode.PubDate.Before(oldest) {
			oldest = episode.PubDate
		}
	}
	return oldest
}

func oldestPubDate(episodes []*model.Episode) time.Time {
	var oldest time.Time
	for _, episode := range episodes {
		if episode.PubDate.IsZero() {
			continue
		}
		if oldest.IsZero() || episode.PubDate.Before(oldest) {
			oldest = episode.PubDate
		}
	}
	return oldest
}

// resurrectEpisodes re-queues previously cleaned (soft-deleted) episodes that match the current
// filters again, e.g. when max_age is increased to cover episodes that were already downloaded
// and deleted. It works off the database records directly, so re-inclusion does not depend on the
// episode still being within the page_size API window.
func (u *Manager) resurrectEpisodes(feedConfig *feed.Config, cleaned, downloaded []*model.Episode, apiEpisodes map[string]*model.Episode) error {
	filters := &feedConfig.Filters

	// Cheap pre-filter on duration/age, which are always present on the record.
	// Title/description filters are applied later, after any missing metadata is recovered.
	var candidates []*model.Episode
	for _, episode := range cleaned {
		if matchDurationAndAge(episode, filters, log.WithField("episode_id", episode.ID)) {
			candidates = append(candidates, episode)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Resurrect only episodes that would survive the cleanup policy, otherwise
	// every update cycle would re-download episodes just for cleanup to delete them again.
	if feedConfig.Clean != nil && feedConfig.Clean.KeepLast > 0 {
		candidates = keepLastSurvivors(candidates, downloaded, feedConfig.Clean.KeepLast)
	}

	for _, episode := range candidates {
		// Episodes cleaned by older versions were stored without a title/description. Recover that
		// metadata from the current feed query (the deep discovery scan brings in-range episodes
		// back into the result), so title/description filters can be evaluated and the re-downloaded
		// episode has metadata for the feed build.
		if episode.Title == "" {
			if api, ok := apiEpisodes[episode.ID]; ok {
				episode.Title = api.Title
				episode.Description = api.Description
			}
		}

		// Without a title the episode cannot be published; leave it cleaned until it reappears in
		// the feed query (e.g. once a deep scan covers it) and we can recover its metadata.
		if episode.Title == "" {
			continue
		}

		if !matchFilters(episode, filters) {
			continue
		}

		log.WithField("episode_id", episode.ID).Info("re-queueing previously cleaned episode that matches filters again")
		title, description := episode.Title, episode.Description
		if err := u.db.UpdateEpisode(feedConfig.ID, episode.ID, func(saved *model.Episode) error {
			saved.Status = model.EpisodeNew
			if saved.Title == "" {
				saved.Title = title
			}
			if saved.Description == "" {
				saved.Description = description
			}
			return nil
		}); err != nil {
			return errors.Wrapf(err, "failed to re-queue cleaned episode %s", episode.ID)
		}
	}

	return nil
}

// keepLastSurvivors returns the subset of candidates that would remain after applying the
// keep_last cleanup policy, given the currently downloaded episodes. Ranking is by PubDate.
func keepLastSurvivors(candidates, downloaded []*model.Episode, keepLast int) []*model.Episode {
	all := make([]*model.Episode, 0, len(downloaded)+len(candidates))
	all = append(all, downloaded...)
	all = append(all, candidates...)
	sort.Slice(all, func(i, j int) bool {
		return all[i].PubDate.After(all[j].PubDate)
	})

	if len(all) > keepLast {
		all = all[:keepLast]
	}

	surviving := make(map[string]struct{}, len(all))
	for _, episode := range all {
		surviving[episode.ID] = struct{}{}
	}

	kept := candidates[:0]
	for _, episode := range candidates {
		if _, ok := surviving[episode.ID]; ok {
			kept = append(kept, episode)
		}
	}
	return kept
}

func (u *Manager) fetchEpisodes(ctx context.Context, feedConfig *feed.Config) ([]*model.Episode, error) {
	var (
		feedID       = feedConfig.ID
		downloadList []*model.Episode
		pageSize     = feedConfig.PageSize
	)

	log.WithField("page_size", pageSize).Info("fetching episodes for download")

	// Build the list of files to download
	err := u.db.WalkEpisodes(ctx, feedID, func(episode *model.Episode) error {
		var (
			logger = log.WithFields(log.Fields{"episode_id": episode.ID})
		)
		if episode.Status != model.EpisodeNew && episode.Status != model.EpisodeError {
			// File already downloaded
			logger.Infof("skipping due to already downloaded")
			return nil
		}

		if !matchFilters(episode, &feedConfig.Filters) {
			return nil
		}

		// Limit the number of episodes downloaded at once
		pageSize--
		if pageSize < 0 {
			return nil
		}

		log.Debugf("adding %s (%q) to queue", episode.ID, episode.Title)
		downloadList = append(downloadList, episode)
		return nil
	})

	if err != nil {
		return nil, errors.Wrapf(err, "failed to build update list")
	}

	return downloadList, nil
}

func (u *Manager) downloadEpisodes(ctx context.Context, feedConfig *feed.Config, downloadList []*model.Episode) error {
	var (
		downloadCount = len(downloadList)
		downloaded    = 0
		feedID        = feedConfig.ID
	)

	if downloadCount > 0 {
		log.Infof("download count: %d", downloadCount)
	} else {
		log.Info("no episodes to download")
		return nil
	}

	// Download pending episodes

	for idx, episode := range downloadList {
		var (
			logger      = log.WithFields(log.Fields{"index": idx, "episode_id": episode.ID})
			episodeName = feed.EpisodeName(feedConfig, episode)
		)

		// Check whether episode already exists
		size, err := u.fs.Size(ctx, fmt.Sprintf("%s/%s", feedID, episodeName))
		if err == nil {
			logger.Infof("episode %q already exists on disk", episode.ID)

			// File already exists, update file status and disk size
			if err := u.db.UpdateEpisode(feedID, episode.ID, func(episode *model.Episode) error {
				episode.Size = size
				episode.Status = model.EpisodeDownloaded
				return nil
			}); err != nil {
				logger.WithError(err).Error("failed to update file info")
				return err
			}

			continue
		} else if os.IsNotExist(err) {
			// Will download, do nothing here
		} else {
			logger.WithError(err).Error("failed to stat file")
			return err
		}

		// Download episode to disk
		// We download the episode to a temp directory first to avoid downloading this file by clients
		// while still being processed by youtube-dl (e.g. a file is being downloaded from YT or encoding in progress)

		logger.Infof("! downloading episode %s", episode.VideoURL)
		tempFile, err := u.downloader.Download(ctx, feedConfig, episode)
		if err != nil {
			// YouTube might block host with HTTP Error 429: Too Many Requests
			// We still need to generate XML, so just stop sending download requests and
			// retry next time
			if err == ytdl.ErrTooManyRequests {
				logger.Warn("server responded with a 'Too Many Requests' error")
				break
			}

			// Execute episode download error hooks
			if len(feedConfig.OnEpisodeDownloadError) > 0 {
				env := []string{
					"FEED_NAME=" + feedID,
					"EPISODE_TITLE=" + episode.Title,
					"ERROR_MESSAGE=" + err.Error(),
				}

				for i, hook := range feedConfig.OnEpisodeDownloadError {
					if hookErr := hook.Invoke(env); hookErr != nil {
						logger.Errorf("failed to execute episode download error hook %d: %v", i+1, hookErr)
					} else {
						logger.Infof("episode download error hook %d executed successfully", i+1)
					}
				}
			}

			if err := u.db.UpdateEpisode(feedID, episode.ID, func(episode *model.Episode) error {
				episode.Status = model.EpisodeError
				return nil
			}); err != nil {
				return err
			}

			continue
		}

		logger.Debug("copying file")
		fileSize, err := u.fs.Create(ctx, fmt.Sprintf("%s/%s", feedID, episodeName), tempFile)
		tempFile.Close()
		if err != nil {
			logger.WithError(err).Error("failed to copy file")
			return err
		}

		// Execute post episode download hooks
		if len(feedConfig.PostEpisodeDownload) > 0 {
			env := []string{
				"EPISODE_FILE=" + fmt.Sprintf("%s/%s", feedID, episodeName),
				"FEED_NAME=" + feedID,
				"EPISODE_TITLE=" + episode.Title,
			}

			for i, hook := range feedConfig.PostEpisodeDownload {
				if err := hook.Invoke(env); err != nil {
					logger.Errorf("failed to execute post episode download hook %d: %v", i+1, err)
				} else {
					logger.Infof("post episode download hook %d executed successfully", i+1)
				}
			}
		}

		// Update file status in database

		logger.Infof("successfully downloaded file %q", episode.ID)
		if err := u.db.UpdateEpisode(feedID, episode.ID, func(episode *model.Episode) error {
			episode.Size = fileSize
			episode.Status = model.EpisodeDownloaded
			return nil
		}); err != nil {
			return err
		}

		downloaded++
	}

	log.Infof("downloaded %d episode(s)", downloaded)
	return nil
}

func (u *Manager) buildXML(ctx context.Context, feedConfig *feed.Config) error {
	f, err := u.db.GetFeed(ctx, feedConfig.ID)
	if err != nil {
		return err
	}

	// Build iTunes XML feed with data received from builder
	log.Debug("building iTunes podcast feed")
	podcast, err := feed.Build(ctx, f, feedConfig, u.hostname)
	if err != nil {
		return err
	}

	var (
		reader  = bytes.NewReader([]byte(podcast.String()))
		xmlName = fmt.Sprintf("%s.xml", feedConfig.ID)
	)

	if _, err := u.fs.Create(ctx, xmlName, reader); err != nil {
		return errors.Wrap(err, "failed to upload new XML feed")
	}

	return nil
}

func (u *Manager) buildOPML(ctx context.Context) error {
	// Build OPML with data received from builder
	log.Debug("building podcast OPML")
	opml, err := feed.BuildOPML(ctx, u.feeds, u.db, u.hostname)
	if err != nil {
		return err
	}

	var (
		reader  = bytes.NewReader([]byte(opml))
		xmlName = fmt.Sprintf("%s.opml", "podsync")
	)

	if _, err := u.fs.Create(ctx, xmlName, reader); err != nil {
		return errors.Wrap(err, "failed to upload OPML")
	}

	return nil
}

func (u *Manager) cleanup(ctx context.Context, feedConfig *feed.Config) error {
	var (
		feedID = feedConfig.ID
		logger = log.WithField("feed_id", feedID)
		list   []*model.Episode
		result *multierror.Error
	)

	if feedConfig.Clean == nil {
		logger.Debug("no cleanup policy configured")
		return nil
	}

	count := feedConfig.Clean.KeepLast
	if count < 1 {
		logger.Info("nothing to clean")
		return nil
	}

	logger.WithField("count", count).Info("running cleaner")
	if err := u.db.WalkEpisodes(ctx, feedConfig.ID, func(episode *model.Episode) error {
		if episode.Status == model.EpisodeDownloaded {
			list = append(list, episode)
		}
		return nil
	}); err != nil {
		return err
	}

	if count > len(list) {
		return nil
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].PubDate.After(list[j].PubDate)
	})

	for _, episode := range list[count:] {
		logger.WithField("episode_id", episode.ID).Infof("deleting %q", episode.Title)

		var (
			episodeName = feed.EpisodeName(feedConfig, episode)
			path        = fmt.Sprintf("%s/%s", feedConfig.ID, episodeName)
		)

		err := u.fs.Delete(ctx, path)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				logger.WithError(err).Errorf("failed to delete episode file: %s", episode.ID)
				result = multierror.Append(result, errors.Wrapf(err, "failed to delete episode: %s", episode.ID))
				continue
			}

			logger.WithField("episode_id", episode.ID).Info("episode was not found - file does not exist")
		}

		// Only the file is removed; the episode metadata is retained so the record
		// can be resurrected later (e.g. when max_age is increased) without losing
		// its title/description. Cleaned episodes are already excluded from the feed
		// by their status, so keeping the metadata has no effect on the output.
		if err := u.db.UpdateEpisode(feedID, episode.ID, func(episode *model.Episode) error {
			episode.Status = model.EpisodeCleaned
			return nil
		}); err != nil {
			result = multierror.Append(result, errors.Wrapf(err, "failed to set state for cleaned episode: %s", episode.ID))
			continue
		}
	}

	return result.ErrorOrNil()
}
