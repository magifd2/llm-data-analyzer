package splitter

import (
	"strings"
	"testing"
)

func TestSplitter(t *testing.T) {
	// This text has 9 tokens according to cl100k_base
	text := "This is a test sentence for the splitter."
	chunkSize := 5

	s, err := NewSplitter(chunkSize)
	if err != nil {
		t.Fatalf("Failed to create splitter: %v", err)
	}

	reader := strings.NewReader(text)
	chunks, err := s.Split(reader)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("Expected 2 chunks, got %d", len(chunks))
	}

	// Check token count of each chunk
	for i, chunk := range chunks {
		tokens := s.tkm.Encode(chunk, nil, nil)
		if len(tokens) > chunkSize {
			t.Errorf("Chunk %d has %d tokens, which is more than the chunk size of %d", i, len(tokens), chunkSize)
		}
	}

	// Check content of chunks
	expectedChunk1 := "This is a test sentence"
	if chunks[0] != expectedChunk1 {
		t.Errorf("Expected chunk 1 to be '%s', got '%s'", expectedChunk1, chunks[0])
	}

	expectedChunk2 := " for the splitter."
	if chunks[1] != expectedChunk2 {
		t.Errorf("Expected chunk 2 to be '%s', got '%s'", expectedChunk2, chunks[1])
	}
}

func TestSplitterJSONL(t *testing.T) {
	jsonlContent := `{"key": "value1", "text": "This is the first line."}
{"key": "value2", "text": "This is the second line, which is a bit longer."}
{"key": "value3", "text": "Third line."}`

	// Let's set chunk size to 25 to make it split into 3 chunks
	chunkSize := 25

	s, err := NewSplitter(chunkSize)
	if err != nil {
		t.Fatalf("Failed to create splitter: %v", err)
	}

	reader := strings.NewReader(jsonlContent)
	chunks, err := s.SplitJSONL(reader)
	if err != nil {
		t.Fatalf("SplitJSONL failed: %v", err)
	}

	if len(chunks) != 3 {
		t.Fatalf("Expected 3 chunks, got %d", len(chunks))
	}

	// The first line should be in the first chunk
	expectedChunk1 := `{"key": "value1", "text": "This is the first line."}
`
	if chunks[0] != expectedChunk1 {
		t.Errorf("Expected chunk 1 to be '%s', got '%s'", expectedChunk1, chunks[0])
	}

	// The second line should be in the second chunk
	expectedChunk2 := `{"key": "value2", "text": "This is the second line, which is a bit longer."}
`
	if chunks[1] != expectedChunk2 {
		t.Errorf("Expected chunk 2 to be '%s', got '%s'", expectedChunk2, chunks[1])
	}

	// The third line should be in the third chunk
	expectedChunk3 := `{"key": "value3", "text": "Third line."}
`
	if chunks[2] != expectedChunk3 {
		t.Errorf("Expected chunk 3 to be '%s', got '%s'", expectedChunk3, chunks[2])
	}
}