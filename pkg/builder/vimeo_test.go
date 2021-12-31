package builder

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/model"
)

var (
	vimeoKey = os.Getenv("VIMEO_TEST_API_KEY")
)

func TestQueryVimeoChannel(t *testing.T) {
	if vimeoKey == "" {
		t.Skip("Vimeo API key is not provided")
	}

	builder, err := NewVimeoBuilder(context.Background(), vimeoKey)
	require.NoError(t, err)

	podcast := &model.Feed{ItemID: "staffpicks", Quality: model.QualityHigh}
	err = builder.queryChannel(podcast)
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
	err = builder.queryGroup(podcast)
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
	err = builder.queryUser(podcast)
	require.NoError(t, err)

	require.Equal(t, "https://vimeo.com/motionarray", podcast.ItemURL)
	assert.Equal(t, "Artlist Ltd", podcast.Title)
	assert.Equal(t, "Artlist Ltd", podcast.Author)
	assert.NotEmpty(t, podcast.Description)
}

func TestQueryVimeoVideos(t *testing.T) {
	if vimeoKey == "" {
		t.Skip("Vimeo API key is not provided")
	}

	builder, err := NewVimeoBuilder(context.Background(), vimeoKey)
	require.NoError(t, err)

	feed := &model.Feed{ItemID: "staffpicks", Quality: model.QualityHigh}

	err = builder.queryVideos(builder.client.Channels.ListVideo, feed)
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
