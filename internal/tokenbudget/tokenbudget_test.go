package tokenbudget

import (
	"testing"

	"goodkind.io/statusline/internal/statuspayload"
)

func TestCeilingUsesPayloadContextWindowSizeWithoutOutputReserve(t *testing.T) {
	t.Setenv("CLAUDE_CODE_MAX_OUTPUT_TOKENS", "128000")

	data := statuspayload.Payload{
		Model: statuspayload.Model{ID: "claude-sonnet-4-1m"},
		ContextWindow: statuspayload.ContextWindow{
			ContextWindowSize: 950_000,
		},
	}

	got := Ceiling(data)
	want := 950_000
	if got != want {
		t.Fatalf("Ceiling() = %d, want %d", got, want)
	}
}

func TestCeilingHonorsFullMillionPayloadWindow(t *testing.T) {
	data := statuspayload.Payload{
		Model: statuspayload.Model{ID: "claude-opus-4-8[1m]"},
		ContextWindow: statuspayload.ContextWindow{
			ContextWindowSize: 1_000_000,
		},
	}

	got := Ceiling(data)
	want := 1_000_000
	if got != want {
		t.Fatalf("Ceiling() = %d, want %d", got, want)
	}
}

func TestCeilingHonorsPayloadWindowSmallerThanDefault(t *testing.T) {
	data := statuspayload.Payload{
		Model: statuspayload.Model{ID: "claude-opus-4-8[1m]"},
		ContextWindow: statuspayload.ContextWindow{
			ContextWindowSize: 120_000,
		},
	}

	got := Ceiling(data)
	want := 120_000
	if got != want {
		t.Fatalf("Ceiling() = %d, want %d", got, want)
	}
}

func TestCeilingFallsBackToDefaultWindow(t *testing.T) {
	data := statuspayload.Payload{
		Model: statuspayload.Model{ID: "claude-sonnet-4"},
	}

	got := Ceiling(data)
	want := 200_000
	if got != want {
		t.Fatalf("Ceiling() = %d, want %d", got, want)
	}
}

func TestCeilingFallsBackToLargeWindowWhenPayloadFlagsLargeContext(t *testing.T) {
	data := statuspayload.Payload{
		Model:             statuspayload.Model{ID: "claude-fable-5"},
		Exceeds200KTokens: true,
	}

	got := Ceiling(data)
	want := 1_000_000
	if got != want {
		t.Fatalf("Ceiling() = %d, want %d", got, want)
	}
}

func TestCeilingPrefersPayloadWindowOverLargeContextFlag(t *testing.T) {
	data := statuspayload.Payload{
		Model:             statuspayload.Model{ID: "claude-fable-5"},
		Exceeds200KTokens: true,
		ContextWindow: statuspayload.ContextWindow{
			ContextWindowSize: 400_000,
		},
	}

	got := Ceiling(data)
	want := 400_000
	if got != want {
		t.Fatalf("Ceiling() = %d, want %d", got, want)
	}
}
