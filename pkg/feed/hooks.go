package feed

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// ExecHook represents a single hook configuration
type ExecHook struct {
	Command []string `toml:"command"`
	Timeout int      `toml:"timeout"` // timeout in seconds, 0 means use default (60s)
}

// Invoke runs a hook with the provided environment variables
func (h *ExecHook) Invoke(env []string) error {
	if h == nil {
		return nil
	}
	if len(h.Command) == 0 {
		return fmt.Errorf("hook command is empty")
	}

	// Set up context with timeout (default 1 minute if not specified)
	timeout := h.Timeout
	if timeout == 0 {
		timeout = 60 // default to 1 minute
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// Create command with context
	var cmd *exec.Cmd
	if len(h.Command) == 1 {
		// Single command, use shell to parse
		cmd = exec.CommandContext(ctx, "/bin/sh", "-c", h.Command[0])
	} else {
		// Multiple arguments, use directly
		cmd = exec.CommandContext(ctx, h.Command[0], h.Command[1:]...)
	}

	// Set up environment variables
	cmd.Env = append(os.Environ(), env...)

	// Execute the command
	data, err := cmd.CombinedOutput()
	output := string(data)

	if err != nil {
		return fmt.Errorf("hook execution failed: %v, output: %s", err, output)
	}

	return nil
}
