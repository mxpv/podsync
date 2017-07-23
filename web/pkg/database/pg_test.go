package database

import (
	"github.com/stretchr/testify/require"
	"testing"
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

	out, err := client.GetFeed(WithUserId("123"))
	require.NoError(t, err)
	require.Equal(t, 1, len(out))
	require.Equal(t, feed.Id, out[0].Id)

	out, err = client.GetFeed(WithHashId("xyz"))
	require.NoError(t, err)
	require.Equal(t, 1, len(out))
	require.Equal(t, feed.Id, out[0].Id)
}

const TestDatabaseConnectionUrl = "postgres://postgres:@localhost/podsync?sslmode=disable"

func createClient(t *testing.T) *PgStorage {
	pg, err := NewPgStorage(&PgConfig{ConnectionUrl: TestDatabaseConnectionUrl})
	require.NoError(t, err)

	_, err = pg.db.Model(&Feed{}).Where("1=1").Delete()
	require.NoError(t, err)

	return pg
}
