package segments_test

import (
	"testing"

	"github.com/donaldgifford/cc-statusline/internal/color"
	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/segments"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

func init() {
	// Disable color for deterministic test output.
	disabled := false
	color.SetEnabled(&disabled)
}

func th() *theme.Theme { return theme.Get("tokyo-night") }

func intPtr(v int) *int { return &v }

func TestCWDEmpty(t *testing.T) {
	t.Parallel()
	got, err := segments.CWD{}.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestCWDPath(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{CWD: "/some/path"}
	got, err := segments.CWD{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "/some/path", got)
}

func TestModelEmpty(t *testing.T) {
	t.Parallel()
	got, err := segments.Model{}.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestModelPresent(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{Model: model.ModelInfo{DisplayName: "Opus 4.6"}}
	got, err := segments.Model{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "[Opus 4.6]", got)
}

func TestContextNil(t *testing.T) {
	t.Parallel()
	got, err := segments.Context{}.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "ctx:--", got)
}

func TestContextNullPercentage(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{ContextWindow: &model.ContextWindow{}}
	got, err := segments.Context{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "ctx:--", got)
}

func TestContextValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pct  int
		want string
	}{
		{"zero", 0, "ctx:0%"},
		{"low", 30, "ctx:30%"},
		{"mid", 50, "ctx:50%"},
		{"high", 80, "ctx:80%"},
		{"critical", 95, "ctx:95%"},
		{"full", 100, "ctx:100%"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := &model.StatusData{
				ContextWindow: &model.ContextWindow{
					UsedPercentage: intPtr(tt.pct),
				},
			}
			got, err := segments.Context{}.Render(data, th())
			assertNoError(t, err)
			assertEqual(t, tt.want, got)
		})
	}
}

func TestCostNil(t *testing.T) {
	t.Parallel()
	got, err := segments.Cost{}.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestCostZero(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{Cost: &model.CostInfo{TotalCostUSD: 0}}
	got, err := segments.Cost{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "$0.00", got)
}

func TestCostValue(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{Cost: &model.CostInfo{TotalCostUSD: 1.234}}
	got, err := segments.Cost{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "$1.23", got)
}

func TestDurationNil(t *testing.T) {
	t.Parallel()
	got, err := segments.Duration{}.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestDurationZero(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{Cost: &model.CostInfo{TotalDurationMS: 0}}
	got, err := segments.Duration{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestDurationMinutes(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{Cost: &model.CostInfo{TotalDurationMS: 720000}}
	got, err := segments.Duration{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "12m", got)
}

func TestDurationHours(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{Cost: &model.CostInfo{TotalDurationMS: 4980000}}
	got, err := segments.Duration{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "1h23m", got)
}

func TestTokensNil(t *testing.T) {
	t.Parallel()
	got, err := segments.Tokens{}.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestTokensZero(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{ContextWindow: &model.ContextWindow{}}
	got, err := segments.Tokens{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestTokensFormatted(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{
		ContextWindow: &model.ContextWindow{
			TotalInputTokens:  15234,
			TotalOutputTokens: 4521,
		},
	}
	got, err := segments.Tokens{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "15k/4k tokens", got)
}

func TestTokensSmall(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{
		ContextWindow: &model.ContextWindow{
			TotalInputTokens:  500,
			TotalOutputTokens: 100,
		},
	}
	got, err := segments.Tokens{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "500/100 tokens", got)
}

func TestLinesNil(t *testing.T) {
	t.Parallel()
	got, err := segments.Lines{}.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestLinesZero(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{Cost: &model.CostInfo{}}
	got, err := segments.Lines{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestLinesPresent(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{
		Cost: &model.CostInfo{TotalLinesAdded: 156, TotalLinesRemoved: 23},
	}
	got, err := segments.Lines{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "+156 -23", got)
}

func TestVimNil(t *testing.T) {
	t.Parallel()
	got, err := segments.Vim{}.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestVimNormal(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{Vim: &model.VimInfo{Mode: "NORMAL"}}
	got, err := segments.Vim{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "NORMAL", got)
}

func TestVimInsert(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{Vim: &model.VimInfo{Mode: "INSERT"}}
	got, err := segments.Vim{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "INSERT", got)
}

func TestAgentNil(t *testing.T) {
	t.Parallel()
	got, err := segments.Agent{}.Render(&model.StatusData{}, th())
	assertNoError(t, err)
	assertEqual(t, "", got)
}

func TestAgentPresent(t *testing.T) {
	t.Parallel()
	data := &model.StatusData{Agent: &model.AgentInfo{Name: "security-reviewer"}}
	got, err := segments.Agent{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "security-reviewer", got)
}

func TestRegistryAll(t *testing.T) {
	t.Parallel()

	all := segments.All()
	expected := []string{
		"cwd", "git_branch", "model", "context", "cost",
		"duration", "tokens", "lines", "vim", "agent",
		"daily_cost", "burn_rate", "model_breakdown",
	}
	for _, name := range expected {
		if _, ok := all[name]; !ok {
			t.Errorf("registry missing segment %q", name)
		}
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertEqual(t *testing.T, want, got string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
