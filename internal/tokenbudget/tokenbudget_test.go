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

// TestCeilingIgnoresExceeds200KTokens locks the documented meaning of the flag:
// it reports usage against a fixed 200k threshold regardless of window size, so
// it must not influence the ceiling in either direction.
func TestCeilingIgnoresExceeds200KTokens(t *testing.T) {
	tests := []struct {
		name              string
		exceeds200KTokens bool
		contextWindowSize int
		want              int
	}{
		{name: "flag set with a reported window", exceeds200KTokens: true, contextWindowSize: 1_000_000, want: 1_000_000},
		{name: "flag clear with a reported window", exceeds200KTokens: false, contextWindowSize: 1_000_000, want: 1_000_000},
		{name: "flag set without a reported window", exceeds200KTokens: true, want: 200_000},
		{name: "flag clear without a reported window", exceeds200KTokens: false, want: 200_000},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := statuspayload.Payload{
				Model:             statuspayload.Model{ID: "claude-fable-5"},
				Exceeds200KTokens: test.exceeds200KTokens,
				ContextWindow: statuspayload.ContextWindow{
					ContextWindowSize: test.contextWindowSize,
				},
			}

			got := Ceiling(data)
			if got != test.want {
				t.Fatalf("Ceiling() = %d, want %d", got, test.want)
			}
		})
	}
}
