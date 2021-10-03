package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/config"
)

func TestSC_BUILDFEED(t *testing.T) {
	builder, err := NewSoundcloudBuilder()
	require.NoError(t, err)

	urls := []string{
		"https://soundcloud.com/moby/sets/remixes",
		"https://soundcloud.com/npr/sets/soundscapes",
	}

	for _, addr := range urls {
		t.Run(addr, func(t *testing.T) {
			feed, err := builder.Build(testCtx, &config.Feed{URL: addr})
			require.NoError(t, err)

			assert.NotEmpty(t, feed.Title)
			assert.NotEmpty(t, feed.Description)
			assert.NotEmpty(t, feed.Author)
			assert.NotEmpty(t, feed.ItemURL)

			assert.NotZero(t, len(feed.Episodes))

			for _, item := range feed.Episodes {
				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.VideoURL)
				assert.NotZero(t, item.Duration)
				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.Thumbnail)
			}
		})
	}
}
