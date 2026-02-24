package usageapi

import (
	"encoding/json"
	"time"

	"github.com/donaldgifford/cc-statusline/internal/cache"
	"github.com/donaldgifford/cc-statusline/internal/errlog"
)

const (
	usageCacheKey     = "usage"
	usageCacheTTL     = 30 * time.Second
	usageGracePeriod  = 5 * time.Minute
	maxConsecFailures = 5
)

// CachedClient wraps Client with caching and grace period fallback.
type CachedClient struct {
	client       *Client
	consecFails  int
	lastGoodData *UsageResponse
	lastGoodTime time.Time
}

// NewCachedClient creates a CachedClient wrapping the given Client.
func NewCachedClient(client *Client) *CachedClient {
	return &CachedClient{client: client}
}

// Fetch returns cached usage data if within TTL, otherwise fetches fresh data.
// On failure, returns stale data within a 5-minute grace period.
func (cc *CachedClient) Fetch() (*UsageResponse, error) {
	// Try cache first.
	if cached := cache.Get(usageCacheKey, ""); cached != nil {
		var usage UsageResponse
		if err := json.Unmarshal(cached, &usage); err == nil {
			return &usage, nil
		}
	}

	// Fetch fresh data.
	usage, err := cc.client.Fetch()
	if err != nil {
		cc.consecFails++
		if cc.consecFails >= maxConsecFailures {
			errlog.Log("usage API: %d consecutive failures; run 'cc-statusline auth --status'", cc.consecFails)
		}

		// Grace period: return last good data if recent enough.
		if cc.lastGoodData != nil && time.Since(cc.lastGoodTime) < usageGracePeriod {
			errlog.Log("usage API: using stale data (grace period)")
			return cc.lastGoodData, nil
		}

		return nil, err
	}

	// Success — reset failure counter and cache.
	cc.consecFails = 0
	cc.lastGoodData = usage
	cc.lastGoodTime = time.Now()

	encoded, marshalErr := json.Marshal(usage)
	if marshalErr == nil {
		cache.Set(usageCacheKey, encoded, usageCacheTTL, time.Time{}) //nolint:errcheck,gosec // best-effort cache
	}

	return usage, nil
}
