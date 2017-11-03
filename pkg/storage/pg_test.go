package storage

import (
	"testing"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
	"github.com/stretchr/testify/require"
)

func TestPgStorage_CreateFeed(t *testing.T) {
	feed := &model.Feed{
		HashID:   "xyz",
		Provider: api.ProviderYoutube,
		LinkType: api.LinkTypeChannel,
		ItemID:   "123",
	}

	client := createClient(t)
	err := client.CreateFeed(feed)
	require.NoError(t, err)
	require.True(t, feed.FeedID > 0)
}

func TestPgStorage_CreateFeedWithDuplicate(t *testing.T) {
	feed := &model.Feed{
		HashID:   "123",
		Provider: api.ProviderYoutube,
		LinkType: api.LinkTypeChannel,
		ItemID:   "123",
	}

	client := createClient(t)
	err := client.CreateFeed(feed)
	require.NoError(t, err)

	// Ensure 1 record
	count, err := client.db.Model(&model.Feed{}).Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// Insert duplicated feed
	err = client.CreateFeed(feed)
	require.NoError(t, err)

	// Check no duplicates inserted
	count, err = client.db.Model(&model.Feed{}).Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestPgStorage_GetFeed(t *testing.T) {
	feed := &model.Feed{
		HashID:   "xyz",
		UserID:   "123",
		Provider: api.ProviderYoutube,
		LinkType: api.LinkTypeChannel,
		ItemID:   "123",
	}

	client := createClient(t)
	client.CreateFeed(feed)

	out, err := client.GetFeed("xyz")
	require.NoError(t, err)
	require.Equal(t, feed.FeedID, out.FeedID)
}

func TestPgStorage_UpdateLastAccess(t *testing.T) {
	feed := &model.Feed{
		HashID:   "xyz",
		UserID:   "123",
		Provider: api.ProviderYoutube,
		LinkType: api.LinkTypeChannel,
		ItemID:   "123",
	}

	client := createClient(t)
	err := client.CreateFeed(feed)
	require.NoError(t, err)

	lastAccess := feed.LastAccess
	require.True(t, lastAccess.Unix() > 0)

	last, err := client.GetFeed("xyz")
	require.NoError(t, err)

	require.NotEmpty(t, last.HashID)
	require.NotEmpty(t, last.UserID)
	require.NotEmpty(t, last.Provider)
	require.NotEmpty(t, last.LinkType)
	require.NotEmpty(t, last.ItemID)

	require.True(t, last.LastAccess.UnixNano() > lastAccess.UnixNano())
}

const TestDatabaseConnectionUrl = "postgres://postgres:@localhost/podsync?sslmode=disable"

func createClient(t *testing.T) *PgStorage {
	pg, err := NewPgStorage(&PgConfig{ConnectionUrl: TestDatabaseConnectionUrl})
	require.NoError(t, err)

	_, err = pg.db.Model(&model.Feed{}).Where("1=1").Delete()
	require.NoError(t, err)

	return pg
}
