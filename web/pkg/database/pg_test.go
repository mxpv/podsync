package database

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

func TestUniqueHashId(t *testing.T) {
	client := createClient(t)

	err := client.CreateFeed(&Feed{HashId: "xyz", URL: "url"})
	require.NoError(t, err)

	err = client.CreateFeed(&Feed{HashId: "xyz", URL: "url"})
	require.Error(t, err)
}

const TestDatabaseConnectionUrl = "postgres://postgres:@localhost/podsync?sslmode=disable"

func createClient(t *testing.T) *PgStorage {
	pg, err := NewPgStorage(&PgConfig{ConnectionUrl: TestDatabaseConnectionUrl})
	require.NoError(t, err)

	_, err = pg.db.Model(&Feed{}).Where("1=1").Delete()
	require.NoError(t, err)

	return pg
}
