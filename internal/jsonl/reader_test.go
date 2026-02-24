package jsonl

import (
	"path/filepath"
	"testing"
	"time"
)

func TestReadFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fixture   string
		wantCount int
		wantErr   bool
		validate  func(t *testing.T, entries []Entry)
	}{
		{
			name:      "valid entries",
			fixture:   "valid.jsonl",
			wantCount: 3,
			validate: func(t *testing.T, entries []Entry) {
				t.Helper()
				if entries[0].SessionID != "sess-001" {
					t.Errorf("entry[0].SessionID = %q, want %q", entries[0].SessionID, "sess-001")
				}
				if entries[0].CostUSD != 0.05 {
					t.Errorf("entry[0].CostUSD = %f, want %f", entries[0].CostUSD, 0.05)
				}
				if entries[0].Message.Model != "claude-sonnet-4-6-20250514" {
					t.Errorf("entry[0].Message.Model = %q, want %q", entries[0].Message.Model, "claude-sonnet-4-6-20250514")
				}
				if entries[0].Message.Usage.InputTokens != 1000 {
					t.Errorf("entry[0].InputTokens = %d, want %d", entries[0].Message.Usage.InputTokens, 1000)
				}
				if entries[2].Message.Model != "claude-opus-4-6-20250514" {
					t.Errorf("entry[2].Message.Model = %q, want %q", entries[2].Message.Model, "claude-opus-4-6-20250514")
				}
			},
		},
		{
			name:      "malformed lines skipped",
			fixture:   "mixed.jsonl",
			wantCount: 3,
			validate: func(t *testing.T, entries []Entry) {
				t.Helper()
				// Should have 3 valid entries, skipping the invalid JSON line,
				// the line with no timestamp structure, and the null message line.
				if entries[0].RequestID != "req-001" {
					t.Errorf("entry[0].RequestID = %q, want %q", entries[0].RequestID, "req-001")
				}
				if entries[1].RequestID != "req-002" {
					t.Errorf("entry[1].RequestID = %q, want %q", entries[1].RequestID, "req-002")
				}
				if entries[2].RequestID != "req-003" {
					t.Errorf("entry[2].RequestID = %q, want %q", entries[2].RequestID, "req-003")
				}
			},
		},
		{
			name:      "duplicates removed",
			fixture:   "duplicates.jsonl",
			wantCount: 2,
			validate: func(t *testing.T, entries []Entry) {
				t.Helper()
				if entries[0].Message.ID != "msg-001" {
					t.Errorf("entry[0].Message.ID = %q, want %q", entries[0].Message.ID, "msg-001")
				}
				if entries[1].Message.ID != "msg-002" {
					t.Errorf("entry[1].Message.ID = %q, want %q", entries[1].Message.ID, "msg-002")
				}
			},
		},
		{
			name:      "empty file",
			fixture:   "empty.jsonl",
			wantCount: 0,
		},
		{
			name:      "multiple sessions",
			fixture:   "multi_session.jsonl",
			wantCount: 4,
			validate: func(t *testing.T, entries []Entry) {
				t.Helper()
				sessions := make(map[string]int)
				for _, e := range entries {
					sessions[e.SessionID]++
				}
				if sessions["sess-001"] != 2 {
					t.Errorf("sess-001 count = %d, want 2", sessions["sess-001"])
				}
				if sessions["sess-002"] != 2 {
					t.Errorf("sess-002 count = %d, want 2", sessions["sess-002"])
				}
			},
		},
		{
			name:    "nonexistent file",
			fixture: "nonexistent.jsonl",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join("..", "..", "testdata", tt.fixture)

			entries, err := ReadFile(path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(entries) != tt.wantCount {
				t.Fatalf("got %d entries, want %d", len(entries), tt.wantCount)
			}
			if tt.validate != nil {
				tt.validate(t, entries)
			}
		})
	}
}

func TestReadFile_TimestampOrder(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "testdata", "multi_session.jsonl")
	entries, err := ReadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Entries should be in file order (which is chronological in our fixture).
	for i := 1; i < len(entries); i++ {
		if entries[i].Timestamp.Before(entries[i-1].Timestamp) {
			t.Errorf("entry[%d].Timestamp (%v) is before entry[%d].Timestamp (%v)",
				i, entries[i].Timestamp, i-1, entries[i-1].Timestamp)
		}
	}
}

func TestReadFile_CacheTokenFields(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "testdata", "valid.jsonl")
	entries, err := ReadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) < 1 {
		t.Fatal("expected at least 1 entry")
	}

	usage := entries[0].Message.Usage
	if usage.CacheCreationInputTokens != 200 {
		t.Errorf("CacheCreationInputTokens = %d, want 200", usage.CacheCreationInputTokens)
	}
	if usage.CacheReadInputTokens != 100 {
		t.Errorf("CacheReadInputTokens = %d, want 100", usage.CacheReadInputTokens)
	}
}

func TestDedupKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		entry   *Entry
		wantKey string
	}{
		{
			name: "both IDs present",
			entry: &Entry{
				RequestID: "req-1",
				Message:   &Message{ID: "msg-1"},
			},
			wantKey: "msg-1|req-1",
		},
		{
			name: "only request ID",
			entry: &Entry{
				RequestID: "req-1",
				Message:   &Message{},
			},
			wantKey: "|req-1",
		},
		{
			name: "only message ID",
			entry: &Entry{
				Message: &Message{ID: "msg-1"},
			},
			wantKey: "msg-1|",
		},
		{
			name:    "neither ID",
			entry:   &Entry{Timestamp: time.Now()},
			wantKey: "",
		},
		{
			name: "nil message",
			entry: &Entry{
				RequestID: "req-1",
			},
			wantKey: "|req-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := dedupKey(tt.entry)
			if got != tt.wantKey {
				t.Errorf("dedupKey() = %q, want %q", got, tt.wantKey)
			}
		})
	}
}
