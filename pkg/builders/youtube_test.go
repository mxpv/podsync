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

func TestQueryYTChannel(t *testing.T) {
	if ytKey == "" {
		t.Skip("YouTube API key is not provided")
	}

	builder, err := NewYouTubeBuilder(ytKey)
	require.NoError(t, err)

	channel, err := builder.listChannels(api.LinkTypeChannel, "UC2yTVSttx7lxAOAzx1opjoA")
	require.NoError(t, err)
	require.Equal(t, "UC2yTVSttx7lxAOAzx1opjoA", channel.Id)

	channel, err = builder.listChannels(api.LinkTypeUser, "fxigr1")
	require.NoError(t, err)
	require.Equal(t, "UCr_fwF-n-2_olTYd-m3n32g", channel.Id)
}

func TestBuildYTFeed(t *testing.T) {
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

	for _, feed := range feeds {
		t.Run(feed.ItemID, func(t *testing.T) {
			podcast, err := builder.Build(feed)

			require.NoError(t, err)

			assert.NotEmpty(t, podcast.Title)
			assert.NotEmpty(t, podcast.IAuthor)
			assert.NotEmpty(t, podcast.Description)

			assert.NotNil(t, podcast.ISummary)
			if podcast.ISummary != nil {
				assert.NotEmpty(t, podcast.ISummary.Text)
			}

			assert.NotZero(t, len(podcast.Items))

			for _, item := range podcast.Items {
				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.Link)
				assert.NotEmpty(t, item.IDuration)

				assert.NotNil(t, item.ISummary)
				if item.ISummary != nil {
					assert.NotEmpty(t, item.ISummary.Text)
				}

				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.IAuthor)
				assert.NotEmpty(t, item.Description)
			}
		})
	}
}
