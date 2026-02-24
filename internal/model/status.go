// Package model defines Go types for the Claude Code stdin JSON payload.
package model

// StatusData represents the full JSON payload piped to stdin by Claude Code.
// Fields marked as pointers are conditionally absent or nullable in the JSON.
type StatusData struct {
	SessionID      string         `json:"session_id"`
	TranscriptPath string         `json:"transcript_path"`
	CWD            string         `json:"cwd"`
	Model          ModelInfo      `json:"model"`
	Workspace      WorkspaceInfo  `json:"workspace"`
	Version        string         `json:"version"`
	OutputStyle    *OutputStyle   `json:"output_style,omitempty"`
	Cost           *CostInfo      `json:"cost,omitempty"`
	ContextWindow  *ContextWindow `json:"context_window,omitempty"`
	Exceeds200K    bool           `json:"exceeds_200k_tokens"`
	Vim            *VimInfo       `json:"vim,omitempty"`
	Agent          *AgentInfo     `json:"agent,omitempty"`
}

// ModelInfo identifies the Claude model for the current session.
type ModelInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// WorkspaceInfo describes the directories Claude Code is operating in.
type WorkspaceInfo struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

// OutputStyle describes the Claude Code output style.
type OutputStyle struct {
	Name string `json:"name"`
}

// CostInfo contains cumulative session cost and activity metrics.
type CostInfo struct {
	TotalCostUSD       float64 `json:"total_cost_usd"`
	TotalDurationMS    int64   `json:"total_duration_ms"`
	TotalAPIDurationMS int64   `json:"total_api_duration_ms"`
	TotalLinesAdded    int     `json:"total_lines_added"`
	TotalLinesRemoved  int     `json:"total_lines_removed"`
}

// ContextWindow describes token usage and context window capacity.
type ContextWindow struct {
	TotalInputTokens    int           `json:"total_input_tokens"`
	TotalOutputTokens   int           `json:"total_output_tokens"`
	ContextWindowSize   int           `json:"context_window_size"`
	UsedPercentage      *int          `json:"used_percentage"`
	RemainingPercentage *int          `json:"remaining_percentage"`
	CurrentUsage        *CurrentUsage `json:"current_usage"`
}

// CurrentUsage contains token counts from the most recent API call.
// This is null before the first API call in a session.
type CurrentUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// VimInfo is present only when vim mode is enabled in Claude Code.
type VimInfo struct {
	Mode string `json:"mode"`
}

// AgentInfo is present only when Claude Code is running with --agent.
type AgentInfo struct {
	Name string `json:"name"`
}
