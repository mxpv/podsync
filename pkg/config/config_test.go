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

[database]
dir = "/home/user/db/"

[feeds]
  [feeds.XYZ]
  url = "https://youtube.com/watch?v=ygIUF678y40"
  page_size = 48
  update_period = "5h"
  format = "audio"
  quality = "low"
  filters = { title = "regex for title here" }
  clean = { keep_last = 10 }
  custom = { cover_art = "http://img", category = "TV", explicit = true, lang = "en" }
`
	path := setup(t, file)
	defer os.Remove(path)

	config, err := LoadConfig(path)
	assert.NoError(t, err)
	assert.NotNil(t, config)

	assert.Equal(t, "test/data/", config.Server.DataDir)
	assert.EqualValues(t, 80, config.Server.Port)

	assert.Equal(t, "/home/user/db/", config.Database.Dir)

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
	assert.EqualValues(t, "regex for title here", feed.Filters.Title)
	assert.EqualValues(t, 10, feed.Clean.KeepLast)

	assert.EqualValues(t, "http://img", feed.Custom.CoverArt)
	assert.EqualValues(t, "TV", feed.Custom.Category)
	assert.True(t, feed.Custom.Explicit)
	assert.EqualValues(t, "en", feed.Custom.Language)

	assert.Nil(t, config.Database.Badger)
}

func TestApplyDefaults(t *testing.T) {
	const file = `
[server]
data_dir = "/data"

[feeds]
  [feeds.A]
  url = "https://youtube.com/watch?v=ygIUF678y40"
`
	path := setup(t, file)
	defer os.Remove(path)

	config, err := LoadConfig(path)
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
		cfg.applyDefaults("")
		assert.Equal(t, "http://localhost", cfg.Server.Hostname)
	})

	t.Run("empty hostname with port", func(t *testing.T) {
		cfg.Server.Hostname = ""
		cfg.Server.Port = 7979
		cfg.applyDefaults("")
		assert.Equal(t, "http://localhost:7979", cfg.Server.Hostname)
	})

	t.Run("skip overwrite", func(t *testing.T) {
		cfg.Server.Hostname = "https://my.host:4443"
		cfg.Server.Port = 80
		cfg.applyDefaults("")
		assert.Equal(t, "https://my.host:4443", cfg.Server.Hostname)
	})
}

func TestDefaultDatabasePath(t *testing.T) {
	cfg := Config{}
	cfg.applyDefaults("/home/user/podsync/config.toml")
	assert.Equal(t, "/home/user/podsync/db", cfg.Database.Dir)
}

func TestLoadBadgerConfig(t *testing.T) {
	const file = `
[server]
data_dir = "/data"

[feeds]
  [feeds.A]
  url = "https://youtube.com/watch?v=ygIUF678y40"

[database]
  badger = { truncate = true, file_io = true }
`
	path := setup(t, file)
	defer os.Remove(path)

	config, err := LoadConfig(path)
	assert.NoError(t, err)
	require.NotNil(t, config)
	require.NotNil(t, config.Database.Badger)

	assert.True(t, config.Database.Badger.Truncate)
	assert.True(t, config.Database.Badger.FileIO)
}

func setup(t *testing.T, file string) string {
	t.Helper()

	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)

	defer f.Close()

	_, err = f.WriteString(file)
	require.NoError(t, err)

	return f.Name()
}
