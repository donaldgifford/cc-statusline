package segments

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/donaldgifford/cc-statusline/internal/cache"
	"github.com/donaldgifford/cc-statusline/internal/jsonl"
	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

const (
	// SourceExperimentalJSONL is the source value for JSONL-based segments.
	SourceExperimentalJSONL = "experimental:jsonl"

	dailyCostCacheKey = "daily_cost"
	dailyCostTTL      = 5 * time.Second
)

// ErrNoTranscripts indicates no JSONL transcript files could be found.
var ErrNoTranscripts = errors.New("no transcript files found")

// DailyCost displays the total cost across all JSONL entries from today (UTC).
type DailyCost struct{}

// Name implements render.Segment.
func (DailyCost) Name() string { return "daily_cost" }

// Source implements render.Segment.
func (DailyCost) Source() string { return SourceExperimentalJSONL }

// Render implements render.Segment.
func (DailyCost) Render(data *model.StatusData, th *theme.Theme) (string, error) {
	cost, err := getDailyCost(data.TranscriptPath)
	if err != nil {
		return "", err
	}
	text := fmt.Sprintf("$%.2f today", cost)
	return th.Colorize("daily_cost", text), nil
}

type dailyCostResult struct {
	Cost float64 `json:"cost"`
}

func getDailyCost(transcriptPath string) (float64, error) {
	// Try cache first.
	if cached := cache.Get(dailyCostCacheKey, transcriptPath); cached != nil {
		var result dailyCostResult
		if err := json.Unmarshal(cached, &result); err == nil {
			return result.Cost, nil
		}
	}

	// Discover all JSONL files and aggregate today's cost.
	files := jsonl.DiscoverFiles()
	if transcriptPath != "" && !slices.Contains(files, transcriptPath) {
		files = append(files, transcriptPath)
	}

	if len(files) == 0 {
		return 0, ErrNoTranscripts
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	var totalCost float64

	for _, f := range files {
		entries, err := jsonl.ReadFile(f)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.Timestamp.UTC().Truncate(24 * time.Hour).Equal(today) {
				totalCost += e.CostUSD
			}
		}
	}

	// Cache the result.
	result := dailyCostResult{Cost: totalCost}
	encoded, err := json.Marshal(result)
	if err == nil {
		var mtime time.Time
		if transcriptPath != "" {
			if info, statErr := os.Stat(transcriptPath); statErr == nil {
				mtime = info.ModTime()
			}
		}
		cache.Set(dailyCostCacheKey, encoded, dailyCostTTL, mtime) //nolint:errcheck,gosec // best-effort cache
	}

	return totalCost, nil
}
