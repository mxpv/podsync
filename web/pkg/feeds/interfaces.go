package feeds

import (
	itunes "github.com/mxpv/podcast"
	"github.com/mxpv/podsync/web/pkg/api"
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

type parser interface {
	ParseURL(link string) (feed *api.Feed, err error)
}
