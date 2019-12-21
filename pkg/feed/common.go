package feed

import (
	"context"
	"errors"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/model"
)

var (
	ErrNotFound      = errors.New("resource not found")
	ErrQuotaExceeded = errors.New("query limit is exceeded")
)

type Builder interface {
	Build(ctx context.Context, cfg *config.Feed) (*model.Feed, error)
}
