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

func (dl YoutubeDl) Download(ctx context.Context, feedConfig *config.Feed, url string, destPath string) (string, error) {
	if feedConfig.Format == model.FormatAudio {
		// Audio
		if feedConfig.Quality == model.QualityHigh {
			// High quality audio (encoded to mp3)
			return dl.exec(ctx,
				"--extract-audio",
				"--audio-format",
				"mp3",
				"--format",
				"bestaudio",
				"--output",
				destPath,
				url,
			)
		} else { //nolint
			// Low quality audio (encoded to mp3)
			return dl.exec(ctx,
				"--extract-audio",
				"--audio-format",
				"mp3",
				"--format",
				"worstaudio",
				"--output",
				destPath,
				url,
			)
		}
	} else {
		/*
			Video
		*/
		if feedConfig.Quality == model.QualityHigh {
			// High quality
			return dl.exec(ctx,
				"--format",
				"bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
				"--output",
				destPath,
				url,
			)
		} else { //nolint
			// Low quality
			return dl.exec(ctx,
				"--format",
				"worstvideo[ext=mp4]+worstaudio[ext=m4a]/worst[ext=mp4]/worst",
				"--output",
				destPath,
				url,
			)
		}
	}
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
