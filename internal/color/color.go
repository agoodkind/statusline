// Package color converts context-usage ratios into status-line colors.
package color

import (
	"fmt"
	"math"
)

const (
	rainbowStartHue = 280.0
	rainbowEndHue   = 0.0
	fillSaturation  = 0.65
	fillValue       = 0.95
	fullRGBValue    = 255
)

// HexForRatio maps a fill position in [0,1] to a hex color on a full rainbow
// sweep: purple and blue at the left, then cyan, green, and yellow, to red.
func HexForRatio(position float64) string {
	position = max(0, min(1, position))
	hue := rainbowStartHue + (rainbowEndHue-rainbowStartHue)*position
	red, green, blue := hsvToRGB(hue, fillSaturation, fillValue)
	return fmt.Sprintf("#%02X%02X%02X", red, green, blue)
}

func hsvToRGB(hue float64, saturation float64, value float64) (int, int, int) {
	chroma := value * saturation
	secondary := chroma * (1 - math.Abs(math.Mod(hue/60.0, 2)-1))
	match := value - chroma

	var red float64
	var green float64
	var blue float64
	switch {
	case hue < 60:
		red, green, blue = chroma, secondary, 0
	case hue < 120:
		red, green, blue = secondary, chroma, 0
	case hue < 180:
		red, green, blue = 0, chroma, secondary
	case hue < 240:
		red, green, blue = 0, secondary, chroma
	case hue < 300:
		red, green, blue = secondary, 0, chroma
	default:
		red, green, blue = chroma, 0, secondary
	}

	return floatToRGB(red, match), floatToRGB(green, match), floatToRGB(blue, match)
}

func floatToRGB(component float64, match float64) int {
	return int((component+match)*fullRGBValue + 0.5)
}
