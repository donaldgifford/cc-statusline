// Package pricing fetches and caches Claude model pricing data.
package pricing

import (
	"context"
	"embed"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/donaldgifford/cc-statusline/internal/cache"
	"github.com/donaldgifford/cc-statusline/internal/errlog"
)

const (
	// litellmURL is the source for model pricing data.
	litellmURL = "https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json"

	// cacheKey is the cache file name for fetched pricing data.
	cacheKey = "pricing"

	// cacheTTL is the time between pricing refreshes.
	cacheTTL = 24 * time.Hour

	// fetchTimeout is the HTTP request timeout.
	fetchTimeout = 10 * time.Second

	// maxResponseBytes caps the response body to avoid OOM on bad responses.
	maxResponseBytes = 10 * 1024 * 1024 // 10MB
)

//go:embed fallback.json
var fallbackFS embed.FS

// ModelPricing holds per-token costs for a single model.
type ModelPricing struct {
	InputCostPerToken           float64 `json:"input_cost_per_token"`
	OutputCostPerToken          float64 `json:"output_cost_per_token"`
	CacheCreationInputTokenCost float64 `json:"cache_creation_input_token_cost"`
	CacheReadInputTokenCost     float64 `json:"cache_read_input_token_cost"`
	MaxInputTokens              int     `json:"max_input_tokens"`
	MaxOutputTokens             int     `json:"max_output_tokens"`
}

// PricingData maps normalized model names to their pricing.
type PricingData map[string]ModelPricing

// Get returns the current pricing data. It checks the cache first, then
// fetches from LiteLLM, falling back to the bundled pricing on failure.
func Get() PricingData {
	// Try cache first (no source path — pricing has no source file to track).
	if cached := cache.Get(cacheKey, ""); cached != nil {
		var data PricingData
		if err := json.Unmarshal(cached, &data); err == nil {
			return data
		}
	}

	// Fetch fresh data.
	if data, err := fetch(); err == nil {
		if encoded, err := json.Marshal(data); err == nil {
			if err := cache.Set(cacheKey, encoded, cacheTTL, time.Time{}); err != nil {
				errlog.Log("pricing cache write: %v", err)
			}
		}
		return data
	}

	// Fallback to bundled pricing.
	errlog.Log("pricing: using bundled fallback")
	return loadFallback()
}

// Lookup returns the pricing for a model ID. It tries exact match first,
// then attempts prefix matching for model variants.
func Lookup(data PricingData, modelID string) (ModelPricing, bool) {
	// Exact match.
	if p, ok := data[modelID]; ok {
		return p, true
	}

	// Try stripping date suffix for prefix match (e.g., "claude-sonnet-4-6-20250514" -> "claude-sonnet-4-6").
	for key, p := range data {
		if strings.HasPrefix(modelID, key) || strings.HasPrefix(key, modelID) {
			return p, true
		}
	}

	return ModelPricing{}, false
}

func fetch() (PricingData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, litellmURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, maxResponseBytes)
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(limited).Decode(&raw); err != nil {
		return nil, err
	}

	return filterClaude(raw), nil
}

func filterClaude(raw map[string]json.RawMessage) PricingData {
	result := make(PricingData)
	for key, val := range raw {
		lower := strings.ToLower(key)
		if !strings.Contains(lower, "claude") {
			continue
		}

		var mp ModelPricing
		if err := json.Unmarshal(val, &mp); err != nil {
			continue
		}

		// Skip entries without pricing data.
		if mp.InputCostPerToken == 0 && mp.OutputCostPerToken == 0 {
			continue
		}

		// Normalize the key to a bare model name.
		normalized := normalizeModelKey(key)
		// Keep the first entry for each normalized name (avoids regional dupes).
		if _, exists := result[normalized]; !exists {
			result[normalized] = mp
		}
	}
	return result
}

// normalizeModelKey strips provider prefixes and version suffixes from
// LiteLLM model keys to produce bare model names.
//
// Examples:
//
//	"anthropic.claude-sonnet-4-6"          -> "claude-sonnet-4-6"
//	"us.anthropic.claude-opus-4-6-v1"      -> "claude-opus-4-6"
//	"azure_ai/claude-haiku-4-5"            -> "claude-haiku-4-5"
//	"anthropic.claude-3-5-sonnet-20241022-v2:0" -> "claude-3-5-sonnet-20241022"
func normalizeModelKey(key string) string {
	// Strip provider prefix: "anthropic.", "us.anthropic.", "azure_ai/", etc.
	if idx := strings.LastIndex(key, "."); idx != -1 {
		key = key[idx+1:]
	}
	if idx := strings.LastIndex(key, "/"); idx != -1 {
		key = key[idx+1:]
	}

	// Strip LiteLLM version suffixes like "-v1:0", "-v2:0", "-v1".
	if idx := strings.LastIndex(key, "-v"); idx != -1 {
		suffix := key[idx:]
		// Only strip if it looks like a version suffix (e.g., "-v1", "-v1:0", "-v2:0").
		if len(suffix) >= 3 && suffix[2] >= '0' && suffix[2] <= '9' {
			key = key[:idx]
		}
	}

	// Strip @date suffixes like "@20251001".
	if idx := strings.Index(key, "@"); idx != -1 {
		key = key[:idx]
	}

	return key
}

func loadFallback() PricingData {
	data, err := fallbackFS.ReadFile("fallback.json")
	if err != nil {
		// Should never happen since the file is embedded.
		return make(PricingData)
	}

	var pricing PricingData
	if err := json.Unmarshal(data, &pricing); err != nil {
		return make(PricingData)
	}
	return pricing
}
