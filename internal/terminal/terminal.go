// Package terminal detects terminal display properties.
package terminal

import (
	"math"
	"os"
	"strconv"

	"golang.org/x/term"
)

const fallbackWidth = 80

// Width returns the first usable terminal width from the supplied file descriptors.
func Width(fileDescriptors ...uintptr) int {
	if width := columnsWidth(os.Getenv("COLUMNS")); width > 0 {
		return width
	}
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

func columnsWidth(value string) int {
	width, err := strconv.Atoi(value)
	if err != nil || width <= 0 {
		return 0
	}
	return width
}
