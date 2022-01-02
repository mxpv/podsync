package feed

import (
	"time"

	"github.com/mxpv/podsync/pkg/model"
)

// Config is a configuration for a feed loaded from TOML
type Config struct {
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
	UpdatePeriod time.Duration `toml:"update_period"`
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
	// List of additional youtube-dl arguments passed at download time
	YouTubeDLArgs []string `toml:"youtube_dl_args"`
	// Included in OPML file
	OPML bool `toml:"opml"`
	// Playlist sort
	PlaylistSort model.Sorting `toml:"playlist_sort"`
}

type Filters struct {
	Title          string `toml:"title"`
	NotTitle       string `toml:"not_title"`
	Description    string `toml:"description"`
	NotDescription string `toml:"not_description"`
	// More filters to be added here
}

type Custom struct {
	CoverArt        string        `toml:"cover_art"`
	CoverArtQuality model.Quality `toml:"cover_art_quality"`
	Category        string        `toml:"category"`
	Subcategories   []string      `toml:"subcategories"`
	Explicit        bool          `toml:"explicit"`
	Language        string        `toml:"lang"`
	Author          string        `toml:"author"`
	Title           string        `toml:"title"`
	Description     string        `toml:"description"`
	OwnerName       string        `toml:"ownerName"`
	OwnerEmail      string        `toml:"ownerEmail"`
}

type Cleanup struct {
	// KeepLast defines how many episodes to keep
	KeepLast int `toml:"keep_last"`
}
