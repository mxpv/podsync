package id

import (
	"testing"

	"github.com/mxpv/podsync/web/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	hid, err := NewIdGenerator()
	require.NoError(t, err)

	feed := &api.Feed{
		UserId:   "1",
		Provider: api.Youtube,
		LinkType: api.Channel,
		ItemId:   "UC2yTVSttx7lxAOAzx1opjoA",
		PageSize: 10,
		Quality:  api.HighQuality,
		Format:   api.AudioFormat,
	}

	hash1, err := hid.Generate(feed)
	require.NoError(t, err)
	require.NotEmpty(t, hash1)

	// Ensure we have same hash for same feed/parameters
	hash2, err := hid.Generate(feed)
	require.NoError(t, err)
	require.Equal(t, hash1, hash2)

	feed.UserId = ""
	hash3, err := hid.Generate(feed)
	require.NoError(t, err)
	require.NotEqual(t, hash1, hash3)
}
