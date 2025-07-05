package builder

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/model"
)

func TestParseYoutubeURL_Playlist(t *testing.T) {
	link, _ := url.ParseRequestURI("https://www.youtube.com/playlist?list=PLCB9F975ECF01953C")
	kind, id, err := parseYoutubeURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypePlaylist, kind)
	require.Equal(t, "PLCB9F975ECF01953C", id)

	link, _ = url.ParseRequestURI("https://www.youtube.com/watch?v=rbCbho7aLYw&list=PLMpEfaKcGjpWEgNtdnsvLX6LzQL0UC0EM")
	kind, id, err = parseYoutubeURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypePlaylist, kind)
	require.Equal(t, "PLMpEfaKcGjpWEgNtdnsvLX6LzQL0UC0EM", id)

	// Test for YouTube handle with /playlists suffix
	link, _ = url.ParseRequestURI("https://www.youtube.com/@AndrejKarpathy/playlists")
	kind, id, err = parseYoutubeURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypePlaylist, kind)
	require.Equal(t, "@AndrejKarpathy", id)
}

func TestParseYoutubeURL_Channel(t *testing.T) {
	link, _ := url.ParseRequestURI("https://www.youtube.com/channel/UC5XPnUk8Vvv_pWslhwom6Og")
	kind, id, err := parseYoutubeURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypeChannel, kind)
	require.Equal(t, "UC5XPnUk8Vvv_pWslhwom6Og", id)

	link, _ = url.ParseRequestURI("https://www.youtube.com/channel/UCrlakW-ewUT8sOod6Wmzyow/videos")
	kind, id, err = parseYoutubeURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypeChannel, kind)
	require.Equal(t, "UCrlakW-ewUT8sOod6Wmzyow", id)

	// Test for YouTube handle with /videos suffix
	link, _ = url.ParseRequestURI("http://www.youtube.com/@aiDotEngineer/videos") // Test with http
	kind, id, err = parseYoutubeURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypeChannel, kind)
	require.Equal(t, "@aiDotEngineer", id)

	// Test for YouTube handle without any suffix (e.g. /videos, /playlists)
	link, _ = url.ParseRequestURI("https://www.youtube.com/@handleWithoutSuffix")
	kind, id, err = parseYoutubeURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypeChannel, kind)
	require.Equal(t, "@handleWithoutSuffix", id)

	// Test for YouTube handle that is just "@" (e.g. /@/videos)
	link, _ = url.ParseRequestURI("https://www.youtube.com/@/videos")
	kind, id, err = parseYoutubeURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypeChannel, kind)
	require.Equal(t, "@", id)
}

func TestParseYoutubeURL_User(t *testing.T) {
	link, _ := url.ParseRequestURI("https://youtube.com/user/fxigr1")
	kind, id, err := parseYoutubeURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypeUser, kind)
	require.Equal(t, "fxigr1", id)
}

func TestParseYoutubeURL_InvalidLink(t *testing.T) {
	link, _ := url.ParseRequestURI("https://www.youtube.com/user///")
	_, _, err := parseYoutubeURL(link)
	require.Error(t, err)

	link, _ = url.ParseRequestURI("https://www.youtube.com/channel//videos")
	_, _, err = parseYoutubeURL(link)
	require.Error(t, err)

	// Test for YouTube handle with an unsupported suffix
	link, _ = url.ParseRequestURI("https://www.youtube.com/@handle/feature")
	_, _, err = parseYoutubeURL(link)
	require.Error(t, err)
	require.EqualError(t, err, "unsupported youtube handle link suffix: /feature for handle @handle")

	// Test for unsupported /c/ style YouTube URLs
	link, _ = url.ParseRequestURI("https://www.youtube.com/c/ІлПідпільнийStandup/videos")
	_, _, err = parseYoutubeURL(link)
	require.Error(t, err)
	require.EqualError(t, err, "unsupported youtube link format")

	// Test for empty path
	link, _ = url.ParseRequestURI("https://www.youtube.com/")
	_, _, err = parseYoutubeURL(link)
	require.Error(t, err)
	require.EqualError(t, err, "youtube URL path is empty or just a slash")

	// Test for invalid playlist URL
	link, _ = url.ParseRequestURI("https://www.youtube.com/playlist")
	_, _, err = parseYoutubeURL(link)
	require.Error(t, err)
	require.EqualError(t, err, "playlist URL is missing 'list' query parameter")

	// Test for watch URL without list param (not a feed type)
	link, _ = url.ParseRequestURI("https://www.youtube.com/watch?v=somevideo")
	_, _, err = parseYoutubeURL(link)
	require.Error(t, err)
	require.EqualError(t, err, "watch URL without a 'list' query parameter is not a supported feed type")
}

// TestParseURL_SchemelessYoutubeHandle tests the top-level ParseURL for schemeless handle input
func TestParseURL_SchemelessYoutubeHandle(t *testing.T) {
	info, err := ParseURL("youtube.com/@handleWithoutSuffix")
	require.NoError(t, err)
	require.Equal(t, model.ProviderYoutube, info.Provider)
	require.Equal(t, model.TypeChannel, info.LinkType)
	require.Equal(t, "@handleWithoutSuffix", info.ItemID)
}

func TestParseVimeoURL_Group(t *testing.T) {
	link, _ := url.ParseRequestURI("https://vimeo.com/groups/109")
	kind, id, err := parseVimeoURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypeGroup, kind)
	require.Equal(t, "109", id)

	link, _ = url.ParseRequestURI("http://vimeo.com/groups/109")
	kind, id, err = parseVimeoURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypeGroup, kind)
	require.Equal(t, "109", id)

	link, _ = url.ParseRequestURI("http://www.vimeo.com/groups/109")
	kind, id, err = parseVimeoURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypeGroup, kind)
	require.Equal(t, "109", id)

	link, _ = url.ParseRequestURI("https://vimeo.com/groups/109/videos/")
	kind, id, err = parseVimeoURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypeGroup, kind)
	require.Equal(t, "109", id)
}

func TestParseVimeoURL_Channel(t *testing.T) {
	link, _ := url.ParseRequestURI("https://vimeo.com/channels/staffpicks")
	kind, id, err := parseVimeoURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypeChannel, kind)
	require.Equal(t, "staffpicks", id)

	link, _ = url.ParseRequestURI("http://vimeo.com/channels/staffpicks/146224925")
	kind, id, err = parseVimeoURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypeChannel, kind)
	require.Equal(t, "staffpicks", id)
}

func TestParseVimeoURL_User(t *testing.T) {
	link, _ := url.ParseRequestURI("https://vimeo.com/awhitelabelproduct")
	kind, id, err := parseVimeoURL(link)
	require.NoError(t, err)
	require.Equal(t, model.TypeUser, kind)
	require.Equal(t, "awhitelabelproduct", id)
}

func TestParseVimeoURL_InvalidLink(t *testing.T) {
	link, _ := url.ParseRequestURI("http://www.apple.com")
	_, _, err := parseVimeoURL(link)
	require.Error(t, err)

	link, _ = url.ParseRequestURI("http://www.vimeo.com")
	_, _, err = parseVimeoURL(link)
	require.Error(t, err)

	link, _ = url.ParseRequestURI("http://www.vimeo.com/groups/") // Missing group ID
	_, _, err = parseVimeoURL(link)
	require.Error(t, err)
	require.EqualError(t, err, "invalid vimeo group link: missing group ID")

	link, _ = url.ParseRequestURI("http://www.vimeo.com/channels/") // Missing channel ID
	_, _, err = parseVimeoURL(link)
	require.Error(t, err)
	require.EqualError(t, err, "invalid vimeo channel link: missing channel ID")
}
