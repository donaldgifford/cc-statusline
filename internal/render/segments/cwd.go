package segments

import (
	"os"
	"strings"

	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

// CWD displays the current working directory, abbreviated with ~ for home.
type CWD struct{}

// Name implements render.Segment.
func (CWD) Name() string { return "cwd" }

// Source implements render.Segment.
func (CWD) Source() string { return SourceStable }

// Render implements render.Segment.
func (CWD) Render(data *model.StatusData, th *theme.Theme) (string, error) {
	dir := data.CWD
	if dir == "" {
		return "", nil
	}
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(dir, home) {
		dir = "~" + dir[len(home):]
	}
	return th.Colorize("cwd", dir), nil
}
