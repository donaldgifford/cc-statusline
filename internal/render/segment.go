// Package render provides the statusline rendering pipeline.
package render

import (
	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

// Segment is a unit of statusline output. Each segment renders a single
// piece of information (e.g., model name, context percentage).
type Segment interface {
	// Name returns the segment identifier used in config (e.g., "model").
	Name() string

	// Source returns the data source category: "stable",
	// "experimental:jsonl", or "experimental:usage_api".
	Source() string

	// Render produces the formatted text for this segment. Returns an empty
	// string when the segment has no data to display (e.g., vim mode when
	// vim is disabled). Returns an error only for experimental segments
	// that encounter failures.
	Render(data *model.StatusData, th *theme.Theme) (string, error)
}
