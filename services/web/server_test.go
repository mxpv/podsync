package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockFileSystem struct{}

func (m *mockFileSystem) Open(name string) (http.File, error) {
	return nil, http.ErrMissingFile
}

func TestDebugEndpointDisabledByDefault(t *testing.T) {
	cfg := Config{
		Port: 8080,
		Path: "feeds",
	}

	srv := New(cfg, &mockFileSystem{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/debug/vars", nil)
	rec := httptest.NewRecorder()

	srv.Handler.ServeHTTP(rec, req)

	// Should return 404 when debug endpoints are disabled
	assert.Equal(t, http.StatusNotFound, rec.Code)
	// Should NOT contain expvar data
	assert.False(t, strings.Contains(rec.Body.String(), "cmdline"))
}

func TestDebugEndpointEnabledWhenConfigured(t *testing.T) {
	cfg := Config{
		Port:           8080,
		Path:           "feeds",
		DebugEndpoints: true,
	}

	srv := New(cfg, &mockFileSystem{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/debug/vars", nil)
	rec := httptest.NewRecorder()

	srv.Handler.ServeHTTP(rec, req)

	// Should return 200 and JSON content when debug endpoints are enabled
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	// Verify it contains expvar data (cmdline is always present)
	assert.True(t, strings.Contains(rec.Body.String(), "cmdline"))
}
