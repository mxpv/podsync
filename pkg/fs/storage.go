package fs

import (
	"context"
	"io"
	"net/http"
)

// Storage is a file system interface to host downloaded episodes and feeds.
type Storage interface {
	// FileSystem must be implemented to in order to pass Storage interface to HTTP file server.
	http.FileSystem

	// Create will create a new file from reader
	Create(ctx context.Context, name string, reader io.Reader) (int64, error)

	// Delete deletes the file
	Delete(ctx context.Context, name string) error

	// Size returns a storage object's size in bytes
	Size(ctx context.Context, name string) (int64, error)
}
