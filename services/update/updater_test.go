package update

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/db"
	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/fs"
	"github.com/mxpv/podsync/pkg/model"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()

	database, err := db.NewBadger(&db.Config{Dir: t.TempDir()})
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	return &Manager{db: database}
}

func seedEpisodes(t *testing.T, manager *Manager, feedID string, episodes []*model.Episode) {
	t.Helper()

	err := manager.db.AddFeed(context.Background(), feedID, &model.Feed{
		ID:       feedID,
		Episodes: episodes,
	})
	require.NoError(t, err)
}

func TestDiscoveryWindowNoMaxAgeStaysShallow(t *testing.T) {
	manager := newTestManager(t)
	ctx := context.Background()

	require.NoError(t, manager.db.AddFeed(ctx, "1", &model.Feed{ID: "1", ScannedThrough: time.Now().AddDate(0, 0, -100)}))

	since, _ := manager.discoveryWindow(ctx, &feed.Config{ID: "1"})
	assert.True(t, since.IsZero(), "no max_age must never trigger a deep scan")
}

func TestDiscoveryWindowNewFeedStaysShallow(t *testing.T) {
	manager := newTestManager(t)
	ctx := context.Background()

	// No feed record exists yet
	since, _ := manager.discoveryWindow(ctx, &feed.Config{ID: "1", Filters: feed.Filters{MaxAge: 200}})
	assert.True(t, since.IsZero(), "a brand new feed must not deep scan its back-catalog")
}

func TestDiscoveryWindowTriggersDeepScanWhenExpandedPastWatermark(t *testing.T) {
	manager := newTestManager(t)
	ctx := context.Background()

	watermark := time.Now().AddDate(0, 0, -100)
	require.NoError(t, manager.db.AddFeed(ctx, "1", &model.Feed{ID: "1", ScannedThrough: watermark}))

	since, wm := manager.discoveryWindow(ctx, &feed.Config{ID: "1", Filters: feed.Filters{MaxAge: 200}})
	require.False(t, since.IsZero(), "max_age reaching past the watermark must trigger a deep scan")
	assert.WithinDuration(t, time.Now().AddDate(0, 0, -200), since, time.Hour)
	assert.Equal(t, watermark.Unix(), wm.Unix())
}

func TestDiscoveryWindowStaysShallowWithinWatermark(t *testing.T) {
	manager := newTestManager(t)
	ctx := context.Background()

	watermark := time.Now().AddDate(0, 0, -100)
	require.NoError(t, manager.db.AddFeed(ctx, "1", &model.Feed{ID: "1", ScannedThrough: watermark}))

	since, _ := manager.discoveryWindow(ctx, &feed.Config{ID: "1", Filters: feed.Filters{MaxAge: 50}})
	assert.True(t, since.IsZero(), "max_age within the scanned window must not deep scan")
}

func TestDiscoveryWindowUpgradePathDerivesWatermarkFromEpisodes(t *testing.T) {
	manager := newTestManager(t)
	ctx := context.Background()

	// Feed predating ScannedThrough: derive the mark from the oldest processed episode
	require.NoError(t, manager.db.AddFeed(ctx, "1", &model.Feed{
		ID: "1",
		Episodes: []*model.Episode{
			{ID: "old", PubDate: time.Now().AddDate(0, 0, -100), Status: model.EpisodeCleaned},
			{ID: "recent", PubDate: time.Now().AddDate(0, 0, -2), Status: model.EpisodeDownloaded},
		},
	}))

	since, _ := manager.discoveryWindow(ctx, &feed.Config{ID: "1", Filters: feed.Filters{MaxAge: 200}})
	require.False(t, since.IsZero(), "expansion past the oldest processed episode must deep scan")
	assert.WithinDuration(t, time.Now().AddDate(0, 0, -200), since, time.Hour)
}

func TestNextScannedThrough(t *testing.T) {
	filters := &feed.Filters{}
	watermark := time.Now().AddDate(0, 0, -100)
	since := time.Now().AddDate(0, 0, -200)

	result := &model.Feed{Episodes: []*model.Episode{
		{ID: "a", PubDate: time.Now().AddDate(0, 0, -10)},
		{ID: "b", PubDate: time.Now().AddDate(0, 0, -150)},
	}}

	t.Run("shallow keeps existing watermark", func(t *testing.T) {
		got := nextScannedThrough(time.Time{}, watermark, result, map[string]struct{}{}, filters)
		assert.Equal(t, watermark, got)
	})

	t.Run("shallow without watermark establishes baseline from oldest episode", func(t *testing.T) {
		got := nextScannedThrough(time.Time{}, time.Time{}, result, map[string]struct{}{}, filters)
		assert.WithinDuration(t, time.Now().AddDate(0, 0, -150), got, time.Hour)
	})

	t.Run("deep scan holds watermark until all matching episodes are processed", func(t *testing.T) {
		// Only "a" is processed; "b" is still pending
		got := nextScannedThrough(since, watermark, result, map[string]struct{}{"a": {}}, filters)
		assert.Equal(t, watermark, got, "must not advance while episodes are still pending download")
	})

	t.Run("deep scan advances once all matching episodes are processed", func(t *testing.T) {
		got := nextScannedThrough(since, watermark, result, map[string]struct{}{"a": {}, "b": {}}, filters)
		assert.Equal(t, since, got)
	})

	t.Run("deep scan advances when only non-matching episodes are pending", func(t *testing.T) {
		// "b" is excluded by the filter, so it must not block catch-up completion
		f := &feed.Filters{MaxDuration: 60}
		r := &model.Feed{Episodes: []*model.Episode{
			{ID: "a", Duration: 30, PubDate: time.Now().AddDate(0, 0, -10)},
			{ID: "b", Duration: 600, PubDate: time.Now().AddDate(0, 0, -150)},
		}}
		got := nextScannedThrough(since, watermark, r, map[string]struct{}{"a": {}}, f)
		assert.Equal(t, since, got)
	})
}

func TestDiscoveryWindowExistingFeedNoHistoryStaysShallow(t *testing.T) {
	manager := newTestManager(t)
	ctx := context.Background()

	// Existing feed record with no ScannedThrough and only un-processed (New) episodes:
	// there is no baseline to compare max_age against, so discovery must stay shallow.
	require.NoError(t, manager.db.AddFeed(ctx, "1", &model.Feed{
		ID:       "1",
		Episodes: []*model.Episode{{ID: "x", Status: model.EpisodeNew, PubDate: time.Now().AddDate(0, 0, -10)}},
	}))

	since, _ := manager.discoveryWindow(ctx, &feed.Config{ID: "1", Filters: feed.Filters{MaxAge: 200}})
	assert.True(t, since.IsZero(), "an existing feed with no scan baseline must stay shallow")
}

func TestOldestProcessedPubDate(t *testing.T) {
	now := time.Now()

	t.Run("ignores unprocessed and undated episodes", func(t *testing.T) {
		episodes := []*model.Episode{
			{ID: "new", Status: model.EpisodeNew, PubDate: now.AddDate(0, 0, -200)}, // not processed
			{ID: "nodate", Status: model.EpisodeDownloaded},                         // zero PubDate
			{ID: "dl", Status: model.EpisodeDownloaded, PubDate: now.AddDate(0, 0, -50)},
			{ID: "cl", Status: model.EpisodeCleaned, PubDate: now.AddDate(0, 0, -100)},
		}
		assert.WithinDuration(t, now.AddDate(0, 0, -100), oldestProcessedPubDate(episodes), time.Hour)
	})

	t.Run("returns zero when nothing processed", func(t *testing.T) {
		episodes := []*model.Episode{{ID: "new", Status: model.EpisodeNew, PubDate: now}}
		assert.True(t, oldestProcessedPubDate(episodes).IsZero())
	})
}

func TestOldestPubDate(t *testing.T) {
	now := time.Now()
	episodes := []*model.Episode{
		{ID: "nodate"}, // zero PubDate is ignored
		{ID: "a", PubDate: now.AddDate(0, 0, -10)},
		{ID: "b", PubDate: now.AddDate(0, 0, -30)},
	}
	assert.WithinDuration(t, now.AddDate(0, 0, -30), oldestPubDate(episodes), time.Hour)
}

func TestCleanupPreservesMetadata(t *testing.T) {
	ctx := context.Background()

	dataDir := t.TempDir()
	storage, err := fs.NewLocal(dataDir, false, false)
	require.NoError(t, err)

	manager := newTestManager(t)
	manager.fs = storage

	var (
		recent = time.Now().AddDate(0, 0, -1)
		old    = time.Now().AddDate(0, 0, -30)
	)
	seedEpisodes(t, manager, "1", []*model.Episode{
		{ID: "recent", Title: "Recent", Description: "Recent desc", PubDate: recent, Status: model.EpisodeDownloaded},
		{ID: "old", Title: "Old", Description: "Old desc", PubDate: old, Status: model.EpisodeDownloaded},
	})

	cfg := &feed.Config{ID: "1", Clean: &feed.Cleanup{KeepLast: 1}}

	// Create the file that cleanup is expected to delete
	_, err = storage.Create(ctx, "1/old.mp4", strings.NewReader("data"))
	require.NoError(t, err)

	require.NoError(t, manager.cleanup(ctx, cfg))

	// The old episode is soft-deleted but keeps its metadata so it can be resurrected later
	old1, err := manager.db.GetEpisode(ctx, "1", "old")
	require.NoError(t, err)
	assert.Equal(t, model.EpisodeCleaned, old1.Status)
	assert.Equal(t, "Old", old1.Title)
	assert.Equal(t, "Old desc", old1.Description)

	// Its file is removed from storage
	_, err = storage.Size(ctx, "1/old.mp4")
	assert.Error(t, err)

	// The recent episode is untouched
	recent1, err := manager.db.GetEpisode(ctx, "1", "recent")
	require.NoError(t, err)
	assert.Equal(t, model.EpisodeDownloaded, recent1.Status)
}

func TestResurrectCleanedEpisodeMatchingFiltersAgain(t *testing.T) {
	manager := newTestManager(t)
	ctx := context.Background()

	pubDate := time.Now().AddDate(0, 0, -30)
	// Cleanup now preserves metadata, so the cleaned record keeps its title/description
	cleaned := []*model.Episode{
		{ID: "old", Title: "Old Episode", Description: "Description", PubDate: pubDate, Status: model.EpisodeCleaned},
	}
	seedEpisodes(t, manager, "1", cleaned)

	cfg := &feed.Config{ID: "1", Filters: feed.Filters{MaxAge: 60}}

	_, err := manager.resurrectEpisodes(cfg, cleaned, nil, nil)
	require.NoError(t, err)

	episode, err := manager.db.GetEpisode(ctx, "1", "old")
	require.NoError(t, err)
	assert.Equal(t, model.EpisodeNew, episode.Status)
	assert.Equal(t, "Old Episode", episode.Title)
	assert.Equal(t, "Description", episode.Description)
}

func TestResurrectRecoversWipedMetadataFromAPIResult(t *testing.T) {
	manager := newTestManager(t)
	ctx := context.Background()

	pubDate := time.Now().AddDate(0, 0, -30)
	// A legacy cleaned episode whose title/description were wiped by older cleanup
	cleaned := []*model.Episode{
		{ID: "old", Duration: 600, PubDate: pubDate, Status: model.EpisodeCleaned},
	}
	seedEpisodes(t, manager, "1", cleaned)

	// The deep-discovery scan brings the episode back into the feed query with fresh metadata
	apiEpisodes := map[string]*model.Episode{
		"old": {ID: "old", Title: "Recovered Title", Description: "Recovered Description"},
	}

	cfg := &feed.Config{ID: "1", Filters: feed.Filters{MaxAge: 60, MinDuration: 120}}

	_, err := manager.resurrectEpisodes(cfg, cleaned, nil, apiEpisodes)
	require.NoError(t, err)

	episode, err := manager.db.GetEpisode(ctx, "1", "old")
	require.NoError(t, err)
	assert.Equal(t, model.EpisodeNew, episode.Status)
	assert.Equal(t, "Recovered Title", episode.Title)
	assert.Equal(t, "Recovered Description", episode.Description)
}

func TestResurrectSkipsWipedEpisodeWithoutAPIMetadata(t *testing.T) {
	manager := newTestManager(t)
	ctx := context.Background()

	pubDate := time.Now().AddDate(0, 0, -30)
	// Legacy cleaned episode with no metadata, and not present in the current feed query
	cleaned := []*model.Episode{
		{ID: "old", Duration: 600, PubDate: pubDate, Status: model.EpisodeCleaned},
	}
	seedEpisodes(t, manager, "1", cleaned)

	cfg := &feed.Config{ID: "1", Filters: feed.Filters{MaxAge: 60, MinDuration: 120}}

	// No apiEpisodes: nothing to title the episode with, so it stays cleaned for now
	_, err := manager.resurrectEpisodes(cfg, cleaned, nil, map[string]*model.Episode{})
	require.NoError(t, err)

	episode, err := manager.db.GetEpisode(ctx, "1", "old")
	require.NoError(t, err)
	assert.Equal(t, model.EpisodeCleaned, episode.Status)
}

func TestResurrectSkipsEpisodeFailingTextFilter(t *testing.T) {
	manager := newTestManager(t)
	ctx := context.Background()

	pubDate := time.Now().AddDate(0, 0, -10)
	// Passes duration/age, but the not_title filter excludes it after metadata is available
	cleaned := []*model.Episode{
		{ID: "old", Title: "Live Stream", PubDate: pubDate, Status: model.EpisodeCleaned},
	}
	seedEpisodes(t, manager, "1", cleaned)

	cfg := &feed.Config{ID: "1", Filters: feed.Filters{MaxAge: 60, NotTitle: "(?i)live"}}

	_, err := manager.resurrectEpisodes(cfg, cleaned, nil, nil)
	require.NoError(t, err)

	episode, err := manager.db.GetEpisode(ctx, "1", "old")
	require.NoError(t, err)
	assert.Equal(t, model.EpisodeCleaned, episode.Status)
}

func TestResurrectSkipsEpisodeNotMatchingFilters(t *testing.T) {
	manager := newTestManager(t)
	ctx := context.Background()

	pubDate := time.Now().AddDate(0, 0, -90)
	cleaned := []*model.Episode{
		{ID: "old", Title: "Old Episode", PubDate: pubDate, Status: model.EpisodeCleaned},
	}
	seedEpisodes(t, manager, "1", cleaned)

	cfg := &feed.Config{ID: "1", Filters: feed.Filters{MaxAge: 60}}

	_, err := manager.resurrectEpisodes(cfg, cleaned, nil, nil)
	require.NoError(t, err)

	episode, err := manager.db.GetEpisode(ctx, "1", "old")
	require.NoError(t, err)
	assert.Equal(t, model.EpisodeCleaned, episode.Status)
}

func TestResurrectSkipsEpisodeThatWouldBeCleanedAgain(t *testing.T) {
	manager := newTestManager(t)
	ctx := context.Background()

	var (
		oldDate    = time.Now().AddDate(0, 0, -30)
		recentDate = time.Now().AddDate(0, 0, -1)
		downloaded = &model.Episode{ID: "recent", Title: "Recent", PubDate: recentDate, Status: model.EpisodeDownloaded}
		old        = &model.Episode{ID: "old", Title: "Old Episode", PubDate: oldDate, Status: model.EpisodeCleaned}
	)

	seedEpisodes(t, manager, "1", []*model.Episode{old, downloaded})

	// keep_last = 1 and a newer downloaded episode exists, so resurrecting
	// the old one would result in cleanup deleting it again
	cfg := &feed.Config{ID: "1", Clean: &feed.Cleanup{KeepLast: 1}}

	_, err := manager.resurrectEpisodes(cfg, []*model.Episode{old}, []*model.Episode{downloaded}, nil)
	require.NoError(t, err)

	episode, err := manager.db.GetEpisode(ctx, "1", "old")
	require.NoError(t, err)
	assert.Equal(t, model.EpisodeCleaned, episode.Status)
}

func TestResurrectKeepsEpisodeWithinKeepLastWindow(t *testing.T) {
	manager := newTestManager(t)
	ctx := context.Background()

	var (
		oldDate    = time.Now().AddDate(0, 0, -30)
		recentDate = time.Now().AddDate(0, 0, -1)
		downloaded = &model.Episode{ID: "recent", Title: "Recent", PubDate: recentDate, Status: model.EpisodeDownloaded}
		old        = &model.Episode{ID: "old", Title: "Old Episode", PubDate: oldDate, Status: model.EpisodeCleaned}
	)

	seedEpisodes(t, manager, "1", []*model.Episode{old, downloaded})

	cfg := &feed.Config{ID: "1", Clean: &feed.Cleanup{KeepLast: 2}}

	_, err := manager.resurrectEpisodes(cfg, []*model.Episode{old}, []*model.Episode{downloaded}, nil)
	require.NoError(t, err)

	episode, err := manager.db.GetEpisode(ctx, "1", "old")
	require.NoError(t, err)
	assert.Equal(t, model.EpisodeNew, episode.Status)
	assert.Equal(t, "Old Episode", episode.Title)
}

func TestResurrectReportsResurrectedIDsHoldingWatermark(t *testing.T) {
	manager := newTestManager(t)

	pubDate := time.Now().AddDate(0, 0, -150)
	old := &model.Episode{ID: "old", Title: "Old Episode", PubDate: pubDate, Status: model.EpisodeCleaned}
	seedEpisodes(t, manager, "1", []*model.Episode{old})

	cfg := &feed.Config{ID: "1", Filters: feed.Filters{MaxAge: 200}}

	resurrected, err := manager.resurrectEpisodes(cfg, []*model.Episode{old}, nil, nil)
	require.NoError(t, err)
	require.Contains(t, resurrected, "old", "a re-queued episode must be reported so the watermark accounts for it")

	// The episode is now pending download again. Removing it from the processed set must keep the
	// deep-scan high-water mark from advancing past it, so the next cycle stays deep and the
	// shallow removal loop cannot prune it before it downloads.
	var (
		watermark = time.Now().AddDate(0, 0, -100)
		since     = time.Now().AddDate(0, 0, -200)
		result    = &model.Feed{Episodes: []*model.Episode{old}}
		processed = map[string]struct{}{"old": {}} // it was cleaned, so initially counted as done
	)
	for id := range resurrected {
		delete(processed, id)
	}

	got := nextScannedThrough(since, watermark, result, processed, &cfg.Filters)
	assert.Equal(t, watermark, got, "watermark must not advance while a resurrected episode is still pending")
}

func TestResurrectWithNoCleanedEpisodesDoesNothing(t *testing.T) {
	manager := newTestManager(t)
	ctx := context.Background()

	pubDate := time.Now().AddDate(0, 0, -1)
	seedEpisodes(t, manager, "1", []*model.Episode{
		{ID: "done", Title: "Done", PubDate: pubDate, Status: model.EpisodeDownloaded},
	})

	cfg := &feed.Config{ID: "1"}

	_, err := manager.resurrectEpisodes(cfg, nil, nil, nil)
	require.NoError(t, err)

	episode, err := manager.db.GetEpisode(ctx, "1", "done")
	require.NoError(t, err)
	assert.Equal(t, model.EpisodeDownloaded, episode.Status)
}
