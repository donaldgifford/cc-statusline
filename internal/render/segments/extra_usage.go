package segments

import (
	"fmt"

	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
	"github.com/donaldgifford/cc-statusline/internal/usageapi"
)

// ExtraUsage displays extra usage spending against the monthly limit.
type ExtraUsage struct {
	UsageFetcher func() (*usageapi.UsageResponse, error)
}

// Name implements render.Segment.
func (ExtraUsage) Name() string { return "extra_usage" }

// Source implements render.Segment.
func (ExtraUsage) Source() string { return SourceExperimentalUsageAPI }

// Render implements render.Segment.
func (s ExtraUsage) Render(_ *model.StatusData, th *theme.Theme) (string, error) {
	if s.UsageFetcher == nil {
		return "", fmt.Errorf("no usage fetcher configured")
	}

	usage, err := s.UsageFetcher()
	if err != nil {
		return "", err
	}

	if usage.ExtraUsage == nil || !usage.ExtraUsage.IsEnabled {
		return "", nil
	}

	used := centsToUSD(usage.ExtraUsage.UsedCredits)
	limit := centsToUSD(usage.ExtraUsage.MonthlyLimit)

	text := fmt.Sprintf("extra: $%.2f / $%.2f", used, limit)
	return th.Colorize("extra_usage", text), nil
}

func centsToUSD(cents *int) float64 {
	if cents == nil {
		return 0
	}
	return float64(*cents) / 100
}
