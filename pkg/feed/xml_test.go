package feed

import (
	"context"
	"testing"
	"time"

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

func TestBuildXMLWithFilenameTemplate(t *testing.T) {
	feed := model.Feed{
		Episodes: []*model.Episode{
			{
				ID:          "video123",
				Status:      model.EpisodeDownloaded,
				Title:       "A title / with chars",
				Description: "description",
				PubDate:     time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	cfg := Config{
		ID:               "test",
		Format:           model.FormatVideo,
		FilenameTemplate: "{{pub_date}}_{{title}}_{{id}}",
	}

	out, err := Build(context.Background(), &feed, &cfg, "http://localhost/")
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	require.NotNil(t, out.Items[0].Enclosure)
	assert.Equal(t, "video123", out.Items[0].GUID)
	assert.Equal(t, "http://localhost/test/2025-12-31_A_title_with_chars_video123.mp4", out.Items[0].Enclosure.URL)
}

func TestEpisodeNameTemplate(t *testing.T) {
	cfg := &Config{
		ID:               "test",
		Format:           model.FormatVideo,
		FilenameTemplate: "{{pub_date}}_{{title}}_{{id}}",
	}

	episode := &model.Episode{
		ID:      "abc123",
		Title:   "My / Video: Title?",
		PubDate: time.Date(2026, 2, 8, 10, 0, 0, 0, time.UTC),
	}

	assert.Equal(t, "2026-02-08_My_Video_Title_abc123.mp4", EpisodeName(cfg, episode))
}

func TestValidateFilenameTemplate(t *testing.T) {
	assert.NoError(t, ValidateFilenameTemplate(""))
	assert.NoError(t, ValidateFilenameTemplate("{{pub_date}}_{{title}}_{{id}}"))
	assert.Error(t, ValidateFilenameTemplate("{{unknown}}_{{id}}"))
}
