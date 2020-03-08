package fs

import (
	"context"
	"io"
)

type Storage interface {
	// Create will create a new file from reader
	Create(ctx context.Context, ns string, fileName string, reader io.Reader) (int64, error)

	// Delete deletes the file
	Delete(ctx context.Context, ns string, fileName string) error

	// Size returns the size of a file in bytes
	Size(ctx context.Context, ns string, fileName string) (int64, error)

	// URL will generate a download link for a file
	URL(ctx context.Context, ns string, fileName string) (string, error)
}
