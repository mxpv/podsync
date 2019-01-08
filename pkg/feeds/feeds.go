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

const (
	MetricQueries   = "queries"
	MetricDownloads = "downloads"
)

type stats interface {
	Inc(metric, hashID string) (int64, error)
	Get(metric, hashID string) (int64, error)
}

type builder interface {
	Build(feed *model.Feed) (podcast *itunes.Podcast, err error)
}

type storage interface {
	SaveFeed(feed *model.Feed) error
	GetFeed(hashID string) (*model.Feed, error)
	GetMetadata(hashID string) (*model.Feed, error)
	Downgrade(userID string, featureLevel int) error
}

type Service struct {
	generator IDGen
	stats     stats
	db        storage
	builders  map[api.Provider]builder
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

func (s Service) BuildFeed(hashID string) (*itunes.Podcast, error) {
	feed, err := s.QueryFeed(hashID)
	if err != nil {
		return nil, err
	}

	count, err := s.stats.Inc(MetricQueries, feed.HashID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update metrics for feed: %s", hashID)
	}

	if feed.PageSize > 150 && count > api.ExtendedPaginationQueryLimit {
		return nil, api.ErrQuotaExceeded
	}

	builder, ok := s.builders[feed.Provider]
	if !ok {
		return nil, errors.Wrapf(err, "failed to get builder for feed: %s", hashID)
	}

	podcast, err := builder.Build(feed)
	if err != nil {
		return nil, err
	}

	return podcast, nil
}

func (s Service) GetMetadata(hashID string) (*api.Metadata, error) {
	feed, err := s.db.GetMetadata(hashID)
	if err != nil {
		return nil, err
	}

	downloads, err := s.stats.Inc(MetricDownloads, hashID)
	if err != nil {
		return nil, err
	}

	return &api.Metadata{
		Provider:  feed.Provider,
		Format:    feed.Format,
		Quality:   feed.Quality,
		Downloads: downloads,
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

type FeedOption func(*Service)

//noinspection GoExportedFuncWithUnexportedType
func WithStorage(db storage) FeedOption {
	return func(service *Service) {
		service.db = db
	}
}

//noinspection GoExportedFuncWithUnexportedType
func WithBuilder(provider api.Provider, builder builder) FeedOption {
	return func(service *Service) {
		service.builders[provider] = builder
	}
}

//noinspection GoExportedFuncWithUnexportedType
func WithStats(m stats) FeedOption {
	return func(service *Service) {
		service.stats = m
	}
}

func NewFeedService(opts ...FeedOption) (*Service, error) {
	idGen, err := NewIDGen()
	if err != nil {
		return nil, err
	}

	svc := &Service{
		generator: idGen,
		builders:  make(map[api.Provider]builder),
	}

	for _, fn := range opts {
		fn(svc)
	}

	return svc, nil
}
