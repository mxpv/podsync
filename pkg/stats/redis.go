package stats

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type RedisStats struct {
	client *redis.Client
}

func (r RedisStats) Inc(metric, hashID string) (int64, error) {
	key := r.makeKey(hashID)
	return r.client.HIncrBy(key, metric, 1).Result()
}

func (r RedisStats) Get(metric, hashID string) (int64, error) {
	key := r.makeKey(hashID)
	return r.client.HGet(key, metric).Int64()
}

func (r RedisStats) Close() error {
	return r.client.Close()
}

func (r RedisStats) makeKey(hashID string) string {
	now := time.Now().UTC()
	return fmt.Sprintf("stats/%d/%d/%s", now.Year(), now.Month(), hashID)
}

func NewRedisStats(redisUrl string) (*RedisStats, error) {
	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)
	if err := client.Ping().Err(); err != nil {
		return nil, err
	}

	return &RedisStats{client}, nil
}
