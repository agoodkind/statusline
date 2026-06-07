// Command statusline renders a Claude Code status line: a context-usage bar
// whose fill is a smooth thermometer gradient, green when the context is nearly
// empty and warming to red as it fills, followed by the session cost.
//
// The bar fills against the usable input ceiling, which is the context window
// minus the tokens reserved for the model's reply, so it reaches full red at the
// point where input actually hits the hard context-window wall. The end label
// still shows the full window size.
//
// It reads the status-line JSON payload from stdin (the object Claude Code
// pipes to the configured statusLine command) and writes one line to stdout.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

// defaultContextWindow is the token budget assumed when the model does not
// advertise a larger window.
const defaultContextWindow = 200_000

// millionTokenWindow is the budget used when the model id marks a 1M context.
const millionTokenWindow = 1_000_000

// defaultMaxOutputTokens is the reply budget assumed when the environment does
// not set CLAUDE_CODE_MAX_OUTPUT_TOKENS.
const defaultMaxOutputTokens = 64_000

// trackColor is the dim gray of the unfilled portion of the bar.
const trackColor = "#3A3A3A"

// fullBlock fills a whole cell. partialBlocks index the eighths of a cell from
// empty to full, so the fill can end on a fractional column and grow smoothly
// rather than one whole cell at a time.
const fullBlock = "█"

var partialBlocks = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}

// coolHue and warmHue bound the fill color in HSV degrees: a low bar reads green
// and the color warms toward red as the bar fills.
const (
	coolHue = 140.0
	warmHue = 0.0
)

// payload is the subset of the Claude Code status-line JSON this program reads.
type payload struct {
	Model          modelInfo `json:"model"`
	TranscriptPath string    `json:"transcript_path"`
	Cost           costInfo  `json:"cost"`
}

type modelInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type costInfo struct {
	TotalCostUSD float64 `json:"total_cost_usd"`
}

// usage is the token-usage block found on transcript message records.
type usage struct {
	InputTokens         int `json:"input_tokens"`
	OutputTokens        int `json:"output_tokens"`
	CacheReadTokens     int `json:"cache_read_input_tokens"`
	CacheCreationTokens int `json:"cache_creation_input_tokens"`
}

// transcriptLine is the subset of a transcript JSONL record this program reads.
type transcriptLine struct {
	Message struct {
		Usage usage `json:"usage"`
	} `json:"message"`
}

func main() {
	slog.Debug("statusline render", slog.String("component", "statusline"))

	// Claude Code runs this command with stdout connected to a pipe, so the
	// terminal-color autodetection would strip color. Force 24-bit color so the
	// pastel fill survives; Claude Code renders the ANSI codes itself.
	lipgloss.SetColorProfile(termenv.TrueColor)

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "statusline: read stdin: %v\n", err)
		os.Exit(1)
	}

	var data payload
	err = json.Unmarshal(raw, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "statusline: json: %v\n", err)
		os.Exit(1)
	}

	total := contextWindow(data.Model.ID)
	used := contextUsed(data.TranscriptPath)
	ceiling := inputCeiling(total)
	width := terminalWidth()

	fmt.Println(render(used, ceiling, data.Cost.TotalCostUSD, width))
}

// inputCeiling returns the usable input budget: the context window minus the
// tokens reserved for the model's reply. The reservation comes from
// CLAUDE_CODE_MAX_OUTPUT_TOKENS when set, which is where input meets the hard
// context-window wall. It falls back to defaultMaxOutputTokens, and never
// returns less than 1.
func inputCeiling(total int) int {
	reserve := defaultMaxOutputTokens
	raw := os.Getenv("CLAUDE_CODE_MAX_OUTPUT_TOKENS")
	if raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil && parsed > 0 {
			reserve = parsed
		}
	}

	ceiling := total - reserve
	if ceiling < 1 {
		return total
	}
	return ceiling
}

// contextWindow returns the token budget for the given model id.
func contextWindow(modelID string) int {
	if strings.Contains(modelID, "[1m]") || strings.Contains(modelID, "1m") {
		return millionTokenWindow
	}
	return defaultContextWindow
}

// contextUsed reads the transcript JSONL at path and returns the token count of
// the most recent record that carries a usage block. It returns 0 when the file
// is absent or carries no usage.
func contextUsed(path string) int {
	if path == "" {
		return 0
	}

	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	used := 0
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 16*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var record transcriptLine
		err = json.Unmarshal(line, &record)
		if err != nil {
			continue
		}
		u := record.Message.Usage
		sum := u.InputTokens + u.CacheReadTokens + u.CacheCreationTokens
		if sum > 0 {
			used = sum
		}
	}
	if err := scanner.Err(); err != nil {
		return 0
	}
	return used
}

// terminalWidth returns the column count of the controlling terminal, or 80
// when it cannot be determined.
func terminalWidth() int {
	for _, fd := range []uintptr{os.Stdout.Fd(), os.Stderr.Fd()} {
		if fd > uintptr(math.MaxInt) {
			continue
		}
		if w, _, err := term.GetSize(int(fd)); err == nil && w > 0 {
			return w
		}
	}
	return 80
}

// render builds the full status line. The label shows used tokens, the suffix
// shows the usable input ceiling and cost, and the bar fills against that same
// ceiling, so the bar reaches full exactly when used reaches the right-hand
// number.
func render(used, ceiling int, cost float64, width int) string {
	// If usage already exceeds the estimated ceiling, show the ceiling as the
	// used value so the right number never reads lower than the left.
	ceiling = max(ceiling, used)

	label := humanTokens(used) + " "
	suffix := " " + humanTokens(ceiling) + " · " + money(cost)

	barWidth := clampBarWidth(width - lipgloss.Width(label) - lipgloss.Width(suffix))

	ratio := 0.0
	if ceiling > 0 {
		ratio = float64(used) / float64(ceiling)
	}
	ratio = min(1.0, max(0.0, ratio))

	return label + bar(barWidth, ratio) + suffix
}

// clampBarWidth keeps the bar within a legible range regardless of terminal
// size.
func clampBarWidth(w int) int {
	const minWidth = 3
	const maxWidth = 48
	return max(minWidth, min(maxWidth, w))
}

// bar renders a width-cell bar filled to the given ratio. Each filled column is
// colored by how far along the bar it sits, so the fill runs green near the
// start and warms toward red as it approaches full, which makes the color track
// how full the context is. The fill ends on a fractional sub-cell block so it
// grows smoothly instead of in whole cells.
func bar(width int, ratio float64) string {
	track := lipgloss.NewStyle().Foreground(lipgloss.Color(trackColor))

	filledCells := ratio * float64(width)
	full := int(filledCells)
	frac := filledCells - float64(full)

	var b strings.Builder
	for i := range width {
		switch {
		case i < full:
			b.WriteString(fillStyle(i, width).Render(fullBlock))
		case i == full && frac > 0:
			idx := max(int(frac*float64(len(partialBlocks))), 1)
			style := fillStyle(i, width).Background(lipgloss.Color(trackColor))
			b.WriteString(style.Render(partialBlocks[idx]))
		default:
			b.WriteString(track.Render(fullBlock))
		}
	}
	return b.String()
}

// fillStyle returns the foreground style for the column at index i of a
// width-cell bar, colored by that column's position so the hue tracks fullness.
func fillStyle(i, width int) lipgloss.Style {
	pos := (float64(i) + 0.5) / float64(width)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(hexForRatio(pos)))
}

// hexForRatio maps a fill position in [0,1] to a hex color running from green at
// 0 to red at 1.
func hexForRatio(pos float64) string {
	pos = max(0, min(1, pos))
	hue := coolHue + (warmHue-coolHue)*pos
	r, g, b := hsvToRGB(hue, 0.65, 0.95)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// hsvToRGB converts an HSV color (hue in degrees, saturation and value in [0,1])
// to 8-bit RGB components.
func hsvToRGB(h, s, v float64) (int, int, int) {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := v - c

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	to8 := func(f float64) int {
		return int((f+m)*255 + 0.5)
	}
	return to8(r), to8(g), to8(b)
}

// humanTokens formats a token count as a short k-suffixed string, so 91000
// becomes "91k" and 1000000 becomes "1.0M".
func humanTokens(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%dk", n/1_000)
	default:
		return strconv.Itoa(n)
	}
}

// money formats a dollar amount to whole cents, so 2.4193 becomes "$2.42".
func money(usd float64) string {
	return fmt.Sprintf("$%.2f", usd)
}
