// Command statusline renders a Claude Code status line.
package main

import (
	"log/slog"
	"os"

	"goodkind.io/statusline/internal/app"
)

func main() {
	slog.Debug("statusline render", slog.String("component", "statusline"))
	os.Exit(app.Run(os.Stdin, os.Stdout, os.Stderr))
}
