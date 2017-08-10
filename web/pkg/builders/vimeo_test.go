package builders

import (
	"os"
	"testing"

	"context"

	itunes "github.com/mxpv/podcast"
	"github.com/mxpv/podsync/web/pkg/api"
	"github.com/stretchr/testify/require"
)

var (
	vimeoKey    = os.Getenv("VIMEO_TEST_API_KEY")
	defaultFeed = &api.Feed{Quality: api.HighQuality}
)

func TestParseVimeoGroupLink(t *testing.T) {
	builder := &VimeoBuilder{}

	kind, id, err := builder.parseUrl("https://vimeo.com/groups/109")
	require.NoError(t, err)
	require.Equal(t, linkTypeGroup, kind)
	require.Equal(t, "109", id)

	kind, id, err = builder.parseUrl("http://vimeo.com/groups/109")
	require.NoError(t, err)
	require.Equal(t, linkTypeGroup, kind)
	require.Equal(t, "109", id)

	kind, id, err = builder.parseUrl("http://www.vimeo.com/groups/109")
	require.NoError(t, err)
	require.Equal(t, linkTypeGroup, kind)
	require.Equal(t, "109", id)

	kind, id, err = builder.parseUrl("https://vimeo.com/groups/109/videos/")
	require.NoError(t, err)
	require.Equal(t, linkTypeGroup, kind)
	require.Equal(t, "109", id)
}

func TestParseVimeoChannelLink(t *testing.T) {
	builder := &VimeoBuilder{}

	kind, id, err := builder.parseUrl("https://vimeo.com/channels/staffpicks")
	require.NoError(t, err)
	require.Equal(t, linkTypeChannel, kind)
	require.Equal(t, "staffpicks", id)

	kind, id, err = builder.parseUrl("http://vimeo.com/channels/staffpicks/146224925")
	require.NoError(t, err)
	require.Equal(t, linkTypeChannel, kind)
	require.Equal(t, "staffpicks", id)
}

func TestParseVimeoUserLink(t *testing.T) {
	builder := &VimeoBuilder{}

	kind, id, err := builder.parseUrl("https://vimeo.com/awhitelabelproduct")
	require.NoError(t, err)
	require.Equal(t, linkTypeUser, kind)
	require.Equal(t, "awhitelabelproduct", id)
}

func TestParseInvalidVimeoLink(t *testing.T) {
	builder := &VimeoBuilder{}

	_, _, err := builder.parseUrl("")
	require.Error(t, err)

	_, _, err = builder.parseUrl("http://www.apple.com")
	require.Error(t, err)

	_, _, err = builder.parseUrl("http://www.vimeo.com")
	require.Error(t, err)
}

func TestQueryVimeoChannel(t *testing.T) {
	builder, err := NewVimeoBuilder(context.Background(), vimeoKey)
	require.NoError(t, err)

	podcast, err := builder.queryChannel("staffpicks", defaultFeed)
	require.NoError(t, err)

	require.Equal(t, "https://vimeo.com/channels/staffpicks", podcast.Link)
	require.Equal(t, "Vimeo Staff Picks", podcast.Title)
	require.Equal(t, "Vimeo Curation", podcast.IAuthor)
	require.NotEmpty(t, podcast.Description)
	require.NotEmpty(t, podcast.Image)
	require.NotEmpty(t, podcast.IImage)
}

func TestQueryVimeoGroup(t *testing.T) {
	builder, err := NewVimeoBuilder(context.Background(), vimeoKey)
	require.NoError(t, err)

	podcast, err := builder.queryGroup("motion", defaultFeed)
	require.NoError(t, err)

	require.Equal(t, "https://vimeo.com/groups/motion", podcast.Link)
	require.Equal(t, "Motion Graphic Artists", podcast.Title)
	require.Equal(t, "Danny Garcia", podcast.IAuthor)
	require.NotEmpty(t, podcast.Description)
	require.NotEmpty(t, podcast.Image)
	require.NotEmpty(t, podcast.IImage)
}

func TestQueryVimeoUser(t *testing.T) {
	builder, err := NewVimeoBuilder(context.Background(), vimeoKey)
	require.NoError(t, err)

	podcast, err := builder.queryUser("motionarray", defaultFeed)
	require.NoError(t, err)

	require.Equal(t, "https://vimeo.com/motionarray", podcast.Link)
	require.Equal(t, "Motion Array", podcast.Title)
	require.Equal(t, "Motion Array", podcast.IAuthor)
	require.NotEmpty(t, podcast.Description)
}

func TestQueryVimeoVideos(t *testing.T) {
	builder, err := NewVimeoBuilder(context.Background(), vimeoKey)
	require.NoError(t, err)

	feed := &itunes.Podcast{}

	err = builder.queryVideos(builder.client.Channels.ListVideo, "staffpicks", feed, &api.Feed{})
	require.NoError(t, err)

	require.Equal(t, vimeoDefaultPageSize, len(feed.Items))

	for _, item := range feed.Items {
		require.NotEmpty(t, item.Title)
		require.NotEmpty(t, item.Link)
		require.NotEmpty(t, item.GUID)
		require.NotEmpty(t, item.IDuration)
		require.NotNil(t, item.Enclosure)
		require.NotEmpty(t, item.Enclosure.URL)
		require.True(t, item.Enclosure.Length > 0)
	}
}
