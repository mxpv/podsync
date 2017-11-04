//go:generate mockgen -source=handler.go -destination=handler_mock_test.go -package=handler

package handler

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	itunes "github.com/mxpv/podcast"
	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/config"
	"github.com/stretchr/testify/require"
)

var cfg = &config.AppConfig{}

func TestCreateFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req := &api.CreateFeedRequest{
		URL:      "https://youtube.com/channel/123",
		PageSize: 55,
		Quality:  api.QualityLow,
		Format:   api.FormatAudio,
	}

	feed := NewMockfeedService(ctrl)
	feed.EXPECT().CreateFeed(gomock.Eq(req), gomock.Any()).Times(1).Return("456", nil)

	patreon := NewMockpatreonService(ctrl)
	patreon.EXPECT().GetFeatureLevel(gomock.Any()).Return(api.DefaultFeatures)

	srv := httptest.NewServer(New(feed, patreon, cfg))
	defer srv.Close()

	query := `{"url": "https://youtube.com/channel/123", "page_size": 55, "quality": "low", "format": "audio"}`
	resp, err := http.Post(srv.URL+"/api/create", "application/json", strings.NewReader(query))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.JSONEq(t, `{"id": "456"}`, readBody(t, resp))
}

func TestCreateInvalidFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	srv := httptest.NewServer(New(NewMockfeedService(ctrl), nil, cfg))
	defer srv.Close()

	query := `{}`
	resp, err := http.Post(srv.URL+"/api/create", "application/json", strings.NewReader(query))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	query = `{"url": "not a url", "page_size": 55, "quality": "low", "format": "audio"}`
	resp, err = http.Post(srv.URL+"/api/create", "application/json", strings.NewReader(query))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	query = `{"url": "https://youtube.com/channel/123", "page_size": 1, "quality": "low", "format": "audio"}`
	resp, err = http.Post(srv.URL+"/api/create", "application/json", strings.NewReader(query))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	query = `{"url": "https://youtube.com/channel/123", "page_size": 151, "quality": "low", "format": "audio"}`
	resp, err = http.Post(srv.URL+"/api/create", "application/json", strings.NewReader(query))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	query = `{"url": "https://youtube.com/channel/123", "page_size": 50, "quality": "xyz", "format": "audio"}`
	resp, err = http.Post(srv.URL+"/api/create", "application/json", strings.NewReader(query))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	query = `{"url": "https://youtube.com/channel/123", "page_size": 50, "quality": "low", "format": "xyz"}`
	resp, err = http.Post(srv.URL+"/api/create", "application/json", strings.NewReader(query))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	query = `{"url": "https://youtube.com/channel/123", "page_size": 50, "quality": "low", "format": ""}`
	resp, err = http.Post(srv.URL+"/api/create", "application/json", strings.NewReader(query))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	query = `{"url": "https://youtube.com/channel/123", "page_size": 50, "quality": "", "format": "audio"}`
	resp, err = http.Post(srv.URL+"/api/create", "application/json", strings.NewReader(query))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	podcast := itunes.New("", "", "", nil, nil)

	feed := NewMockfeedService(ctrl)
	feed.EXPECT().BuildFeed("123").Return(&podcast, nil)

	srv := httptest.NewServer(New(feed, nil, cfg))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/123")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	feed := NewMockfeedService(ctrl)
	feed.EXPECT().GetMetadata("123").Times(1).Return(&api.Metadata{}, nil)

	srv := httptest.NewServer(New(feed, nil, cfg))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/metadata/123")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func readBody(t *testing.T, resp *http.Response) string {
	buf, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	require.NoError(t, err)

	return string(buf)
}
