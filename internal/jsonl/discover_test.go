package jsonl

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestDiscoverFiles(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv().

	// Create a temp directory structure mimicking Claude Code's config.
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "projects", "myproject")
	if err := os.MkdirAll(projectDir, 0o750); err != nil {
		t.Fatal(err)
	}

	// Create a JSONL file in the project directory.
	jsonlPath := filepath.Join(projectDir, "transcript.jsonl")
	if err := os.WriteFile(jsonlPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Set CLAUDE_CONFIG_DIR to point to our temp dir.
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	files := DiscoverFiles()
	if !slices.Contains(files, jsonlPath) {
		t.Errorf("DiscoverFiles() did not find %q in results: %v", jsonlPath, files)
	}
}

func TestDiscoverFiles_NoDirs(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv().

	// Point to a nonexistent directory.
	t.Setenv("CLAUDE_CONFIG_DIR", "/nonexistent/path")
	t.Setenv("XDG_CONFIG_HOME", "/nonexistent/xdg")
	t.Setenv("HOME", "/nonexistent/home")

	files := DiscoverFiles()
	if len(files) != 0 {
		t.Errorf("expected no files, got %v", files)
	}
}

func TestConfigDirs(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv().

	t.Setenv("CLAUDE_CONFIG_DIR", "/custom/claude")
	t.Setenv("XDG_CONFIG_HOME", "/custom/xdg")

	dirs := configDirs()
	if len(dirs) < 2 {
		t.Fatalf("expected at least 2 dirs, got %d: %v", len(dirs), dirs)
	}
	if dirs[0] != "/custom/claude" {
		t.Errorf("dirs[0] = %q, want %q", dirs[0], "/custom/claude")
	}
	wantXDG := "/custom/xdg/claude"
	if dirs[1] != wantXDG {
		t.Errorf("dirs[1] = %q, want %q", dirs[1], wantXDG)
	}
}
