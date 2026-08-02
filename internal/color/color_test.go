package color

import (
	"fmt"
	"testing"
)

// barCellCount is the widest bar the renderer draws, so it is the tightest
// spacing the gradient has to stay smooth across.
const barCellCount = 24

// maxChannelStep is the largest 8-bit jump allowed between neighbouring cells.
// Bigger steps show up as a hard band where two cells meet.
const maxChannelStep = 24

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

func TestHexForRatioStepsSmoothlyAcrossAFullBar(t *testing.T) {
	previousRed, previousGreen, previousBlue := parseHex(t, HexForRatio(0))

	for cell := 1; cell < barCellCount; cell++ {
		position := (float64(cell) + 0.5) / float64(barCellCount)
		hex := HexForRatio(position)
		red, green, blue := parseHex(t, hex)

		steps := []struct {
			channel  string
			previous int
			current  int
		}{
			{channel: "red", previous: previousRed, current: red},
			{channel: "green", previous: previousGreen, current: green},
			{channel: "blue", previous: previousBlue, current: blue},
		}
		for _, step := range steps {
			delta := step.current - step.previous
			if delta < 0 {
				delta = -delta
			}
			if delta > maxChannelStep {
				t.Fatalf("cell %d %s jumped %d (from %d to %d, %q), want at most %d",
					cell, step.channel, delta, step.previous, step.current, hex, maxChannelStep)
			}
		}

		previousRed, previousGreen, previousBlue = red, green, blue
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

func TestOklchToRGBStaysInGamut(t *testing.T) {
	for cell := range barCellCount {
		position := (float64(cell) + 0.5) / float64(barCellCount)
		hue := rainbowStartHue + (rainbowEndHue-rainbowStartHue)*position
		red, green, blue := oklchToRGB(hue, fillChroma, fillLightness)

		channels := map[string]int{"red": red, "green": green, "blue": blue}
		for name, value := range channels {
			if value <= 0 || value >= fullRGBValue {
				t.Fatalf("hue %.1f clipped %s to %d, want strictly inside 0..%d",
					hue, name, value, fullRGBValue)
			}
		}
	}
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
