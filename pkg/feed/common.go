package feed

import (
	"context"

	"github.com/pkg/errors"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/model"
)

type Builder interface {
	Build(ctx context.Context, cfg *config.Feed) (*model.Feed, error)
}

func New(ctx context.Context, cfg *config.Feed, keys map[model.Provider]KeyProvider) (Builder, error) {
	info, err := ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	keyProvider, ok := keys[info.Provider]
	if !ok {
		return nil, errors.Errorf("unknown key provider: %s", info.Provider)
	}

	switch info.Provider {
	case model.ProviderYoutube:
		return NewYouTubeBuilder(keyProvider.Get())
	case model.ProviderVimeo:
		return NewVimeoBuilder(ctx, keyProvider.Get())
	default:
		return nil, errors.Errorf("unsupported provider %q", info.Provider)
	}
}

type feedProvider interface {
	GetFeed(ctx context.Context, feedID string) (*model.Feed, error)
}

type urlProvider interface {
	URL(ctx context.Context, ns string, fileName string) (string, error)
}
