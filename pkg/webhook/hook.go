package webhook

import (
	"fmt"
	"strconv"

	"github.com/go-pg/pg"
	"github.com/mxpv/patreon-go"
	"github.com/mxpv/podsync/pkg/models"
	"github.com/pkg/errors"
)

const (
	EventHeader     = "X-Patreon-Event"
	SignatureHeader = "X-Patreon-Signature"

	EventNameCreatePledge = "pledges:create"
	EventNameUpdatePledge = "pledges:update"
	EventNameDeletePledge = "pledges:delete"
)

type Handler struct {
	db *pg.DB
}

func (h Handler) toModel(pledge *patreon.Pledge) (*models.Pledge, error) {
	pledgeID, err := strconv.ParseInt(pledge.Id, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse pledge id: %s", pledge.Id)
	}

	patronID, err := strconv.ParseInt(pledge.Relationships.Patron.Data.Id, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse patron id: %s", pledge.Relationships.Patron.Data.Id)
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

func (h Handler) Handle(pledge *patreon.Pledge, event string) error {
	model, err := h.toModel(pledge)
	if err != nil {
		return err
	}

	switch event {
	case EventNameCreatePledge:
		return h.db.Insert(model)
	case EventNameUpdatePledge:
		return h.db.Update(model)
	case EventNameDeletePledge:
		return h.db.Delete(model)
	default:
		return fmt.Errorf("unknown event: %s", event)
	}
}

func NewHookHandler(db *pg.DB) *Handler {
	return &Handler{db: db}
}
