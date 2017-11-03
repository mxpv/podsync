package support

import (
	"fmt"
	"log"
	"strconv"

	"github.com/go-pg/pg"
	"github.com/mxpv/patreon-go"
	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/models"
	"github.com/pkg/errors"
)

const (
	creatorID = "2822191"
)

type Patreon struct {
	db *pg.DB
}

func (h Patreon) toModel(pledge *patreon.Pledge) (*models.Pledge, error) {
	pledgeID, err := strconv.ParseInt(pledge.ID, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse pledge id: %s", pledge.ID)
	}

	patronID, err := strconv.ParseInt(pledge.Relationships.Patron.Data.ID, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse patron id: %s", pledge.Relationships.Patron.Data.ID)
	}

	model := &models.Pledge{
		PledgeID:    pledgeID,
		PatronID:    patronID,
		AmountCents: pledge.Attributes.AmountCents,
	}

	if pledge.Attributes.CreatedAt.Valid {
		model.CreatedAt = pledge.Attributes.CreatedAt.Time
	}

	if pledge.Attributes.DeclinedSince.Valid {
		model.DeclinedSince = pledge.Attributes.DeclinedSince.Time
	}

	// Read optional fields

	if pledge.Attributes.TotalHistoricalAmountCents != nil {
		model.TotalHistoricalAmountCents = *pledge.Attributes.TotalHistoricalAmountCents
	}

	if pledge.Attributes.OutstandingPaymentAmountCents != nil {
		model.OutstandingPaymentAmountCents = *pledge.Attributes.OutstandingPaymentAmountCents
	}

	if pledge.Attributes.IsPaused != nil {
		model.IsPaused = *pledge.Attributes.IsPaused
	}

	return model, nil
}

func (h Patreon) Hook(pledge *patreon.Pledge, event string) error {
	model, err := h.toModel(pledge)
	if err != nil {
		return err
	}

	switch event {
	case patreon.EventCreatePledge:
		return h.db.Insert(model)
	case patreon.EventUpdatePledge:
		return h.db.Update(model)
	case patreon.EventDeletePledge:
		err := h.db.Delete(model)
		if err == pg.ErrNoRows {
			return nil
		}

		return err
	default:
		return fmt.Errorf("unknown event: %s", event)
	}
}

func (h Patreon) FindPledge(patronID string) (*models.Pledge, error) {
	p := &models.Pledge{}
	return p, h.db.Model(p).Where("patron_id = ?", patronID).Limit(1).Select()
}

func (h Patreon) GetFeatureLevel(patronID string) (level int) {
	level = api.DefaultFeatures

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
		// Check the amount of pledge
		if pledge.AmountCents >= 100 {
			level = api.ExtendedFeatures
			return
		}
	}

	return
}

func NewPatreon(db *pg.DB) *Patreon {
	return &Patreon{db: db}
}
