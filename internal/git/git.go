// Package git provides helpers for running git commands as subprocesses.
package git

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

const timeout = 500 * time.Millisecond

// Branch returns the current git branch for the given directory.
// Returns an empty string on error, timeout, or if the directory is not
// a git repository.
func Branch(dir string) string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
