package display

import (
	"reflect"
	"testing"
)

func TestModelLabelVariantsKeepsSimpleLabels(t *testing.T) {
	got := ModelLabelVariants("Opus")
	want := []string{"Opus"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ModelLabelVariants() = %v, want %v", got, want)
	}
}

func TestModelLabelVariantsShortensContextSuffix(t *testing.T) {
	got := ModelLabelVariants("Opus (1M Context)")
	want := []string{"Opus (1M Context)", "Opus (1M)", "Opus"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ModelLabelVariants() = %v, want %v", got, want)
	}
}

func TestModelLabelVariantsKeepsAlreadyShortParenthetical(t *testing.T) {
	got := ModelLabelVariants("Opus (1M)")
	want := []string{"Opus (1M)", "Opus"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ModelLabelVariants() = %v, want %v", got, want)
	}
}

func TestModelLabelVariantsReturnsNilForEmptyLabel(t *testing.T) {
	got := ModelLabelVariants("")
	if got != nil {
		t.Fatalf("ModelLabelVariants() = %v, want nil", got)
	}
}
