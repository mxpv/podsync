package builders

import (
	"os"
	"testing"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

var ytKey = os.Getenv("YOUTUBE_TEST_API_KEY")

func TestQueryYTChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping YT test in short mode")
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
	if testing.Short() {
		t.Skip("skipping YT test in short mode")
	}

	builder, err := NewYouTubeBuilder(ytKey)
	require.NoError(t, err)

	feeds := []*model.Feed{
		{
			Provider: api.ProviderYoutube,
			LinkType: api.LinkTypeChannel,
			ItemID:   "UCupvZG-5ko_eiXAupbDfxWw",
			PageSize: maxYoutubeResults,
		},
		{
			Provider: api.ProviderYoutube,
			LinkType: api.LinkTypePlaylist,
			ItemID: "PLfVk3KMh3VX1yJShGRsJmsqAjvMIviJYQ",
			PageSize: maxYoutubeResults,
		},
	}

	for _, feed := range feeds {
		t.Run(feed.ItemID, func(t *testing.T) {
			podcast, err := builder.Build(feed)

			require.NoError(t, err)

			assert.NotEmpty(t, podcast.Title)
			assert.NotEmpty(t, podcast.IAuthor)
			assert.NotEmpty(t, podcast.Description)

			require.NotNil(t, podcast.ISummary)
			assert.NotEmpty(t, podcast.ISummary.Text)

			assert.Equal(t, 50, len(podcast.Items))

			for _, item := range podcast.Items {
				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.Link)
				assert.NotEmpty(t, item.IDuration)

				require.NotNil(t, item.ISummary)
				assert.NotEmpty(t, item.ISummary.Text)

				assert.NotEmpty(t, item.Title)
				assert.NotEmpty(t, item.IAuthor)
				assert.NotEmpty(t, item.Description)
			}
		})
	}
}
