package feed

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/model"
)

func TestBuildOPML(t *testing.T) {
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
	<head>
		<title>Podsync feeds</title>
	</head>
	<body>
		<outline text="desc" type="rss" xmlUrl="https://url/1.xml" title="1"></outline>
	</body>
</opml>`

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlMock := NewMockurlProvider(ctrl)
	urlMock.EXPECT().URL(gomock.Any(), "", "1.xml").Return("https://url/1.xml", nil)

	dbMock := NewMockfeedProvider(ctrl)
	dbMock.EXPECT().GetFeed(gomock.Any(), "1").Return(&model.Feed{Title: "1", Description: "desc"}, nil)

	cfg := config.Config{
		Feeds: map[string]*config.Feed{
			"any": {ID: "1", OPML: true},
		},
	}

	out, err := BuildOPML(context.Background(), &cfg, dbMock, urlMock)
	assert.NoError(t, err)
	assert.Equal(t, expected, out)
}
