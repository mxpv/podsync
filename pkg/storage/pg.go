package storage

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/proxy"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
)

type Postgres struct {
	db *pg.DB
}

func NewPG(connectionURL string, ping bool) (Postgres, error) {
	opts, err := pg.ParseURL(connectionURL)
	if err != nil {
		return Postgres{}, err
	}

	// If host format is "projection:region:host", than use Google SQL Proxy
	// See https://github.com/go-pg/pg/issues/576
	if strings.Count(opts.Addr, ":") == 2 {
		log.Print("using GCP SQL proxy")
		opts.Dialer = func(network, addr string) (net.Conn, error) {
			return proxy.Dial(addr)
		}
	}

	db := pg.Connect(opts)

	// Check database connectivity
	if ping {
		if _, err := db.ExecOne("SELECT 1"); err != nil {
			_ = db.Close()
			return Postgres{}, errors.Wrap(err, "failed to check database connectivity")
		}
	}

	return Postgres{db: db}, nil
}

func (p Postgres) SaveFeed(feed *model.Feed) error {
	_, err := p.db.Model(feed).Insert()
	if err != nil {
		return errors.Wrap(err, "failed to save feed to database")
	}
	return err
}

func (p Postgres) GetFeed(hashID string) (*model.Feed, error) {
	lastAccess := time.Now().UTC()

	feed := &model.Feed{}
	res, err := p.db.Model(feed).
		Set("last_access = ?", lastAccess).
		Where("hash_id = ?", hashID).
		Returning("*").
		Update()

	if err != nil {
		return nil, errors.Wrapf(err, "failed to query feed: %s", hashID)
	}

	if res.RowsAffected() != 1 {
		return nil, api.ErrNotFound
	}

	return feed, nil
}

func (p Postgres) GetMetadata(hashID string) (*model.Feed, error) {
	feed := &model.Feed{}
	err := p.db.
		Model(feed).
		Where("hash_id = ?", hashID).
		Column("provider", "format", "quality", "user_id").
		Select()

	if err != nil {
		return nil, err
	}

	return feed, nil
}

func (p Postgres) Downgrade(patronID string, featureLevel int) error {
	if featureLevel > api.ExtendedFeatures {
		return nil
	}

	if featureLevel == api.ExtendedFeatures {
		const maxPages = 150
		_, err := p.db.
			Model(&model.Feed{}).
			Set("page_size = ?", maxPages).
			Where("user_id = ? AND page_size > ?", patronID, maxPages).
			Update()

		if err != nil {
			return errors.Wrapf(err, "failed to reduce page sizes for patron '%s'", patronID)
		}

		_, err = p.db.
			Model(&model.Feed{}).
			Set("feature_level = ?", api.ExtendedFeatures).
			Where("user_id = ?", patronID, maxPages).
			Update()

		if err != nil {
			return errors.Wrapf(err, "failed to downgrade patron '%s' to feature level %d", patronID, featureLevel)
		}

		return nil
	}

	if featureLevel == api.DefaultFeatures {
		_, err := p.db.
			Model(&model.Feed{}).
			Set("page_size = ?", 50).
			Set("feature_level = ?", api.DefaultFeatures).
			Set("format = ?", api.FormatVideo).
			Set("quality = ?", api.QualityHigh).
			Where("user_id = ?", patronID).
			Update()

		if err != nil {
			return errors.Wrapf(err, "failed to downgrade patron '%s' to feature level %d", patronID, featureLevel)
		}

		return nil
	}

	return errors.New("unsupported downgrade type")
}

func (p Postgres) AddPledge(pledge *model.Pledge) error {
	return p.db.Insert(pledge)
}

func (p Postgres) UpdatePledge(patronID string, pledge *model.Pledge) error {
	updateColumns := []string{
		"declined_since",
		"amount_cents",
		"total_historical_amount_cents",
		"outstanding_payment_amount_cents",
		"is_paused",
	}

	res, err := p.db.Model(pledge).Column(updateColumns...).Where("patron_id = ?", patronID).Update()
	if err != nil {
		return errors.Wrapf(err, "failed to update pledge %d for user %s: %v", pledge.PledgeID, patronID, err)
	}

	if res.RowsAffected() != 1 {
		return errors.Wrapf(err, "unexpected number of updated rows: %d for user %s", res.RowsAffected(), patronID)
	}

	return nil
}

func (p Postgres) DeletePledge(pledge *model.Pledge) error {
	err := p.db.Delete(pledge)
	if err == pg.ErrNoRows {
		return nil
	}

	return err
}

func (p Postgres) GetPledge(patronID string) (*model.Pledge, error) {
	pledge := &model.Pledge{}
	err := p.db.Model(pledge).Where("patron_id = ?", patronID).Limit(1).Select()
	if err != nil {
		return nil, err
	}

	return pledge, nil
}

func (p Postgres) GetAllPledges() (list []*model.Pledge, err error) {
	err = p.db.Model(&list).Select()
	return
}

func (p Postgres) GetAllFeeds() (list []*model.Feed, err error) {
	err = p.db.Model(&list).Select()
	return
}

func (p Postgres) Close() error {
	return p.db.Close()
}
