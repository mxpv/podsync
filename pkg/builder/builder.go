package builder

import (
	"context"

	"github.com/pkg/errors"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/model"
)

type Builder interface {
	Build(ctx context.Context, cfg *config.Feed) (*model.Feed, error)
}

func New(ctx context.Context, provider model.Provider, key string) (Builder, error) {
	switch provider {
	case model.ProviderYoutube:
		return NewYouTubeBuilder(key)
	case model.ProviderVimeo:
		return NewVimeoBuilder(ctx, key)
	case model.ProviderSoundcloud:
		return NewSoundcloudBuilder()
	default:
		return nil, errors.Errorf("unsupported provider %q", provider)
	}
}
