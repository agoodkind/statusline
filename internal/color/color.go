// Package color converts context-usage ratios into status-line colors.
package color

import (
	"fmt"

	"github.com/lucasb-eyer/go-colorful"
)

// The gradient sweeps hue in CIE LCh, a perceptually uniform space. Each hue
// takes the most chroma sRGB can hold at its lightness, so no hue is capped by
// what a weaker one allows.
const (
	rainbowStartHue = 315.0 // purple
	rainbowEndHue   = 25.0  // red
)

// Lightness is flat for most of the sweep and eases down over the last stretch.
// How much chroma a hue can hold depends on its lightness, and red peaks far
// below the rest: at lightness 0.74 it reaches chroma 0.41, at 0.50 it reaches
// 0.86. Dropping lightness only at that end is what lets the red saturate,
// while every hue before the knee keeps the lightness the rest of the sweep is
// built around.
const (
	baseLightness = 0.74
	endLightness  = 0.50
	lightnessKnee = 0.80
)

// The gamut boundary has no closed form, so the chroma ceiling is found by
// bisection. The ceiling covers the most chromatic hue in sRGB with room to
// spare, and the step count resolves it far finer than 8-bit output can show.
const (
	chromaSearchCeil  = 1.8
	chromaSearchSteps = 24
)

// HexForRatio maps a fill position in [0,1] to a hex color on a rainbow sweep:
// purple and blue at the left, then cyan, green, and yellow, to a deep red.
func HexForRatio(position float64) string {
	position = max(0, min(1, position))
	hue := rainbowStartHue + (rainbowEndHue-rainbowStartHue)*position
	lightness := lightnessAt(position)
	return fmt.Sprintf("%s", colorful.Hcl(hue, maxChroma(hue, lightness), lightness).Clamped().Hex())
}

// lightnessAt holds the base lightness until the knee, then eases to the end
// lightness. The ease means the two stretches join without a visible corner.
func lightnessAt(position float64) float64 {
	if position <= lightnessKnee {
		return baseLightness
	}
	progress := (position - lightnessKnee) / (1 - lightnessKnee)
	return baseLightness + (endLightness-baseLightness)*smoothStep(progress)
}

// smoothStep eases a 0-to-1 ramp so it starts and ends flat.
func smoothStep(progress float64) float64 {
	return progress * progress * (3 - 2*progress)
}

// maxChroma returns the largest chroma that stays inside sRGB at this hue and
// lightness. Anything larger clips, and a clipped color flattens against its
// neighbour instead of advancing the gradient.
func maxChroma(hue float64, lightness float64) float64 {
	low, high := 0.0, chromaSearchCeil
	for range chromaSearchSteps {
		mid := (low + high) / 2
		if colorful.Hcl(hue, mid, lightness).IsValid() {
			low = mid
		} else {
			high = mid
		}
	}
	return low
}
