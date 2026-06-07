// Package terminal detects terminal display properties.
package terminal

import (
	"math"

	"golang.org/x/term"
)

const fallbackWidth = 80

// Width returns the first usable terminal width from the supplied file descriptors.
func Width(fileDescriptors ...uintptr) int {
	for _, fileDescriptor := range fileDescriptors {
		if fileDescriptor > uintptr(math.MaxInt) {
			continue
		}
		width, _, err := term.GetSize(int(fileDescriptor))
		if err == nil && width > 0 {
			return width
		}
	}
	return fallbackWidth
}
