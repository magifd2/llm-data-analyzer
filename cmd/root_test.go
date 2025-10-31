package cmd

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRootCmd(t *testing.T) {
	// 1. Create a mock LLM server
	analysisCount := 0
	summaryCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		if strings.Contains(string(body), "Summarize the following analysis") {
			// This is the summary call
			summaryCount++
			w.Write([]byte(`{"choices": [{"message": {"content": "Final summary."}}]}`))
		} else {
			// This is an analysis call
			analysisCount++
			w.Write([]byte(`{"choices": [{"message": {"content": "chunk summary"}}]}`))
		}
	}))
	defer mockServer.Close()

	// 2. Create a temporary config file
	configFile, _ := ioutil.TempFile("", "config-*.yaml")
	defer os.Remove(configFile.Name())
	configContent := `
endpoints:
  - name: test-endpoint
    endpoint_url: "` + mockServer.URL + `"
    api_key_env: "TEST_API_KEY"
    model: "test-model"
    context_window_size: 100
    chunk_size: 100
`
	configFile.Write([]byte(configContent))
	configFile.Close()

	// 3. Create dummy prompt and input files
	analysisPromptFile, _ := ioutil.TempFile("", "analysis-*.txt")
	defer os.Remove(analysisPromptFile.Name())
	analysisPromptFile.Write([]byte("Analyze this:"))
	analysisPromptFile.Close()

	summaryPromptFile, _ := ioutil.TempFile("", "summary-*.txt")
	defer os.Remove(summaryPromptFile.Name())
	summaryPromptFile.Write([]byte("Summarize the following analysis:"))
	summaryPromptFile.Close()

	inputFile, _ := ioutil.TempFile("", "input-*.txt")
	defer os.Remove(inputFile.Name())
	inputFile.Write([]byte(strings.Repeat("This is a test sentence. ", 20))) // >100 tokens
	inputFile.Close()

	// Set dummy API key
	os.Setenv("TEST_API_KEY", "dummy-key")

	// 4. Run the command with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rootCmd.SetArgs([]string{
		"--config", configFile.Name(),
		"--endpoint-name", "test-endpoint",
		"--analysis-prompt-file", analysisPromptFile.Name(),
		"--summary-prompt-file", summaryPromptFile.Name(),
		"--verbose",
		inputFile.Name(),
	})
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	// 5. Check assertions
	if analysisCount != 2 { // 140 tokens / 100 chunk size = 2 chunks
		t.Errorf("expected 2 analysis calls, got %d", analysisCount)
	}
	if summaryCount != 1 {
		t.Errorf("expected 1 summary call, got %d", summaryCount)
	}
}