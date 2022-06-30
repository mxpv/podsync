package builder

import (
	"testing"

	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRss_BuildFeed(t *testing.T) {
	builder, err := NewRssBuilder()
	require.NoError(t, err)

	urls := []string{
		"https://rsshub.app/bilibili/user/video/2267573",
	}

	for _, addr := range urls {
		t.Run(addr, func(t *testing.T) {
			feed, err := builder.Build(testCtx, &feed.Config{URL: addr, PageSize: 10})
			require.NoError(t, err)

			assert.Equal(t, feed.Provider, model.ProviderRss)
			assert.NotEmpty(t, feed.Title)
			assert.NotEmpty(t, feed.Description)
			assert.NotEmpty(t, feed.Author)
			assert.NotEmpty(t, feed.ItemURL)
			assert.NotEmpty(t, feed.CoverArt)

			assert.NotZero(t, len(feed.Episodes))

			for _, item := range feed.Episodes {
				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.VideoURL)
				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.Thumbnail)
			}
		})
	}
}
