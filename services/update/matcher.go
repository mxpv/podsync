package update

import (
	"regexp"

	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/model"
	log "github.com/sirupsen/logrus"
)

func matchDurationFilter(durationLimit int64, episodeDuration int64, operator string, logger log.FieldLogger) bool {
	if durationLimit != 0 {
		if operator == "min" && episodeDuration < durationLimit {
			logger.Info("skipping due to duration filter")
			return false
		} else if operator == "max" && episodeDuration > durationLimit {
			logger.Info("skipping due to duration filter")
			return false
		}
	}
	return true
}

func matchRegexpFilter(pattern, str string, negative bool, logger log.FieldLogger) bool {
	if pattern != "" {
		matched, err := regexp.MatchString(pattern, str)
		if err != nil {
			logger.Warnf("pattern %q is not a valid")
		} else {
			if matched == negative {
				logger.Infof("skipping due to regexp mismatch")
				return false
			}
		}
	}
	return true
}

func matchFilters(episode *model.Episode, filters *feed.Filters) bool {
	logger := log.WithFields(log.Fields{"episode_id": episode.ID})
	if !matchRegexpFilter(filters.Title, episode.Title, false, logger.WithField("filter", "title")) {
		return false
	}

	if !matchRegexpFilter(filters.NotTitle, episode.Title, true, logger.WithField("filter", "not_title")) {
		return false
	}

	if !matchRegexpFilter(filters.Description, episode.Description, false, logger.WithField("filter", "description")) {
		return false
	}

	if !matchRegexpFilter(filters.NotDescription, episode.Description, true, logger.WithField("filter", "not_description")) {
		return false
	}

	if !matchDurationFilter(filters.MinDuration, episode.Duration, "min", logger.WithField("filter", "min_duration")) {
		return false
	}

	if !matchDurationFilter(filters.MaxDuration, episode.Duration, "max", logger.WithField("filter", "max_duration")) {
		return false
	}

	return true
}
