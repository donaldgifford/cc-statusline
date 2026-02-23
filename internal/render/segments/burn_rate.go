package segments

import (
	"encoding/json"
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
	burnRateCacheKey = "burn_rate"
	burnRateTTL      = 5 * time.Second

	// blockGap is the minimum gap between entries that defines a new activity
	// block. Entries within this gap are considered part of the same block.
	blockGap = 5 * time.Hour
)

// BurnRate calculates cost/hour from the current activity block.
type BurnRate struct{}

// Name implements render.Segment.
func (BurnRate) Name() string { return "burn_rate" }

// Source implements render.Segment.
func (BurnRate) Source() string { return SourceExperimentalJSONL }

// Render implements render.Segment.
func (BurnRate) Render(data *model.StatusData, th *theme.Theme) (string, error) {
	rate, err := getBurnRate(data.TranscriptPath)
	if err != nil {
		return "", err
	}
	text := fmt.Sprintf("$%.2f/hr", rate)
	return th.Colorize("burn_rate", text), nil
}

type burnRateResult struct {
	Rate float64 `json:"rate"`
}

func getBurnRate(transcriptPath string) (float64, error) {
	// Try cache first.
	if cached := cache.Get(burnRateCacheKey, transcriptPath); cached != nil {
		var result burnRateResult
		if err := json.Unmarshal(cached, &result); err == nil {
			return result.Rate, nil
		}
	}

	files := jsonl.DiscoverFiles()
	if transcriptPath != "" && !slices.Contains(files, transcriptPath) {
		files = append(files, transcriptPath)
	}

	if len(files) == 0 {
		return 0, ErrNoTranscripts
	}

	// Collect all entries from all files.
	var allEntries []jsonl.Entry
	for _, f := range files {
		entries, err := jsonl.ReadFile(f)
		if err != nil {
			continue
		}
		allEntries = append(allEntries, entries...)
	}

	rate := calculateBurnRate(allEntries)

	// Cache the result.
	result := burnRateResult{Rate: rate}
	encoded, err := json.Marshal(result)
	if err == nil {
		var mtime time.Time
		if transcriptPath != "" {
			if info, statErr := os.Stat(transcriptPath); statErr == nil {
				mtime = info.ModTime()
			}
		}
		cache.Set(burnRateCacheKey, encoded, burnRateTTL, mtime) //nolint:errcheck,gosec // best-effort cache
	}

	return rate, nil
}

// calculateBurnRate finds the current activity block (ending at the most recent
// entry) and computes cost per hour within that block.
func calculateBurnRate(entries []jsonl.Entry) float64 {
	if len(entries) == 0 {
		return 0
	}

	// Find the start of the current block by walking backward from the end.
	// A block boundary is defined as a gap of > blockGap between consecutive entries.
	blockStart := 0
	for i := len(entries) - 1; i > 0; i-- {
		gap := entries[i].Timestamp.Sub(entries[i-1].Timestamp)
		if gap > blockGap {
			blockStart = i
			break
		}
	}

	// Sum cost within the block.
	var blockCost float64
	for i := blockStart; i < len(entries); i++ {
		blockCost += entries[i].CostUSD
	}

	// Calculate duration of the block.
	first := entries[blockStart].Timestamp
	last := entries[len(entries)-1].Timestamp
	duration := last.Sub(first)

	if duration <= 0 {
		// Single entry or zero-duration block — return the cost as instantaneous.
		return blockCost
	}

	return blockCost / duration.Hours()
}
