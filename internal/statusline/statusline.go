// Package statusline provides the core entry point for rendering the statusline.
package statusline

import (
	"io"

	"github.com/donaldgifford/cc-statusline/internal/config"
	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render"
	"github.com/donaldgifford/cc-statusline/internal/render/segments"
)

// Run reads session JSON from in, renders the statusline, and writes to out.
// This is the core function that all tests exercise.
func Run(in io.Reader, out, _ io.Writer) error {
	return RunWithConfig(in, out, nil)
}

// RunWithConfig reads session JSON and renders the statusline using the
// provided configuration. If cfg is nil, the default config is used.
// Callers should set color.SetEnabled before calling this function.
func RunWithConfig(in io.Reader, out io.Writer, cfg *config.Config) error {
	data, err := model.ReadStatus(in)
	if err != nil {
		return err
	}

	if cfg == nil {
		cfg = config.Default()
	}

	rcfg := render.Config{
		Lines:                cfg.ResolvedLines(),
		Separator:            cfg.Separator,
		ThemeName:            cfg.Theme,
		ExperimentalJSONL:    cfg.Experimental.JSONL,
		ExperimentalUsageAPI: cfg.Experimental.UsageAPI,
	}

	r := render.New(rcfg, segments.All())
	return r.Render(out, data)
}
