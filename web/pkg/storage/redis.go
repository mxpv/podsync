package storage

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/mxpv/podsync/web/pkg/api"
	"github.com/pkg/errors"
)

const expiration = 24 * time.Hour * 90

// Backward compatible Redis storage for feeds
type RedisStorage struct {
	client *redis.Client
}

func (r *RedisStorage) makeURL(m map[string]string) (string, error) {
	provider := m["provider"]
	linkType := m["type"]
	id := m["id"]

	if provider == "" || linkType == "" || id == "" {
		return "", errors.New("failed to query URL data from storage")
	}

	url := ""

	if strings.EqualFold(provider, "youtube") {
		if strings.EqualFold(linkType, "channel") {
			url = "https://youtube.com/channel/" + id
		} else if strings.EqualFold(linkType, "playlist") {
			url = "https://youtube.com/playlist?list=" + id
		} else if strings.EqualFold(linkType, "user") {
			url = "https://youtube.com/user/" + id
		}
	} else if strings.EqualFold(provider, "vimeo") {
		if strings.EqualFold(linkType, "channel") {
			url = "https://vimeo.com/channels/" + id
		} else if strings.EqualFold(linkType, "user") {
			url = "https://vimeo.com/" + id
		} else if strings.EqualFold(linkType, "group") {
			url = "https://vimeo.com/groups/" + id
		}
	}

	if url == "" {
		return "", fmt.Errorf("failed to query URL (provider: %s, type: %s, id: %s)", provider, linkType, id)
	}

	return url, nil
}

func (r *RedisStorage) parsePageSize(m map[string]string) (int, error) {
	str, ok := m["pagesize"]
	if !ok {
		return 50, nil
	}

	size, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, err
	}

	if size > 150 {
		return 0, errors.New("invalid page size")
	}

	return int(size), nil
}

func (r *RedisStorage) parseFormat(m map[string]string) (api.Format, api.Quality, error) {
	quality, ok := m["quality"]
	if !ok {
		return api.VideoFormat, api.HighQuality, nil
	}

	if quality == "VideoHigh" {
		return api.VideoFormat, api.HighQuality, nil
	} else if quality == "VideoLow" {
		return api.VideoFormat, api.LowQuality, nil
	} else if quality == "AudioHigh" {
		return api.AudioFormat, api.HighQuality, nil
	} else if quality == "AudioLow" {
		return api.AudioFormat, api.LowQuality, nil
	}

	return "", "", fmt.Errorf("unsupported formmat %s", quality)
}

func (r *RedisStorage) GetFeed(hashId string) (*api.Feed, error) {
	result, err := r.client.HGetAll(hashId).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query feed with id %s", hashId)
	}

	// Expire after 3 month if no use
	if err := r.client.Expire(hashId, expiration).Err(); err != nil {
		return nil, errors.Wrap(err, "failed query update feed")
	}

	feed := &api.Feed{
		PageSize: api.DefaultPageSize,
		Quality:  api.DefaultQuality,
		Format:   api.DefaultFormat,
	}

	m := make(map[string]string, len(result))
	for key, val := range result {
		m[strings.ToLower(key)] = val
	}

	j, ok := m["json"]
	if ok {
		if err := json.Unmarshal([]byte(j), feed); err != nil {
			return nil, err
		}

		return feed, nil
	}

	// Construct URL data
	url, err := r.makeURL(m)
	if err != nil {
		return nil, err
	}

	// Fetch user id
	patreonId, ok := m["patreonid"]
	if ok {
		feed.UserId = patreonId
	}

	// Unpack page size
	pageSize, err := r.parsePageSize(m)
	if err != nil {
		return nil, err
	}

	if patreonId == "" && pageSize > 50 {
		return nil, errors.New("wrong feed data")
	}

	// Parse feed's format and quality
	format, quality, err := r.parseFormat(m)
	if err != nil {
		return nil, err
	}

	feed.Format = format
	quality = quality

	return feed, nil
}

func (r *RedisStorage) CreateFeed(feed *api.Feed) error {
	raw, err := json.Marshal(feed)
	if err != nil {
		return err
	}

	fields := map[string]interface{}{"json": string(raw)}
	if err := r.client.HMSet(feed.HashId, fields).Err(); err != nil {
		return errors.Wrap(err, "failed to save feed")
	}

	return r.client.Expire(feed.HashId, expiration).Err()
}

func (r *RedisStorage) keys() ([]string, error) {
	return r.client.Keys("*").Result()
}

func NewRedisStorage(redisUrl string) (*RedisStorage, error) {
	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)
	return &RedisStorage{client}, nil
}
