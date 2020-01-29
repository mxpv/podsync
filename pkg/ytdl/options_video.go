package ytdl

import (
	"fmt"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/model"
)

type OptionsVideo struct {
	quality   model.Quality
	maxHeight int
}

func NewOptionsVideo(feedConfig *config.Feed) *OptionsVideo {
	options := &OptionsVideo{}

	options.quality = feedConfig.Quality
	options.maxHeight = feedConfig.MaxHeight

	return options
}

func (options OptionsVideo) GetConfig() []string {
	var (
		arguments []string
		format    string
	)

	switch options.quality {
	// I think after enabling MaxHeight param QualityLow option don't need.
	// If somebody want download video in low quality then can set MaxHeight to 360p
	// ¯\_(ツ)_/¯
	case model.QualityLow:
		format = "worstvideo[ext=mp4]+worstaudio[ext=m4a]/worst[ext=mp4]/worst"
	default:
		format = "bestvideo%s[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best"

		if options.maxHeight > 0 {
			format = fmt.Sprintf(format, fmt.Sprintf("[height<=%d]", options.maxHeight))
		} else {
			// unset replace pattern
			format = fmt.Sprintf(format, "")
		}
	}

	arguments = append(arguments, "--format", format)

	return arguments
}
