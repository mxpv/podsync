package update

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/model"
)

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
