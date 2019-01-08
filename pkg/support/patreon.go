package support

import (
	"fmt"
	"strconv"

	patreon "github.com/mxpv/patreon-go"
	"github.com/pkg/errors"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"

	log "github.com/sirupsen/logrus"
)

const (
	creatorID = "2822191"
)

type storage interface {
	AddPledge(pledge *model.Pledge) error
	UpdatePledge(patronID string, pledge *model.Pledge) error
	DeletePledge(pledge *model.Pledge) error
	GetPledge(patronID string) (*model.Pledge, error)
}

type Patreon struct {
	db storage
}

func ToModel(pledge *patreon.Pledge) (*model.Pledge, error) {
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
	logger := log.WithFields(log.Fields{
		"module":       "hook",
		"pledge_id":    pledge.ID,
		"pledge_event": event,
	})

	obj, err := ToModel(pledge)
	if err != nil {
		logger.WithError(err).Error("failed to convert pledge to model")
		return err
	}

	switch event {
	case patreon.EventCreatePledge:
		return h.db.AddPledge(obj)
	case patreon.EventUpdatePledge:
		// Update comes with different PledgeID from Patreon, so do update by user ID
		patronID := pledge.Relationships.Patron.Data.ID
		if err := h.db.UpdatePledge(patronID, obj); err != nil {
			return err
		}

		return nil
	case patreon.EventDeletePledge:
		return h.db.DeletePledge(obj)
	default:
		return fmt.Errorf("unknown event: %s", event)
	}
}

func (h Patreon) FindPledge(patronID string) (*model.Pledge, error) {
	return h.db.GetPledge(patronID)
}

func (h Patreon) GetFeatureLevelByID(patronID string) (level int) {
	level = api.DefaultFeatures

	if patronID == "" {
		return
	}

	if patronID == creatorID {
		level = api.PodcasterFeatures
		return
	}

	pledge, err := h.FindPledge(patronID)
	if err != nil {
		log.WithError(err).WithField("user_id", patronID).Error("can't find pledge for user")
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

func NewPatreon(db storage) *Patreon {
	return &Patreon{db: db}
}
