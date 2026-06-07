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
func Ceiling(data statuspayload.Payload) int {
	if data.ContextWindow.ContextWindowSize > 0 {
		return data.ContextWindow.ContextWindowSize
	}
	return ContextWindow(data.Model.ID)
}

// ContextWindow returns the fallback context limit for the given model id.
func ContextWindow(modelID string) int {
	if strings.Contains(modelID, "[1m]") || strings.Contains(modelID, "1m") {
		return practicalMillionTokenLimit
	}
	return defaultContextWindow
}
