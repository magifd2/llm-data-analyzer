package splitter

import (
	"testing"
)

func TestEncode(t *testing.T) {
	s, err := NewSplitter(100)
	if err != nil {
		t.Fatalf("Failed to create splitter: %v", err)
	}
	s.Encode("hello world")
}
