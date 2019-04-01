package cache

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack"
)

var ErrNotFound = errors.New("not found")

// RedisCache implements caching layer for feeds using Redis
type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(redisURL string) (RedisCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return RedisCache{}, err
	}

	client := redis.NewClient(opts)
	if err := client.Ping().Err(); err != nil {
		return RedisCache{}, err
	}

	return RedisCache{client: client}, nil
}

func (c RedisCache) Set(key, value string, ttl time.Duration) error {
	return c.client.Set(key, value, ttl).Err()
}

func (c RedisCache) Get(key string) (string, error) {
	val, err := c.client.Get(key).Result()
	if err == redis.Nil {
		return "", ErrNotFound
	} else {
		return val, err
	}
}

func (c RedisCache) SaveItem(key string, item interface{}, exp time.Duration) error {
	data, err := msgpack.Marshal(item)
	if err != nil {
		return err
	}

	return c.client.Set(key, data, exp).Err()
}

func (c RedisCache) GetItem(key string, item interface{}) error {
	data, err := c.client.Get(key).Bytes()
	if err == redis.Nil {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	if err := msgpack.Unmarshal(data, item); err != nil {
		return err
	}

	return nil
}

func (c RedisCache) Invalidate(key... string) error {
	return c.client.Del(key...).Err()
}

func (c RedisCache) Close() error {
	return c.client.Close()
}
