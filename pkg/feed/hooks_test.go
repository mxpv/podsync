package feed

import (
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
	hook := &ExecHook{
		Command: []string{"curl -s -X POST -d \"$EPISODE_TITLE\" httpbin.org/post"},
		Timeout: 10,
	}

	env := []string{
		"EPISODE_TITLE=Test Episode for Webhook",
		"FEED_NAME=test-podcast",
		"EPISODE_FILE=test-podcast/episode001.mp3",
	}

	err := hook.Invoke(env)
	assert.NoError(t, err, "Curl webhook should execute successfully")
}
