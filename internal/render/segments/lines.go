package segments

import (
	"fmt"

	"github.com/donaldgifford/cc-statusline/internal/color"
	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

// Lines displays lines added/removed with color coding.
type Lines struct{}

// Name implements render.Segment.
func (Lines) Name() string { return "lines" }

// Source implements render.Segment.
func (Lines) Source() string { return SourceStable }

// Render implements render.Segment.
func (Lines) Render(data *model.StatusData, th *theme.Theme) (string, error) {
	_ = th
	if data.Cost == nil {
		return "", nil
	}
	added := data.Cost.TotalLinesAdded
	removed := data.Cost.TotalLinesRemoved
	if added == 0 && removed == 0 {
		return "", nil
	}
	addText := color.Colorize(fmt.Sprintf("+%d", added), color.FgGreen)
	remText := color.Colorize(fmt.Sprintf("-%d", removed), color.FgRed)
	return addText + " " + remText, nil
}
