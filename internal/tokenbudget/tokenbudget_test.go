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

func TestCeilingDoesNotRender872KForOneMillionModelWith128KOutputReserve(t *testing.T) {
	t.Setenv("CLAUDE_CODE_MAX_OUTPUT_TOKENS", "128000")

	data := statuspayload.Payload{
		Model: statuspayload.Model{ID: "claude-sonnet-4-1m"},
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

func TestCeilingHonorsPayloadWindowSmallerThanPracticalLimit(t *testing.T) {
	data := statuspayload.Payload{
		Model: statuspayload.Model{ID: "claude-opus-4-8[1m]"},
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

func TestContextWindowUsesDefaultWithoutMillionMarker(t *testing.T) {
	got := ContextWindow("claude-sonnet-4")
	want := 200_000
	if got != want {
		t.Fatalf("ContextWindow() = %d, want %d", got, want)
	}
}

func TestContextWindowRecognizesFableWithoutMillionMarker(t *testing.T) {
	got := ContextWindow("claude-fable-5")
	want := 950_000
	if got != want {
		t.Fatalf("ContextWindow() = %d, want %d", got, want)
	}
}

func TestCeilingHonorsPayloadWindowLargerThanModelFallback(t *testing.T) {
	data := statuspayload.Payload{
		Model: statuspayload.Model{ID: "claude-fable-5"},
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

func TestCeilingFallsBackToModelWindowWithoutPayloadSize(t *testing.T) {
	data := statuspayload.Payload{
		Model: statuspayload.Model{ID: "claude-fable-5"},
	}

	got := Ceiling(data)
	want := 950_000
	if got != want {
		t.Fatalf("Ceiling() = %d, want %d", got, want)
	}
}
