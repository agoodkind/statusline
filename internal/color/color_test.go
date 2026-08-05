package color

import (
	"fmt"
	"math"
	"testing"

	"github.com/lucasb-eyer/go-colorful"
)

// barColorCount is how many colors the widest bar shows. The renderer paints
// two per cell, so this is twice the widest bar in cells.
const barColorCount = 48

// labScale converts go-colorful's distances, whose lightness runs 0 to 1, into
// the 0 to 100 lightness the CIE Lab figures below are quoted in.
const labScale = 100.0

// maxAverageStep bounds the typical perceptual step between neighbouring
// colors, which is what governs whether the sweep reads as a gradient or as
// bands. Measured 7.7. Painting one color per cell instead of two doubles it
// to about 15 and bands visibly.
const maxAverageStep = 9.0

// maxStepRatio bounds how uneven those steps are, largest over smallest.
// Measured 4.9, with the peak at the teal-to-green transition where the gamut
// widens fastest. It is bounded to catch a sweep that stalls or lurches, not
// held near 1: evenness matters far less than the steps being small.
const maxStepRatio = 6.0

func TestHexForRatioClampsPositions(t *testing.T) {
	low := HexForRatio(-1)
	zero := HexForRatio(0)
	if low != zero {
		t.Fatalf("HexForRatio(-1) = %q, want %q", low, zero)
	}

	high := HexForRatio(2)
	one := HexForRatio(1)
	if high != one {
		t.Fatalf("HexForRatio(2) = %q, want %q", high, one)
	}
}

func TestHexForRatioReturnsHexTriplet(t *testing.T) {
	got := HexForRatio(0.5)
	if len(got) != 7 {
		t.Fatalf("len(HexForRatio()) = %d, want 7", len(got))
	}
	if got[0] != '#' {
		t.Fatalf("HexForRatio()[0] = %q, want #", got[0])
	}
}

// colorPosition mirrors how the renderer samples the gradient: once per half
// cell, at that half's centre.
func colorPosition(index int) float64 {
	return (float64(index) + 0.5) / float64(barColorCount)
}

func TestHexForRatioStepsSmoothlyAcrossAFullBar(t *testing.T) {
	previous := colorAt(t, 0)
	smallest, largest, total := math.Inf(1), 0.0, 0.0

	for index := 1; index < barColorCount; index++ {
		current := colorAt(t, index)
		step := current.DistanceLab(previous) * labScale
		smallest = math.Min(smallest, step)
		largest = math.Max(largest, step)
		total += step
		previous = current
	}

	if smallest <= 0 {
		t.Fatalf("smallest step = %v, want a gradient that advances every color", smallest)
	}
	if ratio := largest / smallest; ratio > maxStepRatio {
		t.Fatalf("step ratio = %.2f (largest %.1f, smallest %.1f), want at most %.2f",
			ratio, largest, smallest, maxStepRatio)
	}
	if average := total / float64(barColorCount-1); average > maxAverageStep {
		t.Fatalf("average step = %.2f, want at most %.2f", average, maxAverageStep)
	}
}

func TestHexForRatioSweepsPurpleBlueGreenRed(t *testing.T) {
	tests := []struct {
		name     string
		position float64
		dominant string
	}{
		{name: "purple at the left edge", position: 0, dominant: "purple"},
		{name: "blue just after the left edge", position: 0.2, dominant: "blue"},
		{name: "green past the middle", position: 0.62, dominant: "green"},
		{name: "red at the right edge", position: 1, dominant: "red"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hex := HexForRatio(test.position)
			red, green, blue := parseHex(t, hex)
			got := dominantChannel(red, green, blue)
			if got != test.dominant {
				t.Fatalf("HexForRatio(%v) = %q (r=%d g=%d b=%d), want %s",
					test.position, hex, red, green, blue, test.dominant)
			}
		})
	}
}

// TestRedEndOutSaturatesFlatLightness is the reason lightness ramps down. Held
// at the base lightness the red end is limited to roughly half the chroma it
// reaches once lightness drops.
func TestRedEndOutSaturatesFlatLightness(t *testing.T) {
	ramped := maxChroma(rainbowEndHue, lightnessAt(1))
	flat := maxChroma(rainbowEndHue, baseLightness)

	if ramped <= flat {
		t.Fatalf("red chroma ramped = %.2f, flat = %.2f, want the ramp to gain chroma",
			ramped, flat)
	}
}

// lightnessTolerance keeps the endpoint check off exact float equality. Go
// permits fusing a multiply and an add into one FMA instruction, so the ramp's
// final value can differ in the last bits between architectures.
const lightnessTolerance = 1e-12

// TestLightnessHoldsBeforeTheKnee keeps the ramp confined to the warm end, so
// the rest of the sweep is unaffected by it.
func TestLightnessHoldsBeforeTheKnee(t *testing.T) {
	for _, position := range []float64{0, 0.25, 0.5, lightnessKnee} {
		if got := lightnessAt(position); got != baseLightness {
			t.Fatalf("lightnessAt(%v) = %v, want %v", position, got, baseLightness)
		}
	}
	if got := lightnessAt(1); math.Abs(got-endLightness) > lightnessTolerance {
		t.Fatalf("lightnessAt(1) = %v, want %v within %v", got, endLightness, lightnessTolerance)
	}
}

// TestGradientStaysInGamut guards against clipping. A clipped color flattens
// against its neighbour, which is the banding this gradient exists to avoid.
func TestGradientStaysInGamut(t *testing.T) {
	for index := range barColorCount {
		position := colorPosition(index)
		hue := rainbowStartHue + (rainbowEndHue-rainbowStartHue)*position
		lightness := lightnessAt(position)
		got := colorful.Hcl(hue, maxChroma(hue, lightness), lightness)
		if !got.IsValid() {
			t.Fatalf("color %d at position %.3f = %v, want a color inside sRGB",
				index, position, got)
		}
	}
}

func colorAt(t *testing.T, index int) colorful.Color {
	t.Helper()
	hex := HexForRatio(colorPosition(index))
	parsed, err := colorful.Hex(hex)
	if err != nil {
		t.Fatalf("Hex(%q) failed: %v", hex, err)
	}
	return parsed
}

func parseHex(t *testing.T, hex string) (int, int, int) {
	t.Helper()
	var red int
	var green int
	var blue int
	_, err := fmt.Sscanf(hex, "#%02X%02X%02X", &red, &green, &blue)
	if err != nil {
		t.Fatalf("Sscanf(%q) failed: %v", hex, err)
	}
	return red, green, blue
}

// dominantChannel names the hue family of an RGB triplet. Purple and blue share
// a dominant blue channel and are told apart by how much red is mixed in.
func dominantChannel(red int, green int, blue int) string {
	if blue >= red && blue >= green {
		if red > green {
			return "purple"
		}
		return "blue"
	}
	if green > red {
		return "green"
	}
	return "red"
}
