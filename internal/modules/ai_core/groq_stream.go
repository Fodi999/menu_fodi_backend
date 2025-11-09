package ai_core

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GroqStreamResponse represents a streaming response chunk
type GroqStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

// GroqStream represents a streaming connection
type GroqStream struct {
	reader   *bufio.Reader
	response *http.Response
}

// CreateChatCompletionStream creates a streaming chat completion
func (c *GroqClient) CreateChatCompletionStream(messages []GroqMessage, temperature float64, maxTokens int) (*GroqStream, error) {
	req := GroqRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      true, // Enable streaming
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("groq API error (status %d): %s", resp.StatusCode, string(body))
	}

	return &GroqStream{
		reader:   bufio.NewReader(resp.Body),
		response: resp,
	}, nil
}

// Recv receives the next streaming chunk
func (s *GroqStream) Recv() (*GroqStreamResponse, error) {
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// SSE format: "data: {...}"
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// Check for [DONE] marker
			if data == "[DONE]" {
				return nil, io.EOF
			}

			var chunk GroqStreamResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				// Skip invalid JSON (sometimes happens with SSE)
				continue
			}

			return &chunk, nil
		}
	}
}

// Close closes the streaming connection
func (s *GroqStream) Close() error {
	return s.response.Body.Close()
}
