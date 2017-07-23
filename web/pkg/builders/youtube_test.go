package builders

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var ytKey = "AIzaSyAp0mB03BFY3fm0Oxaxk96-mnE0D3MeUp4"

func TestParsePlaylist(t *testing.T) {
	builder := &YouTubeBuilder{}

	kind, id, err := builder.parseUrl("https://www.youtube.com/playlist?list=PLCB9F975ECF01953C")
	require.NoError(t, err)
	require.Equal(t, linkTypePlaylist, kind)
	require.Equal(t, "PLCB9F975ECF01953C", id)
}

func TestParseChannel(t *testing.T) {
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

func TestParseUser(t *testing.T) {
	builder := &YouTubeBuilder{}

	kind, id, err := builder.parseUrl("https://youtube.com/user/fxigr1")
	require.NoError(t, err)
	require.Equal(t, linkTypeUser, kind)
	require.Equal(t, "fxigr1", id)
}

func TestHandleInvalidLink(t *testing.T) {
	builder := &YouTubeBuilder{}

	_, _, err := builder.parseUrl("https://www.youtube.com/user///")
	require.Error(t, err)

	_, _, err = builder.parseUrl("https://www.youtube.com/channel//videos")
	require.Error(t, err)
}

func TestQueryChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping YT test in short mode")
	}

	builder, err := NewYouTubeBuilder(ytKey)
	require.NoError(t, err)

	channel, err := builder.queryChannel("UC2yTVSttx7lxAOAzx1opjoA", "")
	require.NoError(t, err)
	require.Equal(t, "UC2yTVSttx7lxAOAzx1opjoA", channel.Id)

	channel, err = builder.queryChannel("", "fxigr1")
	require.NoError(t, err)
	require.Equal(t, "UCr_fwF-n-2_olTYd-m3n32g", channel.Id)
}

func TestBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping YT test in short mode")
	}

	builder, err := NewYouTubeBuilder(ytKey)
	require.NoError(t, err)

	podcast, err := builder.Build("https://youtube.com/channel/UCupvZG-5ko_eiXAupbDfxWw", maxResults)
	require.NoError(t, err)

	require.Equal(t, "CNN", podcast.Title)
	require.NotEmpty(t, podcast.Description)

	require.Equal(t, 50, len(podcast.Items))

	for _, item := range podcast.Items {
		require.NotEmpty(t, item.Title)
		require.NotEmpty(t, item.Link)
	}
}
