package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnalyze(t *testing.T) {
	// Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and headers
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("Invalid Authorization header: %s", r.Header.Get("Authorization"))
		}

		// Send a mock response
		w.Header().Set("Content-Type", "application/json")
		resp := ChatCompletionResponse{
			Choices: []Choice{
				{
					Message: Message{
						Role:    "assistant",
						Content: "This is a test response.",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// Create a client with the mock server's URL
	client := NewClient(mockServer.URL, "test-api-key", "test-model")

	// Call the Analyze function
	prompt := "This is a test prompt."
	response, err := client.Analyze(context.Background(), prompt)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Check the response
	expectedResponse := "This is a test response."
	if response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, response)
	}
}