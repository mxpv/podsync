package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	feed := &Feed{
		HashId: "xyz",
		URL:    "http://youtube.com",
	}

	client := createClient(t)
	err := client.CreateFeed(feed)
	require.NoError(t, err)
	require.True(t, feed.Id > 0)
}

func TestCreateDuplicate(t *testing.T) {
	feed := &Feed{
		HashId: "123",
		URL:    "http://youtube.com",
	}

	client := createClient(t)
	err := client.CreateFeed(feed)
	require.NoError(t, err)

	// Ensure 1 record
	count, err := client.db.Model(&Feed{}).Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// Insert duplicated feed
	err = client.CreateFeed(feed)
	require.NoError(t, err)

	// Check no duplicates inserted
	count, err = client.db.Model(&Feed{}).Count()
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestGetFeed(t *testing.T) {
	feed := &Feed{
		HashId: "xyz",
		UserId: "123",
		URL:    "http://youtube.com",
	}

	client := createClient(t)
	client.CreateFeed(feed)

	out, err := client.GetFeed("xyz")
	require.NoError(t, err)
	require.Equal(t, feed.Id, out.Id)
}

func TestUpdateLastAccess(t *testing.T) {
	feed := &Feed{
		HashId: "xyz",
		UserId: "123",
		URL:    "http://youtube.com",
	}

	client := createClient(t)
	err := client.CreateFeed(feed)
	require.NoError(t, err)

	lastAccess := feed.LastAccess
	require.True(t, lastAccess.Unix() > 0)

	last, err := client.GetFeed("xyz")
	require.NoError(t, err)

	require.NotEmpty(t, last.HashId)
	require.NotEmpty(t, last.UserId)
	require.NotEmpty(t, last.URL)

	require.True(t, last.LastAccess.Unix() > lastAccess.Unix())
}

const TestDatabaseConnectionUrl = "postgres://postgres:@localhost/podsync?sslmode=disable"

func createClient(t *testing.T) *PgStorage {
	pg, err := NewPgStorage(&PgConfig{ConnectionUrl: TestDatabaseConnectionUrl})
	require.NoError(t, err)

	_, err = pg.db.Model(&Feed{}).Where("1=1").Delete()
	require.NoError(t, err)

	return pg
}
