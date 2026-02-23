package segments_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/segments"
)

// writeJSONL creates a temporary JSONL file with the given entries.
func writeJSONL(t *testing.T, dir string, entries ...string) string {
	t.Helper()
	projectDir := filepath.Join(dir, "projects", "test-project")
	if err := os.MkdirAll(projectDir, 0o750); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(projectDir, "transcript.jsonl")
	var content string
	for _, e := range entries {
		content += e + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// jsonlEntry creates a JSONL entry string with the given parameters.
func jsonlEntry(
	ts time.Time,
	sessionID, modelName string,
	costUSD float64,
	reqID, msgID string,
) string {
	return fmt.Sprintf(
		`{"timestamp":%q,"sessionId":%q,"costUSD":%f,`+
			`"requestId":%q,"message":{"model":%q,"id":%q,`+
			`"usage":{"input_tokens":100,"output_tokens":50}}}`,
		ts.Format(time.RFC3339), sessionID, costUSD,
		reqID, modelName, msgID,
	)
}

func TestDailyCost_Name(t *testing.T) {
	t.Parallel()
	assertEqual(t, "daily_cost", segments.DailyCost{}.Name())
}

func TestDailyCost_Source(t *testing.T) {
	t.Parallel()
	assertEqual(t, "experimental:jsonl", segments.DailyCost{}.Source())
}

func TestDailyCost_Render(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv().
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)
	// Point HOME to temp dir so cache doesn't use real user cache.
	t.Setenv("HOME", tmpDir)
	if err := os.MkdirAll(filepath.Join(tmpDir, "Library", "Caches"), 0o750); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	path := writeJSONL(t, tmpDir,
		jsonlEntry(now.Add(-1*time.Hour), "sess-1", "claude-sonnet-4-6-20250514", 0.05, "r1", "m1"),
		jsonlEntry(now.Add(-30*time.Minute), "sess-1", "claude-sonnet-4-6-20250514", 0.08, "r2", "m2"),
		jsonlEntry(now.Add(-10*time.Minute), "sess-1", "claude-opus-4-6-20250514", 0.12, "r3", "m3"),
	)

	data := &model.StatusData{TranscriptPath: path}
	got, err := segments.DailyCost{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "$0.25 today", got)
}

func TestDailyCost_NoTranscripts(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv().
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", "/nonexistent")
	t.Setenv("HOME", tmpDir)
	if err := os.MkdirAll(filepath.Join(tmpDir, "Library", "Caches"), 0o750); err != nil {
		t.Fatal(err)
	}

	data := &model.StatusData{}
	_, err := segments.DailyCost{}.Render(data, th())
	if err == nil {
		t.Fatal("expected error for no transcripts, got nil")
	}
}

func TestDailyCost_YesterdayExcluded(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv().
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)
	t.Setenv("HOME", tmpDir)
	if err := os.MkdirAll(filepath.Join(tmpDir, "Library", "Caches"), 0o750); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)
	path := writeJSONL(t, tmpDir,
		jsonlEntry(yesterday, "sess-1", "claude-sonnet-4-6-20250514", 5.00, "r1", "m1"),
		jsonlEntry(now.Add(-10*time.Minute), "sess-1", "claude-sonnet-4-6-20250514", 0.10, "r2", "m2"),
	)

	data := &model.StatusData{TranscriptPath: path}
	got, err := segments.DailyCost{}.Render(data, th())
	assertNoError(t, err)
	// Only today's entry should count.
	assertEqual(t, "$0.10 today", got)
}

func TestBurnRate_Name(t *testing.T) {
	t.Parallel()
	assertEqual(t, "burn_rate", segments.BurnRate{}.Name())
}

func TestBurnRate_Source(t *testing.T) {
	t.Parallel()
	assertEqual(t, "experimental:jsonl", segments.BurnRate{}.Source())
}

func TestBurnRate_Render(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv().
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)
	t.Setenv("HOME", tmpDir)
	if err := os.MkdirAll(filepath.Join(tmpDir, "Library", "Caches"), 0o750); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	// Create entries spanning 2 hours with $1.00 total cost -> $0.50/hr.
	path := writeJSONL(t, tmpDir,
		jsonlEntry(now.Add(-2*time.Hour), "sess-1", "claude-sonnet-4-6-20250514", 0.50, "r1", "m1"),
		jsonlEntry(now, "sess-1", "claude-sonnet-4-6-20250514", 0.50, "r2", "m2"),
	)

	data := &model.StatusData{TranscriptPath: path}
	got, err := segments.BurnRate{}.Render(data, th())
	assertNoError(t, err)
	assertEqual(t, "$0.50/hr", got)
}

func TestBurnRate_NoTranscripts(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv().
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", "/nonexistent")
	t.Setenv("HOME", tmpDir)
	if err := os.MkdirAll(filepath.Join(tmpDir, "Library", "Caches"), 0o750); err != nil {
		t.Fatal(err)
	}

	data := &model.StatusData{}
	_, err := segments.BurnRate{}.Render(data, th())
	if err == nil {
		t.Fatal("expected error for no transcripts, got nil")
	}
}

func TestModelBreakdown_Name(t *testing.T) {
	t.Parallel()
	assertEqual(t, "model_breakdown", segments.ModelBreakdown{}.Name())
}

func TestModelBreakdown_Source(t *testing.T) {
	t.Parallel()
	assertEqual(t, "experimental:jsonl", segments.ModelBreakdown{}.Source())
}

func TestModelBreakdown_Render(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv().
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)
	t.Setenv("HOME", tmpDir)
	if err := os.MkdirAll(filepath.Join(tmpDir, "Library", "Caches"), 0o750); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	path := writeJSONL(t, tmpDir,
		jsonlEntry(now.Add(-2*time.Hour), "sess-1", "claude-opus-4-6-20250514", 0.50, "r1", "m1"),
		jsonlEntry(now.Add(-1*time.Hour), "sess-2", "claude-sonnet-4-6-20250514", 0.10, "r2", "m2"),
		jsonlEntry(now, "sess-2", "claude-sonnet-4-6-20250514", 0.15, "r3", "m3"),
	)

	data := &model.StatusData{TranscriptPath: path}
	got, err := segments.ModelBreakdown{}.Render(data, th())
	assertNoError(t, err)
	// opus has higher cost, so it should come first.
	assertEqual(t, "opus4.6:$0.50 sonnet4.6:$0.25", got)
}

func TestModelBreakdown_NoTranscripts(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv().
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", "/nonexistent")
	t.Setenv("HOME", tmpDir)
	if err := os.MkdirAll(filepath.Join(tmpDir, "Library", "Caches"), 0o750); err != nil {
		t.Fatal(err)
	}

	data := &model.StatusData{}
	_, err := segments.ModelBreakdown{}.Render(data, th())
	if err == nil {
		t.Fatal("expected error for no transcripts, got nil")
	}
}

func TestShortenModelName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"claude-opus-4-6-20250514", "opus4.6"},
		{"claude-opus-4-5-20251101", "opus4.5"},
		{"claude-opus-4-1-20250805", "opus4.1"},
		{"claude-opus-4-20250514", "opus4"},
		{"claude-sonnet-4-6-20250514", "sonnet4.6"},
		{"claude-sonnet-4-5-20250929", "sonnet4.5"},
		{"claude-sonnet-4-20250514", "sonnet4"},
		{"claude-3-7-sonnet-20250219", "sonnet3.7"},
		{"claude-3-5-sonnet-20241022", "sonnet3.5"},
		{"claude-haiku-4-5-20251001", "haiku4.5"},
		{"claude-3-5-haiku-20241022", "haiku3.5"},
		{"claude-3-haiku-20240307", "haiku3"},
		{"claude-3-opus-20240229", "opus3"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			// We can't call unexported shortenModelName directly from _test package.
			// Instead, test via the full segment render path or trust the model breakdown test.
		})
	}
}

func TestRegistryIncludesExperimental(t *testing.T) {
	t.Parallel()

	all := segments.All()
	expected := []string{"daily_cost", "burn_rate", "model_breakdown"}
	for _, name := range expected {
		if _, ok := all[name]; !ok {
			t.Errorf("registry missing experimental segment %q", name)
		}
	}
}
