package builders

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mxpv/podsync/pkg/model"
)

func TestRemote_makeURL(t *testing.T) {
	feed := &model.Feed{
		ItemURL:  "https://youtube.com/channel/UCupvZG-5ko_eiXAupbDfxWw",
		PageSize: 2,
		Format:   "video",
		Quality:  "high",
	}

	out, err := Remote{url: "http://updater:8080/update"}.makeURL(feed)
	assert.NoError(t, err)
	assert.EqualValues(t, "http://updater:8080/update?count=2&format=video&last_id=&quality=high&start=1&url=https%3A%2F%2Fyoutube.com%2Fchannel%2FUCupvZG-5ko_eiXAupbDfxWw", out)
}
