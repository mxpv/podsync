package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/go-multierror"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"

	"github.com/mxpv/podsync/pkg/db"
	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/model"
	"github.com/mxpv/podsync/pkg/server"
	"github.com/mxpv/podsync/pkg/ytdl"
)

type Config struct {
	// Server is the web server configuration
	Server server.Config `toml:"server"`
	// Log is the optional logging configuration
	Log Log `toml:"log"`
	// Database configuration
	Database db.Config `toml:"database"`
	// Feeds is a list of feeds to host by this app.
	// ID will be used as feed ID in http://podsync.net/{FEED_ID}.xml
	Feeds map[string]*feed.Config
	// Tokens is API keys to use to access YouTube/Vimeo APIs.
	Tokens map[model.Provider]StringSlice `toml:"tokens"`
	// Downloader (youtube-dl) configuration
	Downloader ytdl.Config `toml:"downloader"`
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

	for id, f := range config.Feeds {
		f.ID = id
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

	if c.Server.Path != "" {
		var pathReg = regexp.MustCompile(model.PathRegex)
		if !pathReg.MatchString(c.Server.Path) {
			result = multierror.Append(result, errors.Errorf("Server handle path must be match %s or empty", model.PathRegex))
		}
	}

	if len(c.Feeds) == 0 {
		result = multierror.Append(result, errors.New("at least one feed must be specified"))
	}

	for id, f := range c.Feeds {
		if f.URL == "" {
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
		if feed.UpdatePeriod == 0 {
			feed.UpdatePeriod = model.DefaultUpdatePeriod
		}

		if feed.Quality == "" {
			feed.Quality = model.DefaultQuality
		}

		if feed.Custom.CoverArtQuality == "" {
			feed.Custom.CoverArtQuality = model.DefaultQuality
		}

		if feed.Format == "" {
			feed.Format = model.DefaultFormat
		}

		if feed.PageSize == 0 {
			feed.PageSize = model.DefaultPageSize
		}

		if feed.PlaylistSort == "" {
			feed.PlaylistSort = model.SortingAsc
		}
	}
}

// StringSlice is a toml extension that lets you to specify either a string
// value (a slice with just one element) or a string slice.
type StringSlice []string

func (s *StringSlice) UnmarshalTOML(v interface{}) error {
	if str, ok := v.(string); ok {
		*s = []string{str}
		return nil
	}

	return errors.New("failed to decode string slice field")
}
