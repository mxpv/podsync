package feeds

import (
	"fmt"
	"time"

	itunes "github.com/mxpv/podcast"
	"github.com/mxpv/podsync/pkg/api"
	"github.com/pkg/errors"
)

const (
	maxPageSize = 150
)

type idService interface {
	Generate(feed *api.Feed) (string, error)
}

type storageService interface {
	CreateFeed(feed *api.Feed) error
	GetFeed(hashId string) (*api.Feed, error)
}

type builder interface {
	Build(feed *api.Feed) (podcast *itunes.Podcast, err error)
}

type service struct {
	id       idService
	storage  storageService
	builders map[api.Provider]builder
}

func (s *service) CreateFeed(req *api.CreateFeedRequest, identity *api.Identity) (string, error) {
	feed, err := parseURL(req.URL)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create feed for URL: %s", req.URL)
	}

	// Make sure builder exists for this provider
	_, ok := s.builders[feed.Provider]
	if !ok {
		return "", fmt.Errorf("failed to get builder for URL: %s", req.URL)
	}

	// Set default fields
	feed.PageSize = api.DefaultPageSize
	feed.Format = api.FormatVideo
	feed.Quality = api.QualityHigh
	feed.FeatureLevel = api.DefaultFeatures
	feed.LastAccess = time.Now().UTC()

	if identity.FeatureLevel > 0 {
		feed.UserId = identity.UserId
		feed.Quality = req.Quality
		feed.Format = req.Format
		feed.FeatureLevel = identity.FeatureLevel
		feed.PageSize = req.PageSize
		if feed.PageSize > maxPageSize {
			feed.PageSize = maxPageSize
		}
	}

	// Generate short id
	hashId, err := s.id.Generate(feed)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate id for feed")
	}

	feed.HashId = hashId

	// Save to database
	if err := s.storage.CreateFeed(feed); err != nil {
		return "", errors.Wrap(err, "failed to save feed to database")
	}

	return hashId, nil
}

func (s *service) GetFeed(hashId string) (*itunes.Podcast, error) {
	feed, err := s.GetMetadata(hashId)
	if err != nil {
		return nil, err
	}

	builder, ok := s.builders[feed.Provider]
	if !ok {
		return nil, errors.Wrapf(err, "failed to get builder for feed: %s", hashId)
	}

	return builder.Build(feed)
}

func (s *service) GetMetadata(hashId string) (*api.Feed, error) {
	return s.storage.GetFeed(hashId)
}

type feedOption func(*service)

func WithStorage(storage storageService) feedOption {
	return func(service *service) {
		service.storage = storage
	}
}

func WithIdGen(id idService) feedOption {
	return func(service *service) {
		service.id = id
	}
}

func WithBuilder(provider api.Provider, builder builder) feedOption {
	return func(service *service) {
		service.builders[provider] = builder
	}
}

func NewFeedService(opts ...feedOption) *service {
	svc := &service{}
	svc.builders = make(map[api.Provider]builder)

	for _, fn := range opts {
		fn(svc)
	}

	return svc
}
