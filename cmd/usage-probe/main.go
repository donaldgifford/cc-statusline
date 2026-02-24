// Command usage-probe is a standalone tool for testing the Anthropic OAuth
// usage API. Not shipped in releases.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	usageURL = "https://api.anthropic.com/api/oauth/usage"
	betaHdr  = "oauth-2025-04-20"
	timeout  = 5 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	token, source, err := findToken()
	if err != nil {
		return fmt.Errorf("finding token: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Using token from: %s\n", source)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, usageURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("anthropic-beta", betaHdr)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Status: %d\n", resp.StatusCode)

	// Pretty-print if valid JSON.
	var pretty json.RawMessage
	if json.Unmarshal(body, &pretty) == nil {
		formatted, fmtErr := json.MarshalIndent(pretty, "", "  ")
		if fmtErr == nil {
			fmt.Println(string(formatted))
			return nil
		}
	}
	fmt.Println(string(body))
	return nil
}

type credentials struct {
	ClaudeAIOAuth *oauthCred `json:"claudeAiOauth"`
}

type oauthCred struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

func findToken() (string, string, error) {
	// 1. Environment variable.
	if t := os.Getenv("CC_STATUSLINE_TOKEN"); t != "" {
		return t, "CC_STATUSLINE_TOKEN env var", nil
	}

	// 2. macOS Keychain.
	if runtime.GOOS == "darwin" {
		token, err := readKeychain()
		if err == nil {
			return token, "macOS Keychain", nil
		}
	}

	// 3. Linux credentials file.
	token, err := readCredsFile()
	if err == nil {
		return token, "~/.claude/.credentials.json", nil
	}

	return "", "", fmt.Errorf("no credentials found: tried env var, keychain, credentials file")
}

func readKeychain() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "security", "find-generic-password",
		"-s", "Claude Code-credentials", "-w")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("keychain read: %w", err)
	}

	var creds credentials
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(out))), &creds); err != nil {
		return "", fmt.Errorf("keychain parse: %w", err)
	}
	if creds.ClaudeAIOAuth == nil || creds.ClaudeAIOAuth.AccessToken == "" {
		return "", fmt.Errorf("no OAuth token in keychain")
	}

	// Check expiry.
	if creds.ClaudeAIOAuth.ExpiresAt > 0 {
		expiresAt := time.UnixMilli(creds.ClaudeAIOAuth.ExpiresAt)
		if time.Now().After(expiresAt) {
			fmt.Fprintf(os.Stderr, "Warning: token expired at %s\n", expiresAt)
		}
	}

	return creds.ClaudeAIOAuth.AccessToken, nil
}

func readCredsFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(home + "/.claude/.credentials.json")
	if err != nil {
		return "", err
	}

	var creds credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", fmt.Errorf("credentials parse: %w", err)
	}
	if creds.ClaudeAIOAuth == nil || creds.ClaudeAIOAuth.AccessToken == "" {
		return "", fmt.Errorf("no OAuth token in credentials file")
	}

	return creds.ClaudeAIOAuth.AccessToken, nil
}
