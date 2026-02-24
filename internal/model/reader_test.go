package model_test

import (
	"os"
	"strings"
	"testing"

	"github.com/donaldgifford/cc-statusline/internal/model"
)

func TestReadStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fixture  string // path relative to testdata/
		wantErr  bool
		validate func(t *testing.T, data *model.StatusData)
	}{
		{
			name:    "all fields present",
			fixture: "basic.json",
			validate: func(t *testing.T, data *model.StatusData) {
				t.Helper()
				assertEqual(t, "session_id", "abc123-def456-ghi789", data.SessionID)
				assertEqual(t, "model.id", "claude-opus-4-6", data.Model.ID)
				assertEqual(t, "model.display_name", "Opus 4.6", data.Model.DisplayName)
				assertEqual(t, "cwd", "/Users/test/code/my-project", data.CWD)
				assertEqual(t, "version", "1.0.88", data.Version)
				assertEqual(t, "workspace.current_dir", "/Users/test/code/my-project", data.Workspace.CurrentDir)
				assertEqual(t, "workspace.project_dir", "/Users/test/code/my-project", data.Workspace.ProjectDir)

				if data.Cost == nil {
					t.Fatal("cost should not be nil")
				}
				assertFloat(t, "cost.total_cost_usd", 0.23, data.Cost.TotalCostUSD)
				assertInt64(t, "cost.total_duration_ms", 45000, data.Cost.TotalDurationMS)
				assertInt(t, "cost.total_lines_added", 156, data.Cost.TotalLinesAdded)
				assertInt(t, "cost.total_lines_removed", 23, data.Cost.TotalLinesRemoved)

				if data.ContextWindow == nil {
					t.Fatal("context_window should not be nil")
				}
				assertInt(t, "context_window.used_percentage", 43, *data.ContextWindow.UsedPercentage)
				assertInt(t, "context_window.remaining_percentage", 57, *data.ContextWindow.RemainingPercentage)
				assertInt(t, "context_window.context_window_size", 200000, data.ContextWindow.ContextWindowSize)

				if data.ContextWindow.CurrentUsage == nil {
					t.Fatal("current_usage should not be nil")
				}
				assertInt(t, "current_usage.input_tokens", 8500, data.ContextWindow.CurrentUsage.InputTokens)

				if data.Vim == nil {
					t.Fatal("vim should not be nil")
				}
				assertEqual(t, "vim.mode", "NORMAL", data.Vim.Mode)

				if data.Agent == nil {
					t.Fatal("agent should not be nil")
				}
				assertEqual(t, "agent.name", "security-reviewer", data.Agent.Name)
			},
		},
		{
			name:    "minimal fields only",
			fixture: "minimal.json",
			validate: func(t *testing.T, data *model.StatusData) {
				t.Helper()
				assertEqual(t, "session_id", "min-session-001", data.SessionID)
				assertEqual(t, "model.id", "claude-sonnet-4-20250514", data.Model.ID)
				assertEqual(t, "model.display_name", "Sonnet 4", data.Model.DisplayName)

				if data.Cost != nil {
					t.Error("cost should be nil for minimal fixture")
				}
				if data.Vim != nil {
					t.Error("vim should be nil for minimal fixture")
				}
				if data.Agent != nil {
					t.Error("agent should be nil for minimal fixture")
				}
			},
		},
		{
			name:    "null values",
			fixture: "nulls.json",
			validate: func(t *testing.T, data *model.StatusData) {
				t.Helper()
				if data.ContextWindow == nil {
					t.Fatal("context_window should not be nil")
				}
				if data.ContextWindow.UsedPercentage != nil {
					t.Errorf("used_percentage should be nil, got %d", *data.ContextWindow.UsedPercentage)
				}
				if data.ContextWindow.RemainingPercentage != nil {
					t.Errorf("remaining_percentage should be nil, got %d", *data.ContextWindow.RemainingPercentage)
				}
				if data.ContextWindow.CurrentUsage != nil {
					t.Error("current_usage should be nil")
				}
			},
		},
		{
			name:    "empty object",
			fixture: "empty.json",
			validate: func(t *testing.T, data *model.StatusData) {
				t.Helper()
				assertEqual(t, "session_id", "", data.SessionID)
				assertEqual(t, "model.id", "", data.Model.ID)
				if data.Cost != nil {
					t.Error("cost should be nil for empty object")
				}
			},
		},
		{
			name:    "malformed JSON",
			fixture: "malformed.json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := os.Open("../../testdata/" + tt.fixture)
			if err != nil {
				t.Fatalf("open fixture: %v", err)
			}
			defer f.Close()

			data, err := model.ReadStatus(f)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, data)
			}
		})
	}
}

func TestReadStatusEmptyInput(t *testing.T) {
	t.Parallel()

	data, err := model.ReadStatus(strings.NewReader(""))
	if err != nil {
		t.Fatalf("empty input should not error: %v", err)
	}
	if data.SessionID != "" {
		t.Errorf("expected empty session_id, got %q", data.SessionID)
	}
}

func TestReadStatusUnknownFields(t *testing.T) {
	t.Parallel()

	input := `{"session_id":"test","unknown_field":"value","nested":{"deep":true}}`
	data, err := model.ReadStatus(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unknown fields should be silently ignored: %v", err)
	}
	assertEqual(t, "session_id", "test", data.SessionID)
}

func assertEqual(t *testing.T, field, want, got string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: want %q, got %q", field, want, got)
	}
}

func assertInt(t *testing.T, field string, want, got int) {
	t.Helper()
	if got != want {
		t.Errorf("%s: want %d, got %d", field, want, got)
	}
}

func assertInt64(t *testing.T, field string, want, got int64) {
	t.Helper()
	if got != want {
		t.Errorf("%s: want %d, got %d", field, want, got)
	}
}

func assertFloat(t *testing.T, field string, want, got float64) {
	t.Helper()
	if got != want {
		t.Errorf("%s: want %f, got %f", field, want, got)
	}
}
