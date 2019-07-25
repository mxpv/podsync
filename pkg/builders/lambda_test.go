package builders

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mxpv/podsync/pkg/model"
)

func TestLambda_Invoke(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping lambda test in short mode")
	}

	lambda, err := NewLambda()
	assert.NoError(t, err)

	feed := &model.Feed{
		ItemURL:  "https://youtube.com/channel/UCupvZG-5ko_eiXAupbDfxWw",
		PageSize: 2,
		Format:   "video",
		Quality:  "high",
		Episodes: []*model.Item{
			{ID: "Test"},
		},
	}

	err = lambda.Build(feed)
	assert.NoError(t, err)

	assert.Len(t, feed.Episodes, 3)
	assert.Equal(t, "Test", feed.Episodes[2].ID)
	assert.NotEmpty(t, feed.LastID)
}
