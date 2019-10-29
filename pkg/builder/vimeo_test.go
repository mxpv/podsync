package builder

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/config"
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

	podcast := &Feed{ItemID: "staffpicks"}
	err = builder.queryChannel(podcast, &config.Feed{Quality: config.QualityHigh})
	require.NoError(t, err)

	require.Equal(t, "https://vimeo.com/channels/staffpicks", podcast.ItemURL)
	require.Equal(t, "Vimeo Staff Picks", podcast.Title)
	require.Equal(t, "Vimeo Curation", podcast.Author)
	require.NotEmpty(t, podcast.Description)
	require.NotEmpty(t, podcast.CoverArt)
}

func TestQueryVimeoGroup(t *testing.T) {
	if vimeoKey == "" {
		t.Skip("Vimeo API key is not provided")
	}

	builder, err := NewVimeoBuilder(context.Background(), vimeoKey)
	require.NoError(t, err)

	podcast := &Feed{ItemID: "motion"}
	err = builder.queryGroup(podcast, &config.Feed{Quality: config.QualityHigh})
	require.NoError(t, err)

	require.Equal(t, "https://vimeo.com/groups/motion", podcast.ItemURL)
	require.Equal(t, "Motion Graphic Artists", podcast.Title)
	require.Equal(t, "Danny Garcia", podcast.Author)
	require.NotEmpty(t, podcast.Description)
	require.NotEmpty(t, podcast.CoverArt)
}

func TestQueryVimeoUser(t *testing.T) {
	if vimeoKey == "" {
		t.Skip("Vimeo API key is not provided")
	}

	builder, err := NewVimeoBuilder(context.Background(), vimeoKey)
	require.NoError(t, err)

	podcast := &Feed{ItemID: "motionarray"}
	err = builder.queryUser(podcast, &config.Feed{Quality: config.QualityHigh})
	require.NoError(t, err)

	require.Equal(t, "https://vimeo.com/motionarray", podcast.ItemURL)
	require.Equal(t, "Motion Array", podcast.Title)
	require.Equal(t, "Motion Array", podcast.Author)
	require.NotEmpty(t, podcast.Description)
}

func TestQueryVimeoVideos(t *testing.T) {
	if vimeoKey == "" {
		t.Skip("Vimeo API key is not provided")
	}

	builder, err := NewVimeoBuilder(context.Background(), vimeoKey)
	require.NoError(t, err)

	feed := &Feed{ItemID: "staffpicks"}

	err = builder.queryVideos(builder.client.Channels.ListVideo, feed, &config.Feed{Quality: config.QualityHigh})
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
