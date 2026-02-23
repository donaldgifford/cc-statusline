// Package segments implements individual statusline segments.
package segments

import (
	"github.com/donaldgifford/cc-statusline/internal/render"
	"github.com/donaldgifford/cc-statusline/internal/usageapi"
)

// SourceStable is the source value for segments that use stable data.
const SourceStable = "stable"

// AllConfig holds optional dependencies for segment initialization.
type AllConfig struct {
	// UsageFetcher provides usage data for experimental:usage_api segments.
	// If nil, usage API segments will return errors when rendered.
	UsageFetcher func() (*usageapi.UsageResponse, error)
}

// All returns a map of all built-in segments keyed by name.
func All(cfg AllConfig) map[string]render.Segment {
	list := []render.Segment{
		CWD{},
		GitBranch{},
		Model{},
		Context{},
		Cost{},
		Duration{},
		Tokens{},
		Lines{},
		Vim{},
		Agent{},
		DailyCost{},
		BurnRate{},
		ModelBreakdown{},
		FiveHour(cfg),
		WeeklyLimits(cfg),
		ExtraUsage(cfg),
	}
	m := make(map[string]render.Segment, len(list))
	for _, s := range list {
		m[s.Name()] = s
	}
	return m
}
