package update

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/model"
)

func TestMultipleKeywordsInNotTitleFilterIssue158(t *testing.T) {
	// https://github.com/mxpv/podsync/issues/158
	// Use regex alternation (|) to exclude episodes matching any of multiple keywords.
	// (?i) makes the match case-insensitive.
	filters := &feed.Filters{
		NotTitle: "(?i)(live|q&a)",
	}

	// Episodes with any of the excluded keywords should be skipped
	assert.False(t, matchFilters(&model.Episode{ID: "1", Title: "Live Stream Q&A Session"}, filters))
	assert.False(t, matchFilters(&model.Episode{ID: "2", Title: "Weekly Q&A with the Team"}, filters))
	assert.False(t, matchFilters(&model.Episode{ID: "3", Title: "LIVE: Year in Review"}, filters))
	assert.False(t, matchFilters(&model.Episode{ID: "4", Title: "live recording from the studio"}, filters))

	// Episodes without any excluded keyword should be included
	assert.True(t, matchFilters(&model.Episode{ID: "5", Title: "Episode 42: Deep Dive into Go"}, filters))
	assert.True(t, matchFilters(&model.Episode{ID: "6", Title: "Interview with a Go developer"}, filters))
}

func TestTitleFilterWithMultipleKeywordsIssue158(t *testing.T) {
	// https://github.com/mxpv/podsync/issues/158
	// Use regex alternation (|) to include only episodes matching at least one keyword.
	filters := &feed.Filters{
		Title: "(?i)(tutorial|how.to|guide)",
	}

	// Episodes matching at least one keyword should be included
	assert.True(t, matchFilters(&model.Episode{ID: "1", Title: "Go Tutorial for Beginners"}, filters))
	assert.True(t, matchFilters(&model.Episode{ID: "2", Title: "How to Write Unit Tests"}, filters))
	assert.True(t, matchFilters(&model.Episode{ID: "3", Title: "A Complete Guide to Docker"}, filters))

	// Episodes not matching any keyword should be skipped
	assert.False(t, matchFilters(&model.Episode{ID: "4", Title: "Weekly News Roundup"}, filters))
	assert.False(t, matchFilters(&model.Episode{ID: "5", Title: "Live Q&A with Experts"}, filters))
}

func TestCombinedTitleAndDurationFiltersIssue158(t *testing.T) {
	// https://github.com/mxpv/podsync/issues/158
	// All filters are combined with AND logic: an episode must satisfy every filter to be included.
	// To include episodes matching a title OR a minimum duration, combine both conditions into
	// a single title filter using regex, or rely on not_title to exclude what you don't want.
	filters := &feed.Filters{
		NotTitle:    "(?i)(short clip|preview|trailer)",
		MinDuration: 600, // 10 minutes
	}

	// Long episodes without excluded keywords should be included
	assert.True(t, matchFilters(&model.Episode{ID: "1", Title: "Full Episode: Go Concurrency", Duration: 3600}, filters))
	assert.True(t, matchFilters(&model.Episode{ID: "2", Title: "Interview: Building APIs", Duration: 1200}, filters))

	// Short episodes (below min_duration) should be excluded regardless of title
	assert.False(t, matchFilters(&model.Episode{ID: "3", Title: "Full Episode: Quick Tips", Duration: 300}, filters))

	// Long episodes with an excluded keyword should be excluded
	assert.False(t, matchFilters(&model.Episode{ID: "4", Title: "Preview: Next Week's Episode", Duration: 3600}, filters))

	// Short episodes with an excluded keyword should be excluded
	assert.False(t, matchFilters(&model.Episode{ID: "5", Title: "Short Clip from the Show", Duration: 120}, filters))
}

func TestNotTitleFilterIssue798(t *testing.T) {
	// https://github.com/mxpv/podsync/issues/798
	filters := &feed.Filters{
		NotTitle:    "(?i)^(holy mass|holy sacrifice|the holy)( |$)",
		MinDuration: 600,
	}

	// Titles starting with pattern should be excluded
	assert.False(t, matchFilters(&model.Episode{ID: "1", Title: "Holy Mass — Tuesday", Duration: 3600}, filters))
	assert.False(t, matchFilters(&model.Episode{ID: "2", Title: "The Holy Sacrifice of the Mass", Duration: 3600}, filters))
	assert.False(t, matchFilters(&model.Episode{ID: "3", Title: "The Holy Mass (Latin)", Duration: 3600}, filters))

	// Titles NOT starting with pattern should be included
	assert.True(t, matchFilters(&model.Episode{ID: "4", Title: "Homily: The Parable of the Good Samaritan", Duration: 1200}, filters))
	assert.True(t, matchFilters(&model.Episode{ID: "5", Title: "Sermon — Love Your Enemies", Duration: 1800}, filters))
	assert.True(t, matchFilters(&model.Episode{ID: "6", Title: "Reflection on Today's Gospel", Duration: 900}, filters))
}
