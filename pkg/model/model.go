package model

import (
	"time"

	"github.com/mxpv/podsync/pkg/api"
)

//noinspection SpellCheckingInspection
type Pledge struct {
	PledgeID                      int64 `sql:",pk"`
	PatronID                      int64
	CreatedAt                     time.Time `dynamodbav:",unixtime"`
	DeclinedSince                 time.Time `dynamodbav:",unixtime"`
	AmountCents                   int
	TotalHistoricalAmountCents    int
	OutstandingPaymentAmountCents int
	IsPaused                      bool
}

type Item struct {
	ID          string
	Title       string
	Description string
	Thumbnail   string
	Duration    int64
	VideoURL    string
	PubDate     time.Time `dynamodbav:",unixtime"`
	Size        int64

	Order string `dynamodbav:"-"`
}

//noinspection SpellCheckingInspection
type Feed struct {
	FeedID         int64  `sql:",pk" dynamodbav:"-"`
	HashID         string // Short human readable feed id for users
	UserID         string // Patreon user id
	ItemID         string
	LinkType       api.LinkType // Either group, channel or user
	Provider       api.Provider // Youtube or Vimeo
	PageSize       int          // The number of episodes to return
	Format         api.Format
	Quality        api.Quality
	FeatureLevel   int
	CreatedAt      time.Time `dynamodbav:",unixtime"`
	LastAccess     time.Time `dynamodbav:",unixtime"`
	ExpirationTime time.Time `sql:"-" dynamodbav:",unixtime"`
	CoverArt       string    `dynamodbav:",omitempty"`
	Explicit       bool
	Language       string `dynamodbav:",omitempty"` // ISO 639
	Title          string
	Description    string
	PubDate        time.Time `dynamodbav:",unixtime"`
	Author         string
	ItemURL        string    // Platform specific URL
	Episodes       []*Item   // Array of episodes
	LastID         string    // Last seen video URL ID (for incremental updater)
	UpdatedAt      time.Time `dynamodbav:",unixtime"`
}
