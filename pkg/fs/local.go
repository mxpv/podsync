package fs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Local struct {
	hostname string
	rootDir  string
}

func NewLocal(rootDir string, hostname string) (*Local, error) {
	if hostname == "" {
		return nil, errors.New("hostname can't be empty")
	}

	hostname = strings.TrimSuffix(hostname, "/")
	if !strings.HasPrefix(hostname, "http") {
		hostname = fmt.Sprintf("http://%s", hostname)
	}

	return &Local{rootDir: rootDir, hostname: hostname}, nil
}

func (l *Local) Create(ctx context.Context, ns string, fileName string, reader io.Reader) (int64, error) {
	var (
		logger  = log.WithField("episode_id", fileName)
		feedDir = filepath.Join(l.rootDir, ns)
	)

	if err := os.MkdirAll(feedDir, 0755); err != nil {
		return 0, errors.Wrapf(err, "failed to create a directory for the feed: %s", feedDir)
	}

	logger.Debugf("creating directory: %s", feedDir)
	if err := os.MkdirAll(feedDir, 0755); err != nil {
		return 0, errors.Wrapf(err, "failed to create feed dir: %s", feedDir)
	}

	var (
		episodePath = filepath.Join(l.rootDir, ns, fileName)
	)

	logger.Debugf("copying to: %s", episodePath)
	written, err := l.copyFile(reader, episodePath)
	if err != nil {
		return 0, errors.Wrap(err, "failed to copy file")
	}

	logger.Debugf("copied %d bytes", written)
	return written, nil
}

func (l *Local) Delete(ctx context.Context, ns string, fileName string) error {
	path := filepath.Join(l.rootDir, ns, fileName)
	return os.Remove(path)
}

func (l *Local) Size(ctx context.Context, ns string, fileName string) (int64, error) {
	path := filepath.Join(l.rootDir, ns, fileName)

	stat, err := os.Stat(path)
	if err == nil {
		return stat.Size(), nil
	}

	return 0, err
}

func (l *Local) URL(ctx context.Context, ns string, fileName string) (string, error) {
	if _, err := l.Size(ctx, ns, fileName); err != nil {
		return "", errors.Wrap(err, "failed to check whether file exists")
	}

	if ns == "" {
		return fmt.Sprintf("%s/%s", l.hostname, fileName), nil
	}

	return fmt.Sprintf("%s/%s/%s", l.hostname, ns, fileName), nil
}

func (l *Local) copyFile(source io.Reader, destinationPath string) (int64, error) {
	dest, err := os.Create(destinationPath)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create destination file")
	}

	defer dest.Close()

	written, err := io.Copy(dest, source)
	if err != nil {
		return 0, errors.Wrap(err, "failed to copy data")
	}

	return written, nil
}
