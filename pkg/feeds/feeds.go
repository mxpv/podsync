package feeds

import (
	"fmt"
	"log"
	"time"

	itunes "github.com/mxpv/podcast"
	"github.com/pkg/errors"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
)

const feedCacheTTL = 15 * time.Minute

type Builder interface {
	Build(feed *model.Feed) (podcast *itunes.Podcast, err error)
}

type storage interface {
	SaveFeed(feed *model.Feed) error
	GetFeed(hashID string) (*model.Feed, error)
	GetMetadata(hashID string) (*model.Feed, error)
	Downgrade(userID string, featureLevel int) error
}

type cacheService interface {
	Set(key, value string, ttl time.Duration) error
	Get(key string) (string, error)
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

func (s Service) BuildFeed(hashID string) ([]byte, error) {
	// Check cached version first
	cached, err := s.cache.Get(hashID)
	if err == nil {
		return []byte(cached), nil
	}

	// Query feed metadata

	feed, err := s.QueryFeed(hashID)
	if err != nil {
		return nil, err
	}

	builder, ok := s.builders[feed.Provider]
	if !ok {
		return nil, errors.Wrapf(err, "failed to get builder for feed: %s", hashID)
	}

	// Rebuild feed using YouTube API

	podcast, err := builder.Build(feed)
	if err != nil {
		return nil, err
	}

	data := podcast.String()

	// Save to cache

	if err := s.cache.Set(hashID, data, feedCacheTTL); err != nil {
		log.Printf("failed to cache feed %q: %+v", hashID, err)
	}

	return []byte(data), nil
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
	log.Printf("Downgrading patron '%s' to feature level %d", patronID, featureLevel)

	if err := s.db.Downgrade(patronID, featureLevel); err != nil {
		log.Printf("! downgrade failed")
		return err
	}

	log.Printf("updated user '%s' to feature level %d", patronID, featureLevel)
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
