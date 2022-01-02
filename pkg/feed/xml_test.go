package feed

import (
	"context"
	"testing"

	itunes "github.com/eduncan911/podcast"
	"github.com/mxpv/podsync/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildXML(t *testing.T) {
	feed := model.Feed{
		Episodes: []*model.Episode{
			{
				ID:          "1",
				Status:      model.EpisodeDownloaded,
				Title:       "title",
				Description: "description",
			},
		},
	}

	cfg := Config{
		ID:     "test",
		Custom: Custom{Description: "description", Category: "Technology", Subcategories: []string{"Gadgets", "Podcasting"}},
	}

	out, err := Build(context.Background(), &feed, &cfg, "http://localhost/")
	assert.NoError(t, err)

	assert.EqualValues(t, "description", out.Description)
	assert.EqualValues(t, "Technology", out.Category)

	require.Len(t, out.ICategories, 1)
	category := out.ICategories[0]
	assert.EqualValues(t, "Technology", category.Text)

	require.Len(t, category.ICategories, 2)
	assert.EqualValues(t, "Gadgets", category.ICategories[0].Text)
	assert.EqualValues(t, "Podcasting", category.ICategories[1].Text)

	require.Len(t, out.Items, 1)
	require.NotNil(t, out.Items[0].Enclosure)
	assert.EqualValues(t, out.Items[0].Enclosure.URL, "http://localhost/test/1.mp4")
	assert.EqualValues(t, out.Items[0].Enclosure.Type, itunes.MP4)
}
