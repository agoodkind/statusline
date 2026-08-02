// Package color converts context-usage ratios into status-line colors.
package color

import (
	"github.com/lucasb-eyer/go-colorful"
)

// The gradient sweeps hue in CIE LCh, a perceptually uniform space, holding
// lightness fixed. Rotating hue in HSV instead makes blue land much darker than
// yellow, and those brightness jumps read as hard bands between adjacent cells.
// Constant LCh lightness keeps every step the same visual size, so the fill
// blends. Hue angles below are LCh angles, not HSV ones.
const (
	rainbowStartHue = 315.0 // purple
	rainbowEndHue   = 25.0  // red
	// Lightness sets how much chroma the sweep can carry. 0.74 is where the
	// tightest hue holds the most: below it the whole sweep clips sooner.
	fillLightness = 0.74
)

// Chroma is the saturation knob and is fixed across the sweep. It sits at the
// highest value every hue in the sweep can hold: the teal region runs out of
// gamut first, so a larger value clips there and flattens those cells against
// each other. Letting each hue take its own maximum instead would saturate the
// green far more, but chroma would then change fast between neighbouring cells
// and reintroduce visible banding, which is what this gradient exists to avoid.
const fillChroma = 0.40

// HexForRatio maps a fill position in [0,1] to a hex color on a rainbow sweep:
// purple and blue at the left, then cyan, green, and yellow, to red.
func HexForRatio(position float64) string {
	position = max(0, min(1, position))
	hue := rainbowStartHue + (rainbowEndHue-rainbowStartHue)*position
	return fillColor(hue).Clamped().Hex()
}

// fillColor returns the gradient color at one hue angle. Clamping is the
// caller's job so tests can check whether the color was in gamut to begin with.
func fillColor(hue float64) colorful.Color {
	return colorful.Hcl(hue, fillChroma, fillLightness)
}
