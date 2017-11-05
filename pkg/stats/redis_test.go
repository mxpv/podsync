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

func createRedisClient(t *testing.T) *RedisStats {
	client, err := NewRedisStats("redis://localhost")
	require.NoError(t, err)

	keys, err := client.client.Keys("*").Result()
	require.NoError(t, err)

	err = client.client.Del(keys...).Err()
	require.NoError(t, err)

	return client
}
