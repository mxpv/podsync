package config

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/model"
)

func TestLoadConfig(t *testing.T) {
	const file = `
[tokens]
youtube = "123"
vimeo = "321"

[server]
port = 80
data_dir = "test/data/"

[feeds]
  [feeds.XYZ]
  url = "https://youtube.com/watch?v=ygIUF678y40"
  page_size = 48
  update_period = "5h"
  format = "audio"
  quality = "low"
`

	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)

	defer os.Remove(f.Name())

	_, err = f.WriteString(file)
	require.NoError(t, err)

	config, err := LoadConfig(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, config)

	assert.Equal(t, "test/data/", config.Server.DataDir)
	assert.EqualValues(t, 80, config.Server.Port)

	assert.Equal(t, "123", config.Tokens.YouTube)
	assert.Equal(t, "321", config.Tokens.Vimeo)

	assert.Len(t, config.Feeds, 1)
	feed, ok := config.Feeds["XYZ"]
	assert.True(t, ok)
	assert.Equal(t, "https://youtube.com/watch?v=ygIUF678y40", feed.URL)
	assert.EqualValues(t, 48, feed.PageSize)
	assert.EqualValues(t, Duration{5 * time.Hour}, feed.UpdatePeriod)
	assert.EqualValues(t, "audio", feed.Format)
	assert.EqualValues(t, "low", feed.Quality)
}

func TestApplyDefaults(t *testing.T) {
	const file = `
[server]
data_dir = "/data"

[feeds]
  [feeds.A]
  url = "https://youtube.com/watch?v=ygIUF678y40"
`
	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)

	defer os.Remove(f.Name())

	_, err = f.WriteString(file)
	require.NoError(t, err)

	config, err := LoadConfig(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, config)

	assert.Len(t, config.Feeds, 1)
	feed, ok := config.Feeds["A"]
	require.True(t, ok)

	assert.EqualValues(t, feed.UpdatePeriod, Duration{model.DefaultUpdatePeriod})
	assert.EqualValues(t, feed.PageSize, 50)
	assert.EqualValues(t, feed.Quality, "high")
	assert.EqualValues(t, feed.Format, "video")
}

func TestDefaultHostname(t *testing.T) {
	cfg := Config{
		Server: Server{},
	}

	t.Run("empty hostname", func(t *testing.T) {
		cfg.applyDefaults()
		assert.Equal(t, "http://localhost", cfg.Server.Hostname)
	})

	t.Run("empty hostname with port", func(t *testing.T) {
		cfg.Server.Hostname = ""
		cfg.Server.Port = 7979
		cfg.applyDefaults()
		assert.Equal(t, "http://localhost:7979", cfg.Server.Hostname)
	})

	t.Run("skip overwrite", func(t *testing.T) {
		cfg.Server.Hostname = "https://my.host:4443"
		cfg.Server.Port = 80
		cfg.applyDefaults()
		assert.Equal(t, "https://my.host:4443", cfg.Server.Hostname)
	})
}
