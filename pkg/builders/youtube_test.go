package builders

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
)

var ytKey = os.Getenv("YOUTUBE_TEST_API_KEY")

func TestYT_QueryChannel(t *testing.T) {
	if ytKey == "" {
		t.Skip("YouTube API key is not provided")
	}

	builder, err := NewYouTubeBuilder(ytKey)
	require.NoError(t, err)

	channel, err := builder.listChannels(api.LinkTypeChannel, "UC2yTVSttx7lxAOAzx1opjoA", "id")
	require.NoError(t, err)
	require.Equal(t, "UC2yTVSttx7lxAOAzx1opjoA", channel.Id)

	channel, err = builder.listChannels(api.LinkTypeUser, "fxigr1", "id")
	require.NoError(t, err)
	require.Equal(t, "UCr_fwF-n-2_olTYd-m3n32g", channel.Id)
}

func TestYT_BuildFeed(t *testing.T) {
	if ytKey == "" {
		t.Skip("YouTube API key is not provided")
	}

	builder, err := NewYouTubeBuilder(ytKey)
	require.NoError(t, err)

	feeds := []*model.Feed{
		{Provider: api.ProviderYoutube, LinkType: api.LinkTypeUser, ItemID: "fxigr1"},
		{Provider: api.ProviderYoutube, LinkType: api.LinkTypeChannel, ItemID: "UCupvZG-5ko_eiXAupbDfxWw"},
		{Provider: api.ProviderYoutube, LinkType: api.LinkTypePlaylist, ItemID: "PLfVk3KMh3VX1yJShGRsJmsqAjvMIviJYQ"},
		{Provider: api.ProviderYoutube, LinkType: api.LinkTypeChannel, ItemID: "UCK9lZ2lHRBgx2LOcqPifukA"},
		{Provider: api.ProviderYoutube, LinkType: api.LinkTypeUser, ItemID: "WylsaLive"},
		{Provider: api.ProviderYoutube, LinkType: api.LinkTypePlaylist, ItemID: "PLUVl5pafUrBydT_gsCjRGeCy0hFHloec8"},
	}

	for _, f := range feeds {
		feed := f
		t.Run(feed.ItemID, func(t *testing.T) {
			err := builder.Build(feed)
			require.NoError(t, err)

			assert.NotEmpty(t, feed.Title)
			assert.NotEmpty(t, feed.Description)
			assert.NotEmpty(t, feed.Author)
			assert.NotEmpty(t, feed.ItemURL)
			assert.NotEmpty(t, feed.LastID)

			assert.NotZero(t, len(feed.Episodes))

			for _, item := range feed.Episodes {
				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.VideoURL)
				assert.NotZero(t, item.Duration)

				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.Description)
				assert.NotEmpty(t, item.Thumbnail)
			}
		})
	}
}

func TestYT_GetVideoCount(t *testing.T) {
	if ytKey == "" {
		t.Skip("YouTube API key is not provided")
	}

	builder, err := NewYouTubeBuilder(ytKey)
	require.NoError(t, err)

	feeds := []*model.Feed{
		{Provider: api.ProviderYoutube, LinkType: api.LinkTypeUser, ItemID: "fxigr1"},
		{Provider: api.ProviderYoutube, LinkType: api.LinkTypeChannel, ItemID: "UCupvZG-5ko_eiXAupbDfxWw"},
		{Provider: api.ProviderYoutube, LinkType: api.LinkTypePlaylist, ItemID: "PLfVk3KMh3VX1yJShGRsJmsqAjvMIviJYQ"},
		{Provider: api.ProviderYoutube, LinkType: api.LinkTypeChannel, ItemID: "UCK9lZ2lHRBgx2LOcqPifukA"},
		{Provider: api.ProviderYoutube, LinkType: api.LinkTypeUser, ItemID: "WylsaLive"},
		{Provider: api.ProviderYoutube, LinkType: api.LinkTypePlaylist, ItemID: "PLUVl5pafUrBydT_gsCjRGeCy0hFHloec8"},
	}

	for _, f := range feeds {
		t.Run(f.ItemID, func(t *testing.T) {
			count, err := builder.GetVideoCount(f)
			assert.NoError(t, err)
			assert.NotZero(t, count)
		})
	}
}
