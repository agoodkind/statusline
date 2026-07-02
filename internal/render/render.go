// Package render builds the visible status-line output.
package render

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"goodkind.io/statusline/internal/color"
	"goodkind.io/statusline/internal/display"
)

const (
	trackColor      = "#3A3A3A"
	fullBlock       = "█"
	minBarWidth     = 3
	maxBarWidth     = 48
	suffixSeparator = " · "
)

var partialBlocks = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}

// UsageLimit contains one optional rate-limit suffix with a remaining percentage.
type UsageLimit struct {
	Label               string
	RemainingPercentage int
}

// Line builds the complete status line.
func Line(used int, ceiling int, cost float64, model string, width int, usageLimits ...UsageLimit) string {
	ceiling = max(ceiling, used)

	label := display.HumanTokens(used) + " "
	suffix := " " + display.HumanTokens(ceiling) + suffixSeparator + display.Money(cost)
	prefix := prefixWithModel(label, suffix, width, model)
	barWidth := clampBarWidth(width - lipgloss.Width(prefix) - lipgloss.Width(suffix))
	suffix = suffixWithUsageLimits(prefix, suffix, width, usageLimits)
	barWidth = clampBarWidth(width - lipgloss.Width(prefix) - lipgloss.Width(suffix))

	ratio := 0.0
	if ceiling > 0 {
		ratio = float64(used) / float64(ceiling)
	}
	ratio = min(1.0, max(0.0, ratio))

	return prefix + bar(barWidth, ratio) + suffix
}

func prefixWithModel(label string, suffix string, width int, model string) string {
	if model == "" {
		return label
	}

	candidate := model + suffixSeparator + label
	lineWidth := lipgloss.Width(candidate) + minBarWidth + lipgloss.Width(suffix)
	if lineWidth > width {
		return label
	}
	return candidate
}

func suffixWithUsageLimits(
	prefix string,
	suffix string,
	width int,
	usageLimits []UsageLimit,
) string {
	for _, usageLimit := range usageLimits {
		nextSuffix := suffix + suffixSeparator + usageLimit.text()
		lineWidth := lipgloss.Width(prefix) + minBarWidth + lipgloss.Width(nextSuffix)
		if lineWidth > width {
			break
		}
		suffix = nextSuffix
	}
	return suffix
}

func (usageLimit UsageLimit) text() string {
	return usageLimit.Label + " " + strconv.Itoa(usageLimit.RemainingPercentage) + "%"
}

func clampBarWidth(width int) int {
	return max(minBarWidth, min(maxBarWidth, width))
}

func bar(width int, ratio float64) string {
	track := lipgloss.NewStyle().Foreground(lipgloss.Color(trackColor))
	filledCells := ratio * float64(width)
	fullCells := int(filledCells)
	fraction := filledCells - float64(fullCells)

	var builder strings.Builder
	for i := range width {
		switch {
		case i < fullCells:
			builder.WriteString(fillStyle(i, width).Render(fullBlock))
		case i == fullCells && fraction > 0:
			blockIndex := max(int(fraction*float64(len(partialBlocks))), 1)
			style := fillStyle(i, width).Background(lipgloss.Color(trackColor))
			builder.WriteString(style.Render(partialBlocks[blockIndex]))
		default:
			builder.WriteString(track.Render(fullBlock))
		}
	}
	return builder.String()
}

func fillStyle(index int, width int) lipgloss.Style {
	position := (float64(index) + 0.5) / float64(width)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color.HexForRatio(position)))
}
