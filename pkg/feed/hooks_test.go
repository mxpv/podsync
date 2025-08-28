package feed

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteHook_WriteEnvToFile(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "env_output.txt")

	hook := &ExecHook{
		Command: []string{"sh", "-c", "printenv | grep '^TEST_VAR=' > " + tempFile},
		Timeout: 5,
	}

	env := []string{
		"TEST_VAR=test-value",
	}

	err := hook.Invoke(env)
	require.NoError(t, err)

	// Read the file and verify contents
	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	output := string(content)
	assert.Contains(t, output, "TEST_VAR=test-value")
}

func TestExecuteHook_CornerCases(t *testing.T) {
	tests := []struct {
		name        string
		hook        *ExecHook
		env         []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil hook",
			hook:        nil,
			env:         []string{"TEST=value"},
			expectError: false,
		},
		{
			name: "empty command",
			hook: &ExecHook{
				Command: []string{},
			},
			env:         []string{"TEST=value"},
			expectError: true,
			errorMsg:    "hook command is empty",
		},
		{
			name: "invalid command",
			hook: &ExecHook{
				Command: []string{"nonexistentcommand12345"},
			},
			env:         []string{"TEST=value"},
			expectError: true,
			errorMsg:    "hook execution failed",
		},
		{
			name: "successful command",
			hook: &ExecHook{
				Command: []string{"echo", "test"},
			},
			env:         []string{"TEST=value"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.hook.Invoke(tt.env)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecuteHook_CurlWebhook(t *testing.T) {
	// Create a local test server to avoid external dependencies
	receivedData := ""
	receivedHeaders := make(map[string]string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the request data for verification
		body, err := io.ReadAll(r.Body)
		if err == nil {
			receivedData = string(body)
		}
		receivedHeaders["Content-Type"] = r.Header.Get("Content-Type")
		receivedHeaders["User-Agent"] = r.Header.Get("User-Agent")

		// Return a simple response
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "ok"}`)
	}))
	defer server.Close()

	// Use the local test server URL instead of external httpbin.org
	hook := &ExecHook{
		Command: []string{fmt.Sprintf("curl -s -X POST -d \"$EPISODE_TITLE\" %s", server.URL)},
		Timeout: 10,
	}

	env := []string{
		"EPISODE_TITLE=Test Episode for Webhook",
		"FEED_NAME=test-podcast",
		"EPISODE_FILE=test-podcast/episode001.mp3",
	}

	err := hook.Invoke(env)
	assert.NoError(t, err, "Curl webhook should execute successfully")

	// Verify that the request was actually made and data was received
	assert.Equal(t, "Test Episode for Webhook", receivedData, "Server should receive the episode title")
	assert.Contains(t, receivedHeaders["User-Agent"], "curl", "Request should be made by curl")
}
