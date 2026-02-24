// Package jsonl parses Claude Code JSONL transcript files.
package jsonl

import (
	"bufio"
	"encoding/json"
	"os"
	"time"
)

// Entry represents a single line in a Claude Code JSONL transcript.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	SessionID string    `json:"sessionId"`
	Version   string    `json:"version"`
	CostUSD   float64   `json:"costUSD"`
	RequestID string    `json:"requestId"`
	CWD       string    `json:"cwd"`
	Message   *Message  `json:"message"`
}

// Message holds the API response data within a JSONL entry.
type Message struct {
	Model string `json:"model"`
	ID    string `json:"id"`
	Usage *Usage `json:"usage"`
}

// Usage holds token counts from the API response.
type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// ReadFile reads and parses a JSONL transcript file, returning valid entries.
// Invalid lines are silently skipped. Duplicates are removed by message ID
// and request ID.
func ReadFile(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck // read-only file

	seen := make(map[string]struct{})
	var entries []Entry

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		// Skip entries missing required fields.
		if entry.Timestamp.IsZero() {
			continue
		}
		if entry.Message == nil || entry.Message.Usage == nil {
			continue
		}

		// Deduplicate by message ID + request ID.
		key := dedupKey(&entry)
		if key != "" {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
		}

		entries = append(entries, entry)
	}

	return entries, scanner.Err()
}

func dedupKey(e *Entry) string {
	msgID := ""
	if e.Message != nil {
		msgID = e.Message.ID
	}
	if msgID == "" && e.RequestID == "" {
		return ""
	}
	return msgID + "|" + e.RequestID
}
