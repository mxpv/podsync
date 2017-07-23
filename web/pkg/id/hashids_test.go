package id

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEncode(t *testing.T) {
	hid, err := NewIdGenerator()
	require.NoError(t, err)

	hash, err := hid.Encode(1)
	require.NoError(t, err)
	require.NotEmpty(t, hash)
}
