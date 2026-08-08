// Command statusline renders a Claude Code status line.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"goodkind.io/statusline/internal/app"
	"goodkind.io/statusline/internal/version"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Fprintln(os.Stdout, version.String())
		os.Exit(0)
	}
	slog.Debug("statusline render", slog.String("component", "statusline"))
	os.Exit(app.Run(os.Stdin, os.Stdout, os.Stderr))
}
