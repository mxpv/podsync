package ytdl

import (
	"fmt"
	"path/filepath"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/model"
)

type Options interface {
	GetConfig() []string
}

type OptionsDl struct{}

func (o OptionsDl) New(feedConfig *config.Feed, episode *model.Episode, feedPath string) []string {

	var (
		arguments []string
		options   Options
	)

	if feedConfig.Format == model.FormatVideo {
		options = NewOptionsVideo(feedConfig)
	} else {
		options = NewOptionsAudio(feedConfig)
	}

	arguments = options.GetConfig()
	arguments = append(arguments, "--output", o.makeOutputTemplate(feedPath, episode), episode.VideoURL)

	return arguments
}

func (o OptionsDl) makeOutputTemplate(feedPath string, episode *model.Episode) string {
	filename := fmt.Sprintf("%s.%s", episode.ID, "%(ext)s")
	return filepath.Join(feedPath, filename)
}
