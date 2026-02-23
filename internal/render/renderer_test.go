package render_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/donaldgifford/cc-statusline/internal/color"
	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render"
	"github.com/donaldgifford/cc-statusline/internal/render/segments"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
	"github.com/donaldgifford/cc-statusline/internal/statusline"
)

func init() {
	disabled := false
	color.SetEnabled(&disabled)
}

func TestRendererSingleLine(t *testing.T) {
	t.Parallel()

	cfg := render.Config{
		Lines:     [][]string{{"model", "context"}},
		Separator: " ",
		ThemeName: theme.DefaultThemeName,
	}
	r := render.New(cfg, segments.All())

	data := &model.StatusData{
		Model:         model.ModelInfo{DisplayName: "Opus 4.6"},
		ContextWindow: &model.ContextWindow{UsedPercentage: intPtr(43)},
	}

	var buf bytes.Buffer
	if err := r.Render(&buf, data); err != nil {
		t.Fatalf("Render: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	want := "[Opus 4.6] ctx:43%"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRendererMultiLine(t *testing.T) {
	t.Parallel()

	cfg := render.Config{
		Lines: [][]string{
			{"model", "context"},
			{"cost", "duration"},
		},
		Separator: " ",
		ThemeName: theme.DefaultThemeName,
	}
	r := render.New(cfg, segments.All())

	data := &model.StatusData{
		Model:         model.ModelInfo{DisplayName: "Opus 4.6"},
		ContextWindow: &model.ContextWindow{UsedPercentage: intPtr(43)},
		Cost:          &model.CostInfo{TotalCostUSD: 0.23, TotalDurationMS: 720000},
	}

	var buf bytes.Buffer
	if err := r.Render(&buf, data); err != nil {
		t.Fatalf("Render: %v", err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "[Opus 4.6] ctx:43%" {
		t.Errorf("line 0: got %q, want %q", lines[0], "[Opus 4.6] ctx:43%")
	}
	if lines[1] != "$0.23 12m" {
		t.Errorf("line 1: got %q, want %q", lines[1], "$0.23 12m")
	}
}

func TestRendererSegmentOrder(t *testing.T) {
	t.Parallel()

	cfg := render.Config{
		Lines:     [][]string{{"context", "model"}},
		Separator: " | ",
		ThemeName: theme.DefaultThemeName,
	}
	r := render.New(cfg, segments.All())

	data := &model.StatusData{
		Model:         model.ModelInfo{DisplayName: "Sonnet"},
		ContextWindow: &model.ContextWindow{UsedPercentage: intPtr(10)},
	}

	var buf bytes.Buffer
	if err := r.Render(&buf, data); err != nil {
		t.Fatalf("Render: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	want := "ctx:10% | [Sonnet]"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRendererEmptySegmentsOmitted(t *testing.T) {
	t.Parallel()

	cfg := render.Config{
		Lines:     [][]string{{"model", "vim", "context"}},
		Separator: " ",
		ThemeName: theme.DefaultThemeName,
	}
	r := render.New(cfg, segments.All())

	// No vim data, so vim segment should be omitted.
	data := &model.StatusData{
		Model:         model.ModelInfo{DisplayName: "Opus"},
		ContextWindow: &model.ContextWindow{UsedPercentage: intPtr(50)},
	}

	var buf bytes.Buffer
	if err := r.Render(&buf, data); err != nil {
		t.Fatalf("Render: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	want := "[Opus] ctx:50%"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRendererUnknownSegmentSkipped(t *testing.T) {
	t.Parallel()

	cfg := render.Config{
		Lines:     [][]string{{"model", "nonexistent", "context"}},
		Separator: " ",
		ThemeName: theme.DefaultThemeName,
	}
	r := render.New(cfg, segments.All())

	data := &model.StatusData{
		Model:         model.ModelInfo{DisplayName: "Opus"},
		ContextWindow: &model.ContextWindow{UsedPercentage: intPtr(10)},
	}

	var buf bytes.Buffer
	if err := r.Render(&buf, data); err != nil {
		t.Fatalf("Render: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	want := "[Opus] ctx:10%"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRunEndToEnd(t *testing.T) {
	t.Parallel()

	input := `{
		"model": {"id": "claude-opus-4-6", "display_name": "Opus 4.6"},
		"cwd": "/some/path",
		"context_window": {"used_percentage": 43, "context_window_size": 200000}
	}`

	var buf bytes.Buffer
	err := statusline.Run(strings.NewReader(input), &buf, &buf)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	// Should contain model and context.
	if !strings.Contains(got, "Opus 4.6") {
		t.Errorf("output should contain model name, got %q", got)
	}
	if !strings.Contains(got, "43%") {
		t.Errorf("output should contain context percentage, got %q", got)
	}
}

func TestRunEmptyInput(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := statusline.Run(strings.NewReader(""), &buf, &buf)
	if err != nil {
		t.Fatalf("Run with empty input: %v", err)
	}
}

func intPtr(v int) *int { return &v }
