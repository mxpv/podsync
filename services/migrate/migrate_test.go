package migrate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mxpv/podsync/pkg/db"
	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/fs"
	"github.com/mxpv/podsync/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testDB struct {
	episodes map[string]map[string]*model.Episode
}

func newTestDB() *testDB {
	return &testDB{episodes: map[string]map[string]*model.Episode{}}
}

func (t *testDB) Close() error          { return nil }
func (t *testDB) Version() (int, error) { return 1, nil }
func (t *testDB) AddFeed(_ context.Context, _ string, _ *model.Feed) error {
	return errors.New("not implemented")
}
func (t *testDB) GetFeed(_ context.Context, _ string) (*model.Feed, error) {
	return nil, errors.New("not implemented")
}
func (t *testDB) WalkFeeds(_ context.Context, _ func(feed *model.Feed) error) error { return nil }
func (t *testDB) DeleteFeed(_ context.Context, _ string) error                      { return errors.New("not implemented") }
func (t *testDB) DeleteEpisode(_ string, _ string) error                            { return errors.New("not implemented") }

func (t *testDB) GetEpisode(_ context.Context, feedID string, episodeID string) (*model.Episode, error) {
	if f, ok := t.episodes[feedID]; ok {
		if ep, ok := f[episodeID]; ok {
			return ep, nil
		}
	}
	return nil, os.ErrNotExist
}

func (t *testDB) UpdateEpisode(feedID string, episodeID string, cb func(episode *model.Episode) error) error {
	ep, err := t.GetEpisode(context.Background(), feedID, episodeID)
	if err != nil {
		return err
	}
	return cb(ep)
}

func (t *testDB) WalkEpisodes(_ context.Context, feedID string, cb func(episode *model.Episode) error) error {
	for _, ep := range t.episodes[feedID] {
		if err := cb(ep); err != nil {
			return err
		}
	}
	return nil
}

var _ db.Storage = (*testDB)(nil)

type flakySizeStorage struct {
	fs.Storage
	targetPath   string
	targetChecks int
}

func (s *flakySizeStorage) Size(ctx context.Context, name string) (int64, error) {
	if name == s.targetPath {
		s.targetChecks++
		if s.targetChecks == 1 {
			return 0, os.ErrNotExist
		}
		if s.targetChecks == 2 {
			return 0, errors.New("transient stat failure")
		}
	}
	return s.Storage.Size(ctx, name)
}

func TestRunMigratesLegacyFilename(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	storage, err := fs.NewLocal(tmpDir, false, false)
	require.NoError(t, err)

	tdb := newTestDB()
	feedID := "A"
	episode := &model.Episode{
		ID:      "abc123",
		Title:   "Title / Needs Cleanup",
		PubDate: time.Date(2026, 2, 8, 9, 0, 0, 0, time.UTC),
		Status:  model.EpisodeDownloaded,
	}
	tdb.episodes[feedID] = map[string]*model.Episode{episode.ID: episode}

	cfg := &feed.Config{
		ID:               feedID,
		Format:           model.FormatVideo,
		FilenameTemplate: "{{pub_date}}_{{title}}_{{id}}",
	}

	legacyName := feed.LegacyEpisodeName(cfg, episode)
	legacyPath := filepath.Join(feedID, legacyName)
	_, err = storage.Create(ctx, legacyPath, strings.NewReader("video-bytes"))
	require.NoError(t, err)

	svc := New(map[string]*feed.Config{feedID: cfg}, tdb, storage, false)
	result, err := svc.Run(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	newPath := filepath.Join(feedID, feed.EpisodeName(cfg, episode))
	_, err = storage.Size(ctx, newPath)
	require.NoError(t, err)

	_, err = storage.Size(ctx, legacyPath)
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	assert.Equal(t, int64(len("video-bytes")), episode.Size)
	assert.Equal(t, model.EpisodeDownloaded, episode.Status)
	assert.Equal(t, 1, result.Migrated)
}

func TestRunDryRunDoesNotWrite(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	storage, err := fs.NewLocal(tmpDir, false, false)
	require.NoError(t, err)

	tdb := newTestDB()
	feedID := "B"
	episode := &model.Episode{
		ID:      "xyz999",
		Title:   "Another Title",
		PubDate: time.Date(2026, 2, 8, 9, 0, 0, 0, time.UTC),
		Status:  model.EpisodeDownloaded,
	}
	tdb.episodes[feedID] = map[string]*model.Episode{episode.ID: episode}

	cfg := &feed.Config{
		ID:               feedID,
		Format:           model.FormatVideo,
		FilenameTemplate: "{{pub_date}}_{{title}}_{{id}}",
	}

	legacyPath := filepath.Join(feedID, feed.LegacyEpisodeName(cfg, episode))
	_, err = storage.Create(ctx, legacyPath, strings.NewReader("video-bytes"))
	require.NoError(t, err)

	svc := New(map[string]*feed.Config{feedID: cfg}, tdb, storage, true)
	result, err := svc.Run(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)

	newPath := filepath.Join(feedID, feed.EpisodeName(cfg, episode))
	_, err = storage.Size(ctx, newPath)
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	_, err = storage.Size(ctx, legacyPath)
	require.NoError(t, err)

	assert.Equal(t, int64(0), episode.Size)
	assert.Equal(t, 1, result.Migrated)
}

func TestRunFailsOnUnexpectedSecondTargetStatError(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	baseStorage, err := fs.NewLocal(tmpDir, false, false)
	require.NoError(t, err)

	tdb := newTestDB()
	feedID := "C"
	episode := &model.Episode{
		ID:      "retry111",
		Title:   "Retry Test",
		PubDate: time.Date(2026, 2, 8, 9, 0, 0, 0, time.UTC),
		Status:  model.EpisodeDownloaded,
	}
	tdb.episodes[feedID] = map[string]*model.Episode{episode.ID: episode}

	cfg := &feed.Config{
		ID:               feedID,
		Format:           model.FormatVideo,
		FilenameTemplate: "{{pub_date}}_{{title}}_{{id}}",
	}

	legacyPath := filepath.Join(feedID, feed.LegacyEpisodeName(cfg, episode))
	_, err = baseStorage.Create(ctx, legacyPath, strings.NewReader("video-bytes"))
	require.NoError(t, err)

	newPath := filepath.Join(feedID, feed.EpisodeName(cfg, episode))
	storage := &flakySizeStorage{Storage: baseStorage, targetPath: newPath}

	svc := New(map[string]*feed.Config{feedID: cfg}, tdb, storage, false)
	_, err = svc.Run(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to stat target file")
	assert.Contains(t, err.Error(), "during migration")

	_, err = baseStorage.Size(ctx, legacyPath)
	require.NoError(t, err)
}
