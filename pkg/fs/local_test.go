package fs

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var (
	testCtx = context.Background()
)

func TestNewLocal(t *testing.T) {
	local, err := NewLocal("", false)
	assert.NoError(t, err)
	assert.NotNil(t, local)
}

func TestLocal_Create(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)

	stor, err := NewLocal(tmpDir, false)
	assert.NoError(t, err)

	written, err := stor.Create(testCtx, "1/test", bytes.NewBuffer([]byte{1, 5, 7, 8, 3}))
	assert.NoError(t, err)
	assert.EqualValues(t, 5, written)

	stat, err := os.Stat(filepath.Join(tmpDir, "1", "test"))
	assert.NoError(t, err)
	assert.EqualValues(t, 5, stat.Size())
}

func TestLocal_Size(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	stor, err := NewLocal(tmpDir, false)
	assert.NoError(t, err)

	_, err = stor.Create(testCtx, "1/test", bytes.NewBuffer([]byte{1, 5, 7, 8, 3}))
	assert.NoError(t, err)

	sz, err := stor.Size(testCtx, "1/test")
	assert.NoError(t, err)
	assert.EqualValues(t, 5, sz)
}

func TestLocal_NoSize(t *testing.T) {
	stor, err := NewLocal("", false)
	assert.NoError(t, err)

	_, err = stor.Size(testCtx, "1/test")
	assert.True(t, os.IsNotExist(err))
}

func TestLocal_Delete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)

	stor, err := NewLocal(tmpDir, false)
	assert.NoError(t, err)

	_, err = stor.Create(testCtx, "1/test", bytes.NewBuffer([]byte{1, 5, 7, 8, 3}))
	assert.NoError(t, err)

	err = stor.Delete(testCtx, "1/test")
	assert.NoError(t, err)

	_, err = stor.Size(testCtx, "1/test")
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(filepath.Join(tmpDir, "1", "test"))
	assert.True(t, os.IsNotExist(err))

	err = stor.Delete(testCtx, "1/test")
	assert.True(t, errors.Is(err, os.ErrNotExist))
}

func TestLocal_copyFile(t *testing.T) {
	reader := bytes.NewReader([]byte{1, 2, 4})
	tmpDir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)

	file := filepath.Join(tmpDir, "1")

	l := &Local{}
	size, err := l.copyFile(reader, file)
	assert.NoError(t, err)
	assert.EqualValues(t, 3, size)

	stat, err := os.Stat(file)
	assert.NoError(t, err)
	assert.EqualValues(t, 3, stat.Size())
}
