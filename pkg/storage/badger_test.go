package storage

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/link"
	"github.com/mxpv/podsync/pkg/model"
)

var testCtx = context.TODO()

func TestNewBadger(t *testing.T) {
	dir, err := ioutil.TempDir("", "podsync-badger-")
	require.NoError(t, err)

	db, err := NewBadger(&config.Database{Dir: dir})
	require.NoError(t, err)

	err = db.Close()
	assert.NoError(t, err)

	err = os.RemoveAll(dir)
	assert.NoError(t, err)
}

func TestBadger_Version(t *testing.T) {
	dir, err := ioutil.TempDir("", "podsync-badger-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := NewBadger(&config.Database{Dir: dir})
	require.NoError(t, err)
	defer db.Close()

	ver, err := db.Version()
	assert.NoError(t, err)
	assert.Equal(t, CurrentVersion, ver)
}

func TestBadger_AddFeed(t *testing.T) {
	dir, err := ioutil.TempDir("", "podsync-badger-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := NewBadger(&config.Database{Dir: dir})
	require.NoError(t, err)
	defer db.Close()

	err = db.AddFeed(testCtx, getFeed())
	assert.NoError(t, err)
}

func TestBadger_GetFeed(t *testing.T) {
	dir, err := ioutil.TempDir("", "podsync-badger-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := NewBadger(&config.Database{Dir: dir})
	require.NoError(t, err)
	defer db.Close()

	feed := getFeed()
	feed.Episodes = nil

	err = db.AddFeed(testCtx, feed)
	require.NoError(t, err)

	actual, err := db.GetFeed(testCtx, feed.ID)
	assert.NoError(t, err)
	assert.Equal(t, feed, actual)
}

func TestBadger_WalkFeeds(t *testing.T) {
	dir, err := ioutil.TempDir("", "podsync-badger-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := NewBadger(&config.Database{Dir: dir})
	require.NoError(t, err)
	defer db.Close()

	feed := getFeed()
	feed.Episodes = nil // These are not serialized to database

	err = db.AddFeed(testCtx, feed)
	assert.NoError(t, err)

	called := 0
	err = db.WalkFeeds(testCtx, func(actual *model.Feed) error {
		assert.EqualValues(t, feed, actual)
		called++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, called, 1)
}

func TestBadger_DeleteFeed(t *testing.T) {
	dir, err := ioutil.TempDir("", "podsync-badger-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := NewBadger(&config.Database{Dir: dir})
	require.NoError(t, err)
	defer db.Close()

	feed := getFeed()
	err = db.AddFeed(testCtx, feed)
	require.NoError(t, err)

	err = db.DeleteFeed(testCtx, feed.ID)
	assert.NoError(t, err)

	called := 0

	err = db.WalkFeeds(testCtx, func(feed *model.Feed) error {
		called++
		return nil
	})
	assert.NoError(t, err)

	err = db.WalkFiles(testCtx, feed.ID, func(file *model.File) error {
		called++
		return nil
	})
	assert.NoError(t, err)

	assert.Equal(t, 0, called)
}

func TestBadger_WalkFiles(t *testing.T) {
	dir, err := ioutil.TempDir("", "podsync-badger-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := NewBadger(&config.Database{Dir: dir})
	require.NoError(t, err)
	defer db.Close()

	feed := getFeed()
	err = db.AddFeed(testCtx, feed)
	assert.NoError(t, err)

	called := 0

	err = db.WalkFiles(testCtx, feed.ID, func(file *model.File) error {
		assert.Equal(t, feed.ID, file.FeedID)
		assert.Equal(t, feed.Episodes[called].ID, file.EpisodeID)
		assert.Equal(t, feed.Episodes[called].Size, file.Size)
		assert.Equal(t, model.EpisodeNew, file.Status)

		called++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, called)
}

func TestBadger_UpdateFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "podsync-badger-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := NewBadger(&config.Database{Dir: dir})
	require.NoError(t, err)
	defer db.Close()

	feed := getFeed()
	err = db.AddFeed(testCtx, feed)
	assert.NoError(t, err)

	err = db.UpdateFile(feed.ID, feed.Episodes[0].ID, func(file *model.File) error {
		file.Size = 333
		file.Status = model.EpisodeDownloaded
		return nil
	})
	assert.NoError(t, err)

	first := true

	err = db.WalkFiles(testCtx, feed.ID, func(file *model.File) error {
		if first {
			assert.Equal(t, feed.ID, file.FeedID)
			assert.Equal(t, feed.Episodes[0].ID, file.EpisodeID)
			assert.EqualValues(t, 333, file.Size)
			assert.Equal(t, model.EpisodeDownloaded, file.Status)
			first = false
		}

		return nil
	})

	assert.NoError(t, err)
}

func getFeed() *model.Feed {
	return &model.Feed{
		ID:             "1",
		ItemID:         "2",
		LinkType:       link.TypeChannel,
		Provider:       link.ProviderVimeo,
		CreatedAt:      time.Now().UTC(),
		LastAccess:     time.Now().UTC(),
		ExpirationTime: time.Now().UTC().Add(1 * time.Hour),
		Format:         "video",
		Quality:        "high",
		PageSize:       50,
		Language:       "en",
		Title:          "Test",
		Description:    "Test",
		PubDate:        time.Now().UTC(),
		Author:         "",
		ItemURL:        "https://vimeo.com",
		Episodes: []*model.Episode{
			{
				ID:          "1",
				Title:       "Episode title 1",
				Description: "Episode description 1",
				Duration:    100,
				VideoURL:    "https://vimeo.com/123",
				PubDate:     time.Now().UTC(),
				Size:        1234,
				Order:       "1",
			},
			{
				ID:          "2",
				Title:       "Episode title 2",
				Description: "Episode description 2",
				Duration:    299,
				VideoURL:    "https://vimeo.com/321",
				PubDate:     time.Now().UTC(),
				Size:        4321,
				Order:       "2",
			},
		},
		UpdatedAt: time.Now().UTC(),
	}
}
