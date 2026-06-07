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
	if got.Workspace.Repo.Owner != "anthropics" {
		t.Fatalf("Workspace.Repo.Owner = %q, want %q", got.Workspace.Repo.Owner, "anthropics")
	}
	if got.Cost.TotalCostUSD != 2.4193 {
		t.Fatalf("Cost.TotalCostUSD = %f, want %f", got.Cost.TotalCostUSD, 2.4193)
	}
	if got.ContextWindow.ContextWindowSize != 950_000 {
		t.Fatalf("ContextWindow.ContextWindowSize = %d, want %d", got.ContextWindow.ContextWindowSize, 950_000)
	}
	if got.ContextWindow.InputTokens() != 15_500 {
		t.Fatalf("ContextWindow.InputTokens() = %d, want %d", got.ContextWindow.InputTokens(), 15_500)
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

func TestContextWindowInputTokensFallsBackToCurrentUsage(t *testing.T) {
	data := ContextWindow{
		CurrentUsage: &Usage{
			InputTokens:         8_500,
			CacheCreationTokens: 5_000,
			CacheReadTokens:     2_000,
		},
	}

	if !data.HasInputTokens() {
		t.Fatal("HasInputTokens() = false, want true")
	}
	got := data.InputTokens()
	want := 15_500
	if got != want {
		t.Fatalf("InputTokens() = %d, want %d", got, want)
	}
}

func TestContextWindowInputTokensTracksPresentZero(t *testing.T) {
	got, err := UnmarshalPayload([]byte(`{"context_window":{"total_input_tokens":0}}`))
	if err != nil {
		t.Fatalf("UnmarshalPayload() returned error: %v", err)
	}

	if !got.ContextWindow.HasInputTokens() {
		t.Fatal("HasInputTokens() = false, want true")
	}
	if got.ContextWindow.InputTokens() != 0 {
		t.Fatalf("InputTokens() = %d, want 0", got.ContextWindow.InputTokens())
	}
}

func TestContextWindowInputTokensReportsAbsentFields(t *testing.T) {
	data := ContextWindow{}

	if data.HasInputTokens() {
		t.Fatal("HasInputTokens() = true, want false")
	}
	if data.InputTokens() != 0 {
		t.Fatalf("InputTokens() = %d, want 0", data.InputTokens())
	}
}

func TestUnmarshalPayloadReturnsJSONError(t *testing.T) {
	_, err := UnmarshalPayload([]byte(`{`))
	if err == nil {
		t.Fatal("UnmarshalPayload() returned nil error, want JSON error")
	}
}
