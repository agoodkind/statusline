// Package transcript reads token usage from Claude Code transcript JSONL files.
package transcript

import (
	"bufio"
	"encoding/json"
	"os"
)

const (
	initialScanBufferSize = 1024 * 1024
	maxScanBufferSize     = 16 * 1024 * 1024
)

type usage struct {
	InputTokens         int `json:"input_tokens"`
	OutputTokens        int `json:"output_tokens"`
	CacheReadTokens     int `json:"cache_read_input_tokens"`
	CacheCreationTokens int `json:"cache_creation_input_tokens"`
}

type transcriptLine struct {
	Message struct {
		Usage usage `json:"usage"`
	} `json:"message"`
}

// Used reads a transcript JSONL file and returns the latest usage token count.
func Used(path string) int {
	if path == "" {
		return 0
	}

	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	used := 0
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, initialScanBufferSize), maxScanBufferSize)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var record transcriptLine
		err = json.Unmarshal(line, &record)
		if err != nil {
			continue
		}
		lineUsage := record.Message.Usage
		sum := lineUsage.InputTokens + lineUsage.CacheReadTokens + lineUsage.CacheCreationTokens
		if sum > 0 {
			used = sum
		}
	}
	if scanner.Err() != nil {
		return 0
	}
	return used
}
