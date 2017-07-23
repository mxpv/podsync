package database

import (
	"github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/proxy"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/pkg/errors"
	"log"
	"net"
	"strings"
)

type PgConfig struct {
	ConnectionUrl string `yaml:"connectionUrl"`
}

type PgStorage struct {
	db *pg.DB
}

func (p *PgStorage) CreateFeed(feed *Feed) error {
	_, err := p.db.Model(feed).Insert()
	if err != nil {
		return errors.Wrap(err, "failed to create feed")
	}

	return nil
}

func (p *PgStorage) GetFeed(q ...WhereFunc) (out []Feed, err error) {
	out = []Feed{}
	err = p.db.Model(&out).Apply(whereFunc(q...)).Select()
	return
}

func whereFunc(where ...WhereFunc) func(*orm.Query) (*orm.Query, error) {
	return func(q *orm.Query) (*orm.Query, error) {
		for _, fn := range where {
			field, value := fn()
			q = q.Where(field+" = ?", value)
		}

		return q, nil
	}
}

func NewPgStorage(config *PgConfig) (*PgStorage, error) {
	opts, err := pg.ParseURL(config.ConnectionUrl)
	if err != nil {
		return nil, err
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
	if _, err := db.ExecOne("SELECT 1"); err != nil {
		db.Close()
		return nil, errors.Wrap(err, "failed to check database connectivity")
	}

	log.Print("running update script")
	if _, err := db.Exec(installScript); err != nil {
		return nil, errors.Wrap(err, "failed to upgrade database structure")
	}

	storage := &PgStorage{db: db}
	return storage, nil
}
