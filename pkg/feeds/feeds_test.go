//go:generate mockgen -source=interfaces.go -destination=interfaces_mock_test.go -package=feeds

package feeds

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mxpv/podsync/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestService_CreateFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := NewMockid(ctrl)
	id.EXPECT().Generate(gomock.Any()).Times(1).Return("123", nil)

	storage := NewMockstorage(ctrl)
	storage.EXPECT().CreateFeed(gomock.Any()).Times(1).Return(nil)

	s := service{
		id:       id,
		storage:  storage,
		builders: map[api.Provider]builder{api.Youtube: nil},
	}

	req := &api.CreateFeedRequest{
		URL:      "youtube.com/channel/123",
		PageSize: 50,
		Quality:  api.HighQuality,
		Format:   api.VideoFormat,
	}

	hashId, err := s.CreateFeed(req, &api.Identity{})
	require.NoError(t, err)
	require.Equal(t, "123", hashId)
}

func TestService_GetFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	feed := &api.Feed{Provider: api.Youtube}

	storage := NewMockstorage(ctrl)
	storage.EXPECT().GetFeed("123").Times(1).Return(feed, nil)

	bld := NewMockbuilder(ctrl)
	bld.EXPECT().Build(feed).Return(nil, nil)

	s := service{
		storage:  storage,
		builders: map[api.Provider]builder{api.Youtube: bld},
	}

	_, err := s.GetFeed("123")
	require.NoError(t, err)
}

func TestService_GetMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := NewMockstorage(ctrl)
	storage.EXPECT().GetFeed("123").Times(1).Return(&api.Feed{}, nil)

	s := service{storage: storage}
	_, err := s.GetMetadata("123")
	require.NoError(t, err)
}
