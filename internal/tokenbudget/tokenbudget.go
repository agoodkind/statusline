// Package tokenbudget calculates context-window display limits.
package tokenbudget

import (
	"strings"

	"goodkind.io/statusline/internal/statuspayload"
)

const (
	defaultContextWindow       = 200_000
	practicalMillionTokenLimit = 950_000
)

// Ceiling returns the prompt/context display limit for the status-line payload.
// The model's practical limit caps the result, so a million-token model whose
// payload reports the full 1,000,000 window still displays the 950,000 practical
// ceiling; a payload reporting a smaller live window is honored as-is.
func Ceiling(data statuspayload.Payload) int {
	limit := ContextWindow(data.Model.ID)
	size := data.ContextWindow.ContextWindowSize
	if size > 0 && size < limit {
		return size
	}
	return limit
}

// ContextWindow returns the fallback context limit for the given model id.
func ContextWindow(modelID string) int {
	if strings.Contains(modelID, "[1m]") || strings.Contains(modelID, "1m") {
		return practicalMillionTokenLimit
	}
	return defaultContextWindow
}
