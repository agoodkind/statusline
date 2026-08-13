// Package display formats values shown in the status line.
package display

import (
	"fmt"
	"strconv"
)

// HumanTokens formats a token count as a compact display string.
func HumanTokens(tokens int) string {
	switch {
	case tokens >= 1_000_000 || tokens <= -1_000_000:
		return fmt.Sprintf("%.1fM", float64(tokens)/1_000_000)
	case tokens >= 1_000 || tokens <= -1_000:
		return fmt.Sprintf("%dk", tokens/1_000)
	default:
		return strconv.Itoa(tokens)
	}
}

// Money formats a dollar amount to whole cents.
func Money(usd float64) string {
	return fmt.Sprintf("$%.2f", usd)
}
