//go:generate mockgen -source=feeds.go -destination=feeds_mock_test.go -package=feeds

package feeds

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
)

var feed = &model.Feed{
	HashID:   "123",
	ItemID:   "xyz",
	Provider: api.ProviderVimeo,
	LinkType: api.LinkTypeChannel,
	PageSize: 50,
	Quality:  api.QualityHigh,
	Format:   api.FormatVideo,
}

func TestService_CreateFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockstorage(ctrl)
	db.EXPECT().SaveFeed(gomock.Any()).Times(1).Return(nil)

	gen, _ := NewIDGen()

	s := Service{
		generator: gen,
		storage:   db,
		builders:  map[api.Provider]Builder{api.ProviderYoutube: nil},
	}

	req := &api.CreateFeedRequest{
		URL:      "youtube.com/channel/123",
		PageSize: 50,
		Quality:  api.QualityHigh,
		Format:   api.FormatVideo,
	}

	hashID, err := s.CreateFeed(req, &api.Identity{})
	require.NoError(t, err)
	require.NotEmpty(t, hashID)
}

func TestService_makeFeed(t *testing.T) {
	req := &api.CreateFeedRequest{
		URL:      "youtube.com/channel/123",
		PageSize: 1000,
		Quality:  api.QualityLow,
		Format:   api.FormatAudio,
	}

	gen, _ := NewIDGen()

	s := Service{
		generator: gen,
	}

	feed, err := s.makeFeed(req, &api.Identity{})
	require.NoError(t, err)
	require.Equal(t, 50, feed.PageSize)
	require.Equal(t, api.QualityHigh, feed.Quality)
	require.Equal(t, api.FormatVideo, feed.Format)

	feed, err = s.makeFeed(req, &api.Identity{FeatureLevel: api.ExtendedFeatures})
	require.NoError(t, err)
	require.Equal(t, 150, feed.PageSize)
	require.Equal(t, api.QualityLow, feed.Quality)
	require.Equal(t, api.FormatAudio, feed.Format)

	feed, err = s.makeFeed(req, &api.Identity{FeatureLevel: api.ExtendedPagination})
	require.NoError(t, err)
	require.Equal(t, 600, feed.PageSize)
	require.Equal(t, api.QualityLow, feed.Quality)
	require.Equal(t, api.FormatAudio, feed.Format)
}

func TestService_QueryFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockstorage(ctrl)
	db.EXPECT().GetFeed("123").Times(1).Return(nil, nil)

	s := Service{storage: db}
	_, err := s.QueryFeed("123")
	require.NoError(t, err)
}

func TestService_GetFromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache := NewMockcacheService(ctrl)
	cache.EXPECT().Get("123").Return("test", nil)

	s := &Service{cache: cache}

	data, err := s.BuildFeed("123")
	assert.NoError(t, err)
	assert.Equal(t, []byte("test"), data)
}

func TestService_BuildFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stor := NewMockstorage(ctrl)
	stor.EXPECT().GetFeed(feed.HashID).Times(1).Return(feed, nil)
	stor.EXPECT().UpdateFeed(feed).Return(nil)

	cache := NewMockcacheService(ctrl)
	cache.EXPECT().Get(feed.HashID).Return("", errors.New("not found"))
	cache.EXPECT().Set(feed.HashID, gomock.Any(), 15*time.Minute).Return(nil)

	builder := NewMockBuilder(ctrl)
	builder.EXPECT().Build(feed).Return(nil)

	s := Service{storage: stor, cache: cache, builders: map[api.Provider]Builder{
		api.ProviderVimeo: builder,
	}}

	_, err := s.BuildFeed(feed.HashID)
	require.NoError(t, err)
}

func TestService_WrongID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache := NewMockcacheService(ctrl)
	cache.EXPECT().Get(gomock.Any()).Return("", errors.New("not found"))

	stor := NewMockstorage(ctrl)
	stor.EXPECT().GetFeed(gomock.Any()).Times(1).Return(nil, errors.New("not found"))

	s := &Service{storage: stor, cache: cache}

	_, err := s.BuildFeed("invalid_feed_id")
	require.Error(t, err)
}

func TestService_GetMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stor := NewMockstorage(ctrl)
	stor.EXPECT().GetMetadata(feed.HashID).Times(1).Return(feed, nil)

	s := &Service{storage: stor}

	m, err := s.GetMetadata(feed.HashID)
	require.NoError(t, err)
	require.EqualValues(t, 0, m.Downloads)
}
