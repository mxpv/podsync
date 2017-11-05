package stats

import (
	"github.com/stretchr/testify/require"
	"testing"
)

const metric = "downloads"

func TestRedisStats_IncAndGet(t *testing.T) {
	t.Skip("run redis tests manually")

	s := createRedisClient(t)

	const hashID = "321"

	v, err := s.Inc(metric, hashID)
	require.NoError(t, err)
	require.Equal(t, int64(1), v)

	v, err = s.Inc(metric, hashID)
	require.NoError(t, err)
	require.Equal(t, int64(2), v)

	v, err = s.Get(metric, hashID)
	require.NoError(t, err)
	require.Equal(t, int64(2), v)
}

func TestRedisStats_Top(t *testing.T) {
	t.Skip("run redis tests manually")

	s := createRedisClient(t)

	// 3
	s.Inc(metric, "123")
	s.Inc(metric, "123")
	s.Inc(metric, "123")

	// 2
	s.Inc(metric, "321")
	s.Inc(metric, "321")

	// 1
	s.Inc(metric, "213")

	top, err := s.Top(metric)
	require.NoError(t, err)
	require.Len(t, top, 3)

	// 3
	h3, ok := top["123"]
	require.True(t, ok)
	require.Equal(t, int64(3), h3)

	// 2
	h2, ok := top["321"]
	require.True(t, ok)
	require.Equal(t, int64(2), h2)

	// 1
	h1, ok := top["213"]
	require.True(t, ok)
	require.Equal(t, int64(1), h1)
}

func createRedisClient(t *testing.T) *RedisStats {
	client, err := NewRedisStats("redis://localhost")
	require.NoError(t, err)

	keys, err := client.client.Keys("*").Result()
	require.NoError(t, err)

	err = client.client.Del(keys...).Err()
	require.NoError(t, err)

	return client
}
