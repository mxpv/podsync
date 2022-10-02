package builder

import (
	"context"
	"os"
	"strings"

	"github.com/mxpv/podsync/pkg/feed"
	"github.com/pkg/errors"

	"github.com/mxpv/podsync/pkg/model"
)

type Builder interface {
	Build(ctx context.Context, cfg *feed.Config) (*model.Feed, error)
}

func UnmarshalKey(key string) string {
	if strings.HasPrefix(key, "env:") {
		return os.Getenv(strings.TrimPrefix(key, "env:"))
	}
	return key
}

func New(ctx context.Context, provider model.Provider, key string) (Builder, error) {
	switch provider {
	case model.ProviderYoutube:
		return NewYouTubeBuilder(UnmarshalKey(key))
	case model.ProviderVimeo:
		return NewVimeoBuilder(ctx, UnmarshalKey(key))
	case model.ProviderSoundcloud:
		return NewSoundcloudBuilder()
	default:
		return nil, errors.Errorf("unsupported provider %q", provider)
	}
}
