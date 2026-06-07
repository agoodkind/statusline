// Package render builds the visible status-line output.
package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"goodkind.io/statusline/internal/color"
	"goodkind.io/statusline/internal/display"
)

const (
	trackColor  = "#3A3A3A"
	fullBlock   = "█"
	minBarWidth = 3
	maxBarWidth = 48
)

var partialBlocks = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}

// Line builds the complete status line.
func Line(used int, ceiling int, cost float64, width int) string {
	ceiling = max(ceiling, used)

	label := display.HumanTokens(used) + " "
	suffix := " " + display.HumanTokens(ceiling) + " · " + display.Money(cost)
	barWidth := clampBarWidth(width - lipgloss.Width(label) - lipgloss.Width(suffix))

	ratio := 0.0
	if ceiling > 0 {
		ratio = float64(used) / float64(ceiling)
	}
	ratio = min(1.0, max(0.0, ratio))

	return label + bar(barWidth, ratio) + suffix
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
