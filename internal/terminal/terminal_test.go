package terminal

import "testing"

func TestWidthPrefersColumns(t *testing.T) {
	t.Setenv("COLUMNS", "132")

	got := Width(^uintptr(0))
	want := 132
	if got != want {
		t.Fatalf("Width() = %d, want %d", got, want)
	}
}

func TestWidthFallsBackToEightyWhenFDsAreInvalid(t *testing.T) {
	t.Setenv("COLUMNS", "")

	got := Width(^uintptr(0))
	want := 80
	if got != want {
		t.Fatalf("Width() = %d, want %d", got, want)
	}
}

func TestWidthIgnoresInvalidColumns(t *testing.T) {
	t.Setenv("COLUMNS", "not-a-width")

	got := Width(^uintptr(0))
	want := 80
	if got != want {
		t.Fatalf("Width() = %d, want %d", got, want)
	}
}
