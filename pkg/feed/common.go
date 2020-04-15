package feed

import (
	"context"

	"github.com/pkg/errors"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/link"
	"github.com/mxpv/podsync/pkg/model"
)

var (
	ErrNotFound      = errors.New("resource not found")
	ErrQuotaExceeded = errors.New("query limit is exceeded")
)

type Builder interface {
	Build(ctx context.Context, cfg *config.Feed) (*model.Feed, error)
}

func New(ctx context.Context, cfg *config.Feed, tokens config.Tokens) (Builder, error) {
	var (
		provider Builder
		err      error
	)

	info, err := link.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}

	switch info.Provider {
	case link.ProviderYoutube:
		provider, err = NewYouTubeBuilder(tokens.YouTube)
	case link.ProviderVimeo:
		provider, err = NewVimeoBuilder(ctx, tokens.Vimeo)
	default:
		return nil, errors.Errorf("unsupported provider %q", info.Provider)
	}

	return provider, err
}

type feedProvider interface {
	GetFeed(ctx context.Context, feedID string) (*model.Feed, error)
}

type urlProvider interface {
	URL(ctx context.Context, ns string, fileName string) (string, error)
}
