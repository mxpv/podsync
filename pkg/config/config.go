package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/naoina/toml"
	"github.com/pkg/errors"

	"github.com/mxpv/podsync/pkg/model"
)

// Feed is a configuration for a feed
type Feed struct {
	ID string `toml:"-"`
	// URL is a full URL of the field
	URL string `toml:"url"`
	// PageSize is the number of pages to query from YouTube API.
	// NOTE: larger page sizes/often requests might drain your API token.
	PageSize int `toml:"page_size"`
	// UpdatePeriod is how often to check for updates.
	// Format is "300ms", "1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// NOTE: too often update check might drain your API token.
	UpdatePeriod Duration `toml:"update_period"`
	// Cron expression format is how often to check update
	// NOTE: too often update check might drain your API token.
	CronSchedule string `toml:"cron_schedule"`
	// Quality to use for this feed
	Quality model.Quality `toml:"quality"`
	// Maximum height of video
	MaxHeight int `toml:"max_height"`
	// Format to use for this feed
	Format model.Format `toml:"format"`
	// Only download episodes that match this regexp (defaults to matching anything)
	Filters Filters `toml:"filters"`
	// Clean is a cleanup policy to use for this feed
	Clean Cleanup `toml:"clean"`
	// Custom is a list of feed customizations
	Custom Custom `toml:"custom"`
	// Included in OPML file
	OPML bool `toml:"opml"`
}

type Filters struct {
	Title string `toml:"title"`
	// More filters to be added here
}

type Custom struct {
	CoverArt string `toml:"cover_art"`
	Category string `toml:"category"`
	Explicit bool   `toml:"explicit"`
	Language string `toml:"lang"`
}

type Server struct {
	// Hostname to use for download links
	Hostname string `toml:"hostname"`
	// Port is a server port to listen to
	Port int `toml:"port"`
	// DataDir is a path to a directory to keep XML feeds and downloaded episodes,
	// that will be available to user via web server for download.
	DataDir string `toml:"data_dir"`
}

type Database struct {
	// Dir is a directory to keep database files
	Dir    string  `toml:"dir"`
	Badger *Badger `toml:"badger"`
}

// Badger represents BadgerDB configuration parameters
// See https://github.com/dgraph-io/badger#memory-usage
type Badger struct {
	Truncate bool `toml:"truncate"`
	FileIO   bool `toml:"file_io"`
}

type Cleanup struct {
	// KeepLast defines how many episodes to keep
	KeepLast int `toml:"keep_last"`
}

type Log struct {
	// Filename to write the log to (instead of stdout)
	Filename string `toml:"filename"`
	// MaxSize is the maximum size of the log file in MB
	MaxSize int `toml:"max_size"`
	// MaxBackups is the maximum number of log file backups to keep after rotation
	MaxBackups int `toml:"max_backups"`
	// MaxAge is the maximum number of days to keep the logs for
	MaxAge int `toml:"max_age"`
	// Compress old backups
	Compress bool `toml:"compress"`
}

// Downloader is a youtube-dl related configuration
type Downloader struct {
	// SelfUpdate toggles self update every 24 hour
	SelfUpdate bool `toml:"self_update"`
}

type Config struct {
	// Server is the web server configuration
	Server Server `toml:"server"`
	// Log is the optional logging configuration
	Log Log `toml:"log"`
	// Database configuration
	Database Database `toml:"database"`
	// Feeds is a list of feeds to host by this app.
	// ID will be used as feed ID in http://podsync.net/{FEED_ID}.xml
	Feeds map[string]*Feed
	// Tokens is API keys to use to access YouTube/Vimeo APIs.
	Tokens map[model.Provider]StringSlice `toml:"tokens"`
	// Downloader (youtube-dl) configuration
	Downloader Downloader `toml:"downloader"`
}

// LoadConfig loads TOML configuration from a file path
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read config file: %s", path)
	}

	config := Config{}
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal toml")
	}

	for id, feed := range config.Feeds {
		feed.ID = id
	}

	config.applyDefaults(path)

	if err := config.validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) validate() error {
	var result *multierror.Error

	if c.Server.DataDir == "" {
		result = multierror.Append(result, errors.New("data directory is required"))
	}

	if len(c.Feeds) == 0 {
		result = multierror.Append(result, errors.New("at least one feed must be speficied"))
	}

	for id, feed := range c.Feeds {
		if feed.URL == "" {
			result = multierror.Append(result, errors.Errorf("URL is required for %q", id))
		}
	}

	return result.ErrorOrNil()
}

func (c *Config) applyDefaults(configPath string) {
	if c.Server.Hostname == "" {
		if c.Server.Port != 0 && c.Server.Port != 80 {
			c.Server.Hostname = fmt.Sprintf("http://localhost:%d", c.Server.Port)
		} else {
			c.Server.Hostname = "http://localhost"
		}
	}

	if c.Log.Filename != "" {
		if c.Log.MaxSize == 0 {
			c.Log.MaxSize = model.DefaultLogMaxSize
		}
		if c.Log.MaxAge == 0 {
			c.Log.MaxAge = model.DefaultLogMaxAge
		}
		if c.Log.MaxBackups == 0 {
			c.Log.MaxBackups = model.DefaultLogMaxBackups
		}
	}

	if c.Database.Dir == "" {
		c.Database.Dir = filepath.Join(filepath.Dir(configPath), "db")
	}

	for _, feed := range c.Feeds {
		if feed.UpdatePeriod.Duration == 0 {
			feed.UpdatePeriod.Duration = model.DefaultUpdatePeriod
		}

		if feed.Quality == "" {
			feed.Quality = model.DefaultQuality
		}

		if feed.Format == "" {
			feed.Format = model.DefaultFormat
		}

		if feed.PageSize == 0 {
			feed.PageSize = model.DefaultPageSize
		}
	}
}
