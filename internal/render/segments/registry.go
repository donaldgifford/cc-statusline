// Package segments implements individual statusline segments.
package segments

import "github.com/donaldgifford/cc-statusline/internal/render"

// SourceStable is the source value for segments that use stable data.
const SourceStable = "stable"

// All returns a map of all built-in segments keyed by name.
func All() map[string]render.Segment {
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
	}
	m := make(map[string]render.Segment, len(list))
	for _, s := range list {
		m[s.Name()] = s
	}
	return m
}
