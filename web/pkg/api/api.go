package api

import "time"

type Quality string
type Format string

const (
	HighQuality = Quality("high")
	LowQuality  = Quality("low")
	AudioFormat = Format("audio")
	VideoFormat = Format("video")
)

const (
	DefaultPageSize = 50
	DefaultFormat   = VideoFormat
	DefaultQuality  = HighQuality
)

type Feed struct {
	Id         int64     `json:"id"`
	HashId     string    `json:"hash_id"`
	UserId     string    `json:"user_id"`
	URL        string    `json:"url"`
	PageSize   int       `json:"page_size"`
	Quality    Quality   `json:"quality"`
	Format     Format    `json:"format"`
	LastAccess time.Time `json:"last_access"`
}
