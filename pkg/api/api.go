package api

import (
	"time"

	"github.com/pkg/errors"
)

var (
	ErrNotFound = errors.New("resource not found")
)

type Provider string

const (
	Youtube = Provider("youtube")
	Vimeo   = Provider("vimeo")
)

type LinkType string

const (
	Channel  = LinkType("channel")
	Playlist = LinkType("playlist")
	User     = LinkType("user")
	Group    = LinkType("group")
)

type Quality string

const (
	HighQuality = Quality("high")
	LowQuality  = Quality("low")
)

type Format string

const (
	AudioFormat = Format("audio")
	VideoFormat = Format("video")
)

const (
	DefaultPageSize = 50
	DefaultFormat   = VideoFormat
	DefaultQuality  = HighQuality
)

type Feed struct {
	Id           int64     `json:"id"`
	HashId       string    `json:"hash_id"` // Short human readable feed id for users
	UserId       string    `json:"user_id"` // Patreon user id
	ItemId       string    `json:"item_id"`
	Provider     Provider  `json:"provider"`  // Youtube or Vimeo
	LinkType     LinkType  `json:"link_type"` // Either group, channel or user
	PageSize     int       `json:"page_size"` // The number of episodes to return
	Format       Format    `json:"format"`
	Quality      Quality   `json:"quality"`
	FeatureLevel int       `json:"feature_level"` // Available features
	LastAccess   time.Time `json:"last_access"`
}

const (
	DefaultFeatures = iota
	ExtendedFeatures
	PodcasterFeature
)

type CreateFeedRequest struct {
	URL      string  `json:"url" binding:"url,required"`
	PageSize int     `json:"page_size" binding:"min=10,max=150,required"`
	Quality  Quality `json:"quality" binding:"eq=high|eq=low"`
	Format   Format  `json:"format" binding:"eq=video|eq=audio"`
}

type Identity struct {
	UserId       string `json:"user_id"`
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	ProfileURL   string `json:"profile_url"`
	FeatureLevel int    `json:"feature_level"`
}
