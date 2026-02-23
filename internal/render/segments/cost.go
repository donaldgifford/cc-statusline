package segments

import (
	"fmt"

	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

// Cost displays the session cost in USD.
type Cost struct{}

// Name implements render.Segment.
func (Cost) Name() string { return "cost" }

// Source implements render.Segment.
func (Cost) Source() string { return SourceStable }

// Render implements render.Segment.
func (Cost) Render(data *model.StatusData, th *theme.Theme) (string, error) {
	if data.Cost == nil {
		return "", nil
	}
	text := fmt.Sprintf("$%.2f", data.Cost.TotalCostUSD)
	return th.Colorize("cost", text), nil
}
