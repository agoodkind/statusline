package transcript

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUsedReturnsLatestUsageTotal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	lines := []string{
		`{"message":{"usage":{"input_tokens":100,"output_tokens":900,"cache_read_input_tokens":20,"cache_creation_input_tokens":5}}}`,
		`not-json`,
		`{"message":{"usage":{"input_tokens":200,"output_tokens":900,"cache_read_input_tokens":30,"cache_creation_input_tokens":7}}}`,
	}
	content := strings.Join(lines, "\n") + "\n"

	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	got := Used(path)
	want := 237
	if got != want {
		t.Fatalf("Used() = %d, want %d", got, want)
	}
}

func TestUsedReturnsZeroForMissingTranscript(t *testing.T) {
	got := Used(filepath.Join(t.TempDir(), "missing.jsonl"))
	if got != 0 {
		t.Fatalf("Used() = %d, want 0", got)
	}
}

func TestUsedReturnsZeroForBlankPath(t *testing.T) {
	got := Used("")
	if got != 0 {
		t.Fatalf("Used() = %d, want 0", got)
	}
}
