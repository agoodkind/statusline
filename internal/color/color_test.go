package color

import (
	"fmt"
	"testing"
)

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

func TestHexForRatioSweepsPurpleBlueGreenRed(t *testing.T) {
	tests := []struct {
		name     string
		position float64
		dominant string
	}{
		{name: "purple at the left edge", position: 0, dominant: "purple"},
		{name: "blue just after the left edge", position: 40.0 / 280.0, dominant: "blue"},
		{name: "green in the middle", position: 160.0 / 280.0, dominant: "green"},
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

func TestHSVToRGBPrimaryColors(t *testing.T) {
	red, green, blue := hsvToRGB(0, 1, 1)
	if red != 255 || green != 0 || blue != 0 {
		t.Fatalf("hsvToRGB(0, 1, 1) = (%d, %d, %d), want (255, 0, 0)", red, green, blue)
	}

	red, green, blue = hsvToRGB(120, 1, 1)
	if red != 0 || green != 255 || blue != 0 {
		t.Fatalf("hsvToRGB(120, 1, 1) = (%d, %d, %d), want (0, 255, 0)", red, green, blue)
	}

	red, green, blue = hsvToRGB(240, 1, 1)
	if red != 0 || green != 0 || blue != 255 {
		t.Fatalf("hsvToRGB(240, 1, 1) = (%d, %d, %d), want (0, 0, 255)", red, green, blue)
	}
}
