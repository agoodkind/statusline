package render

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestLineRaisesCeilingToUsage(t *testing.T) {
	got := Line(950_000, 872_000, 20.57, 80)

	if !strings.HasPrefix(got, "950k ") {
		t.Fatalf("Line() prefix = %q, want prefix %q", got, "950k ")
	}
	if !strings.HasSuffix(got, " 950k · $20.57") {
		t.Fatalf("Line() suffix = %q, want suffix %q", got, " 950k · $20.57")
	}
}

func TestLineAddsUsageLimitsProgressively(t *testing.T) {
	usageLimits := []UsageLimit{
		{Label: "5h", RemainingPercentage: 76},
		{Label: "7d", RemainingPercentage: 59},
	}

	got := Line(500_000, 950_000, 20.57, 80, usageLimits...)
	wantSuffix := " 950k · $20.57 · 5h 76%"
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("Line() suffix = %q, want suffix %q", got, wantSuffix)
	}
	if strings.Contains(got, "7d") {
		t.Fatalf("Line() = %q, want seven-day limit omitted", got)
	}

	got = Line(500_000, 950_000, 20.57, 90, usageLimits...)
	wantSuffix = " 950k · $20.57 · 5h 76% · 7d 59%"
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("Line() suffix = %q, want suffix %q", got, wantSuffix)
	}
}

func TestClampBarWidthBounds(t *testing.T) {
	if got := clampBarWidth(1); got != 3 {
		t.Fatalf("clampBarWidth(1) = %d, want 3", got)
	}
	if got := clampBarWidth(90); got != 48 {
		t.Fatalf("clampBarWidth(90) = %d, want 48", got)
	}
}

func TestBarUsesRequestedDisplayWidth(t *testing.T) {
	got := bar(5, 0.5)
	if width := lipgloss.Width(got); width != 5 {
		t.Fatalf("lipgloss.Width(bar()) = %d, want 5", width)
	}
}
