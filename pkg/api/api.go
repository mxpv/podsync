package api

import (
	"github.com/pkg/errors"
)

var (
	ErrNotFound      = errors.New("resource not found")
	ErrQuotaExceeded = errors.New("query limit is exceeded")
)

type Provider string

const (
	ProviderYoutube = Provider("youtube")
	ProviderVimeo   = Provider("vimeo")
)

type LinkType string

const (
	LinkTypeChannel  = LinkType("channel")
	LinkTypePlaylist = LinkType("playlist")
	LinkTypeUser     = LinkType("user")
	LinkTypeGroup    = LinkType("group")
)

type Quality string

const (
	QualityHigh = Quality("high")
	QualityLow  = Quality("low")
)

type Format string

const (
	FormatAudio = Format("audio")
	FormatVideo = Format("video")
)

const (
	DefaultPageSize              = 50
	DefaultFormat                = FormatVideo
	DefaultQuality               = QualityHigh
	ExtendedPaginationQueryLimit = 5000
)

type Metadata struct {
	Provider  Provider `json:"provider"`
	Format    Format   `json:"format"`
	Quality   Quality  `json:"quality"`
	Downloads int64    `json:"downloads"`
}

const (
	// DefaultFeatures represent features for Anonymous user
	// Page size: 50
	// Format: video
	// Quality: high
	DefaultFeatures = iota

	// ExtendedFeatures represent features for 1$ pledges
	// Max page size: 150
	// Format: any
	// Quality: any
	ExtendedFeatures

	// ExtendedPagination represent extended pagination feature set
	// Max page size: 600
	// Format: any
	// Quality: any
	ExtendedPagination

	// PodcasterFeatures reserved for future
	PodcasterFeatures
)

type CreateFeedRequest struct {
	URL      string  `json:"url" binding:"url,required"`
	PageSize int     `json:"page_size" binding:"min=10,max=600,required"`
	Quality  Quality `json:"quality" binding:"eq=high|eq=low"`
	Format   Format  `json:"format" binding:"eq=video|eq=audio"`
}

type Identity struct {
	UserID       string `json:"user_id"`
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	ProfileURL   string `json:"profile_url"`
	FeatureLevel int    `json:"feature_level"`
}
