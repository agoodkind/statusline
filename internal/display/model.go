package display

import "strings"

// ModelLabelVariants returns progressively shorter model labels for width fitting.
func ModelLabelVariants(label string) []string {
	if label == "" {
		return []string{label}
	}

	variants := []string{label}
	base, inner, hasParenthetical := splitTrailingParenthetical(label)
	if !hasParenthetical {
		return variants
	}

	if shortInner, ok := strings.CutSuffix(inner, " Context"); ok {
		shortLabel := base + " (" + shortInner + ")"
		if shortLabel != label {
			variants = append(variants, shortLabel)
		}
	}

	if base != "" {
		variants = append(variants, base)
	}
	return variants
}

func splitTrailingParenthetical(label string) (base string, inner string, ok bool) {
	open := strings.LastIndex(label, " (")
	if open < 0 {
		return label, "", false
	}
	if !strings.HasSuffix(label, ")") {
		return label, "", false
	}

	base = label[:open]
	inner = label[open+2 : len(label)-1]
	return base, inner, true
}
