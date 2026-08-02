// Package tokenbudget calculates context-window display limits.
package tokenbudget

import (
	"goodkind.io/statusline/internal/statuspayload"
)

// defaultContextWindow is the window Claude Code documents as the default, and
// the only sensible guess when a payload omits the field. Claude Code always
// sends context_window_size, so this is a defensive fallback rather than a
// routine path. Deliberately not derived from exceeds_200k_tokens: that flag
// reports usage against a fixed 200k threshold regardless of window size, so it
// cannot tell a large window from a small one.
const defaultContextWindow = 200_000

// Ceiling returns the prompt/context display limit for the status-line payload.
// The window the payload reports is authoritative, because it is the session's
// live limit.
func Ceiling(data statuspayload.Payload) int {
	size := data.ContextWindow.ContextWindowSize
	if size > 0 {
		return size
	}
	return defaultContextWindow
}
