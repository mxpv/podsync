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
	Build(feed *model.Feed) error
}

type storage interface {
	SaveFeed(feed *model.Feed) error
	GetFeed(hashID string) (*model.Feed, error)
	UpdateFeed(feed *model.Feed) error
	GetMetadata(hashID string) (*model.Feed, error)
	Downgrade(userID string, featureLevel int) ([]string, error)
}

type cacheService interface {
	Set(key, value string, ttl time.Duration) error
	Get(key string) (string, error)

	SaveItem(key string, item interface{}, exp time.Duration) error
	GetItem(key string, item interface{}) error

	Invalidate(key ...string) error
}

type Service struct {
	generator IDGen
	storage   storage
	builders  map[api.Provider]Builder
	cache     cacheService
}

func NewFeedService(db storage, cache cacheService, builders map[api.Provider]Builder) (*Service, error) {
	idGen, err := NewIDGen()
	if err != nil {
		return nil, err
	}

	svc := &Service{
		generator: idGen,
		storage:   db,
		builders:  builders,
		cache:     cache,
	}

	return svc, nil
}

func (s *Service) CreateFeed(req *api.CreateFeedRequest, identity *api.Identity) (string, error) {
	feed, err := s.makeFeed(req, identity)
	if err != nil {
		return "", err
	}

	// Make sure builder exists for this provider
	_, ok := s.builders[feed.Provider]
	if !ok {
		return "", fmt.Errorf("failed to get builder for URL: %s", req.URL)
	}

	if err := s.storage.SaveFeed(feed); err != nil {
		return "", err
	}

	return feed.HashID, nil
}

func (s *Service) makeFeed(req *api.CreateFeedRequest, identity *api.Identity) (*model.Feed, error) {
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

func (s Service) QueryFeed(hashID string) (*model.Feed, error) {
	return s.storage.GetFeed(hashID)
}

func makeEnclosure(feed *model.Feed, id string, lengthInBytes int64) (string, itunes.EnclosureType, int64) {
	ext := "mp4"
	contentType := itunes.MP4
	if feed.Format == api.FormatAudio {
		ext = "m4a"
		contentType = itunes.M4A
	}

	url := fmt.Sprintf("http://podsync.net/download/%s/%s.%s", feed.HashID, id, ext)
	return url, contentType, lengthInBytes
}

func (s *Service) updateFromRedis(hashID string, feed *model.Feed) {
	key := fmt.Sprintf("feeds/%s", hashID)
	redisFeed := &model.Feed{}
	if err := s.cache.GetItem(key, redisFeed); err != nil {
		return
	}

	feed.Episodes = redisFeed.Episodes
	feed.UpdatedAt = redisFeed.UpdatedAt
	feed.LastID = redisFeed.LastID
	feed.ItemURL = redisFeed.ItemURL
	feed.Author = redisFeed.Author
	feed.PubDate = redisFeed.PubDate
	feed.Description = redisFeed.Description
	feed.Title = redisFeed.Title
	feed.Language = redisFeed.Language
	feed.Explicit = redisFeed.Explicit
	feed.CoverArt = redisFeed.CoverArt
}

func (s *Service) BuildFeed(hashID string) ([]byte, error) {
	var (
		logger = log.WithField("hash_id", hashID)
		now    = time.Now().UTC()
		errKey = "err/" + hashID
	)

	cached, err := s.cache.Get(errKey)
	if err == nil {
		return []byte(cached), nil
	}

	feed, err := s.QueryFeed(hashID)
	if err != nil {
		logger.WithError(err).Error("failed to query feed from dynamodb")
		return nil, err
	}

	s.updateFromRedis(hashID, feed)

	oldLastID := feed.LastID

	const (
		updateTTL = 15 * time.Minute
	)

	if now.Sub(feed.UpdatedAt) < updateTTL {
		if podcast, err := s.buildPodcast(feed); err != nil {
			return nil, err
		} else {
			return []byte(podcast.String()), nil
		}
	}

	builderType := feed.Provider
	if feed.LastID != "" {
		// Use incremental Lambda updater if have seen this feed before
		builderType = api.ProviderGeneric
	}

	builder, ok := s.builders[builderType]
	if !ok {
		return nil, errors.Wrapf(err, "failed to get builder for feed: %s", hashID)
	}

	logger.Info("building feed")

	if err := builder.Build(feed); err != nil {
		logger.WithError(err).Error("failed to build feed")

		// Save error to cache to avoid requests spamming
		_ = s.cache.Set(errKey, err.Error(), updateTTL)

		return nil, err
	}

	// Format podcast

	if feed.PageSize < len(feed.Episodes) {
		feed.Episodes = feed.Episodes[:feed.PageSize]
	}

	feed.LastAccess = time.Now().UTC()

	// Don't zero last seen ID if failed to get updates
	if feed.LastID == "" {
		feed.LastID = oldLastID

		logger.Warnf("failed to get updates for %s (builder: %s)", hashID, builderType)
	}

	podcast, err := s.buildPodcast(feed)
	if err != nil {
		return nil, err
	}

	// Save to storage

	if oldLastID != feed.LastID {
		logger.Infof("updating feed %s (last id: %q, new id: %q)", hashID, oldLastID, feed.LastID)
		if err := s.storage.UpdateFeed(feed); err != nil {
			logger.WithError(err).Error("failed to save feed")
			return nil, err
		}
	}

	return []byte(podcast.String()), nil
}

func (s *Service) buildPodcast(feed *model.Feed) (*itunes.Podcast, error) {
	const (
		podsyncGenerator = "Podsync generator"
		defaultCategory  = "TV & Film"
	)

	now := time.Now().UTC()

	p := itunes.New(feed.Title, feed.ItemURL, feed.Description, &feed.PubDate, &now)
	p.Generator = podsyncGenerator
	p.AddSubTitle(feed.Title)
	p.AddCategory(defaultCategory, nil)
	p.AddImage(feed.CoverArt)
	p.IAuthor = feed.Title
	p.AddSummary(feed.Description)

	if feed.Explicit {
		p.IExplicit = "yes"
	} else {
		p.IExplicit = "no"
	}

	if feed.Language != "" {
		p.Language = feed.Language
	}

	for i, episode := range feed.Episodes {
		item := itunes.Item{
			GUID:        episode.ID,
			Link:        episode.VideoURL,
			Title:       episode.Title,
			Description: episode.Description,
			ISubtitle:   episode.Title,
			IOrder:      strconv.Itoa(i),
		}

		pubDate := time.Time(episode.PubDate)
		if pubDate.IsZero() {
			ts := now
			if i > 0 {
				if prev := time.Time(feed.Episodes[i-1].PubDate); !prev.IsZero() {
					ts = prev
				}
			}

			// HACK: some dates are cached incorrectly resulting incorrect behavior of some podcast clients.
			// Use this hack to have sequence ordered by date.
			pubDate = ts.Add(-time.Duration(i) * time.Hour)
		}

		item.AddPubDate(&pubDate)

		item.AddSummary(episode.Description)
		item.AddImage(episode.Thumbnail)
		item.AddDuration(episode.Duration)
		item.AddEnclosure(makeEnclosure(feed, episode.ID, episode.Size))

		// p.AddItem requires description to be not empty, use workaround
		if item.Description == "" {
			item.Description = " "
		}

		if feed.Explicit {
			item.IExplicit = "yes"
		} else {
			item.IExplicit = "no"
		}

		_, err := p.AddItem(item)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to add item to podcast (id %q)", episode.ID)
		}
	}

	return &p, nil
}

func (s Service) GetMetadata(hashID string) (*api.Metadata, error) {
	feed, err := s.storage.GetMetadata(hashID)
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

	ids, err := s.storage.Downgrade(patronID, featureLevel)
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
