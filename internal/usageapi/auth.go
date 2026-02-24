// Package usageapi provides the Anthropic OAuth usage API client.
package usageapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const keychainTimeout = 2 * time.Second

// CommandRunner executes a system command and returns its output.
// Used for testing to mock keychain access.
type CommandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// TokenResult holds a resolved OAuth token and metadata.
type TokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Source       string
	Expired      bool
}

// ErrNoCredentials indicates that no credential source could provide a token.
var ErrNoCredentials = errors.New("no credentials found")

// AuthConfig configures credential resolution.
type AuthConfig struct {
	// RunCommand executes a system command. Defaults to os/exec.
	RunCommand CommandRunner
}

type storedCredentials struct {
	ClaudeAIOAuth *oauthToken `json:"claudeAiOauth"`
}

type oauthToken struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

// ReadCredentials resolves an OAuth token from available sources.
// Priority: 1) CC_STATUSLINE_TOKEN env, 2) macOS Keychain,
// 3) ~/.claude/.credentials.json, 4) ~/.config/cc-statusline/auth.json.
func ReadCredentials(cfg AuthConfig) (*TokenResult, error) {
	var tried []string

	// 1. Environment variable.
	if t := os.Getenv("CC_STATUSLINE_TOKEN"); t != "" {
		return &TokenResult{
			AccessToken: t,
			Source:      "CC_STATUSLINE_TOKEN",
		}, nil
	}
	tried = append(tried, "CC_STATUSLINE_TOKEN env var: not set")

	// 2. macOS Keychain.
	if runtime.GOOS == "darwin" {
		result, err := readKeychain(cfg)
		if err == nil {
			return result, nil
		}
		tried = append(tried, fmt.Sprintf("macOS Keychain: %v", err))
	}

	// 3. Claude Code credentials file.
	result, err := readClaudeCredentials()
	if err == nil {
		return result, nil
	}
	tried = append(tried, fmt.Sprintf("~/.claude/.credentials.json: %v", err))

	// 4. Manual fallback file.
	result, err = readManualAuth()
	if err == nil {
		return result, nil
	}
	tried = append(tried, fmt.Sprintf("~/.config/cc-statusline/auth.json: %v", err))

	return nil, fmt.Errorf("%w: tried %s", ErrNoCredentials, strings.Join(tried, "; "))
}

func readKeychain(cfg AuthConfig) (*TokenResult, error) {
	runner := cfg.RunCommand
	if runner == nil {
		return nil, errors.New("no command runner configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), keychainTimeout)
	defer cancel()

	out, err := runner(ctx, "security", "find-generic-password",
		"-s", "Claude Code-credentials", "-w")
	if err != nil {
		return nil, fmt.Errorf("keychain read: %w", err)
	}

	return parseCredentials(strings.TrimSpace(string(out)), "macOS Keychain")
}

func readClaudeCredentials() (*TokenResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(home, ".claude", ".credentials.json"))
	if err != nil {
		return nil, err
	}
	return parseCredentials(string(data), "~/.claude/.credentials.json")
}

func readManualAuth() (*TokenResult, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(configDir, "cc-statusline", "auth.json"))
	if err != nil {
		return nil, err
	}
	return parseCredentials(string(data), "~/.config/cc-statusline/auth.json")
}

func parseCredentials(raw, source string) (*TokenResult, error) {
	var creds storedCredentials
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	if creds.ClaudeAIOAuth == nil || creds.ClaudeAIOAuth.AccessToken == "" {
		return nil, errors.New("no OAuth token found")
	}

	result := &TokenResult{
		AccessToken:  creds.ClaudeAIOAuth.AccessToken,
		RefreshToken: creds.ClaudeAIOAuth.RefreshToken,
		Source:       source,
	}

	if creds.ClaudeAIOAuth.ExpiresAt > 0 {
		result.ExpiresAt = time.UnixMilli(creds.ClaudeAIOAuth.ExpiresAt)
		result.Expired = time.Now().After(result.ExpiresAt)
	}

	return result, nil
}
