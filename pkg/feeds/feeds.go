package feeds

import (
	"fmt"
	"strconv"
	"time"

	itunes "github.com/mxpv/podcast"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
)

type Builder interface {
	Build(feed *model.Feed) (podcast *itunes.Podcast, err error)
	GetVideoCount(feed *model.Feed) (uint64, error)
}

type storage interface {
	SaveFeed(feed *model.Feed) error
	GetFeed(hashID string) (*model.Feed, error)
	GetMetadata(hashID string) (*model.Feed, error)
	Downgrade(userID string, featureLevel int) ([]string, error)
}

type cacheService interface {
	Set(key, value string, ttl time.Duration) error
	Get(key string) (string, error)

	SaveItem(key string, item interface{}, exp time.Duration) error
	GetItem(key string, item interface{}) error

	SetMap(key string, fields map[string]interface{}, exp time.Duration) error
	GetMap(key string, fields ...string) (map[string]string, error)

	Invalidate(key ...string) error
}

type Service struct {
	generator IDGen
	db        storage
	builders  map[api.Provider]Builder
	cache     cacheService
}

func (s Service) makeFeed(req *api.CreateFeedRequest, identity *api.Identity) (*model.Feed, error) {
	feed, err := parseURL(req.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create feed for URL: %s", req.URL)
	}

	now := time.Now().UTC()

	feed.UserID = identity.UserID
	feed.FeatureLevel = identity.FeatureLevel
	feed.Quality = req.Quality
	feed.Format = req.Format
	feed.PageSize = req.PageSize
	feed.CreatedAt = now
	feed.LastAccess = now

	switch {
	case identity.FeatureLevel >= api.ExtendedPagination:
		if feed.PageSize > 600 {
			feed.PageSize = 600
		}
	case identity.FeatureLevel == api.ExtendedFeatures:
		if feed.PageSize > 150 {
			feed.PageSize = 150
		}
	default:
		feed.Quality = api.QualityHigh
		feed.Format = api.FormatVideo
		feed.PageSize = 50
	}

	// Generate short id
	hashID, err := s.generator.Generate()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate id for feed")
	}

	feed.HashID = hashID

	return feed, nil
}

func (s Service) CreateFeed(req *api.CreateFeedRequest, identity *api.Identity) (string, error) {
	feed, err := s.makeFeed(req, identity)
	if err != nil {
		return "", err
	}

	// Make sure builder exists for this provider
	_, ok := s.builders[feed.Provider]
	if !ok {
		return "", fmt.Errorf("failed to get builder for URL: %s", req.URL)
	}

	if err := s.db.SaveFeed(feed); err != nil {
		return "", err
	}

	return feed.HashID, nil
}

func (s Service) QueryFeed(hashID string) (*model.Feed, error) {
	return s.db.GetFeed(hashID)
}

func (s Service) getVideoCount(feed *model.Feed, builder Builder) (uint64, bool) {
	videoCount, err := builder.GetVideoCount(feed)
	if err != nil {
		return 0, false
	}

	return videoCount, true
}

const (
	feedKey       = "feed"
	updatedAtKey  = "updatedAt"
	videoCountKey = "videoCount"
)

func (s Service) getCachedFeed(hashID string) ([]byte, time.Time, uint64, error) {
	cached, err := s.cache.GetMap(hashID, feedKey, updatedAtKey, videoCountKey)
	if err != nil {
		return nil, time.Time{}, 0, err
	}

	// Get feed body

	body := []byte(cached[feedKey])

	// Parse updated at

	unixTime, err := strconv.ParseInt(cached[updatedAtKey], 10, 64)
	if err != nil {
		return nil, time.Time{}, 0, err
	}

	updatedAt := time.Unix(unixTime, 0)

	// Parse video count

	var videoCount uint64

	if str, ok := cached[videoCountKey]; ok {
		count, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return nil, time.Time{}, 0, err
		}

		videoCount = count
	}

	// OK
	return body, updatedAt, videoCount, nil
}

func (s Service) BuildFeed(hashID string) ([]byte, error) {
	const (
		feedRecordTTL   = 15 * 24 * time.Hour
		cacheRecheckTTL = 10 * time.Minute
	)

	var (
		now         = time.Now().UTC()
		verifyCache bool
	)

	// Check cached version first
	body, updatedAt, currentCount, err := s.getCachedFeed(hashID)
	if err == nil {
		// We've succeded to retrieve data from Redis, check if it's up to date

		// 1. If cached less than 15 minutes ago, just return data
		if now.Sub(updatedAt) < cacheRecheckTTL {
			return body, nil
		}

		// 2. Verify cache integrity by querying the number of episodes from YouTube
		verifyCache = true
	}

	// Query feed metadata from DynamoDB

	feed, err := s.QueryFeed(hashID)
	if err != nil {
		return nil, err
	}

	builder, ok := s.builders[feed.Provider]
	if !ok {
		return nil, errors.Wrapf(err, "failed to get builder for feed: %s", hashID)
	}

	// 2. Check if cached version is still valid

	// Query YouTube and check the number of videos.
	// Most likely it'll remain the same, so we can return previously cached feed.
	videoCount, videoOk := s.getVideoCount(feed, builder)

	if verifyCache && videoOk {

		if currentCount == videoCount {
			// Cache is up to date, renew and save
			err = s.cache.SetMap(hashID, map[string]interface{}{updatedAtKey: now.Unix()}, feedRecordTTL)
			if err != nil {
				return nil, errors.Wrap(err, "failed to cache item")
			}

			return body, nil
		}
	}

	// Rebuild feed using YouTube API

	log.Infof("building new feed %q", hashID)

	podcast, err := builder.Build(feed)
	if err != nil {
		log.WithError(err).WithField("feed_id", hashID).Error("failed to build cache")

		// If there is cached version - return it
		if verifyCache {
			return body, nil
		}

		return nil, err
	}

	newBody := podcast.String()

	// Save to cache

	data := map[string]interface{}{
		feedKey: newBody,
		updatedAtKey: now.Unix(),
		videoCountKey: strconv.FormatUint(videoCount, 10),
	}

	err = s.cache.SetMap(hashID, data, feedRecordTTL)

	if err != nil {
		log.WithError(err).Warnf("failed to save new feed %q to cache", hashID)
	}

	return []byte(newBody), nil
}

func (s Service) GetMetadata(hashID string) (*api.Metadata, error) {
	feed, err := s.db.GetMetadata(hashID)
	if err != nil {
		return nil, err
	}

	return &api.Metadata{
		Provider: feed.Provider,
		Format:   feed.Format,
		Quality:  feed.Quality,
	}, nil
}

func (s Service) Downgrade(patronID string, featureLevel int) error {
	logger := log.WithFields(log.Fields{
		"user_id": patronID,
		"level":   featureLevel,
	})

	logger.Info("downgrading patron")

	ids, err := s.db.Downgrade(patronID, featureLevel)
	if err != nil {
		logger.WithError(err).Error("database error while downgrading patron")
		return err
	}

	if s.cache.Invalidate(ids...) != nil {
		logger.WithError(err).Error("failed to invalidate cached feeds")
		return err
	}

	logger.Info("successfully updated user")
	return nil
}

func NewFeedService(db storage, cache cacheService, builders map[api.Provider]Builder) (*Service, error) {
	idGen, err := NewIDGen()
	if err != nil {
		return nil, err
	}

	svc := &Service{
		generator: idGen,
		db:        db,
		builders:  builders,
		cache:     cache,
	}

	return svc, nil
}
