package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/model"
)

const (
	versionPath   = "podsync/version"
	feedPrefix    = "feed/"
	feedPath      = "feed/%s"
	episodePrefix = "episode/%s/"
	episodePath   = "episode/%s/%s" // FeedID + EpisodeID
	filePrefix    = "file/%s/"
	filePath      = "file/%s/%s" // FeedID + EpisodeID
)

type Badger struct {
	db *badger.DB
}

var _ Storage = (*Badger)(nil)

func NewBadger(config *config.Database) (*Badger, error) {
	var (
		dir = config.Dir
	)

	log.Infof("opening database %q", dir)

	// Make sure database directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, errors.Wrap(err, "could not mkdir database dir")
	}

	opts := badger.DefaultOptions(dir)
	opts.Logger = log.New()

	db, err := badger.Open(opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	storage := &Badger{db: db}

	if err := db.Update(func(txn *badger.Txn) error {
		return storage.setObj(txn, []byte(versionPath), CurrentVersion, false)
	}); err != nil {
		return nil, errors.Wrap(err, "failed to read database version")
	}

	return &Badger{db: db}, nil
}

func (b *Badger) Close() error {
	return b.db.Close()
}

func (b *Badger) Version() (int, error) {
	var (
		version = -1
	)

	err := b.db.View(func(txn *badger.Txn) error {
		return b.getObj(txn, []byte(versionPath), &version)
	})

	return version, err
}

func (b *Badger) AddFeed(_ context.Context, feed *model.Feed) error {
	return b.db.Update(func(txn *badger.Txn) error {
		// Insert or update feed info
		feedKey := b.getKey(feedPath, feed.ID)
		if err := b.setObj(txn, feedKey, feed, true); err != nil {
			return err
		}

		// Append new episodes
		for _, episode := range feed.Episodes {
			episodeKey := b.getKey(episodePath, feed.ID, episode.ID)
			err := b.setObj(txn, episodeKey, episode, false)
			if err == nil || err == ErrAlreadyExists {
				// Do nothing
			} else {
				return errors.Wrapf(err, "failed to save episode %q", feed.ID)
			}
		}

		// Update download file statuses
		for _, episode := range feed.Episodes {
			fileKey := b.getKey(filePath, feed.ID, episode.ID)
			file := &model.File{
				EpisodeID: episode.ID,
				FeedID:    feed.ID,
				Size:      episode.Size, // Use estimated file size
				Status:    model.EpisodeNew,
			}

			err := b.setObj(txn, fileKey, file, false)
			if err != nil && err != ErrAlreadyExists {
				return errors.Wrapf(err, "failed to set %q status for %q", model.EpisodeNew, episode.ID)
			}
		}

		return nil
	})
}

func (b *Badger) GetFeed(_ context.Context, feedID string) (*model.Feed, error) {
	var (
		feed    = model.Feed{}
		feedKey = b.getKey(feedPath, feedID)
	)

	if err := b.db.View(func(txn *badger.Txn) error {
		// Query feed
		if err := b.getObj(txn, feedKey, &feed); err != nil {
			return err
		}

		// Query episodes
		opts := badger.DefaultIteratorOptions
		opts.Prefix = b.getKey(episodePrefix, feedID)
		opts.PrefetchValues = true
		if err := b.iterator(txn, opts, func(item *badger.Item) error {
			episode := &model.Episode{}
			if err := b.getObj(txn, item.Key(), &episode); err != nil {
				return err
			}

			feed.Episodes = append(feed.Episodes, episode)
			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &feed, nil
}

func (b *Badger) WalkFeeds(_ context.Context, cb func(feed *model.Feed) error) error {
	return b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = b.getKey(feedPrefix)
		opts.PrefetchValues = true
		return b.iterator(txn, opts, func(item *badger.Item) error {
			feed := &model.Feed{}
			if err := b.unmarshalObj(item, feed); err != nil {
				return err
			}

			return cb(feed)
		})
	})
}

func (b *Badger) DeleteFeed(_ context.Context, feedID string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		// Feed
		feedKey := b.getKey(feedPath, feedID)
		if err := txn.Delete(feedKey); err != nil {
			return errors.Wrapf(err, "failed to delete feed %q", feedID)
		}

		// Episodes
		opts := badger.DefaultIteratorOptions
		opts.Prefix = b.getKey(episodePrefix, feedID)
		opts.PrefetchValues = false
		if err := b.iterator(txn, opts, func(item *badger.Item) error {
			return txn.Delete(item.KeyCopy(nil))
		}); err != nil {
			return errors.Wrapf(err, "failed to iterate episodes for feed %q", feedID)
		}

		// Files
		opts = badger.DefaultIteratorOptions
		opts.Prefix = b.getKey(filePrefix, feedID)
		opts.PrefetchValues = false
		if err := b.iterator(txn, opts, func(item *badger.Item) error {
			return txn.Delete(item.KeyCopy(nil))
		}); err != nil {
			return errors.Wrapf(err, "failed to iterate files for feed %q", feedID)
		}

		return nil
	})
}

func (b *Badger) GetEpisode(_ context.Context, feedID string, episodeID string) (*model.Episode, error) {
	var (
		episode model.Episode
		err     error
		key     = b.getKey(episodePath, feedID, episodeID)
	)

	err = b.db.View(func(txn *badger.Txn) error {
		return b.getObj(txn, key, &episode)
	})

	return &episode, err
}

func (b *Badger) WalkFiles(_ context.Context, feedID string, cb func(file *model.File) error) error {
	opts := badger.DefaultIteratorOptions
	opts.Prefix = b.getKey(filePrefix, feedID)
	opts.PrefetchValues = true

	return b.db.View(func(txn *badger.Txn) error {
		return b.iterator(txn, opts, func(item *badger.Item) error {
			file := &model.File{}
			if err := b.unmarshalObj(item, file); err != nil {
				return err
			}

			return cb(file)
		})
	})
}

func (b *Badger) UpdateFile(feedID string, episodeID string, cb func(file *model.File) error) error {
	var (
		key  = b.getKey(filePath, feedID, episodeID)
		file = &model.File{}
	)

	return b.db.Update(func(txn *badger.Txn) error {
		if err := b.getObj(txn, key, file); err != nil {
			return err
		}

		if err := cb(file); err != nil {
			return err
		}

		if file.FeedID != feedID {
			return errors.New("can't change feed ID")
		}

		if file.EpisodeID != episodeID {
			return errors.New("can't change episode ID")
		}

		return b.setObj(txn, key, file, true)
	})
}

func (b *Badger) iterator(txn *badger.Txn, opts badger.IteratorOptions, callback func(item *badger.Item) error) error {
	iter := txn.NewIterator(opts)
	defer iter.Close()

	for iter.Rewind(); iter.Valid(); iter.Next() {
		item := iter.Item()

		if err := callback(item); err != nil {
			return err
		}
	}

	return nil
}

func (b *Badger) getKey(format string, a ...interface{}) []byte {
	resourcePath := fmt.Sprintf(format, a...)
	fullPath := fmt.Sprintf("podsync/v%d/%s", CurrentVersion, resourcePath)

	return []byte(fullPath)
}

func (b *Badger) setObj(txn *badger.Txn, key []byte, obj interface{}, overwrite bool) error {
	if !overwrite {
		// Overwrites are not allowed, make sure there is no object with the given key
		_, err := txn.Get(key)
		if err == nil {
			return ErrAlreadyExists
		} else if err == badger.ErrKeyNotFound {
			// Key not found, do nothing
		} else {
			return errors.Wrap(err, "failed to check whether key exists")
		}
	}

	data, err := b.marshalObj(obj)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize object for key %q", key)
	}

	return txn.Set(key, data)
}

func (b *Badger) getObj(txn *badger.Txn, key []byte, out interface{}) error {
	item, err := txn.Get(key)
	if err != nil {
		return err
	}

	return b.unmarshalObj(item, out)
}

func (b *Badger) marshalObj(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

func (b *Badger) unmarshalObj(item *badger.Item, out interface{}) error {
	return item.Value(func(val []byte) error {
		return json.Unmarshal(val, out)
	})
}
