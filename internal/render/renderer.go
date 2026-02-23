package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

// Config holds rendering configuration.
type Config struct {
	// Lines is the segment layout. Each inner slice is one status bar row.
	Lines [][]string
	// Separator is the string placed between segments on a line.
	Separator string
	// ThemeName selects the color theme.
	ThemeName string
	// Experimental flags control which experimental sources are enabled.
	ExperimentalJSONL    bool
	ExperimentalUsageAPI bool
}

// DefaultConfig returns the default rendering configuration.
func DefaultConfig() Config {
	return Config{
		Lines:     [][]string{{"cwd", "git_branch", "model", "context"}},
		Separator: " ",
		ThemeName: theme.DefaultThemeName,
	}
}

// Renderer produces formatted statusline output.
type Renderer struct {
	cfg      Config
	segments map[string]Segment
	theme    *theme.Theme
}

// New creates a Renderer with the given configuration and segment registry.
func New(cfg Config, segments map[string]Segment) *Renderer {
	return &Renderer{
		cfg:      cfg,
		segments: segments,
		theme:    theme.Get(cfg.ThemeName),
	}
}

// Render writes the formatted statusline to w using the given status data.
func (r *Renderer) Render(w io.Writer, data *model.StatusData) error {
	for _, line := range r.cfg.Lines {
		var parts []string
		for _, name := range line {
			seg, ok := r.segments[name]
			if !ok {
				continue
			}
			if !r.sourceAllowed(seg.Source()) {
				continue
			}
			text, err := seg.Render(data, r.theme)
			if err != nil {
				continue
			}
			if text != "" {
				parts = append(parts, text)
			}
		}
		if len(parts) > 0 {
			if _, err := fmt.Fprintln(w, strings.Join(parts, r.cfg.Separator)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Renderer) sourceAllowed(source string) bool {
	switch source {
	case "experimental:jsonl":
		return r.cfg.ExperimentalJSONL
	case "experimental:usage_api":
		return r.cfg.ExperimentalUsageAPI
	default:
		return true
	}
}
