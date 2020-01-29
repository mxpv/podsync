package model

import (
	"time"

	"github.com/mxpv/podsync/pkg/link"
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

type Episode struct {
	// ID of episode
	ID          string
	Title       string
	Description string
	Thumbnail   string
	Duration    int64
	VideoURL    string
	PubDate     time.Time
	Size        int64
	Order       string
}

type Feed struct {
	FeedID         string
	ItemID         string
	LinkType       link.Type     // Either group, channel or user
	Provider       link.Provider // Youtube or Vimeo
	CreatedAt      time.Time
	LastAccess     time.Time
	ExpirationTime time.Time
	Format         Format
	Quality        Quality
	PageSize       int
	CoverArt       string
	Explicit       bool
	Language       string // ISO 639
	Title          string
	Description    string
	PubDate        time.Time
	Author         string
	ItemURL        string     // Platform specific URL
	Episodes       []*Episode // Array of episodes, serialized as gziped EpisodesData in DynamoDB
	UpdatedAt      time.Time
}
