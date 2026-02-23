// Package statusline provides the core entry point for rendering the statusline.
package statusline

import (
	"encoding/json"
	"io"
)

// Run reads session JSON from in, renders the statusline, and writes to out.
// This is the core function that all tests exercise.
func Run(in io.Reader, out, _ io.Writer) error {
	// Consume stdin to avoid broken pipe when Claude Code writes to us.
	var raw json.RawMessage
	if err := json.NewDecoder(in).Decode(&raw); err != nil {
		// Empty or missing stdin is not an error for the statusline.
		return nil //nolint:nilerr // empty stdin is expected when no data is piped
	}

	// Phase 1 will implement: parse JSON, build segments, render output.
	_ = out

	return nil
}
