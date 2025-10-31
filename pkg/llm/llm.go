package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client represents a client for an OpenAI-compatible LLM API.
type Client struct {
	EndpointURL string
	APIKey      string
	Model       string
	HTTPClient  *http.Client
}

// NewClient creates a new LLM client.
func NewClient(endpointURL, apiKey, model string) *Client {
	return &Client{
		EndpointURL: endpointURL,
		APIKey:      apiKey,
		Model:       model,
		HTTPClient:  &http.Client{},
	}
}

// ChatCompletionRequest represents the request payload for a chat completion.
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message represents a single message in a chat completion request.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents the response from a chat completion.
type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice represents a single choice in a chat completion response.
type Choice struct {
	Message Message `json:"message"`
}

// Analyze sends a prompt to the LLM and returns the response.
func (c *Client) Analyze(ctx context.Context, prompt string) (string, error) {
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	reqPayload := ChatCompletionRequest{
		Model:    c.Model,
		Messages: messages,
	}

	reqBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.EndpointURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	var respPayload ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return "", fmt.Errorf("failed to decode response payload: %w", err)
	}

	if len(respPayload.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return respPayload.Choices[0].Message.Content, nil
}
