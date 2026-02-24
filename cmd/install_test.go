package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallFreshSettings(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, ".claude", "settings.json")

	settings := make(map[string]any)
	settings["statusLine"] = map[string]any{
		"type":    "command",
		"command": "/usr/local/bin/cc-statusline",
		"padding": 0,
	}

	if err := writeSettingsFile(path, settings); err != nil {
		t.Fatalf("writeSettingsFile: %v", err)
	}

	// Verify file was created.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	sl, ok := result["statusLine"].(map[string]any)
	if !ok {
		t.Fatal("statusLine not found or wrong type")
	}
	if sl["type"] != "command" {
		t.Errorf("type = %v, want command", sl["type"])
	}
	if sl["command"] != "/usr/local/bin/cc-statusline" {
		t.Errorf("command = %v", sl["command"])
	}
}

func TestInstallPreservesExistingKeys(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	// Write existing settings with other keys.
	existing := map[string]any{
		"otherKey": "otherValue",
		"nested":   map[string]any{"a": 1},
	}
	data, err := json.Marshal(existing)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	settings, err := readSettingsFile(path)
	if err != nil {
		t.Fatalf("readSettingsFile: %v", err)
	}
	settings["statusLine"] = map[string]any{"type": "command", "command": "/bin/test", "padding": 0}
	if err := writeSettingsFile(path, settings); err != nil {
		t.Fatalf("writeSettingsFile: %v", err)
	}

	result, err := readSettingsFile(path)
	if err != nil {
		t.Fatalf("readSettingsFile after write: %v", err)
	}

	if result["otherKey"] != "otherValue" {
		t.Error("existing key 'otherKey' was not preserved")
	}
	if _, ok := result["statusLine"]; !ok {
		t.Error("statusLine not found after install")
	}
}

func TestUninstallRemovesEntry(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	initial := map[string]any{
		"statusLine": map[string]any{"type": "command", "command": "/bin/test"},
		"otherKey":   "keep",
	}
	data, err := json.Marshal(initial)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	settings, err := readSettingsFile(path)
	if err != nil {
		t.Fatalf("readSettingsFile: %v", err)
	}
	delete(settings, "statusLine")
	if err := writeSettingsFile(path, settings); err != nil {
		t.Fatalf("writeSettingsFile: %v", err)
	}

	result, err := readSettingsFile(path)
	if err != nil {
		t.Fatalf("readSettingsFile after uninstall: %v", err)
	}

	if _, ok := result["statusLine"]; ok {
		t.Error("statusLine should have been removed")
	}
	if result["otherKey"] != "keep" {
		t.Error("other keys should be preserved")
	}
}

func TestUninstallNoFile(t *testing.T) {
	t.Parallel()

	settings, err := readSettingsFile("/nonexistent/settings.json")
	if err != nil {
		t.Fatalf("readSettingsFile for missing file should not error: %v", err)
	}
	if len(settings) != 0 {
		t.Error("missing file should return empty map")
	}
}

func TestReadSettingsFileMissing(t *testing.T) {
	t.Parallel()

	settings, err := readSettingsFile("/no/such/file.json")
	if err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
	if len(settings) != 0 {
		t.Errorf("expected empty map, got %v", settings)
	}
}

func TestVersionOutput(t *testing.T) {
	t.Parallel()

	SetVersionInfo("v0.1.0", "abc1234")
	if buildVersion != "v0.1.0" {
		t.Errorf("buildVersion = %q, want %q", buildVersion, "v0.1.0")
	}
	if buildCommit != "abc1234" {
		t.Errorf("buildCommit = %q, want %q", buildCommit, "abc1234")
	}
}
