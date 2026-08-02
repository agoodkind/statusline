// Package statuspayload decodes the Claude Code status-line JSON payload.
package statuspayload

import (
	"encoding/json"
	"fmt"
)

// percentageScale converts a documented 0-to-100 percentage into a 0-to-1 ratio.
const percentageScale = 100.0

// Payload is the documented Claude Code status-line JSON sent on stdin.
type Payload struct {
	CWD               string        `json:"cwd"`
	SessionID         string        `json:"session_id"`
	SessionName       string        `json:"session_name"`
	TranscriptPath    string        `json:"transcript_path"`
	Model             Model         `json:"model"`
	Workspace         Workspace     `json:"workspace"`
	Version           string        `json:"version"`
	OutputStyle       OutputStyle   `json:"output_style"`
	Cost              Cost          `json:"cost"`
	ContextWindow     ContextWindow `json:"context_window"`
	Exceeds200KTokens bool          `json:"exceeds_200k_tokens"`
	Effort            Effort        `json:"effort"`
	Thinking          Thinking      `json:"thinking"`
	RateLimits        RateLimits    `json:"rate_limits"`
	Vim               Vim           `json:"vim"`
	Agent             Agent         `json:"agent"`
	PR                PullRequest   `json:"pr"`
	Worktree          Worktree      `json:"worktree"`
}

// Model contains the current model identifier and display name.
type Model struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// Label returns the model name to show in the status line.
func (model Model) Label() string {
	if model.DisplayName != "" {
		return model.DisplayName
	}
	return model.ID
}

// UsesDisplayName reports whether the renderer may shorten the model label.
// Display names support progressive shortening; raw model.id values stay as-is or
// are omitted when they do not fit.
func (model Model) UsesDisplayName() bool {
	return model.DisplayName != ""
}

// Workspace contains the current workspace paths and repository identity.
type Workspace struct {
	CurrentDir  string   `json:"current_dir"`
	ProjectDir  string   `json:"project_dir"`
	AddedDirs   []string `json:"added_dirs"`
	GitWorktree string   `json:"git_worktree"`
	Repo        Repo     `json:"repo"`
}

// Repo contains repository identity parsed from the origin remote.
type Repo struct {
	Host  string `json:"host"`
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

// OutputStyle contains the current Claude Code output style name.
type OutputStyle struct {
	Name string `json:"name"`
}

// Cost contains session cost, timing, and line-change totals.
type Cost struct {
	TotalCostUSD       float64 `json:"total_cost_usd"`
	TotalDurationMS    int     `json:"total_duration_ms"`
	TotalAPIDurationMS int     `json:"total_api_duration_ms"`
	TotalLinesAdded    int     `json:"total_lines_added"`
	TotalLinesRemoved  int     `json:"total_lines_removed"`
}

// ContextWindow contains live context-window token data. Claude Code documents
// TotalInputTokens as the tokens currently in the window, already including
// cache reads and writes, and ContextWindowSize as the window maximum. Both are
// always sent. UsedPercentage and RemainingPercentage may be null early in a
// session.
type ContextWindow struct {
	TotalInputTokens    int      `json:"total_input_tokens"`
	TotalOutputTokens   int      `json:"total_output_tokens"`
	ContextWindowSize   int      `json:"context_window_size"`
	UsedPercentage      *float64 `json:"used_percentage"`
	RemainingPercentage *float64 `json:"remaining_percentage"`
	CurrentUsage        *Usage   `json:"current_usage"`
}

// Usage contains per-category token usage from the latest API call.
type Usage struct {
	InputTokens         int `json:"input_tokens"`
	OutputTokens        int `json:"output_tokens"`
	CacheCreationTokens int `json:"cache_creation_input_tokens"`
	CacheReadTokens     int `json:"cache_read_input_tokens"`
}

// Effort contains the current reasoning effort level.
type Effort struct {
	Level string `json:"level"`
}

// Thinking contains the extended-thinking state.
type Thinking struct {
	Enabled bool `json:"enabled"`
}

// RateLimits contains available Claude.ai subscription limit windows.
type RateLimits struct {
	FiveHour *RateLimit `json:"five_hour"`
	SevenDay *RateLimit `json:"seven_day"`
}

// RateLimit contains percentage usage and reset time for a limit window.
type RateLimit struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64   `json:"resets_at"`
}

// Vim contains the current vim mode when vim mode is enabled.
type Vim struct {
	Mode string `json:"mode"`
}

// Agent contains the active agent name when an agent is configured.
type Agent struct {
	Name string `json:"name"`
}

// PullRequest contains open pull request metadata for the current branch.
type PullRequest struct {
	Number      int    `json:"number"`
	URL         string `json:"url"`
	ReviewState string `json:"review_state"`
}

// Worktree contains Claude Code worktree session metadata.
type Worktree struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	Branch         string `json:"branch"`
	OriginalCWD    string `json:"original_cwd"`
	OriginalBranch string `json:"original_branch"`
}

// UsedRatio returns the share of the context window in use, from 0 to 1.
// Claude Code pre-calculates UsedPercentage, so the status line reports that
// figure rather than dividing token counts itself. The field is null early in a
// session, which the Claude Code examples read as zero.
func (contextWindow ContextWindow) UsedRatio() float64 {
	if contextWindow.UsedPercentage == nil {
		return 0
	}
	return *contextWindow.UsedPercentage / percentageScale
}

// UnmarshalPayload returns a status-line payload decoded from JSON bytes.
func UnmarshalPayload(raw []byte) (Payload, error) {
	var payload Payload
	err := payload.UnmarshalJSON(raw)
	if err != nil {
		return Payload{}, err
	}
	return payload, nil
}

// UnmarshalJSON decodes JSON bytes into a status-line payload.
func (payload *Payload) UnmarshalJSON(raw []byte) error {
	type payloadAlias Payload
	var decoded payloadAlias
	err := json.Unmarshal(raw, &decoded)
	if err != nil {
		return fmt.Errorf("unmarshal status payload: %w", err)
	}
	*payload = Payload(decoded)
	return nil
}
