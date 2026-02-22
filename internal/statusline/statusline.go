// Package statusline provides the core entry point for rendering the statusline.
package statusline

import (
	"io"
)

// Run reads session JSON from in, renders the statusline, and writes to out.
// This is the core function that all tests exercise.
func Run(in io.Reader, out io.Writer, errOut io.Writer) error {
	// Phase 1 will implement: parse JSON, build segments, render output.
	return nil
}
