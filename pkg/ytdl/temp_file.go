package ytdl

import (
	"os"

	log "github.com/sirupsen/logrus"
)

type tempFile struct {
	*os.File
	dir string
}

func (f *tempFile) Close() error {
	err := f.File.Close()
	err1 := os.RemoveAll(f.dir)
	if err1 != nil {
		log.Errorf("could not remove temp dir: %v", err1)
	}
	return err
}
