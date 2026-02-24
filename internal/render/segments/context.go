package segments

import (
	"fmt"

	"github.com/donaldgifford/cc-statusline/internal/color"
	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

// Context displays context window usage as a percentage with color coding.
type Context struct{}

// Name implements render.Segment.
func (Context) Name() string { return "context" }

// Source implements render.Segment.
func (Context) Source() string { return SourceStable }

// Render implements render.Segment.
func (Context) Render(data *model.StatusData, th *theme.Theme) (string, error) {
	_ = th
	if data.ContextWindow == nil || data.ContextWindow.UsedPercentage == nil {
		return color.Colorize("ctx:--", color.FgGreen), nil
	}
	pct := *data.ContextWindow.UsedPercentage
	text := fmt.Sprintf("ctx:%d%%", pct)
	return color.Colorize(text, contextColor(pct)), nil
}

func contextColor(pct int) string {
	switch {
	case pct > 80:
		return color.FgRed
	case pct >= 50:
		return color.FgYellow
	default:
		return color.FgGreen
	}
}
