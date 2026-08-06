package render

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// testStatus builds the status every label test shares, with the ratio supplied
// separately the way Claude Code supplies it.
func testStatus(model string, shortenModel bool) Status {
	return Status{
		UsedTokens:   500_000,
		UsedRatio:    0.5263,
		CostUSD:      20.57,
		Model:        model,
		ShortenModel: shortenModel,
	}
}

func TestLineIncludesModelWhenColumnsFit(t *testing.T) {
	got := Line(testStatus("Opus", true), 80)
	wantPrefix := "Opus · 500k "
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("Line() prefix = %q, want prefix %q", got, wantPrefix)
	}
	if !strings.HasSuffix(got, " $20.57") {
		t.Fatalf("Line() suffix = %q, want suffix %q", got, " $20.57")
	}
}

func TestLineShortensDisplayModelLabelProgressively(t *testing.T) {
	model := "Opus (1M Context)"
	label := "500k "
	suffix := " $20.57"

	fullWidth := lipgloss.Width("Opus (1M Context) · "+label) + minBarWidth + lipgloss.Width(suffix)
	got := Line(testStatus(model, true), fullWidth)
	if !strings.HasPrefix(got, "Opus (1M Context) · 500k ") {
		t.Fatalf("Line() prefix = %q, want full model label", got)
	}

	mediumWidth := lipgloss.Width("Opus (1M) · "+label) + minBarWidth + lipgloss.Width(suffix)
	got = Line(testStatus(model, true), mediumWidth)
	if !strings.HasPrefix(got, "Opus (1M) · 500k ") {
		t.Fatalf("Line() prefix = %q, want shortened context label", got)
	}
	if strings.Contains(got, "Context") {
		t.Fatalf("Line() = %q, want Context omitted", got)
	}

	shortWidth := lipgloss.Width("Opus · "+label) + minBarWidth + lipgloss.Width(suffix)
	got = Line(testStatus(model, true), shortWidth)
	if !strings.HasPrefix(got, "Opus · 500k ") {
		t.Fatalf("Line() prefix = %q, want base model name", got)
	}

	got = Line(testStatus(model, true), shortWidth-1)
	if !strings.HasPrefix(got, "500k ") {
		t.Fatalf("Line() prefix = %q, want model omitted", got)
	}
}

func TestLineDoesNotShortenDirectModelID(t *testing.T) {
	model := "Opus (1M Context)"
	label := "500k "
	suffix := " $20.57"

	shortWidth := lipgloss.Width("Opus · "+label) + minBarWidth + lipgloss.Width(suffix)
	got := Line(testStatus(model, false), shortWidth)
	if !strings.HasPrefix(got, "500k ") {
		t.Fatalf("Line() prefix = %q, want model omitted", got)
	}
	if strings.Contains(got, "Opus") {
		t.Fatalf("Line() = %q, want direct model ID unchanged or omitted", got)
	}
}

func TestLineOmitsModelWhenColumnsAreTooNarrow(t *testing.T) {
	got := Line(testStatus("claude-opus-4-extended-thinking", false), 40)
	wantPrefix := "500k "
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("Line() prefix = %q, want prefix %q", got, wantPrefix)
	}
	if strings.Contains(got, "claude-opus-4-extended-thinking") {
		t.Fatalf("Line() = %q, want model omitted", got)
	}
}

func TestLineAddsUsageLimitsProgressively(t *testing.T) {
	usageLimits := []UsageLimit{
		{Label: "5h", RemainingPercentage: 76},
		{Label: "7d", RemainingPercentage: 59},
	}

	got := Line(testStatus("", false), 80, usageLimits...)
	wantSuffix := " $20.57 · 5h 76% · 7d 59%"
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("Line() suffix = %q, want suffix %q", got, wantSuffix)
	}

	got = Line(testStatus("", false), 32, usageLimits...)
	wantSuffix = " $20.57 · 5h 76%"
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("Line() suffix = %q, want suffix %q", got, wantSuffix)
	}
	if strings.Contains(got, "7d") {
		t.Fatalf("Line() = %q, want seven-day limit omitted", got)
	}

	got = Line(testStatus("", false), 90, usageLimits...)
	wantSuffix = " $20.57 · 5h 76% · 7d 59%"
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("Line() suffix = %q, want suffix %q", got, wantSuffix)
	}
}

func TestLineShowsModelAndRateLimitsWhenColumnsFit(t *testing.T) {
	usageLimits := []UsageLimit{
		{Label: "5h", RemainingPercentage: 76},
	}

	got := Line(testStatus("Opus", true), 95, usageLimits...)
	wantPrefix := "Opus · 500k "
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("Line() prefix = %q, want prefix %q", got, wantPrefix)
	}
	wantSuffix := " $20.57 · 5h 76%"
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("Line() suffix = %q, want suffix %q", got, wantSuffix)
	}
}

func TestClampBarWidthBounds(t *testing.T) {
	if got := clampBarWidth(1, 200); got != minBarWidth {
		t.Fatalf("clampBarWidth(1, 200) = %d, want %d", got, minBarWidth)
	}
	if got := clampBarWidth(180, 200); got != maxBarWidth {
		t.Fatalf("clampBarWidth(180, 200) = %d, want %d", got, maxBarWidth)
	}
}

func TestClampBarWidthScalesWithTerminalWidth(t *testing.T) {
	if got := clampBarWidth(60, 40); got != 10 {
		t.Fatalf("clampBarWidth(60, 40) = %d, want 10", got)
	}
	if got := clampBarWidth(60, 80); got != 20 {
		t.Fatalf("clampBarWidth(60, 80) = %d, want 20", got)
	}
}

func TestLineBarShrinksWithNarrowerTerminal(t *testing.T) {
	fixed := lipgloss.Width("500k ") + lipgloss.Width(" $20.57")

	wide := lipgloss.Width(Line(testStatus("", false), 120)) - fixed
	if wide != maxBarWidth {
		t.Fatalf("bar width at 120 columns = %d, want %d", wide, maxBarWidth)
	}

	narrow := lipgloss.Width(Line(testStatus("", false), 60)) - fixed
	if narrow != 60/barWidthDivisor {
		t.Fatalf("bar width at 60 columns = %d, want %d", narrow, 60/barWidthDivisor)
	}
}

func TestBarUsesRequestedDisplayWidth(t *testing.T) {
	got := bar(5, 0.5)
	if width := lipgloss.Width(got); width != 5 {
		t.Fatalf("lipgloss.Width(bar()) = %d, want 5", width)
	}
}
