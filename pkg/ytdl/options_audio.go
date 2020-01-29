package ytdl

import (
	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/model"
)

type OptionsAudio struct {
	quality model.Quality
}

func NewOptionsAudio(feedConfig *config.Feed) *OptionsAudio {
	options := &OptionsAudio{}
	options.quality = feedConfig.Quality

	return options
}

func (options OptionsAudio) GetConfig() []string {
	var arguments []string

	arguments = append(arguments, "--extract-audio", "--audio-format", "mp3")

	switch options.quality {
	case model.QualityLow:
		// really? somebody use it?
		arguments = append(arguments, "--format", "worstaudio")
	default:
		arguments = append(arguments, "--format", "bestaudio")
	}

	return arguments
}
