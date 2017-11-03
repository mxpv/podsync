package model

import "time"

type Pledge struct {
	PledgeID                      int64 `sql:",pk"`
	PatronID                      int64
	CreatedAt                     time.Time
	DeclinedSince                 time.Time
	AmountCents                   int
	TotalHistoricalAmountCents    int
	OutstandingPaymentAmountCents int
	IsPaused                      bool
}
