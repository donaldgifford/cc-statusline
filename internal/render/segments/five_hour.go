package segments

import (
	"fmt"
	"time"

	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
	"github.com/donaldgifford/cc-statusline/internal/usageapi"
)

const (
	// SourceExperimentalUsageAPI is the source value for usage API segments.
	SourceExperimentalUsageAPI = "experimental:usage_api"
)

// FiveHour displays the 5-hour usage window status.
type FiveHour struct {
	UsageFetcher func() (*usageapi.UsageResponse, error)
}

// Name implements render.Segment.
func (FiveHour) Name() string { return "five_hour" }

// Source implements render.Segment.
func (FiveHour) Source() string { return SourceExperimentalUsageAPI }

// Render implements render.Segment.
func (s FiveHour) Render(_ *model.StatusData, th *theme.Theme) (string, error) {
	if s.UsageFetcher == nil {
		return "", fmt.Errorf("no usage fetcher configured")
	}

	usage, err := s.UsageFetcher()
	if err != nil {
		return "", err
	}

	if usage.FiveHour == nil {
		return "", nil
	}

	pct := usage.FiveHour.UtilizationFloat() * 100
	text := fmt.Sprintf("5h: %s (%.0f%%)", formatTimeRemaining(usage.FiveHour.ResetsAt), pct)

	return th.Colorize("five_hour", text), nil
}

func formatTimeRemaining(resetsAt *string) string {
	if resetsAt == nil {
		return "--"
	}

	resetTime, err := time.Parse(time.RFC3339, *resetsAt)
	if err != nil {
		return "--"
	}

	remaining := time.Until(resetTime)
	if remaining <= 0 {
		return "0m left"
	}

	hours := int(remaining.Hours())
	mins := int(remaining.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm left", hours, mins)
	}
	return fmt.Sprintf("%dm left", mins)
}
