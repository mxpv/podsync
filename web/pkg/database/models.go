package database

import "time"

type Quality string
type Format string

const (
	HighQuality = Quality("high")
	LowQuality  = Quality("low")
	AudioFormat = Format("audio")
	VideoFormat = Format("video")
)

type Feed struct {
	Id         int64
	HashId     string
	UserId     string
	URL        string
	PageSize   int
	Quality    Quality
	Format     Format
	LastAccess time.Time
}
