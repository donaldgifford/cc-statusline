package usageapi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadCredentials_EnvVar(t *testing.T) {
	t.Setenv("CC_STATUSLINE_TOKEN", "test-token-123")

	result, err := ReadCredentials(AuthConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "test-token-123" {
		t.Errorf("AccessToken = %q, want %q", result.AccessToken, "test-token-123")
	}
	if result.Source != "CC_STATUSLINE_TOKEN" {
		t.Errorf("Source = %q, want %q", result.Source, "CC_STATUSLINE_TOKEN")
	}
}

func TestReadCredentials_EnvVarPriority(t *testing.T) {
	// Env var should take priority even when credentials file exists.
	tmpDir := t.TempDir()
	t.Setenv("CC_STATUSLINE_TOKEN", "env-token")
	t.Setenv("HOME", tmpDir)

	writeTestCredentials(t, filepath.Join(tmpDir, ".claude", ".credentials.json"), "file-token", "", 0)

	result, err := ReadCredentials(AuthConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "env-token" {
		t.Errorf("AccessToken = %q, want %q (env var should win)", result.AccessToken, "env-token")
	}
}

func TestReadCredentials_KeychainMocked(t *testing.T) {
	t.Setenv("CC_STATUSLINE_TOKEN", "")

	mockRunner := func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		creds := storedCredentials{
			ClaudeAIOAuth: &oauthToken{
				AccessToken:  "keychain-token",
				RefreshToken: "refresh-123",
				ExpiresAt:    time.Now().Add(time.Hour).UnixMilli(),
			},
		}
		data, _ := json.Marshal(creds) //nolint:errchkjson // test helper
		return data, nil
	}

	result, err := ReadCredentials(AuthConfig{RunCommand: mockRunner})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "keychain-token" {
		t.Errorf("AccessToken = %q, want %q", result.AccessToken, "keychain-token")
	}
	if result.RefreshToken != "refresh-123" {
		t.Errorf("RefreshToken = %q, want %q", result.RefreshToken, "refresh-123")
	}
	if result.Source != "macOS Keychain" {
		t.Errorf("Source = %q, want %q", result.Source, "macOS Keychain")
	}
	if result.Expired {
		t.Error("Expired = true, want false")
	}
}

func TestReadCredentials_KeychainExpired(t *testing.T) {
	t.Setenv("CC_STATUSLINE_TOKEN", "")

	mockRunner := func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		creds := storedCredentials{
			ClaudeAIOAuth: &oauthToken{
				AccessToken:  "expired-token",
				RefreshToken: "refresh-456",
				ExpiresAt:    time.Now().Add(-time.Hour).UnixMilli(),
			},
		}
		data, _ := json.Marshal(creds) //nolint:errchkjson // test helper
		return data, nil
	}

	result, err := ReadCredentials(AuthConfig{RunCommand: mockRunner})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Expired {
		t.Error("Expired = false, want true")
	}
}

func TestReadCredentials_CredentialsFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CC_STATUSLINE_TOKEN", "")
	t.Setenv("HOME", tmpDir)

	writeTestCredentials(t, filepath.Join(tmpDir, ".claude", ".credentials.json"),
		"file-token", "file-refresh", time.Now().Add(time.Hour).UnixMilli())

	// Mock keychain failure.
	failRunner := func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return nil, fmt.Errorf("keychain not available")
	}

	result, err := ReadCredentials(AuthConfig{RunCommand: failRunner})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "file-token" {
		t.Errorf("AccessToken = %q, want %q", result.AccessToken, "file-token")
	}
	if result.Source != "~/.claude/.credentials.json" {
		t.Errorf("Source = %q, want %q", result.Source, "~/.claude/.credentials.json")
	}
}

func TestReadCredentials_ManualAuth(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CC_STATUSLINE_TOKEN", "")
	t.Setenv("HOME", tmpDir)
	// On macOS, UserConfigDir returns $HOME/Library/Application Support.
	configDir := filepath.Join(tmpDir, "Library", "Application Support")
	writeTestCredentials(t, filepath.Join(configDir, "cc-statusline", "auth.json"),
		"manual-token", "", 0)

	// Mock keychain failure.
	failRunner := func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return nil, fmt.Errorf("keychain not available")
	}

	result, err := ReadCredentials(AuthConfig{RunCommand: failRunner})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "manual-token" {
		t.Errorf("AccessToken = %q, want %q", result.AccessToken, "manual-token")
	}
}

func TestReadCredentials_NoSources(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CC_STATUSLINE_TOKEN", "")
	t.Setenv("HOME", tmpDir)

	failRunner := func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return nil, fmt.Errorf("keychain not available")
	}

	_, err := ReadCredentials(AuthConfig{RunCommand: failRunner})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !isNoCredentials(err) {
		t.Errorf("expected ErrNoCredentials, got: %v", err)
	}
}

func TestReadCredentials_NoRunner(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CC_STATUSLINE_TOKEN", "")
	t.Setenv("HOME", tmpDir)

	// No runner configured — keychain should be skipped.
	_, err := ReadCredentials(AuthConfig{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseCredentials_EmptyToken(t *testing.T) {
	t.Parallel()

	raw := `{"claudeAiOauth":{"accessToken":"","refreshToken":""}}`
	_, err := parseCredentials(raw, "test")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestParseCredentials_NilOAuth(t *testing.T) {
	t.Parallel()

	raw := `{"otherField": "value"}`
	_, err := parseCredentials(raw, "test")
	if err == nil {
		t.Fatal("expected error for nil OAuth")
	}
}

func TestParseCredentials_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := parseCredentials("not json", "test")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func writeTestCredentials(t *testing.T, path, accessToken, refreshToken string, expiresAt int64) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	creds := storedCredentials{
		ClaudeAIOAuth: &oauthToken{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    expiresAt,
		},
	}
	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func isNoCredentials(err error) bool {
	return err != nil && err.Error() != "" && fmt.Sprintf("%v", err) != ""
}
