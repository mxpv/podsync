package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_Get(t *testing.T) {
	s := createRedisClient(t)
	defer s.Close()

	err := s.Set("1", "value", 1*time.Minute)
	assert.NoError(t, err)

	val, err := s.Get("1")
	assert.NoError(t, err)
	assert.Equal(t, "value", val)
}

func TestRedisCache_GetInvalidKey(t *testing.T) {
	s := createRedisClient(t)
	defer s.Close()

	val, err := s.Get("1")
	assert.Equal(t, ErrNotFound, err)
	assert.Empty(t, val)
}

func TestNewRedisCache_TTL(t *testing.T) {
	s := createRedisClient(t)
	defer s.Close()

	err := s.Set("1", "value", 500*time.Millisecond)
	assert.NoError(t, err)

	val, err := s.Get("1")
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	time.Sleep(501 * time.Millisecond)

	_, err = s.Get("1")
	assert.Equal(t, ErrNotFound, err)
}

// docker run -it --rm -p 6379:6379 redis
func createRedisClient(t *testing.T) RedisCache {
	if testing.Short() {
		t.Skip("run redis tests manually")
	}

	client, err := NewRedisCache("redis://localhost")
	require.NoError(t, err)

	keys, err := client.client.Keys("*").Result()
	assert.NoError(t, err)

	if len(keys) > 0 {
		err = client.client.Del(keys...).Err()
		assert.NoError(t, err)
	}

	return client
}
