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
	trackColor    = "#3A3A3A"
	fullBlock     = "█"
	leftHalfBlock = "▌"
	// A cell shows two colors, one per half, so positions are counted in
	// halves rather than cells.
	halvesPerCell = 2
	leftHalf      = 0
	rightHalf     = 1
	// The bar takes a share of the terminal up to a hard ceiling, so it stays
	// proportionate on narrow terminals and never dominates wide ones.
	minBarWidth     = 3
	maxBarWidth     = 24
	barWidthDivisor = 4
	suffixSeparator = " · "
)

var partialBlocks = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}

// UsageLimit contains one optional rate-limit suffix with a remaining percentage.
type UsageLimit struct {
	Label               string
	RemainingPercentage int
}

// Status carries the values one status line reports, taken from the fields
// Claude Code sends rather than derived from each other.
type Status struct {
	UsedTokens   int
	UsedRatio    float64
	CostUSD      float64
	Model        string
	ShortenModel bool
}

// Line builds the complete status line.
func Line(status Status, width int, usageLimits ...UsageLimit) string {
	label := display.HumanTokens(status.UsedTokens) + " "
	suffix := suffixSeparator + display.Money(status.CostUSD)
	modelLabel := modelPrefix(label, suffix, width, status.Model, status.ShortenModel)

	prefix := modelLabel + label
	suffix = suffixWithUsageLimits(prefix, suffix, width, usageLimits)
	available := width - lipgloss.Width(prefix) - lipgloss.Width(suffix)
	barWidth := clampBarWidth(available, width)

	// The bar cannot draw past its own ends, so the ratio is bounded for
	// drawing. This bounds the drawing only; it does not adjust the reported
	// figures.
	ratio := min(1.0, max(0.0, status.UsedRatio))

	return prefix + bar(barWidth, ratio) + suffix
}

// modelPrefix returns the widest model label that leaves room for the rest of
// the line, with its trailing separator. It returns an empty string when even
// the shortest variant does not fit, so the caller can test the model label on
// its own without the usage figure mixed in.
func modelPrefix(label string, suffix string, width int, model string, shortenModel bool) string {
	variants := modelLabelVariants(model, shortenModel)
	for _, variant := range variants {
		candidate := variant + suffixSeparator
		lineWidth := lipgloss.Width(candidate+label) + minBarWidth + lipgloss.Width(suffix)
		if lineWidth <= width {
			return candidate
		}
	}
	return ""
}

func modelLabelVariants(model string, shortenModel bool) []string {
	if model == "" {
		return nil
	}
	if shortenModel {
		return display.ModelLabelVariants(model)
	}
	return []string{model}
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

func clampBarWidth(available int, terminalWidth int) int {
	ceiling := min(maxBarWidth, terminalWidth/barWidthDivisor)
	return max(minBarWidth, min(ceiling, available))
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
			builder.WriteString(gradientCell(i, width))
		case i == fullCells && fraction > 0:
			blockIndex := max(int(fraction*float64(len(partialBlocks))), 1)
			style := fillStyle(halfPosition(i, leftHalf, width)).
				Background(lipgloss.Color(trackColor))
			builder.WriteString(style.Render(partialBlocks[blockIndex]))
		default:
			builder.WriteString(track.Render(fullBlock))
		}
	}
	return builder.String()
}

// gradientCell paints one filled cell with two colors. A left half block sets
// the left half from the foreground and leaves the right half showing the
// background, so a bar of N cells carries 2N colors. At the widths this bar
// runs, that halves the perceptual distance between neighbouring colors, which
// is what stops the sweep reading as a row of bands.
func gradientCell(index int, width int) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color.HexForRatio(halfPosition(index, leftHalf, width)))).
		Background(lipgloss.Color(color.HexForRatio(halfPosition(index, rightHalf, width)))).
		Render(leftHalfBlock)
}

// halfPosition returns the gradient position at the centre of one half of a
// cell. Sampling at the centre keeps the first and last colors inside the
// sweep rather than at its extremes.
func halfPosition(index int, half int, width int) float64 {
	halves := halvesPerCell * width
	return (float64(halvesPerCell*index+half) + 0.5) / float64(halves)
}

func fillStyle(position float64) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color.HexForRatio(position)))
}
