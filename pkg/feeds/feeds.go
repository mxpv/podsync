package feeds

import (
	"fmt"
	"time"

	"github.com/go-pg/pg"
	itunes "github.com/mxpv/podcast"
	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
	"github.com/pkg/errors"
	"github.com/ventu-io/go-shortid"
)

const (
	maxPageSize = 150
)

type builder interface {
	Build(feed *model.Feed) (podcast *itunes.Podcast, err error)
}

type Service struct {
	sid      *shortid.Shortid
	db       *pg.DB
	builders map[api.Provider]builder
}

func (s Service) CreateFeed(req *api.CreateFeedRequest, identity *api.Identity) (string, error) {
	feed, err := parseURL(req.URL)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create feed for URL: %s", req.URL)
	}

	// Make sure builder exists for this provider
	_, ok := s.builders[feed.Provider]
	if !ok {
		return "", fmt.Errorf("failed to get builder for URL: %s", req.URL)
	}

	now := time.Now().UTC()

	// Set default fields
	feed.PageSize = api.DefaultPageSize
	feed.Format = api.FormatVideo
	feed.Quality = api.QualityHigh
	feed.FeatureLevel = api.DefaultFeatures
	feed.CreatedAt = now
	feed.LastAccess = now

	if identity.FeatureLevel > 0 {
		feed.UserID = identity.UserId
		feed.Quality = req.Quality
		feed.Format = req.Format
		feed.FeatureLevel = identity.FeatureLevel
		feed.PageSize = req.PageSize
		if feed.PageSize > maxPageSize {
			feed.PageSize = maxPageSize
		}
	}

	// Generate short id
	hashId, err := s.sid.Generate()
	if err != nil {
		return "", errors.Wrap(err, "failed to generate id for feed")
	}

	feed.HashID = hashId

	// Save to database
	_, err = s.db.Model(feed).Insert()
	if err != nil {
		return "", errors.Wrap(err, "failed to save feed to database")
	}

	return hashId, nil
}

func (s Service) QueryFeed(hashID string) (*model.Feed, error) {
	lastAccess := time.Now().UTC()

	feed := &model.Feed{}
	res, err := s.db.Model(feed).
		Set("last_access = ?", lastAccess).
		Where("hash_id = ?", hashID).
		Returning("*").
		Update()

	if err != nil {
		return nil, errors.Wrapf(err, "failed to query feed: %s", hashID)
	}

	if res.RowsAffected() != 1 {
		return nil, api.ErrNotFound
	}

	return feed, nil
}

func (s Service) BuildFeed(hashID string) (*itunes.Podcast, error) {
	feed, err := s.QueryFeed(hashID)
	if err != nil {
		return nil, err
	}

	builder, ok := s.builders[feed.Provider]
	if !ok {
		return nil, errors.Wrapf(err, "failed to get builder for feed: %s", hashID)
	}

	return builder.Build(feed)
}

func (s Service) GetMetadata(hashID string) (*api.Metadata, error) {
	feed := &model.Feed{}
	err := s.db.
		Model(feed).
		Where("hash_id = ?", hashID).
		Column("provider", "format", "quality").
		Select()

	if err != nil {
		return nil, err
	}

	return &api.Metadata{
		Provider: feed.Provider,
		Format:   feed.Format,
		Quality:  feed.Quality,
	}, nil
}

type feedOption func(*Service)

//noinspection GoExportedFuncWithUnexportedType
func WithPostgres(db *pg.DB) feedOption {
	return func(service *Service) {
		service.db = db
	}
}

//noinspection GoExportedFuncWithUnexportedType
func WithBuilder(provider api.Provider, builder builder) feedOption {
	return func(service *Service) {
		service.builders[provider] = builder
	}
}

func NewFeedService(opts ...feedOption) (*Service, error) {
	sid, err := shortid.New(1, shortid.DefaultABC, uint64(time.Now().UnixNano()))
	if err != nil {
		return nil, err
	}

	svc := &Service{
		sid:      sid,
		builders: make(map[api.Provider]builder),
	}

	for _, fn := range opts {
		fn(svc)
	}

	return svc, nil
}
