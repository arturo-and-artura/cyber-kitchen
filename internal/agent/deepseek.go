package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultDeepSeekBaseURL = "https://api.deepseek.com"
	DefaultDeepSeekModel   = "deepseek-chat"
)

type DeepSeekConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Client  *http.Client
}

type DeepSeekModel struct {
	apiKey      string
	endpointURL string
	model       string
	client      *http.Client
}

func NewDeepSeekModel(config DeepSeekConfig) (*DeepSeekModel, error) {
	if strings.TrimSpace(config.APIKey) == "" {
		return nil, errors.New("deepseek API key is required")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if baseURL == "" {
		baseURL = DefaultDeepSeekBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("deepseek base URL must be an absolute URL")
	}
	model := strings.TrimSpace(config.Model)
	if model == "" {
		model = DefaultDeepSeekModel
	}
	client := config.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &DeepSeekModel{
		apiKey:      config.APIKey,
		endpointURL: baseURL + "/chat/completions",
		model:       model,
		client:      client,
	}, nil
}

type deepSeekRequest struct {
	Model          string            `json:"model"`
	Messages       []deepSeekMessage `json:"messages"`
	ResponseFormat deepSeekFormat    `json:"response_format"`
	Temperature    float64           `json:"temperature"`
}

type deepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekFormat struct {
	Type string `json:"type"`
}

type deepSeekResponse struct {
	Choices []struct {
		Message deepSeekMessage `json:"message"`
	} `json:"choices"`
}

func (m *DeepSeekModel) Complete(ctx context.Context, prompt string) ([]byte, error) {
	payload, err := json.Marshal(deepSeekRequest{
		Model: m.model,
		Messages: []deepSeekMessage{
			{Role: "system", Content: "Return only the requested JSON object. Do not use markdown fences or prose."},
			{Role: "user", Content: prompt},
		},
		ResponseFormat: deepSeekFormat{Type: "json_object"},
		Temperature:    0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("encode deepseek request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, m.endpointURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build deepseek request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+m.apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := m.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call deepseek: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
		return nil, fmt.Errorf("deepseek returned HTTP %d", response.StatusCode)
	}

	var completion deepSeekResponse
	decoder := json.NewDecoder(io.LimitReader(response.Body, 2<<20))
	if err := decoder.Decode(&completion); err != nil {
		return nil, fmt.Errorf("decode deepseek response: %w", err)
	}
	if len(completion.Choices) == 0 || strings.TrimSpace(completion.Choices[0].Message.Content) == "" {
		return nil, errors.New("deepseek returned no completion")
	}
	return []byte(completion.Choices[0].Message.Content), nil
}
