package config

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/hashicorp/go-multierror"
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
	// Quality to use for this feed
	Quality model.Quality `toml:"quality"`
	// Format to use for this feed
	Format model.Format `toml:"format"`
	// Custom image to use
	CoverArt string `toml:"cover_art"`
}

type Tokens struct {
	// YouTube API key.
	// See https://developers.google.com/youtube/registering_an_application
	YouTube string `toml:"youtube"`
	// Vimeo developer key.
	// See https://developer.vimeo.com/api/guides/start#generate-access-token
	Vimeo string `toml:"vimeo"`
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

type Config struct {
	// Server is the web server configuration
	Server Server `toml:"server"`
	// Feeds is a list of feeds to host by this app.
	// ID will be used as feed ID in http://podsync.net/{FEED_ID}.xml
	Feeds map[string]*Feed
	// Tokens is API keys to use to access YouTube/Vimeo APIs.
	Tokens Tokens `toml:"tokens"`
}

// LoadConfig loads TOML configuration from a file path
func LoadConfig(path string) (*Config, error) {
	config := Config{}
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config file")
	}

	for id, feed := range config.Feeds {
		feed.ID = id
	}

	config.applyDefaults()

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

func (c *Config) applyDefaults() {
	if c.Server.Hostname == "" {
		if c.Server.Port != 0 && c.Server.Port != 80 {
			c.Server.Hostname = fmt.Sprintf("http://localhost:%d", c.Server.Port)
		} else {
			c.Server.Hostname = "http://localhost"
		}
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

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
