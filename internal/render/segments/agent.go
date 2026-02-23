package segments

import (
	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

// Agent displays the agent name when running with --agent.
type Agent struct{}

// Name implements render.Segment.
func (Agent) Name() string { return "agent" }

// Source implements render.Segment.
func (Agent) Source() string { return SourceStable }

// Render implements render.Segment.
func (Agent) Render(data *model.StatusData, th *theme.Theme) (string, error) {
	if data.Agent == nil {
		return "", nil
	}
	return th.Colorize("agent", data.Agent.Name), nil
}
