package storage

import (
	"context"
	"errors"

	"github.com/mxpv/podsync/pkg/model"
)

type Version int

const (
	CurrentVersion = 1
)

var (
	ErrAlreadyExists = errors.New("object already exists")
)

type Storage interface {
	Close() error
	Version() (int, error)

	// AddFeed will:
	// - Insert or update feed info
	// - Append new episodes to the existing list of episodes
	// - Insert File model for each new episode
	AddFeed(ctx context.Context, feed *model.Feed) error

	// WalkFeeds iterates over feeds saved to database
	WalkFeeds(ctx context.Context, cb func(feed *model.Feed) error) error

	// DeleteFeed deletes feed and all related data from database
	DeleteFeed(ctx context.Context, feedID string) error

	// WalkFiles walks all files for the given feed ID
	WalkFiles(ctx context.Context, feedID string, cb func(file *model.File) error) error

	// UpdateFile updates file's status and (optionally) size.
	// Callback can be used to rollback update transaction.
	UpdateFile(file *model.File, cb func() error) error
}
