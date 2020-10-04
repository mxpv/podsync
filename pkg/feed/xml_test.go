package feed

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/model"
)

func TestBuildXML(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlMock := NewMockurlProvider(ctrl)

	feed := model.Feed{}

	cfg := config.Feed{
		Custom: config.Custom{Description: "description", Category: "Technology", Subcategories: []string{"Gadgets", "Podcasting"}},
	}

	out, err := Build(context.Background(), &feed, &cfg, urlMock)
	assert.NoError(t, err)

	assert.EqualValues(t, "description", out.Description)
	assert.EqualValues(t, "Technology", out.Category)

	require.Len(t, out.ICategories, 1)
	category := out.ICategories[0]
	assert.EqualValues(t, "Technology", category.Text)
	require.Len(t, category.ICategories, 2)
	assert.EqualValues(t, "Gadgets", category.ICategories[0].Text)
	assert.EqualValues(t, "Podcasting", category.ICategories[1].Text)
}
