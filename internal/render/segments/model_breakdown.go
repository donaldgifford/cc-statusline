package segments

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/donaldgifford/cc-statusline/internal/cache"
	"github.com/donaldgifford/cc-statusline/internal/jsonl"
	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

const (
	modelBreakdownCacheKey = "model_breakdown"
	modelBreakdownTTL      = 5 * time.Second
)

// ModelBreakdown groups session costs by model ID and displays them.
type ModelBreakdown struct{}

// Name implements render.Segment.
func (ModelBreakdown) Name() string { return "model_breakdown" }

// Source implements render.Segment.
func (ModelBreakdown) Source() string { return SourceExperimentalJSONL }

// Render implements render.Segment.
func (ModelBreakdown) Render(data *model.StatusData, th *theme.Theme) (string, error) {
	breakdown, err := getModelBreakdown(data.TranscriptPath)
	if err != nil {
		return "", err
	}
	if len(breakdown) == 0 {
		return "", nil
	}

	// Sort by cost descending.
	sort.Slice(breakdown, func(i, j int) bool {
		return breakdown[i].Cost > breakdown[j].Cost
	})

	var parts []string
	for _, m := range breakdown {
		text := fmt.Sprintf("%s:$%.2f", m.DisplayName, m.Cost)
		parts = append(parts, text)
	}
	result := strings.Join(parts, " ")
	return th.Colorize("model_breakdown", result), nil
}

type modelCost struct {
	DisplayName string  `json:"display_name"`
	Cost        float64 `json:"cost"`
}

func getModelBreakdown(transcriptPath string) ([]modelCost, error) {
	// Try cache first.
	if cached := cache.Get(modelBreakdownCacheKey, transcriptPath); cached != nil {
		var result []modelCost
		if err := json.Unmarshal(cached, &result); err == nil {
			return result, nil
		}
	}

	files := jsonl.DiscoverFiles()
	if transcriptPath != "" && !slices.Contains(files, transcriptPath) {
		files = append(files, transcriptPath)
	}

	if len(files) == 0 {
		return nil, ErrNoTranscripts
	}

	// Aggregate costs by model.
	costs := make(map[string]float64)
	for _, f := range files {
		entries, err := jsonl.ReadFile(f)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.Message != nil && e.Message.Model != "" {
				costs[e.Message.Model] += e.CostUSD
			}
		}
	}

	// Convert to slice with display names.
	var breakdown []modelCost
	for modelID, cost := range costs {
		breakdown = append(breakdown, modelCost{
			DisplayName: shortenModelName(modelID),
			Cost:        cost,
		})
	}

	// Cache the result.
	encoded, err := json.Marshal(breakdown)
	if err == nil {
		var mtime time.Time
		if transcriptPath != "" {
			if info, statErr := os.Stat(transcriptPath); statErr == nil {
				mtime = info.ModTime()
			}
		}
		cache.Set(modelBreakdownCacheKey, encoded, modelBreakdownTTL, mtime) //nolint:errcheck,gosec // best-effort cache
	}

	return breakdown, nil
}

// shortenModelName converts full model IDs to short display names.
func shortenModelName(modelID string) string {
	// Map of known model family prefixes to short names.
	replacements := []struct {
		prefix string
		short  string
	}{
		{"claude-opus-4-6", "opus4.6"},
		{"claude-opus-4-5", "opus4.5"},
		{"claude-opus-4-1", "opus4.1"},
		{"claude-opus-4", "opus4"},
		{"claude-sonnet-4-6", "sonnet4.6"},
		{"claude-sonnet-4-5", "sonnet4.5"},
		{"claude-sonnet-4", "sonnet4"},
		{"claude-3-7-sonnet", "sonnet3.7"},
		{"claude-3-5-sonnet", "sonnet3.5"},
		{"claude-3-sonnet", "sonnet3"},
		{"claude-haiku-4-5", "haiku4.5"},
		{"claude-3-5-haiku", "haiku3.5"},
		{"claude-3-haiku", "haiku3"},
		{"claude-3-opus", "opus3"},
	}

	for _, r := range replacements {
		if strings.HasPrefix(modelID, r.prefix) {
			return r.short
		}
	}

	// Fallback: strip "claude-" prefix and date suffix if present.
	name := modelID
	name = strings.TrimPrefix(name, "claude-")
	// Remove trailing date (e.g., "-20250514").
	if idx := strings.LastIndex(name, "-20"); idx != -1 && idx > 0 {
		name = name[:idx]
	}
	return name
}
