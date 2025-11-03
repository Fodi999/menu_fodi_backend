package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// GroqClient клиент для работы с Groq API (Llama 3-70B)
type GroqClient struct {
	APIKey     string
	Model      string
	BaseURL    string
	HTTPClient *http.Client
}

// GroqMessage сообщение для Groq API
type GroqMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

// GroqRequest запрос к Groq API
type GroqRequest struct {
	Model       string        `json:"model"`
	Messages    []GroqMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// GroqResponse ответ от Groq API
type GroqResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewGroqClient создаёт новый Groq клиент
func NewGroqClient() *GroqClient {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		apiKey = "demo-key" // fallback для dev
	}

	model := os.Getenv("GROQ_MODEL")
	if model == "" {
		model = "llama3-70b-8192" // default модель
	}

	return &GroqClient{
		APIKey:  apiKey,
		Model:   model,
		BaseURL: "https://api.groq.com/openai/v1",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Chat отправляет запрос к Groq AI
func (c *GroqClient) Chat(messages []GroqMessage, temperature float64, maxTokens int) (*GroqResponse, error) {
	req := GroqRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      false,
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

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("groq API error (status %d): %s", resp.StatusCode, string(body))
	}

	var groqResp GroqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &groqResp, nil
}

// SimpleChat упрощённая версия для быстрого запроса
func (c *GroqClient) SimpleChat(systemPrompt, userMessage string) (string, error) {
	messages := []GroqMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	resp, err := c.Chat(messages, 0.7, 1024)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from Groq")
	}

	return resp.Choices[0].Message.Content, nil
}
