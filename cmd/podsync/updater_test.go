package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/config"
)

func TestUpdater_hostname(t *testing.T) {
	u := Updater{
		config: &config.Config{
			Server: config.Server{
				Hostname: "localhost",
				Port:     7979,
			},
		},
	}

	assert.Equal(t, "http://localhost", u.hostname())

	// Trim end slash
	u.config.Server.Hostname = "https://localhost:8080/"
	assert.Equal(t, "https://localhost:8080", u.hostname())
}

func TestCopyFile(t *testing.T) {
	reader := bytes.NewReader([]byte{1, 2, 4})

	tmpDir, err := ioutil.TempDir("", "podsync-test-")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	file := filepath.Join(tmpDir, "1")

	size, err := copyFile(reader, file)
	assert.NoError(t, err)
	assert.EqualValues(t, 3, size)

	stat, err := os.Stat(file)
	assert.NoError(t, err)
	assert.EqualValues(t, 3, stat.Size())
}
