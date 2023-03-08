package update

import (
	"regexp"
	"time"

	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/model"
	log "github.com/sirupsen/logrus"
)

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

	if filters.MaxDuration > 0 && episode.Duration > filters.MaxDuration {
		logger.WithField("filter", "max_duration").Infof("skipping due to duration filter (%ds)", episode.Duration)
		return false
	}

	if filters.MinDuration > 0 && episode.Duration < filters.MinDuration {
		logger.WithField("filter", "min_duration").Infof("skipping due to duration filter (%ds)", episode.Duration)
		return false
	}

	if filters.MaxAge > 0 {
		dateDiff := int(time.Since(episode.PubDate).Hours()) / 24
		if dateDiff > filters.MaxAge {
			logger.WithField("filter", "max_age").Infof("skipping due to max_age filter (%dd > %dd)", dateDiff, filters.MaxAge)
			return false
		}
	}

	return true
}
