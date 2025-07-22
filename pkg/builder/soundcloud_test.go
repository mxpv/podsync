package builder

import (
	"context"
	"testing"

	"github.com/mxpv/podsync/pkg/feed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCtx = context.Background()

func TestSoundCloud_BuildFeed(t *testing.T) {
	builder, err := NewSoundcloudBuilder()
	require.NoError(t, err)

	urls := []string{
		"https://soundcloud.com/moby/sets/remixes",
		"https://soundcloud.com/npr/sets/soundscapes",
	}

	for _, addr := range urls {
		t.Run(addr, func(t *testing.T) {
			_feed, err := builder.Build(testCtx, &feed.Config{URL: addr})
			require.NoError(t, err)

			assert.NotEmpty(t, _feed.Title)
			assert.NotEmpty(t, _feed.Description)
			assert.NotEmpty(t, _feed.Author)
			assert.NotEmpty(t, _feed.ItemURL)

			assert.NotZero(t, len(_feed.Episodes))

			for _, item := range _feed.Episodes {
				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.VideoURL)
				assert.NotZero(t, item.Duration)
				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.Thumbnail)
			}
		})
	}
}
