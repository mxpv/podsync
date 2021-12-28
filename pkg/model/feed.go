package model

import (
	"time"
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

// Playlist sorting style
type Sorting string

const (
	SortingDesc = Sorting("desc")
	SortingAsc  = Sorting("asc")
)

type Episode struct {
	// ID of episode
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Thumbnail   string        `json:"thumbnail"`
	Duration    int64         `json:"duration"`
	VideoURL    string        `json:"video_url"`
	PubDate     time.Time     `json:"pub_date"`
	Size        int64         `json:"size"`
	Order       string        `json:"order"`
	Status      EpisodeStatus `json:"status"` // Disk status
}

type Feed struct {
	ID              string     `json:"feed_id"`
	ItemID          string     `json:"item_id"`
	LinkType        Type       `json:"link_type"` // Either group, channel or user
	Provider        Provider   `json:"provider"`  // Youtube or Vimeo
	CreatedAt       time.Time  `json:"created_at"`
	LastAccess      time.Time  `json:"last_access"`
	ExpirationTime  time.Time  `json:"expiration_time"`
	Format          Format     `json:"format"`
	Quality         Quality    `json:"quality"`
	CoverArtQuality Quality    `json:"cover_art_quality"`
	PageSize        int        `json:"page_size"`
	CoverArt        string     `json:"cover_art"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	PubDate         time.Time  `json:"pub_date"`
	Author          string     `json:"author"`
	ItemURL         string     `json:"item_url"` // Platform specific URL
	Episodes        []*Episode `json:"-"`        // Array of episodes
	UpdatedAt       time.Time  `json:"updated_at"`
	PlaylistSort    Sorting    `json:"playlist_sort"`
}

type EpisodeStatus string

const (
	EpisodeNew        = EpisodeStatus("new")        // New episode received via API
	EpisodeDownloaded = EpisodeStatus("downloaded") // Downloaded, encoded and available for download
	EpisodeError      = EpisodeStatus("error")      // Could not download, will retry
	EpisodeCleaned    = EpisodeStatus("cleaned")    // Downloaded and later removed from disk due to update strategy
)
