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

func TestRedisCache_SaveItem(t *testing.T) {
	type test struct {
		Feed      []byte    `msgpack:"feed"`
		UpdatedAt time.Time `msgpack:"updated_at"`
	}

	s := createRedisClient(t)
	defer s.Close()

	item := &test{
		Feed:      []byte("123"),
		UpdatedAt: time.Now().UTC(),
	}

	err := s.SaveItem("test", item, time.Minute)
	assert.NoError(t, err)

	var out test
	err = s.GetItem("test", &out)
	assert.NoError(t, err)

	assert.EqualValues(t, item.Feed, &out.Feed)
	assert.EqualValues(t, item.UpdatedAt.Unix(), out.UpdatedAt.Unix())
}

func TestRedisCache_Map(t *testing.T) {
	s := createRedisClient(t)
	defer s.Close()

	data := map[string]interface{}{
		"1": "123",
		"2": "test",
	}

	err := s.SetMap("2", data, time.Minute)
	assert.NoError(t, err)

	out, err := s.GetMap("2", "1", "2")
	assert.NoError(t, err)
	assert.EqualValues(t, data, out)
}

func TestRedisCache_GetMapInvalidKey(t *testing.T) {
	s := createRedisClient(t)
	defer s.Close()

	_, err := s.GetMap("unknown_key", "1", "2")
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
