//go:generate mockgen -source=feeds.go -destination=feeds_mock_test.go -package=feeds

package feeds

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
	"github.com/stretchr/testify/require"
	"github.com/ventu-io/go-shortid"
)

func TestService_CreateFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := NewMockstorageService(ctrl)
	storage.EXPECT().CreateFeed(gomock.Any()).Times(1).Return(nil)

	s := Service{
		sid:      shortid.GetDefault(),
		storage:  storage,
		builders: map[api.Provider]builder{api.ProviderYoutube: nil},
	}

	req := &api.CreateFeedRequest{
		URL:      "youtube.com/channel/123",
		PageSize: 50,
		Quality:  api.QualityHigh,
		Format:   api.FormatVideo,
	}

	hashId, err := s.CreateFeed(req, &api.Identity{})
	require.NoError(t, err)
	require.NotEmpty(t, hashId)
}

func TestService_GetFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	feed := &model.Feed{Provider: api.ProviderYoutube}

	storage := NewMockstorageService(ctrl)
	storage.EXPECT().GetFeed("123").Times(1).Return(feed, nil)

	bld := NewMockbuilder(ctrl)
	bld.EXPECT().Build(feed).Return(nil, nil)

	s := Service{
		storage:  storage,
		builders: map[api.Provider]builder{api.ProviderYoutube: bld},
	}

	_, err := s.GetFeed("123")
	require.NoError(t, err)
}

func TestService_GetMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := NewMockstorageService(ctrl)
	storage.EXPECT().GetFeed("123").Times(1).Return(&model.Feed{}, nil)

	s := Service{storage: storage}
	_, err := s.GetMetadata("123")
	require.NoError(t, err)
}
