package summarizer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"llm-data-analyzer/pkg/llm"
	"github.com/spf13/cobra"
)

func TestSummarizer(t *testing.T) {
	// 1. Create a mock LLM server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"choices": [{"message": {"content": "summary"}}]}`))
	}))
	defer mockServer.Close()

	// 2. Create a client and summarizer
	client := llm.NewClient(mockServer.URL, "test-key", "test-model")
	dummyCmd := &cobra.Command{}
	summarizer, err := NewSummarizer(client, 10, false, dummyCmd) // small chunk size to force recursion
	if err != nil {
		t.Fatalf("failed to create summarizer: %v", err)
	}

	// 3. Summarize a long text
	longText := strings.Repeat("This is a test sentence. ", 20)
	prompt := "Summarize this:"

	result, err := summarizer.Summarize(context.Background(), longText, prompt)
	if err != nil {
		t.Fatalf("Summarize failed: %v", err)
	}

	// 4. Check result
	if result != "summary" {
		t.Errorf("unexpected summary result: %s", result)
	}
}
