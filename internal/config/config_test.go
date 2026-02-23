package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/donaldgifford/cc-statusline/internal/config"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

func TestDefaultValues(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	if cfg.Theme != theme.DefaultThemeName {
		t.Errorf("Theme = %q, want %q", cfg.Theme, theme.DefaultThemeName)
	}
	if cfg.Separator != " " {
		t.Errorf("Separator = %q, want %q", cfg.Separator, " ")
	}
	if cfg.Color != nil {
		t.Error("Color should be nil by default")
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load("/nonexistent/path/.cc-statusline.yaml")
	if err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
	if cfg.Theme != theme.DefaultThemeName {
		t.Errorf("Theme = %q, want default", cfg.Theme)
	}
}

func TestLoadFullConfig(t *testing.T) {
	t.Parallel()

	content := `
theme: rose-pine
separator: " | "
color: false
lines:
  - [cwd, model]
  - [cost, duration]
experimental:
  jsonl: true
  usage_api: false
`
	path := writeTemp(t, content)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Theme != "rose-pine" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "rose-pine")
	}
	if cfg.Separator != " | " {
		t.Errorf("Separator = %q, want %q", cfg.Separator, " | ")
	}
	if cfg.Color == nil || *cfg.Color {
		t.Error("Color should be false")
	}
	if len(cfg.Lines) != 2 {
		t.Fatalf("Lines count = %d, want 2", len(cfg.Lines))
	}
	if cfg.Lines[0][0] != "cwd" || cfg.Lines[0][1] != "model" {
		t.Errorf("Lines[0] = %v, want [cwd, model]", cfg.Lines[0])
	}
	if !cfg.Experimental.JSONL {
		t.Error("Experimental.JSONL should be true")
	}
	if cfg.Experimental.UsageAPI {
		t.Error("Experimental.UsageAPI should be false")
	}
}

func TestLoadSegmentsShorthand(t *testing.T) {
	t.Parallel()

	content := `
segments: [model, context]
`
	path := writeTemp(t, content)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	lines := cfg.ResolvedLines()
	if len(lines) != 1 {
		t.Fatalf("ResolvedLines count = %d, want 1", len(lines))
	}
	if len(lines[0]) != 2 || lines[0][0] != "model" || lines[0][1] != "context" {
		t.Errorf("ResolvedLines[0] = %v, want [model, context]", lines[0])
	}
}

func TestLoadMalformedYAML(t *testing.T) {
	t.Parallel()

	path := writeTemp(t, "theme: [invalid yaml")
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("malformed YAML should error")
	}
}

func TestResolvedLinesDefault(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	lines := cfg.ResolvedLines()
	if len(lines) != 1 {
		t.Fatalf("default ResolvedLines count = %d, want 1", len(lines))
	}
	if len(lines[0]) != len(config.DefaultSegments) {
		t.Errorf("default line has %d segments, want %d", len(lines[0]), len(config.DefaultSegments))
	}
}

func TestResolvedLinesPreference(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Lines:    [][]string{{"a", "b"}, {"c"}},
		Segments: []string{"x", "y"},
	}
	// Lines takes precedence over Segments.
	lines := cfg.ResolvedLines()
	if len(lines) != 2 {
		t.Fatalf("Lines should take precedence, got %d lines", len(lines))
	}
}

func TestColorEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		color *bool
		want  bool
	}{
		{"nil defaults to true", nil, true},
		{"explicit true", boolPtr(true), true},
		{"explicit false", boolPtr(false), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &config.Config{Color: tt.color}
			got := cfg.ColorEnabled()
			if got != tt.want {
				t.Errorf("ColorEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

func boolPtr(v bool) *bool { return &v }
