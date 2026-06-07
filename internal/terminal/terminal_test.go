package terminal

import "testing"

func TestWidthFallsBackToEightyWhenFDsAreInvalid(t *testing.T) {
	got := Width(^uintptr(0))
	want := 80
	if got != want {
		t.Fatalf("Width() = %d, want %d", got, want)
	}
}
