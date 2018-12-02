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
}
