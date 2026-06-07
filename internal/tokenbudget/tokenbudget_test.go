package tokenbudget

import (
	"testing"

	"goodkind.io/statusline/internal/statuspayload"
)

func TestCeilingUsesPayloadContextWindowSizeAndOutputReserve(t *testing.T) {
	t.Setenv("CLAUDE_CODE_MAX_OUTPUT_TOKENS", "64000")

	data := statuspayload.Payload{
		Model: statuspayload.Model{ID: "claude-sonnet-4-1m"},
		ContextWindow: statuspayload.ContextWindow{
			ContextWindowSize: 950_000,
		},
	}

	got := Ceiling(data)
	want := 886_000
	if got != want {
		t.Fatalf("Ceiling() = %d, want %d", got, want)
	}
}

func TestCeilingFallsBackToModelAndOutputReserve(t *testing.T) {
	t.Setenv("CLAUDE_CODE_MAX_OUTPUT_TOKENS", "64000")

	data := statuspayload.Payload{
		Model: statuspayload.Model{ID: "claude-sonnet-4-1m"},
	}

	got := Ceiling(data)
	want := 936_000
	if got != want {
		t.Fatalf("Ceiling() = %d, want %d", got, want)
	}
}

func TestInputCeilingUsesDefaultReserveWhenEnvIsInvalid(t *testing.T) {
	t.Setenv("CLAUDE_CODE_MAX_OUTPUT_TOKENS", "not-a-number")

	got := InputCeiling(200_000)
	want := 136_000
	if got != want {
		t.Fatalf("InputCeiling() = %d, want %d", got, want)
	}
}

func TestInputCeilingKeepsTotalWhenReserveConsumesWindow(t *testing.T) {
	t.Setenv("CLAUDE_CODE_MAX_OUTPUT_TOKENS", "300000")

	got := InputCeiling(200_000)
	want := 200_000
	if got != want {
		t.Fatalf("InputCeiling() = %d, want %d", got, want)
	}
}

func TestContextWindowUsesDefaultWithoutMillionMarker(t *testing.T) {
	got := ContextWindow("claude-sonnet-4")
	want := 200_000
	if got != want {
		t.Fatalf("ContextWindow() = %d, want %d", got, want)
	}
}
