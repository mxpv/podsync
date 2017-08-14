package storage

import (
	"strconv"
	"testing"
	"time"

	"github.com/mxpv/podsync/web/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestRedisStorage_GetFeed(t *testing.T) {
	t.Skip("run redis tests manually")

	client := createRedisClient(t)

	keys, err := client.keys()
	require.NoError(t, err)

	require.True(t, len(keys) > 0)

	for idx, key := range keys {
		if key == "keygen" {
			continue
		}

		feed, err := client.GetFeed(key)
		require.NoError(t, err, "feed %s (id = %d) failed", key, idx)
		require.NotNil(t, feed)
	}
}

func TestRedisStorage_CreateFeed(t *testing.T) {
	t.Skip("run redis tests manually")

	client := createRedisClient(t)

	hashId := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

	err := client.CreateFeed(&api.Feed{
		Id:       123,
		HashId:   hashId,
		UserId:   "321",
		Provider: api.Youtube,
		LinkType: api.Channel,
		ItemId:   "123",
		PageSize: 45,
		Quality:  api.LowQuality,
		Format:   api.AudioFormat,
	})

	require.NoError(t, err)

	feed, err := client.GetFeed(hashId)
	require.NoError(t, err)

	require.Equal(t, int64(123), feed.Id)
	require.Equal(t, hashId, feed.HashId)
	require.Equal(t, "321", feed.UserId)
	require.Equal(t, api.Youtube, feed.Provider)
	require.Equal(t, api.Channel, feed.LinkType)
	require.Equal(t, "123", feed.ItemId)
	require.Equal(t, 45, feed.PageSize)
	require.Equal(t, api.LowQuality, feed.Quality)
	require.Equal(t, api.AudioFormat, feed.Format)
}

func createRedisClient(t *testing.T) *RedisStorage {
	client, err := NewRedisStorage("redis://localhost")
	require.NoError(t, err)

	return client
}
