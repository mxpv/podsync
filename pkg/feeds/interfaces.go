package feeds

import (
	itunes "github.com/mxpv/podcast"
	"github.com/mxpv/podsync/pkg/api"
)

type id interface {
	Generate(feed *api.Feed) (string, error)
}

type storage interface {
	CreateFeed(feed *api.Feed) error
	GetFeed(hashId string) (*api.Feed, error)
}

type builder interface {
	Build(feed *api.Feed) (podcast *itunes.Podcast, err error)
}
