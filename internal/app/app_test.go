package app

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

const minStatuslineBarWidth = 3

// trackColorSequence is the ANSI foreground escape the renderer emits for an
// unfilled bar cell. Its absence means the bar is full.
const trackColorSequence = "\x1b[38;2;58;58;58m"

type failingReader struct{}

func (failingReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestRunReportsPayloadContextWindowTokens(t *testing.T) {
	input := `{"cost":{"total_cost_usd":2.4193},"context_window":{"total_input_tokens":500,"context_window_size":900}}`
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(strings.NewReader(input), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exitCode = %d, want 0, stderr %q", exitCode, stderr.String())
	}
	if !strings.HasPrefix(stdout.String(), "500 ") {
		t.Fatalf("stdout = %q, want prefix %q", stdout.String(), "500 ")
	}
	if !strings.HasSuffix(stdout.String(), " 900 · $2.42\n") {
		t.Fatalf("stdout = %q, want suffix %q", stdout.String(), " 900 · $2.42\\n")
	}
}

// TestRunIgnoresTranscriptPath proves the status line reads token counts only
// from the payload. Claude Code sends total_input_tokens on every payload and
// reports 0 before the first API response, so an empty count is the real state
// rather than a gap to fill from the transcript file.
func TestRunIgnoresTranscriptPath(t *testing.T) {
	transcriptPath := writeTranscript(t, 100)
	input := `{"transcript_path":"` + transcriptPath + `","cost":{"total_cost_usd":2.4193},"context_window":{"total_input_tokens":0,"context_window_size":900}}`
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(strings.NewReader(input), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exitCode = %d, want 0, stderr %q", exitCode, stderr.String())
	}
	if !strings.HasPrefix(stdout.String(), "0 ") {
		t.Fatalf("stdout = %q, want prefix %q", stdout.String(), "0 ")
	}
}

func TestRunRendersExtendedContextWindowFromPayload(t *testing.T) {
	input := `{"cost":{"total_cost_usd":2.4193},"model":{"id":"claude-sonnet-4"},"context_window":{"total_input_tokens":0,"context_window_size":1000000}}`
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(strings.NewReader(input), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exitCode = %d, want 0, stderr %q", exitCode, stderr.String())
	}
	if !strings.HasSuffix(stdout.String(), " 1.0M · $2.42\n") {
		t.Fatalf("stdout = %q, want suffix %q", stdout.String(), " 1.0M · $2.42\\n")
	}
}

// TestRunDropsWindowSizeCarriedByTheModelName avoids printing the same figure
// twice. A model named "Opus 5 (1M context)" already states the window, so the
// suffix carries the cost alone.
func TestRunDropsWindowSizeCarriedByTheModelName(t *testing.T) {
	t.Setenv("COLUMNS", "100")

	input := `{"cost":{"total_cost_usd":2.4193},"model":{"id":"claude-opus-5","display_name":"Opus 5 (1M context)"},"context_window":{"total_input_tokens":0,"context_window_size":1000000}}`
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(strings.NewReader(input), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exitCode = %d, want 0, stderr %q", exitCode, stderr.String())
	}
	if !strings.HasPrefix(stdout.String(), "Opus 5 (1M context) · 0 ") {
		t.Fatalf("stdout = %q, want the model label to state the window", stdout.String())
	}
	if !strings.HasSuffix(stdout.String(), " $2.42\n") {
		t.Fatalf("stdout = %q, want suffix %q", stdout.String(), " $2.42\\n")
	}
	if strings.Contains(stdout.String(), "1.0M") {
		t.Fatalf("stdout = %q, want the window size stated once", stdout.String())
	}
}

// TestRunKeepsWindowSizeWhenTheModelLabelIsDropped covers the narrow terminal.
// Once the label no longer fits, the window size is the only place the figure
// can appear, so it must come back.
func TestRunKeepsWindowSizeWhenTheModelLabelIsDropped(t *testing.T) {
	t.Setenv("COLUMNS", "24")

	input := `{"cost":{"total_cost_usd":2.4193},"model":{"id":"claude-opus-5","display_name":"Opus 5 (1M context)"},"context_window":{"total_input_tokens":0,"context_window_size":1000000}}`
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(strings.NewReader(input), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exitCode = %d, want 0, stderr %q", exitCode, stderr.String())
	}
	if strings.Contains(stdout.String(), "Opus") {
		t.Fatalf("stdout = %q, want the model label dropped at this width", stdout.String())
	}
	if !strings.HasSuffix(stdout.String(), " 1.0M · $2.42\n") {
		t.Fatalf("stdout = %q, want the window size restored", stdout.String())
	}
}

// TestRunFillsBarFromUsedPercentage checks that the bar follows the percentage
// Claude Code calculates. The token counts here would give a different fill if
// the status line divided them itself.
func TestRunFillsBarFromUsedPercentage(t *testing.T) {
	t.Setenv("COLUMNS", "80")

	input := `{"cost":{"total_cost_usd":2.4193},"context_window":{"total_input_tokens":10,"context_window_size":1000,"used_percentage":100}}`
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(strings.NewReader(input), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exitCode = %d, want 0, stderr %q", exitCode, stderr.String())
	}
	if strings.Contains(stdout.String(), trackColorSequence) {
		t.Fatalf("stdout = %q, want a full bar with no unfilled cells", stdout.String())
	}
}

func TestRunIncludesRateLimitRemainingPercentagesWhenColumnsFit(t *testing.T) {
	t.Setenv("COLUMNS", "90")

	input := `{"cost":{"total_cost_usd":2.4193},"context_window":{"total_input_tokens":500,"context_window_size":900},"rate_limits":{"five_hour":{"used_percentage":24.4},"seven_day":{"used_percentage":40.6}}}`
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(strings.NewReader(input), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exitCode = %d, want 0, stderr %q", exitCode, stderr.String())
	}
	wantSuffix := " 900 · $2.42 · 5h 76% · 7d 59%\n"
	if !strings.HasSuffix(stdout.String(), wantSuffix) {
		t.Fatalf("stdout = %q, want suffix %q", stdout.String(), wantSuffix)
	}
}

func TestRunIncludesModelNameWhenColumnsFit(t *testing.T) {
	t.Setenv("COLUMNS", "80")

	input := `{"cost":{"total_cost_usd":2.4193},"model":{"id":"claude-opus-4-8","display_name":"Opus"},"context_window":{"total_input_tokens":500,"context_window_size":900}}`
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(strings.NewReader(input), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exitCode = %d, want 0, stderr %q", exitCode, stderr.String())
	}
	wantPrefix := "Opus · 500 "
	if !strings.HasPrefix(stdout.String(), wantPrefix) {
		t.Fatalf("stdout = %q, want prefix %q", stdout.String(), wantPrefix)
	}
	wantSuffix := " 900 · $2.42\n"
	if !strings.HasSuffix(stdout.String(), wantSuffix) {
		t.Fatalf("stdout = %q, want suffix %q", stdout.String(), wantSuffix)
	}
}

func TestRunShortensModelContextLabelWhenColumnsAreNarrow(t *testing.T) {
	mediumWidth := lipgloss.Width("Opus (1M) · 500 ") + minStatuslineBarWidth + lipgloss.Width(" 900 · $2.42")
	t.Setenv("COLUMNS", strconv.Itoa(mediumWidth))

	input := `{"cost":{"total_cost_usd":2.4193},"model":{"id":"claude-opus-4-8","display_name":"Opus (1M Context)"},"context_window":{"total_input_tokens":500,"context_window_size":900}}`
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(strings.NewReader(input), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exitCode = %d, want 0, stderr %q", exitCode, stderr.String())
	}
	wantPrefix := "Opus (1M) · 500 "
	if !strings.HasPrefix(stdout.String(), wantPrefix) {
		t.Fatalf("stdout = %q, want prefix %q", stdout.String(), wantPrefix)
	}
}

func TestRunDoesNotShortenDirectModelID(t *testing.T) {
	mediumWidth := lipgloss.Width("Opus · 500 ") + minStatuslineBarWidth + lipgloss.Width(" 900 · $2.42")
	t.Setenv("COLUMNS", strconv.Itoa(mediumWidth))

	input := `{"cost":{"total_cost_usd":2.4193},"model":{"id":"Opus (1M Context)"},"context_window":{"total_input_tokens":500,"context_window_size":900}}`
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(strings.NewReader(input), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exitCode = %d, want 0, stderr %q", exitCode, stderr.String())
	}
	if !strings.HasPrefix(stdout.String(), "500 ") {
		t.Fatalf("stdout = %q, want model omitted without shortening", stdout.String())
	}
	if strings.Contains(stdout.String(), "Opus") {
		t.Fatalf("stdout = %q, want direct model ID unchanged or omitted", stdout.String())
	}
}

func TestRunFallsBackToModelIDWhenDisplayNameIsAbsent(t *testing.T) {
	t.Setenv("COLUMNS", "90")

	input := `{"cost":{"total_cost_usd":2.4193},"model":{"id":"claude-sonnet-4-1m"},"context_window":{"total_input_tokens":500,"context_window_size":900}}`
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(strings.NewReader(input), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exitCode = %d, want 0, stderr %q", exitCode, stderr.String())
	}
	wantPrefix := "claude-sonnet-4-1m · 500 "
	if !strings.HasPrefix(stdout.String(), wantPrefix) {
		t.Fatalf("stdout = %q, want prefix %q", stdout.String(), wantPrefix)
	}
	wantSuffix := " 900 · $2.42\n"
	if !strings.HasSuffix(stdout.String(), wantSuffix) {
		t.Fatalf("stdout = %q, want suffix %q", stdout.String(), wantSuffix)
	}
}

func TestRemainingPercentageClampsAndRounds(t *testing.T) {
	testCases := []struct {
		name string
		used float64
		want int
	}{
		{name: "rounds", used: 24.4, want: 76},
		{name: "clamps low", used: 150, want: 0},
		{name: "clamps high", used: -10, want: 100},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := remainingPercentage(testCase.used)
			if got != testCase.want {
				t.Fatalf("remainingPercentage() = %d, want %d", got, testCase.want)
			}
		})
	}
}

func TestRunReturnsErrorForInvalidJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(strings.NewReader("{"), &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("Run() exitCode = %d, want 1", exitCode)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "statusline: json:") {
		t.Fatalf("stderr = %q, want JSON error", stderr.String())
	}
}

func TestRunReturnsErrorForStdinReadFailure(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(failingReader{}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("Run() exitCode = %d, want 1", exitCode)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "statusline: read stdin: read failed") {
		t.Fatalf("stderr = %q, want stdin read error", stderr.String())
	}
}

func writeTranscript(t *testing.T, inputTokens int) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	content := `{"message":{"usage":{"input_tokens":` + strconv.Itoa(inputTokens) + `}}}`
	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}
	return path
}
