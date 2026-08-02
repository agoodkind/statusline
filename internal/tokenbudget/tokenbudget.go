// Package tokenbudget calculates context-window display limits.
package tokenbudget

import (
	"strings"

	"goodkind.io/statusline/internal/statuspayload"
)

const (
	defaultContextWindow       = 200_000
	practicalMillionTokenLimit = 950_000
	millionTokenWindow         = 1_000_000
)

// millionTokenModelMarkers are substrings of model ids whose context window is a
// million tokens even when the id carries no explicit width marker.
var millionTokenModelMarkers = []string{"[1m]", "1m", "fable", "mythos"}

// Ceiling returns the prompt/context display limit for the status-line payload.
// A live window reported by the payload wins, because it reflects the session's
// actual limit; a full million-token window is displayed as the 950,000
// practical ceiling. Without a reported window the model id decides.
func Ceiling(data statuspayload.Payload) int {
	size := data.ContextWindow.ContextWindowSize
	if size <= 0 {
		return ContextWindow(data.Model.ID)
	}
	if size >= millionTokenWindow {
		return practicalMillionTokenLimit
	}
	return size
}

// ContextWindow returns the fallback context limit for the given model id.
func ContextWindow(modelID string) int {
	for _, marker := range millionTokenModelMarkers {
		if strings.Contains(modelID, marker) {
			return practicalMillionTokenLimit
		}
	}
	return defaultContextWindow
}
