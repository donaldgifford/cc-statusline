package segments

import (
	"fmt"
	"strings"
	"time"

	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
	"github.com/donaldgifford/cc-statusline/internal/usageapi"
)

// WeeklyLimits displays per-model and overall weekly usage limits.
type WeeklyLimits struct {
	UsageFetcher func() (*usageapi.UsageResponse, error)
}

// Name implements render.Segment.
func (WeeklyLimits) Name() string { return "weekly_limits" }

// Source implements render.Segment.
func (WeeklyLimits) Source() string { return SourceExperimentalUsageAPI }

// Render implements render.Segment.
func (s WeeklyLimits) Render(_ *model.StatusData, th *theme.Theme) (string, error) {
	if s.UsageFetcher == nil {
		return "", fmt.Errorf("no usage fetcher configured")
	}

	usage, err := s.UsageFetcher()
	if err != nil {
		return "", err
	}

	var parts []string
	if usage.SevenDaySonnet != nil {
		pct := usage.SevenDaySonnet.UtilizationFloat() * 100
		reset := formatResetTime(usage.SevenDaySonnet.ResetsAt)
		parts = append(parts, fmt.Sprintf("sonnet %.0f%% (%s)", pct, reset))
	}
	if usage.SevenDay != nil {
		pct := usage.SevenDay.UtilizationFloat() * 100
		reset := formatResetTime(usage.SevenDay.ResetsAt)
		parts = append(parts, fmt.Sprintf("all %.0f%% (%s)", pct, reset))
	}

	if len(parts) == 0 {
		return "", nil
	}

	text := "wk: " + strings.Join(parts, " / ")
	return th.Colorize("weekly_limits", text), nil
}

// formatResetTime converts an RFC3339 UTC time to a short local representation.
func formatResetTime(resetsAt *string) string {
	if resetsAt == nil {
		return "--"
	}

	resetTime, err := time.Parse(time.RFC3339, *resetsAt)
	if err != nil {
		return "--"
	}

	local := resetTime.Local()
	day := local.Format("Mon")
	hour := local.Format("3p")

	return fmt.Sprintf("resets %s %s", day, strings.ToLower(hour))
}
