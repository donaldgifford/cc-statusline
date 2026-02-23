package segments

import (
	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

// Model displays the model display name in brackets.
type Model struct{}

// Name implements render.Segment.
func (Model) Name() string { return "model" }

// Source implements render.Segment.
func (Model) Source() string { return SourceStable }

// Render implements render.Segment.
func (Model) Render(data *model.StatusData, th *theme.Theme) (string, error) {
	name := data.Model.DisplayName
	if name == "" {
		return "", nil
	}
	return th.Colorize("model", "["+name+"]"), nil
}
