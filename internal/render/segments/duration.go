package segments

import (
	"fmt"
	"time"

	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

// Duration displays the session duration as human-readable time.
type Duration struct{}

// Name implements render.Segment.
func (Duration) Name() string { return "duration" }

// Source implements render.Segment.
func (Duration) Source() string { return SourceStable }

// Render implements render.Segment.
func (Duration) Render(data *model.StatusData, th *theme.Theme) (string, error) {
	if data.Cost == nil || data.Cost.TotalDurationMS == 0 {
		return "", nil
	}
	d := time.Duration(data.Cost.TotalDurationMS) * time.Millisecond
	text := formatDuration(d)
	return th.Colorize("duration", text), nil
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
