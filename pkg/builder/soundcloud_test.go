package builder

import (
	"context"
	"testing"

	"github.com/mxpv/podsync/pkg/feed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCtx = context.Background()

// newSoundcloudBuilderSafe attempts to create a SoundCloud builder,
// returning nil if initialization fails (including panics from the library).
func newSoundcloudBuilderSafe() (builder *SoundCloudBuilder) {
	defer func() {
		if r := recover(); r != nil {
			builder = nil
		}
	}()

	var err error
	builder, err = NewSoundcloudBuilder()
	if err != nil {
		return nil
	}
	return builder
}

func TestSoundCloud_BuildFeed(t *testing.T) {
	builder := newSoundcloudBuilderSafe()
	if builder == nil {
		t.Skip("Skipping SoundCloud test: unable to initialize SoundCloud client (service may be unavailable)")
	}

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
