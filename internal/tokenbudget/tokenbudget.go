// Package tokenbudget calculates usable context-window ceilings.
package tokenbudget

import (
	"os"
	"strconv"
	"strings"

	"goodkind.io/statusline/internal/statuspayload"
)

const (
	defaultContextWindow   = 200_000
	millionTokenWindow     = 1_000_000
	defaultMaxOutputTokens = 64_000
	maxOutputTokensEnv     = "CLAUDE_CODE_MAX_OUTPUT_TOKENS"
)

// Ceiling returns the usable input ceiling for the status-line payload.
func Ceiling(data statuspayload.Payload) int {
	contextWindowSize := data.ContextWindow.ContextWindowSize
	if contextWindowSize == 0 {
		contextWindowSize = ContextWindow(data.Model.ID)
	}
	return InputCeiling(contextWindowSize)
}

// InputCeiling returns the usable input budget after reserving reply tokens.
func InputCeiling(total int) int {
	reserve := defaultMaxOutputTokens
	raw := os.Getenv(maxOutputTokensEnv)
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

// ContextWindow returns the token budget for the given model id.
func ContextWindow(modelID string) int {
	if strings.Contains(modelID, "[1m]") || strings.Contains(modelID, "1m") {
		return millionTokenWindow
	}
	return defaultContextWindow
}
