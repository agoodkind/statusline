// Package color converts context-usage ratios into status-line colors.
package color

import (
	"fmt"
	"math"
)

// The gradient sweeps hue in Oklch, a perceptually uniform space, holding
// lightness and chroma fixed. Rotating hue in HSV instead makes blue land much
// darker than yellow, and those brightness jumps read as hard bands between
// adjacent cells. Constant Oklch lightness keeps every step the same visual
// size, so the fill blends.
const (
	rainbowStartHue = 320.0 // purple
	rainbowEndHue   = 29.0  // red
	fillLightness   = 0.75
	fillChroma      = 0.10
	fullRGBValue    = 255
)

// HexForRatio maps a fill position in [0,1] to a hex color on a rainbow sweep:
// purple and blue at the left, then cyan, green, and yellow, to red.
func HexForRatio(position float64) string {
	position = max(0, min(1, position))
	hue := rainbowStartHue + (rainbowEndHue-rainbowStartHue)*position
	red, green, blue := oklchToRGB(hue, fillChroma, fillLightness)
	return fmt.Sprintf("#%02X%02X%02X", red, green, blue)
}

// oklchToRGB converts an Oklch color to 8-bit sRGB channels. Hue is in degrees.
func oklchToRGB(hue float64, chroma float64, lightness float64) (int, int, int) {
	radians := hue * math.Pi / 180.0
	greenRedAxis := chroma * math.Cos(radians)
	blueYellowAxis := chroma * math.Sin(radians)

	longCone := lightness + 0.3963377774*greenRedAxis + 0.2158037573*blueYellowAxis
	mediumCone := lightness - 0.1055613458*greenRedAxis - 0.0638541728*blueYellowAxis
	shortCone := lightness - 0.0894841775*greenRedAxis - 1.2914855480*blueYellowAxis

	longCone *= longCone * longCone
	mediumCone *= mediumCone * mediumCone
	shortCone *= shortCone * shortCone

	red := 4.0767416621*longCone - 3.3077115913*mediumCone + 0.2309699292*shortCone
	green := -1.2684380046*longCone + 2.6097574011*mediumCone - 0.3413193965*shortCone
	blue := -0.0041960863*longCone - 0.7034186147*mediumCone + 1.7076147010*shortCone

	return encodeChannel(red), encodeChannel(green), encodeChannel(blue)
}

// encodeChannel gamma-encodes one linear-light channel into an 8-bit sRGB value.
func encodeChannel(linear float64) int {
	linear = max(0, min(1, linear))

	var encoded float64
	if linear <= 0.0031308 {
		encoded = linear * 12.92
	} else {
		encoded = 1.055*math.Pow(linear, 1.0/2.4) - 0.055
	}

	return int(encoded*fullRGBValue + 0.5)
}
