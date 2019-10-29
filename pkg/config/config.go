package config

import (
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// Quality to use when downloading episodes
type Quality string

const (
	QualityHigh = Quality("high")
	QualityLow  = Quality("low")
)

// Format to convert episode when downloading episodes
type Format string

const (
	FormatAudio = Format("audio")
	FormatVideo = Format("video")
)

const (
	DefaultFormat       = FormatVideo
	DefaultQuality      = QualityHigh
	DefaultPageSize     = 50
	DefaultUpdatePeriod = 24 * time.Hour
)

// Feed is a configuration for a feed
type Feed struct {
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
	Quality Quality `toml:"quality"`
	// Format to use for this feed
	Format Format `toml:"format"`
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

	// Apply defaults
	for _, feed := range config.Feeds {
		if feed.UpdatePeriod.Duration == 0 {
			feed.UpdatePeriod.Duration = DefaultUpdatePeriod
		}

		if feed.Quality == "" {
			feed.Quality = DefaultQuality
		}

		if feed.Format == "" {
			feed.Format = DefaultFormat
		}

		if feed.PageSize == 0 {
			feed.PageSize = DefaultPageSize
		}
	}

	return &config, nil
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
