package builders

import (
	"fmt"

	itunes "github.com/mxpv/podcast"
	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
)

const (
	podsyncGenerator = "Podsync generator"
	defaultCategory  = "TV & Film"
)

func makeEnclosure(feed *model.Feed, id string, lengthInBytes int64) (string, itunes.EnclosureType, int64) {
	ext := "mp4"
	contentType := itunes.MP4
	if feed.Format == api.FormatAudio {
		ext = "m4a"
		contentType = itunes.M4A
	}

	url := fmt.Sprintf("http://podsync.net/download/%s/%s.%s", feed.HashID, id, ext)
	return url, contentType, lengthInBytes
}
