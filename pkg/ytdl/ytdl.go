package ytdl

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/model"
)

const DownloadTimeout = 10 * time.Minute

var (
	ErrTooManyRequests = errors.New(http.StatusText(http.StatusTooManyRequests))
)

type YoutubeDl struct {
	path string
}

type tempFile struct {
	*os.File
	dir string
}

func New(ctx context.Context) (*YoutubeDl, error) {
	path, err := exec.LookPath("youtube-dl")
	if err != nil {
		return nil, errors.Wrap(err, "youtube-dl binary not found")
	}

	log.Debugf("found youtube-dl binary at %q", path)

	ytdl := &YoutubeDl{
		path: path,
	}

	// Make sure youtube-dl exists
	version, err := ytdl.exec(ctx, "--version")
	if err != nil {
		return nil, errors.Wrap(err, "could not find youtube-dl")
	}

	log.Infof("using youtube-dl %s", version)

	if err := ytdl.ensureDependencies(ctx); err != nil {
		return nil, err
	}

	return ytdl, nil
}

func (dl YoutubeDl) ensureDependencies(ctx context.Context) error {
	found := false

	if path, err := exec.LookPath("ffmpeg"); err == nil {
		found = true

		output, err := exec.CommandContext(ctx, path, "-version").CombinedOutput()
		if err != nil {
			return errors.Wrap(err, "could not get ffmpeg version")
		}

		log.Infof("found ffmpeg: %s", output)
	}

	if path, err := exec.LookPath("avconv"); err == nil {
		found = true

		output, err := exec.CommandContext(ctx, path, "-version").CombinedOutput()
		if err != nil {
			return errors.Wrap(err, "could not get avconv version")
		}

		log.Infof("found avconv: %s", output)
	}

	if !found {
		return errors.New("either ffmpeg or avconv required to run Podsync")
	}

	return nil
}

func (dl YoutubeDl) Download(ctx context.Context, feedConfig *config.Feed, episode *model.Episode) (r io.ReadCloser, err error) {
	tmpDir, err := ioutil.TempDir("", "podsync-")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get temp dir for download")
	}

	defer func() {
		if err != nil {
			err1 := os.RemoveAll(tmpDir)
			if err1 != nil {
				log.Errorf("could not remove temp dir: %v", err1)
			}
		}
	}()

	// filePath with YoutubeDl template format
	filePath := filepath.Join(tmpDir, fmt.Sprintf("%s.%s", episode.ID, "%(ext)s"))

	args := buildArgs(feedConfig, episode, filePath)
	output, err := dl.exec(ctx, args...)
	if err != nil {
		log.WithError(err).Errorf("youtube-dl error: %s", filePath)

		// YouTube might block host with HTTP Error 429: Too Many Requests
		if strings.Contains(output, "HTTP Error 429") {
			return nil, ErrTooManyRequests
		}

		log.Error(output)

		return nil, errors.New(output)
	}

	ext := "mp4"
	if feedConfig.Format == model.FormatAudio {
		ext = "mp3"
	}
	// filePath now with the final extension
	filePath = filepath.Join(tmpDir, fmt.Sprintf("%s.%s", episode.ID, ext))
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open downloaded file")
	}

	return &tempFile{File: f, dir: tmpDir}, nil
}

// SelfUpdate of youtube-dl
func (dl YoutubeDl) SelfUpdate(ctx context.Context) (string, error) {
	output, err := dl.exec(ctx, "--update", "--verbose")

	log.Debugf("youtube-dl --update --verbose\n%s", output)

	if err != nil {
		return output, errors.Wrap(err, "could not update youtube-dl")
	}
	return output, nil
}

func (dl YoutubeDl) exec(ctx context.Context, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, DownloadTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, dl.path, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), errors.Wrap(err, "failed to execute youtube-dl")
	}

	return string(output), nil
}

func buildArgs(feedConfig *config.Feed, episode *model.Episode, outputFilePath string) []string {
	var args []string

	if feedConfig.Format == model.FormatVideo {
		// Video, mp4, high by default

		format := "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best"

		if feedConfig.Quality == model.QualityLow {
			format = "worstvideo[ext=mp4]+worstaudio[ext=m4a]/worst[ext=mp4]/worst"
		} else if feedConfig.Quality == model.QualityHigh && feedConfig.MaxHeight > 0 {
			format = fmt.Sprintf("bestvideo[height<=%d][ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best", feedConfig.MaxHeight)
		}

		args = append(args, "--format", format)
	} else {
		// Audio, mp3, high by default
		format := "bestaudio"
		if feedConfig.Quality == model.QualityLow {
			format = "worstaudio"
		}

		args = append(args, "--extract-audio", "--audio-format", "mp3", "--format", format)
	}

	args = append(args, "--output", outputFilePath, episode.VideoURL)
	return args
}

func (f *tempFile) Close() error {
	err := f.File.Close()
	err1 := os.RemoveAll(f.dir)
	if err1 != nil {
		log.Errorf("could not remove temp dir: %v", err1)
	}
	return err
}
