package id

import (
	"testing"

	"github.com/mxpv/podsync/web/pkg/storage"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	hid, err := NewIdGenerator()
	require.NoError(t, err)

	feed := &storage.Feed{
		UserId:   "1",
		URL:      "https://www.youtube.com/channel/UC2yTVSttx7lxAOAzx1opjoA",
		PageSize: 10,
		Quality:  storage.HighQuality,
		Format:   storage.AudioFormat,
	}

	hash1, err := hid.Encode(feed)
	require.NoError(t, err)
	require.NotEmpty(t, hash1)

	// Ensure we have same hash for same feed/parameters
	hash2, err := hid.Encode(feed)
	require.NoError(t, err)
	require.Equal(t, hash1, hash2)

	feed.UserId = ""
	hash3, err := hid.Encode(feed)
	require.NoError(t, err)
	require.NotEqual(t, hash1, hash3)
}
