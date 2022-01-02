package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/model"
	"github.com/mxpv/podsync/pkg/server"
)

func TestLoadConfig(t *testing.T) {
	const file = `
[tokens]
youtube = "123"
vimeo = ["321", "456"]

[server]
port = 80
data_dir = "test/data/"

[database]
dir = "/home/user/db/"

[downloader]
self_update = true
timeout = 15

[feeds]
  [feeds.XYZ]
  url = "https://youtube.com/watch?v=ygIUF678y40"
  page_size = 48
  update_period = "5h"
  format = "audio"
  quality = "low"
  filters = { title = "regex for title here" }
  playlist_sort = "desc"
  clean = { keep_last = 10 }
  [feeds.XYZ.custom]
  cover_art = "http://img"
  cover_art_quality = "high"
  category = "TV"
  subcategories = ["1", "2"]
  explicit = true
  lang = "en"
  author = "Mrs. Smith (mrs@smith.org)"
  ownerName = "Mrs. Smith"
  ownerEmail = "mrs@smith.org"
`
	path := setup(t, file)
	defer os.Remove(path)

	config, err := LoadConfig(path)
	assert.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, "test/data/", config.Server.DataDir)
	assert.EqualValues(t, 80, config.Server.Port)

	assert.Equal(t, "/home/user/db/", config.Database.Dir)

	require.Len(t, config.Tokens["youtube"], 1)
	assert.Equal(t, "123", config.Tokens["youtube"][0])
	require.Len(t, config.Tokens["vimeo"], 2)
	assert.Equal(t, "321", config.Tokens["vimeo"][0])
	assert.Equal(t, "456", config.Tokens["vimeo"][1])

	assert.Len(t, config.Feeds, 1)
	feed, ok := config.Feeds["XYZ"]
	assert.True(t, ok)
	assert.Equal(t, "https://youtube.com/watch?v=ygIUF678y40", feed.URL)
	assert.EqualValues(t, 48, feed.PageSize)
	assert.EqualValues(t, 5*time.Hour, feed.UpdatePeriod)
	assert.EqualValues(t, "audio", feed.Format)
	assert.EqualValues(t, "low", feed.Quality)
	assert.EqualValues(t, "regex for title here", feed.Filters.Title)
	assert.EqualValues(t, 10, feed.Clean.KeepLast)
	assert.EqualValues(t, model.SortingDesc, feed.PlaylistSort)

	assert.EqualValues(t, "http://img", feed.Custom.CoverArt)
	assert.EqualValues(t, "high", feed.Custom.CoverArtQuality)
	assert.EqualValues(t, "TV", feed.Custom.Category)
	assert.True(t, feed.Custom.Explicit)
	assert.EqualValues(t, "en", feed.Custom.Language)
	assert.EqualValues(t, "Mrs. Smith (mrs@smith.org)", feed.Custom.Author)
	assert.EqualValues(t, "Mrs. Smith", feed.Custom.OwnerName)
	assert.EqualValues(t, "mrs@smith.org", feed.Custom.OwnerEmail)

	assert.EqualValues(t, feed.Custom.Subcategories, []string{"1", "2"})

	assert.Nil(t, config.Database.Badger)

	assert.True(t, config.Downloader.SelfUpdate)
	assert.EqualValues(t, 15, config.Downloader.Timeout)
}

func TestLoadEmptyKeyList(t *testing.T) {
	const file = `
[tokens]
vimeo = []

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
	require.NotNil(t, config)

	require.Len(t, config.Tokens, 1)
	require.Len(t, config.Tokens["vimeo"], 0)
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

	assert.EqualValues(t, feed.UpdatePeriod, model.DefaultUpdatePeriod)
	assert.EqualValues(t, feed.PageSize, 50)
	assert.EqualValues(t, feed.Quality, "high")
	assert.EqualValues(t, feed.Custom.CoverArtQuality, "high")
	assert.EqualValues(t, feed.Format, "video")
}

func TestHttpServerListenAddress(t *testing.T) {
	const file = `
[server]
bind_address = "172.20.10.2"
port = 8080
path = "test"
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
	require.NotNil(t, config.Server.BindAddress)
	require.NotNil(t, config.Server.Path)
}

func TestDefaultHostname(t *testing.T) {
	cfg := Config{
		Server: server.Config{},
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
