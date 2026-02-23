package segments

import (
	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

// Vim displays the vim mode when vim mode is enabled.
type Vim struct{}

// Name implements render.Segment.
func (Vim) Name() string { return "vim" }

// Source implements render.Segment.
func (Vim) Source() string { return SourceStable }

// Render implements render.Segment.
func (Vim) Render(data *model.StatusData, th *theme.Theme) (string, error) {
	if data.Vim == nil {
		return "", nil
	}
	return th.Colorize("vim", data.Vim.Mode), nil
}
