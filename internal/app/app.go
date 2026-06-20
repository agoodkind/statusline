// Package app wires input decoding, usage collection, and rendering.
package app

import (
	"fmt"
	"io"
	"math"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"goodkind.io/statusline/internal/render"
	"goodkind.io/statusline/internal/statuspayload"
	"goodkind.io/statusline/internal/terminal"
	"goodkind.io/statusline/internal/tokenbudget"
	"goodkind.io/statusline/internal/transcript"
)

const (
	successExitCode = 0
	failureExitCode = 1
)

// Run reads a status-line payload, renders one status line, and returns an exit code.
func Run(stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	lipgloss.SetColorProfile(termenv.TrueColor)

	raw, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(stderr, "statusline: read stdin: %v\n", err)
		return failureExitCode
	}

	data, err := statuspayload.UnmarshalPayload(raw)
	if err != nil {
		fmt.Fprintf(stderr, "statusline: json: %v\n", err)
		return failureExitCode
	}

	used := data.ContextWindow.InputTokens()
	if !data.ContextWindow.HasInputTokens() {
		used = transcript.Used(data.TranscriptPath)
	}
	ceiling := tokenbudget.Ceiling(data)
	width := terminal.Width(os.Stdout.Fd(), os.Stderr.Fd())
	usageLimits := usageLimitSegments(data.RateLimits)
	fmt.Fprintln(stdout, render.Line(used, ceiling, data.Cost.TotalCostUSD, width, usageLimits...))

	return successExitCode
}

func usageLimitSegments(rateLimits statuspayload.RateLimits) []render.UsageLimit {
	usageLimits := make([]render.UsageLimit, 0, 2)
	if rateLimits.FiveHour != nil {
		usageLimits = append(usageLimits, usageLimitSegment("5h", *rateLimits.FiveHour))
	}
	if rateLimits.SevenDay != nil {
		usageLimits = append(usageLimits, usageLimitSegment("7d", *rateLimits.SevenDay))
	}
	return usageLimits
}

func usageLimitSegment(label string, rateLimit statuspayload.RateLimit) render.UsageLimit {
	return render.UsageLimit{
		Label:          label,
		UsedPercentage: usedPercentagePercent(rateLimit.UsedPercentage),
	}
}

func usedPercentagePercent(usedPercentage float64) int {
	usedPercentage = max(0.0, min(100.0, usedPercentage))
	return int(math.Round(usedPercentage))
}
