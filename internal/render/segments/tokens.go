package segments

import (
	"fmt"

	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

// Tokens displays input/output token counts.
type Tokens struct{}

// Name implements render.Segment.
func (Tokens) Name() string { return "tokens" }

// Source implements render.Segment.
func (Tokens) Source() string { return SourceStable }

// Render implements render.Segment.
func (Tokens) Render(data *model.StatusData, th *theme.Theme) (string, error) {
	if data.ContextWindow == nil {
		return "", nil
	}
	in := data.ContextWindow.TotalInputTokens
	out := data.ContextWindow.TotalOutputTokens
	if in == 0 && out == 0 {
		return "", nil
	}
	text := fmt.Sprintf("%s/%s tokens", formatTokens(in), formatTokens(out))
	return th.Colorize("tokens", text), nil
}

func formatTokens(n int) string {
	const kilo = 1000
	if n >= kilo {
		return fmt.Sprintf("%dk", n/kilo)
	}
	return fmt.Sprintf("%d", n)
}
