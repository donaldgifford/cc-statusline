package usageapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/donaldgifford/cc-statusline/internal/errlog"
)

const (
	usageURL       = "https://api.anthropic.com/api/oauth/usage"
	betaHeader     = "oauth-2025-04-20"
	requestTimeout = 2 * time.Second
	retryDelay     = 500 * time.Millisecond
)

// UsageResponse holds the parsed usage API response.
type UsageResponse struct {
	FiveHour       *UsageWindow `json:"five_hour"`
	SevenDay       *UsageWindow `json:"seven_day"`
	SevenDayOpus   *UsageWindow `json:"seven_day_opus"`
	SevenDaySonnet *UsageWindow `json:"seven_day_sonnet"`
	ExtraUsage     *ExtraUsage  `json:"extra_usage"`
}

// UsageWindow holds a single usage window's data.
type UsageWindow struct {
	Utilization json.Number `json:"utilization"`
	ResetsAt    *string     `json:"resets_at"`
}

// ExtraUsage holds extra usage (overage) information.
type ExtraUsage struct {
	IsEnabled    bool         `json:"is_enabled"`
	MonthlyLimit *int         `json:"monthly_limit"`
	UsedCredits  *int         `json:"used_credits"`
	Utilization  *json.Number `json:"utilization"`
}

// UtilizationFloat returns the utilization as a float64.
// Returns 0 if the value cannot be parsed.
func (w *UsageWindow) UtilizationFloat() float64 {
	if w == nil {
		return 0
	}
	f, err := w.Utilization.Float64()
	if err != nil {
		return 0
	}
	return f
}

// Client fetches usage data from the Anthropic OAuth API.
type Client struct {
	token      string
	httpClient *http.Client
}

// NewClient creates a usage API client with the given OAuth token.
func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// Fetch retrieves the current usage data. Retries once on 5xx errors.
func (c *Client) Fetch() (*UsageResponse, error) {
	resp, err := c.doRequest()
	if err != nil {
		return nil, err
	}

	// Retry once on 5xx.
	if resp.StatusCode >= http.StatusInternalServerError {
		resp.Body.Close() //nolint:gosec // closing before retry
		time.Sleep(retryDelay)
		resp, err = c.doRequest()
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		errlog.Log("usage API: %d — run 'cc-statusline auth' to re-authenticate", resp.StatusCode)
		return nil, fmt.Errorf("usage API: unauthorized (HTTP %d); run 'cc-statusline auth'", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		errlog.Log("usage API: unexpected status %d", resp.StatusCode)
		return nil, fmt.Errorf("usage API: HTTP %d", resp.StatusCode)
	}

	var usage UsageResponse
	if err := json.NewDecoder(resp.Body).Decode(&usage); err != nil {
		return nil, fmt.Errorf("usage API: parse response: %w", err)
	}

	return &usage, nil
}

func (c *Client) doRequest() (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, usageURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("anthropic-beta", betaHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	return resp, nil
}
