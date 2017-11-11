package support

import (
	"fmt"
	"log"
	"strconv"

	"github.com/go-pg/pg"
	"github.com/mxpv/patreon-go"
	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
	"github.com/pkg/errors"
)

const (
	creatorID = "2822191"
)

type Patreon struct {
	db *pg.DB
}

func (h Patreon) toModel(pledge *patreon.Pledge) (*model.Pledge, error) {
	pledgeID, err := strconv.ParseInt(pledge.ID, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse pledge id: %s", pledge.ID)
	}

	patronID, err := strconv.ParseInt(pledge.Relationships.Patron.Data.ID, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse patron id: %s", pledge.Relationships.Patron.Data.ID)
	}

	m := &model.Pledge{
		PledgeID:    pledgeID,
		PatronID:    patronID,
		AmountCents: pledge.Attributes.AmountCents,
	}

	if pledge.Attributes.CreatedAt.Valid {
		m.CreatedAt = pledge.Attributes.CreatedAt.Time
	}

	if pledge.Attributes.DeclinedSince.Valid {
		m.DeclinedSince = pledge.Attributes.DeclinedSince.Time
	}

	// Read optional fields

	if pledge.Attributes.TotalHistoricalAmountCents != nil {
		m.TotalHistoricalAmountCents = *pledge.Attributes.TotalHistoricalAmountCents
	}

	if pledge.Attributes.OutstandingPaymentAmountCents != nil {
		m.OutstandingPaymentAmountCents = *pledge.Attributes.OutstandingPaymentAmountCents
	}

	if pledge.Attributes.IsPaused != nil {
		m.IsPaused = *pledge.Attributes.IsPaused
	}

	return m, nil
}

func (h Patreon) Hook(pledge *patreon.Pledge, event string) error {
	obj, err := h.toModel(pledge)
	if err != nil {
		return err
	}

	switch event {
	case patreon.EventCreatePledge:
		return h.db.Insert(obj)
	case patreon.EventUpdatePledge:
		err := h.db.Update(obj)
		if err == pg.ErrNoRows {
			log.Printf(
				"! ignoring update for not existing pledge %s for user %s",
				pledge.ID,
				pledge.Relationships.Patron.Data.ID)

			return nil
		}

		return err
	case patreon.EventDeletePledge:
		err := h.db.Delete(obj)
		if err == pg.ErrNoRows {
			return nil
		}

		return err
	default:
		return fmt.Errorf("unknown event: %s", event)
	}
}

func (h Patreon) FindPledge(patronID string) (*model.Pledge, error) {
	p := &model.Pledge{}
	return p, h.db.Model(p).Where("patron_id = ?", patronID).Limit(1).Select()
}

func (h Patreon) GetFeatureLevelByID(patronID string) (level int) {
	level = api.DefaultFeatures

	if patronID == "" {
		return
	}

	if patronID == creatorID {
		level = api.PodcasterFeature
		return
	}

	pledge, err := h.FindPledge(patronID)
	if err != nil {
		log.Printf("! can't find pledge for user %s: %v", patronID, err)
		return
	}

	// Check pledge is valid
	if pledge.DeclinedSince.IsZero() && !pledge.IsPaused {
		level = h.GetFeatureLevelFromAmount(pledge.AmountCents)
		return
	}

	return
}

func (h Patreon) GetFeatureLevelFromAmount(amount int) int {
	// Check the amount of pledge
	if amount >= 300 {
		return api.ExtendedPagination
	}

	if amount >= 100 {
		return api.ExtendedFeatures
	}

	return api.DefaultFeatures
}

func NewPatreon(db *pg.DB) *Patreon {
	return &Patreon{db: db}
}
