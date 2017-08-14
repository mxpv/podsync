package feeds

import (
	"context"
	"fmt"
	"time"

	itunes "github.com/mxpv/podcast"
	"github.com/mxpv/podsync/web/pkg/api"
	"github.com/pkg/errors"
)

type service struct {
	id       id
	storage  storage
	builders map[api.Provider]builder
}

func (s *service) CreateFeed(ctx context.Context, req *api.CreateFeedRequest) (string, error) {
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
	feed.Format = api.VideoFormat
	feed.Quality = api.HighQuality
	feed.FeatureLevel = api.DefaultFeatures
	feed.LastAccess = time.Now().UTC()

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
	feed, err := s.storage.GetFeed(hashId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query feed: %s", hashId)
	}

	return feed, nil
}
