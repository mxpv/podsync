package builders

import (
	"testing"

	"os"

	"github.com/mxpv/podsync/web/pkg/database"
	"github.com/stretchr/testify/require"
)

var ytKey = os.Getenv("YOUTUBE_TEST_API_KEY")

func TestParseYTPlaylist(t *testing.T) {
	builder := &YouTubeBuilder{}

	kind, id, err := builder.parseUrl("https://www.youtube.com/playlist?list=PLCB9F975ECF01953C")
	require.NoError(t, err)
	require.Equal(t, linkTypePlaylist, kind)
	require.Equal(t, "PLCB9F975ECF01953C", id)
}

func TestParseYTChannel(t *testing.T) {
	builder := &YouTubeBuilder{}

	kind, id, err := builder.parseUrl("https://www.youtube.com/channel/UC5XPnUk8Vvv_pWslhwom6Og")
	require.NoError(t, err)
	require.Equal(t, linkTypeChannel, kind)
	require.Equal(t, "UC5XPnUk8Vvv_pWslhwom6Og", id)

	kind, id, err = builder.parseUrl("https://www.youtube.com/channel/UCrlakW-ewUT8sOod6Wmzyow/videos")
	require.NoError(t, err)
	require.Equal(t, linkTypeChannel, kind)
	require.Equal(t, "UCrlakW-ewUT8sOod6Wmzyow", id)
}

func TestParseYTUser(t *testing.T) {
	builder := &YouTubeBuilder{}

	kind, id, err := builder.parseUrl("https://youtube.com/user/fxigr1")
	require.NoError(t, err)
	require.Equal(t, linkTypeUser, kind)
	require.Equal(t, "fxigr1", id)
}

func TestHandleInvalidYTLink(t *testing.T) {
	builder := &YouTubeBuilder{}

	_, _, err := builder.parseUrl("https://www.youtube.com/user///")
	require.Error(t, err)

	_, _, err = builder.parseUrl("https://www.youtube.com/channel//videos")
	require.Error(t, err)
}

func TestQueryYTChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping YT test in short mode")
	}

	builder, err := NewYouTubeBuilder(ytKey)
	require.NoError(t, err)

	channel, err := builder.listChannels(linkTypeChannel, "UC2yTVSttx7lxAOAzx1opjoA")
	require.NoError(t, err)
	require.Equal(t, "UC2yTVSttx7lxAOAzx1opjoA", channel.Id)

	channel, err = builder.listChannels(linkTypeUser, "fxigr1")
	require.NoError(t, err)
	require.Equal(t, "UCr_fwF-n-2_olTYd-m3n32g", channel.Id)
}

func TestBuildYTFeed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping YT test in short mode")
	}

	builder, err := NewYouTubeBuilder(ytKey)
	require.NoError(t, err)

	podcast, err := builder.Build(&database.Feed{
		URL:      "https://youtube.com/channel/UCupvZG-5ko_eiXAupbDfxWw",
		PageSize: maxYoutubeResults,
	})
	require.NoError(t, err)

	require.Equal(t, "CNN", podcast.Title)
	require.NotEmpty(t, podcast.Description)

	require.Equal(t, 50, len(podcast.Items))

	for _, item := range podcast.Items {
		require.NotEmpty(t, item.Title)
		require.NotEmpty(t, item.Link)
		require.NotEmpty(t, item.IDuration)
	}
}
