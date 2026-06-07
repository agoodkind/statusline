package color

import "testing"

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
