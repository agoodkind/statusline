// Package display formats values shown in the status line.
package display

import (
	"fmt"
	"strconv"
	"strings"
)

// HumanTokens formats a token count as a compact display string.
func HumanTokens(tokens int) string {
	switch {
	case tokens >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(tokens)/1_000_000)
	case tokens >= 1_000:
		return fmt.Sprintf("%dk", tokens/1_000)
	default:
		return strconv.Itoa(tokens)
	}
}

// TokensAppearIn reports whether text already states this token count, so a
// caller can avoid printing the same figure twice. Claude Code writes the
// window into the model name as "1M", while HumanTokens writes "1.0M", so the
// comparison drops a trailing ".0" and ignores case.
func TokensAppearIn(text string, tokens int) bool {
	compact := strings.TrimSuffix(HumanTokens(tokens), ".0M")
	if compact != HumanTokens(tokens) {
		compact += "M"
	}
	return strings.Contains(strings.ToLower(text), strings.ToLower(compact))
}

// Money formats a dollar amount to whole cents.
func Money(usd float64) string {
	return fmt.Sprintf("$%.2f", usd)
}
