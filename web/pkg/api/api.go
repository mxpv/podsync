package api

import "time"

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
	URL      string  `json:"url"`
	PageSize int     `json:"page_size"`
	Quality  Quality `json:"quality"`
	Format   Format  `json:"format"`
}
