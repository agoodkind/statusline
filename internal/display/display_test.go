package display

import "testing"

func TestHumanTokens(t *testing.T) {
	testCases := []struct {
		name string
		in   int
		want string
	}{
		{name: "plain", in: 999, want: "999"},
		{name: "negative plain", in: -999, want: "-999"},
		{name: "thousands", in: 91_000, want: "91k"},
		{name: "millions", in: 1_000_000, want: "1.0M"},
		{name: "negative thousands", in: -91_000, want: "-91k"},
		{name: "negative millions", in: -1_000_000, want: "-1.0M"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := HumanTokens(testCase.in)
			if got != testCase.want {
				t.Fatalf("HumanTokens() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestMoney(t *testing.T) {
	got := Money(2.4193)
	want := "$2.42"
	if got != want {
		t.Fatalf("Money() = %q, want %q", got, want)
	}
}
