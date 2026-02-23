package usageapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/donaldgifford/cc-statusline/internal/errlog"
)

const (
	refreshURL     = "https://console.anthropic.com/v1/oauth/token"
	refreshTimeout = 5 * time.Second
	authFilePerms  = 0o600
	authDirPerms   = 0o700
)

type refreshRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// RefreshToken attempts to refresh an expired OAuth token. On success, caches
// the new token at ~/.config/cc-statusline/auth.json (does NOT write back to
// the macOS Keychain or Claude Code's credentials file).
func RefreshToken(refreshTok string) (*TokenResult, error) {
	if refreshTok == "" {
		return nil, fmt.Errorf("no refresh token available; run 'claude auth' to re-authenticate")
	}

	body, err := json.Marshal(refreshRequest{
		GrantType:    "refresh_token",
		RefreshToken: refreshTok,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal refresh request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), refreshTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, refreshURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errlog.Log("token refresh failed with status %d; run 'claude auth' to re-authenticate", resp.StatusCode)
		return nil, fmt.Errorf("refresh failed (HTTP %d); run 'claude auth' to re-authenticate", resp.StatusCode)
	}

	var refreshResp refreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&refreshResp); err != nil {
		return nil, fmt.Errorf("parse refresh response: %w", err)
	}

	if refreshResp.AccessToken == "" {
		return nil, fmt.Errorf("empty access token in refresh response")
	}

	result := &TokenResult{
		AccessToken:  refreshResp.AccessToken,
		RefreshToken: refreshResp.RefreshToken,
		Source:       "token refresh",
	}
	if refreshResp.ExpiresIn > 0 {
		result.ExpiresAt = time.Now().Add(time.Duration(refreshResp.ExpiresIn) * time.Second)
	}

	// Cache the refreshed token for subsequent invocations.
	if err := cacheAuthToken(result); err != nil {
		errlog.Log("failed to cache refreshed token: %v", err)
	}

	return result, nil
}

func cacheAuthToken(result *TokenResult) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(configDir, "cc-statusline")
	if err := os.MkdirAll(dir, authDirPerms); err != nil {
		return err
	}

	creds := storedCredentials{
		ClaudeAIOAuth: &oauthToken{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			ExpiresAt:    result.ExpiresAt.UnixMilli(),
		},
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "auth.json"), data, authFilePerms)
}
