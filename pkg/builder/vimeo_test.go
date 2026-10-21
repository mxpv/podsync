package builder

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/silentsokolov/go-vimeo/vimeo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/model"
)

var (
	vimeoKey = os.Getenv("VIMEO_TEST_API_KEY")
)

// retryOnVimeoRateLimit calls fn again after the rate limit resets when the
// live Vimeo API answers with HTTP 429. CI jobs from several pull requests
// share one token, so short bursts of rate limiting are expected.
func retryOnVimeoRateLimit(t *testing.T, fn func() error) error {
	t.Helper()

	const maxAttempts = 3

	for attempt := 1; ; attempt++ {
		err := fn()

		var rateErr *vimeo.RateLimitError
		if err == nil || attempt == maxAttempts || !errors.As(err, &rateErr) {
			return err
		}

		wait := time.Until(rateErr.Rate.Reset) + time.Second
		if wait < time.Second {
			wait = time.Second
		} else if wait > 90*time.Second {
			wait = 90 * time.Second
		}

		t.Logf("Vimeo rate limit hit, retrying in %s (attempt %d/%d)", wait.Round(time.Second), attempt, maxAttempts)
		time.Sleep(wait)
	}
}

func TestQueryVimeoChannel(t *testing.T) {
	if vimeoKey == "" {
		t.Skip("Vimeo API key is not provided")
	}

	builder, err := NewVimeoBuilder(context.Background(), vimeoKey)
	require.NoError(t, err)

	podcast := &model.Feed{ItemID: "staffpicks", Quality: model.QualityHigh}
	err = retryOnVimeoRateLimit(t, func() error { return builder.queryChannel(podcast) })
	require.NoError(t, err)

	assert.Equal(t, "https://vimeo.com/channels/staffpicks", podcast.ItemURL)
	assert.Equal(t, "Vimeo Staff Picks", podcast.Title)
	assert.Equal(t, "Vimeo Curation", podcast.Author)
	assert.NotEmpty(t, podcast.Description)
	assert.NotEmpty(t, podcast.CoverArt)
}

func TestQueryVimeoGroup(t *testing.T) {
	if vimeoKey == "" {
		t.Skip("Vimeo API key is not provided")
	}

	builder, err := NewVimeoBuilder(context.Background(), vimeoKey)
	require.NoError(t, err)

	podcast := &model.Feed{ItemID: "motion", Quality: model.QualityHigh}
	err = retryOnVimeoRateLimit(t, func() error { return builder.queryGroup(podcast) })
	require.NoError(t, err)

	assert.Equal(t, "https://vimeo.com/groups/motion", podcast.ItemURL)
	assert.Equal(t, "Motion Graphic Artists", podcast.Title)
	assert.Equal(t, "Danny Garcia", podcast.Author)
	assert.NotEmpty(t, podcast.Description)
	assert.NotEmpty(t, podcast.CoverArt)
}

func TestQueryVimeoUser(t *testing.T) {
	if vimeoKey == "" {
		t.Skip("Vimeo API key is not provided")
	}

	builder, err := NewVimeoBuilder(context.Background(), vimeoKey)
	require.NoError(t, err)

	podcast := &model.Feed{ItemID: "motionarray", Quality: model.QualityHigh}
	err = retryOnVimeoRateLimit(t, func() error { return builder.queryUser(podcast) })
	require.NoError(t, err)

	require.Equal(t, "https://vimeo.com/motionarray", podcast.ItemURL)
	assert.NotEmpty(t, podcast.Title)
	assert.NotEmpty(t, podcast.Author)
	assert.NotEmpty(t, podcast.Description)
}

func TestQueryVimeoVideos(t *testing.T) {
	if vimeoKey == "" {
		t.Skip("Vimeo API key is not provided")
	}

	builder, err := NewVimeoBuilder(context.Background(), vimeoKey)
	require.NoError(t, err)

	feed := &model.Feed{ItemID: "staffpicks", Quality: model.QualityHigh}

	err = retryOnVimeoRateLimit(t, func() error {
		feed.Episodes = nil
		return builder.queryVideos(builder.client.Channels.ListVideo, feed)
	})
	require.NoError(t, err)

	require.Equal(t, vimeoDefaultPageSize, len(feed.Episodes))

	for _, item := range feed.Episodes {
		require.NotEmpty(t, item.Title)
		require.NotEmpty(t, item.VideoURL)
		require.NotEmpty(t, item.ID)
		require.NotEmpty(t, item.Thumbnail)
		require.NotZero(t, item.Duration)
		require.NotZero(t, item.Size)
	}
}
