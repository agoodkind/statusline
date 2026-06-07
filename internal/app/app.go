// Package app wires input decoding, usage collection, and rendering.
package app

import (
	"fmt"
	"io"
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
	fmt.Fprintln(stdout, render.Line(used, ceiling, data.Cost.TotalCostUSD, width))

	return successExitCode
}
