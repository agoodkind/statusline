package statuspayload

import "testing"

func TestUnmarshalPayloadReadsDocumentedStatusLineFields(t *testing.T) {
	raw := []byte(`{"cwd":"/current/working/directory","session_id":"abc123","session_name":"my-session","transcript_path":"/tmp/session.jsonl","model":{"id":"claude-opus-4-8","display_name":"Opus"},"workspace":{"current_dir":"/current/working/directory","project_dir":"/original/project/directory","added_dirs":["/extra"],"git_worktree":"feature-xyz","repo":{"host":"github.com","owner":"anthropics","name":"claude-code"}},"version":"2.1.90","output_style":{"name":"default"},"cost":{"total_cost_usd":2.4193,"total_duration_ms":45000,"total_api_duration_ms":2300,"total_lines_added":156,"total_lines_removed":23},"context_window":{"total_input_tokens":15500,"total_output_tokens":1200,"context_window_size":950000,"used_percentage":8,"remaining_percentage":92,"current_usage":{"input_tokens":8500,"output_tokens":1200,"cache_creation_input_tokens":5000,"cache_read_input_tokens":2000}},"exceeds_200k_tokens":false,"effort":{"level":"high"},"thinking":{"enabled":true},"rate_limits":{"five_hour":{"used_percentage":23.5,"resets_at":1738425600},"seven_day":{"used_percentage":41.2,"resets_at":1738857600}},"vim":{"mode":"NORMAL"},"agent":{"name":"security-reviewer"},"pr":{"number":1234,"url":"https://github.com/anthropics/claude-code/pull/1234","review_state":"pending"},"worktree":{"name":"my-feature","path":"/path/to/.claude/worktrees/my-feature","branch":"worktree-my-feature","original_cwd":"/path/to/project","original_branch":"main"}}`)

	got, err := UnmarshalPayload(raw)
	if err != nil {
		t.Fatalf("UnmarshalPayload() returned error: %v", err)
	}

	if got.CWD != "/current/working/directory" {
		t.Fatalf("CWD = %q, want %q", got.CWD, "/current/working/directory")
	}
	if got.SessionID != "abc123" {
		t.Fatalf("SessionID = %q, want %q", got.SessionID, "abc123")
	}
	if got.Model.ID != "claude-opus-4-8" {
		t.Fatalf("Model.ID = %q, want %q", got.Model.ID, "claude-opus-4-8")
	}
	if got.Model.Label() != "Opus" {
		t.Fatalf("Model.Label() = %q, want %q", got.Model.Label(), "Opus")
	}
	if got.Workspace.Repo.Owner != "anthropics" {
		t.Fatalf("Workspace.Repo.Owner = %q, want %q", got.Workspace.Repo.Owner, "anthropics")
	}
	if got.Cost.TotalCostUSD != 2.4193 {
		t.Fatalf("Cost.TotalCostUSD = %f, want %f", got.Cost.TotalCostUSD, 2.4193)
	}
	if got.ContextWindow.ContextWindowSize != 950_000 {
		t.Fatalf("ContextWindow.ContextWindowSize = %d, want %d", got.ContextWindow.ContextWindowSize, 950_000)
	}
	if got.ContextWindow.TotalInputTokens != 15_500 {
		t.Fatalf("ContextWindow.TotalInputTokens = %d, want %d", got.ContextWindow.TotalInputTokens, 15_500)
	}
	if got.ContextWindow.UsedPercentage == nil || *got.ContextWindow.UsedPercentage != 8 {
		t.Fatalf("ContextWindow.UsedPercentage = %v, want 8", got.ContextWindow.UsedPercentage)
	}
	if got.RateLimits.FiveHour == nil || got.RateLimits.FiveHour.ResetsAt != 1_738_425_600 {
		t.Fatalf("RateLimits.FiveHour = %#v, want resets_at", got.RateLimits.FiveHour)
	}
	if got.Worktree.Branch != "worktree-my-feature" {
		t.Fatalf("Worktree.Branch = %q, want %q", got.Worktree.Branch, "worktree-my-feature")
	}
}

func TestModelLabelFallsBackToID(t *testing.T) {
	model := Model{ID: "claude-sonnet-4-1m"}
	if got := model.Label(); got != "claude-sonnet-4-1m" {
		t.Fatalf("Label() = %q, want %q", got, "claude-sonnet-4-1m")
	}
	if model.UsesDisplayName() {
		t.Fatal("UsesDisplayName() = true, want false")
	}
}

func TestModelUsesDisplayName(t *testing.T) {
	model := Model{ID: "claude-opus-4-8", DisplayName: "Opus"}
	if !model.UsesDisplayName() {
		t.Fatal("UsesDisplayName() = false, want true")
	}
}

func TestContextWindowUsedRatioConvertsDocumentedPercentage(t *testing.T) {
	got, err := UnmarshalPayload([]byte(`{"context_window":{"used_percentage":8}}`))
	if err != nil {
		t.Fatalf("UnmarshalPayload() returned error: %v", err)
	}

	if ratio := got.ContextWindow.UsedRatio(); ratio != 0.08 {
		t.Fatalf("UsedRatio() = %v, want 0.08", ratio)
	}
}

// TestContextWindowUsedRatioTreatsNullAsZero covers the early-session payload,
// where Claude Code documents used_percentage as null.
func TestContextWindowUsedRatioTreatsNullAsZero(t *testing.T) {
	got, err := UnmarshalPayload([]byte(`{"context_window":{"used_percentage":null}}`))
	if err != nil {
		t.Fatalf("UnmarshalPayload() returned error: %v", err)
	}

	if ratio := got.ContextWindow.UsedRatio(); ratio != 0 {
		t.Fatalf("UsedRatio() = %v, want 0", ratio)
	}
}

// TestContextWindowUsedRatioIgnoresCurrentUsage proves the ratio comes from the
// percentage Claude Code calculates, not from token counts the status line
// might otherwise divide itself.
func TestContextWindowUsedRatioIgnoresCurrentUsage(t *testing.T) {
	data := ContextWindow{
		TotalInputTokens:  15_500,
		ContextWindowSize: 200_000,
		CurrentUsage: &Usage{
			InputTokens:         8_500,
			CacheCreationTokens: 5_000,
			CacheReadTokens:     2_000,
		},
	}

	if ratio := data.UsedRatio(); ratio != 0 {
		t.Fatalf("UsedRatio() = %v, want 0 without used_percentage", ratio)
	}
}

func TestUnmarshalPayloadReturnsJSONError(t *testing.T) {
	_, err := UnmarshalPayload([]byte(`{`))
	if err == nil {
		t.Fatal("UnmarshalPayload() returned nil error, want JSON error")
	}
}
