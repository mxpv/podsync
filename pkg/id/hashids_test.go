package id

import (
	"testing"

	"github.com/mxpv/podsync/pkg/model"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	hid, err := NewIdGenerator()
	require.NoError(t, err)

	feed := &model.Feed{}

	hash, err := hid.Generate(feed)
	require.NoError(t, err)
	require.NotEmpty(t, hash)
}
