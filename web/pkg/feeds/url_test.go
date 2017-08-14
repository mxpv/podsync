package feeds

import (
	"net/url"
	"testing"

	"github.com/mxpv/podsync/web/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestParseYTPlaylist(t *testing.T) {
	link, _ := url.ParseRequestURI("https://www.youtube.com/playlist?list=PLCB9F975ECF01953C")
	kind, id, err := parseYoutubeURL(link)
	require.NoError(t, err)
	require.Equal(t, api.Playlist, kind)
	require.Equal(t, "PLCB9F975ECF01953C", id)
}

func TestParseYTChannel(t *testing.T) {
	link, _ := url.ParseRequestURI("https://www.youtube.com/channel/UC5XPnUk8Vvv_pWslhwom6Og")
	kind, id, err := parseYoutubeURL(link)
	require.NoError(t, err)
	require.Equal(t, api.Channel, kind)
	require.Equal(t, "UC5XPnUk8Vvv_pWslhwom6Og", id)

	link, _ = url.ParseRequestURI("https://www.youtube.com/channel/UCrlakW-ewUT8sOod6Wmzyow/videos")
	kind, id, err = parseYoutubeURL(link)
	require.NoError(t, err)
	require.Equal(t, api.Channel, kind)
	require.Equal(t, "UCrlakW-ewUT8sOod6Wmzyow", id)
}

func TestParseYTUser(t *testing.T) {
	link, _ := url.ParseRequestURI("https://youtube.com/user/fxigr1")
	kind, id, err := parseYoutubeURL(link)
	require.NoError(t, err)
	require.Equal(t, api.User, kind)
	require.Equal(t, "fxigr1", id)
}

func TestHandleInvalidYTLink(t *testing.T) {
	link, _ := url.ParseRequestURI("https://www.youtube.com/user///")
	_, _, err := parseYoutubeURL(link)
	require.Error(t, err)

	link, _ = url.ParseRequestURI("https://www.youtube.com/channel//videos")
	_, _, err = parseYoutubeURL(link)
	require.Error(t, err)
}

func TestParseVimeoGroupLink(t *testing.T) {
	link, _ := url.ParseRequestURI("https://vimeo.com/groups/109")
	kind, id, err := parseVimeoURL(link)
	require.NoError(t, err)
	require.Equal(t, api.Group, kind)
	require.Equal(t, "109", id)

	link, _ = url.ParseRequestURI("http://vimeo.com/groups/109")
	kind, id, err = parseVimeoURL(link)
	require.NoError(t, err)
	require.Equal(t, api.Group, kind)
	require.Equal(t, "109", id)

	link, _ = url.ParseRequestURI("http://www.vimeo.com/groups/109")
	kind, id, err = parseVimeoURL(link)
	require.NoError(t, err)
	require.Equal(t, api.Group, kind)
	require.Equal(t, "109", id)

	link, _ = url.ParseRequestURI("https://vimeo.com/groups/109/videos/")
	kind, id, err = parseVimeoURL(link)
	require.NoError(t, err)
	require.Equal(t, api.Group, kind)
	require.Equal(t, "109", id)
}

func TestParseVimeoChannelLink(t *testing.T) {
	link, _ := url.ParseRequestURI("https://vimeo.com/channels/staffpicks")
	kind, id, err := parseVimeoURL(link)
	require.NoError(t, err)
	require.Equal(t, api.Channel, kind)
	require.Equal(t, "staffpicks", id)

	link, _ = url.ParseRequestURI("http://vimeo.com/channels/staffpicks/146224925")
	kind, id, err = parseVimeoURL(link)
	require.NoError(t, err)
	require.Equal(t, api.Channel, kind)
	require.Equal(t, "staffpicks", id)
}

func TestParseVimeoUserLink(t *testing.T) {
	link, _ := url.ParseRequestURI("https://vimeo.com/awhitelabelproduct")
	kind, id, err := parseVimeoURL(link)
	require.NoError(t, err)
	require.Equal(t, api.User, kind)
	require.Equal(t, "awhitelabelproduct", id)
}

func TestParseInvalidVimeoLink(t *testing.T) {
	link, _ := url.ParseRequestURI("http://www.apple.com")
	_, _, err := parseVimeoURL(link)
	require.Error(t, err)

	link, _ = url.ParseRequestURI("http://www.vimeo.com")
	_, _, err = parseVimeoURL(link)
	require.Error(t, err)
}
