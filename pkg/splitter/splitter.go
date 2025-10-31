package splitter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

// Splitter handles splitting text into chunks based on token count.
type Splitter struct {
	chunkSize int
	tkm       *tiktoken.Tiktoken
}

// NewSplitter creates a new Splitter.
func NewSplitter(chunkSize int) (*Splitter, error) {
	tkm, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("failed to get tiktoken encoding: %w", err)
	}
	return &Splitter{
		chunkSize: chunkSize,
		tkm:       tkm,
	}, nil
}

// Encode returns the token IDs for a given text.
func (s *Splitter) Encode(text string) []int {
	return s.tkm.Encode(text, nil, nil)
}

// Split reads from an io.Reader and returns a slice of strings, where each string
// is a chunk of text with at most s.chunkSize tokens.
func (s *Splitter) Split(reader io.Reader) ([]string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	tokens := s.Encode(string(content))

	var chunks []string
	for i := 0; i < len(tokens); i += s.chunkSize {
		end := i + s.chunkSize
		if end > len(tokens) {
			end = len(tokens)
		}
		chunks = append(chunks, s.tkm.Decode(tokens[i:end]))
	}

	return chunks, nil
}

// SplitJSONL reads a JSONL file from an io.Reader and groups lines into chunks.
func (s *Splitter) SplitJSONL(reader io.Reader) ([]string, error) {
	var chunks []string
	var currentChunk strings.Builder
	var currentTokenCount int

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()

		// Validate JSONL line
		var js json.RawMessage
		if err := json.Unmarshal([]byte(line), &js); err != nil {
			return nil, fmt.Errorf("invalid JSON in line: %s", line)
		}

		lineTokens := s.Encode(line)
		lineTokenCount := len(lineTokens)

		if lineTokenCount > s.chunkSize {
			return nil, fmt.Errorf("line is too long to fit in a chunk: %d tokens", lineTokenCount)
		}

		if currentTokenCount+lineTokenCount > s.chunkSize {
			// Finalize the current chunk
			chunks = append(chunks, currentChunk.String())
			// Start a new chunk
			currentChunk.Reset()
			currentChunk.WriteString(line)
			currentChunk.WriteString("\n")
			currentTokenCount = lineTokenCount
		} else {
			currentChunk.WriteString(line)
			currentChunk.WriteString("\n")
			currentTokenCount += lineTokenCount
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	// Add the last chunk if it's not empty
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks, nil
}
