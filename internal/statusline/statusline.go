// Package statusline provides the core entry point for rendering the statusline.
package statusline

import (
	"io"

	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render"
	"github.com/donaldgifford/cc-statusline/internal/render/segments"
)

// Run reads session JSON from in, renders the statusline, and writes to out.
// This is the core function that all tests exercise.
func Run(in io.Reader, out, _ io.Writer) error {
	data, err := model.ReadStatus(in)
	if err != nil {
		return err
	}

	cfg := render.DefaultConfig()
	r := render.New(cfg, segments.All())
	return r.Render(out, data)
}
