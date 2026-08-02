// Package tokenbudget calculates context-window display limits.
package tokenbudget

import (
	"goodkind.io/statusline/internal/statuspayload"
)

const (
	defaultContextWindow = 200_000
	largeContextWindow   = 1_000_000
)

// Ceiling returns the prompt/context display limit for the status-line payload.
// The window the payload reports is authoritative, because it is the session's
// live limit. When the payload reports no window, its own large-context flag
// picks between the two window sizes Claude Code offers.
func Ceiling(data statuspayload.Payload) int {
	size := data.ContextWindow.ContextWindowSize
	if size > 0 {
		return size
	}
	if data.Exceeds200KTokens {
		return largeContextWindow
	}
	return defaultContextWindow
}
