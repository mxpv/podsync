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
	now := time.Now().UTC()

	key := r.makeKey(now, hashID)
	top := r.makeTop(now, metric)

	var cmd *redis.IntCmd
	_, err := r.client.TxPipelined(func(p redis.Pipeliner) error {
		cmd = p.HIncrBy(key, metric, 1)
		p.ZIncrBy(top, 1, hashID)
		return nil
	})

	if err != nil {
		return 0, err
	}

	return cmd.Result()
}

func (r RedisStats) Get(metric, hashID string) (int64, error) {
	now := time.Now().UTC()
	key := r.makeKey(now, hashID)
	return r.client.HGet(key, metric).Int64()
}

func (r RedisStats) Top(metric string) (map[string]int64, error) {
	now := time.Now().UTC()
	top := r.makeTop(now, metric)

	zrange, err := r.client.ZRevRangeWithScores(top, 0, 10).Result()
	if err != nil {
		return nil, err
	}

	ret := make(map[string]int64)
	for _, x := range zrange {
		key := x.Member.(string)
		val := int64(x.Score)

		ret[key] = val
	}

	return ret, nil
}

func (r RedisStats) makeKey(now time.Time, hashID string) string {
	return fmt.Sprintf("stats/%d/%d/%s", now.Year(), now.Month(), hashID)
}

func (r RedisStats) makeTop(now time.Time, metric string) string {
	return fmt.Sprintf("stats/top/%d/%d/%s", now.Year(), now.Month(), metric)
}

func (r RedisStats) Close() error {
	return r.client.Close()
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
