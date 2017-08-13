package builders

import (
	"fmt"

	itunes "github.com/mxpv/podcast"
	"github.com/mxpv/podsync/web/pkg/api"
)

const (
	podsyncGenerator = "Podsync generator"
	defaultCategory  = "TV & Film"
)

func makeEnclosure(feed *api.Feed, id string, lengthInBytes int64) (string, itunes.EnclosureType, int64) {
	ext := "mp4"
	contentType := itunes.MP4
	if feed.Format == api.AudioFormat {
		ext = "mp3"
		contentType = itunes.MP3
	}

	url := fmt.Sprintf("http://podsync.net/download/%s/%s.%s", feed.HashId, id, ext)
	return url, contentType, lengthInBytes
}
