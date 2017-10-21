package id

import (
	"testing"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	hid, err := NewIdGenerator()
	require.NoError(t, err)

	feed := &api.Feed{}

	hash, err := hid.Generate(feed)
	require.NoError(t, err)
	require.NotEmpty(t, hash)
}
