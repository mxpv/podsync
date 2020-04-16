package db

import (
	"context"

	"github.com/mxpv/podsync/pkg/model"
)

type Version int

const (
	CurrentVersion = 1
)

type Storage interface {
	Close() error
	Version() (int, error)

	// AddFeed will:
	// - Insert or update feed info
	// - Append new episodes to the existing list of episodes (existing episodes are not overwritten!)
	AddFeed(ctx context.Context, feedID string, feed *model.Feed) error

	// GetFeed gets a feed by ID
	GetFeed(ctx context.Context, feedID string) (*model.Feed, error)

	// WalkFeeds iterates over feeds saved to database
	WalkFeeds(ctx context.Context, cb func(feed *model.Feed) error) error

	// DeleteFeed deletes feed and all related data from database
	DeleteFeed(ctx context.Context, feedID string) error

	// GetEpisode gets episode by identifier
	GetEpisode(ctx context.Context, feedID string, episodeID string) (*model.Episode, error)

	// UpdateEpisode updates episode fields
	UpdateEpisode(feedID string, episodeID string, cb func(episode *model.Episode) error) error

	// WalkEpisodes iterates over episodes that belong to the given feed ID
	WalkEpisodes(ctx context.Context, feedID string, cb func(episode *model.Episode) error) error
}
