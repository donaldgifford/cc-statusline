// Package errlog provides error logging for experimental features.
// Errors are appended to ~/.cache/cc-statusline/error.log with
// ISO 8601 timestamps. The file is rotated when it exceeds 1MB.
package errlog

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	maxLogSize = 1 << 20 // 1MB
	dirPerms   = 0o700
	filePerms  = 0o600
)

// Log appends an error message to the error log file.
func Log(format string, args ...any) {
	path := logPath()
	if path == "" {
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), dirPerms); err != nil {
		return
	}

	rotateIfNeeded(path)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePerms)
	if err != nil {
		return
	}
	defer f.Close() //nolint:errcheck // best-effort logging

	msg := fmt.Sprintf(format, args...)
	ts := time.Now().UTC().Format(time.RFC3339)
	//nolint:errcheck // best-effort log write
	fmt.Fprintf(f, "%s %s\n", ts, msg)
}

func rotateIfNeeded(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.Size() > maxLogSize {
		_ = os.Truncate(path, 0) //nolint:errcheck // best-effort rotation
	}
}

func logPath() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(cacheDir, "cc-statusline", "error.log")
}
