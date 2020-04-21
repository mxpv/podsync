package feed

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFixedKey(t *testing.T) {
	key, err := NewFixedKey("123")
	assert.NoError(t, err)

	assert.EqualValues(t, "123", key.Get())
	assert.EqualValues(t, "123", key.Get())
}

func TestNewRotatedKeys(t *testing.T) {
	key, err := NewRotatedKeys([]string{"123", "456"})
	assert.NoError(t, err)

	assert.EqualValues(t, "123", key.Get())
	assert.EqualValues(t, "456", key.Get())

	assert.EqualValues(t, "123", key.Get())
	assert.EqualValues(t, "456", key.Get())
}
