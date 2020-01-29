package ytdl

import (
	"context"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/model"
)

const DownloadTimeout = 10 * time.Minute

type YoutubeDl struct{}

func New(ctx context.Context) (*YoutubeDl, error) {
	ytdl := &YoutubeDl{}

	// Make sure youtube-dl exists
	version, err := ytdl.exec(ctx, "--version")
	if err != nil {
		return nil, errors.Wrap(err, "could not find youtube-dl")
	}

	log.Infof("using youtube-dl %s", version)

	// Make sure ffmpeg exists
	output, err := exec.CommandContext(ctx, "ffmpeg", "-version").CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, "could not find ffmpeg")
	}

	log.Infof("using ffmpeg %s", output)

	return ytdl, nil
}

func (dl YoutubeDl) Download(ctx context.Context, feedConfig *config.Feed, episode *model.Episode, feedPath string) (string, error) {
	options := &OptionsDl{}

	params := options.New(feedConfig, episode, feedPath)

	return dl.exec(ctx, params...)

}

func (YoutubeDl) exec(ctx context.Context, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, DownloadTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "youtube-dl", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), errors.Wrap(err, "failed to execute youtube-dl")
	}

	return string(output), nil
}
