package llminference

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LLMClient provides interface for LLM inference
type LLMClient interface {
	Generate(ctx context.Context, prompt string, systemPrompt string) (string, error)
	GenerateWithContext(ctx context.Context, prompt string, context map[string]interface{}) (string, error)
}

// OllamaClient implements LLMClient for Ollama
type OllamaClient struct {
	Endpoint string
	Model    string
	Client   *http.Client
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(endpoint, model string) *OllamaClient {
	return &OllamaClient{
		Endpoint: endpoint,
		Model:    model,
		Client:   &http.Client{Timeout: 60 * time.Second},
	}
}

// Generate generates text using Ollama
func (c *OllamaClient) Generate(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	request := map[string]interface{}{
		"model":  c.Model,
		"prompt": prompt,
		"stream": false,
	}
	if systemPrompt != "" {
		request["system"] = systemPrompt
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	resp, err := c.Client.Post(
		fmt.Sprintf("%s/api/generate", c.Endpoint),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	response, ok := result["response"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	return response, nil
}

// GenerateWithContext generates text with additional context
func (c *OllamaClient) GenerateWithContext(ctx context.Context, prompt string, context map[string]interface{}) (string, error) {
	// Build context string
	contextStr := ""
	for k, v := range context {
		contextStr += fmt.Sprintf("%s: %v\n", k, v)
	}

	fullPrompt := fmt.Sprintf("Context:\n%s\n\nPrompt: %s", contextStr, prompt)
	return c.Generate(ctx, fullPrompt, "")
}

// OpenAIClient implements LLMClient for OpenAI API
type OpenAIClient struct {
	APIKey string
	Model  string
	Client *http.Client
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		APIKey: apiKey,
		Model:  model,
		Client: &http.Client{Timeout: 60 * time.Second},
	}
}

// Generate generates text using OpenAI API
func (c *OpenAIClient) Generate(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	messages := []map[string]interface{}{
		{"role": "user", "content": prompt},
	}
	if systemPrompt != "" {
		messages = append([]map[string]interface{}{
			{"role": "system", "content": systemPrompt},
		}, messages...)
	}

	request := map[string]interface{}{
		"model":    c.Model,
		"messages": messages,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid choice format")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid content format")
	}

	return content, nil
}

// GenerateWithContext generates text with additional context
func (c *OpenAIClient) GenerateWithContext(ctx context.Context, prompt string, context map[string]interface{}) (string, error) {
	contextStr := ""
	for k, v := range context {
		contextStr += fmt.Sprintf("%s: %v\n", k, v)
	}

	fullPrompt := fmt.Sprintf("Context:\n%s\n\nPrompt: %s", contextStr, prompt)
	return c.Generate(ctx, fullPrompt, "")
}
