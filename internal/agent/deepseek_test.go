package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeepSeekModelSendsBoundedJSONCompletionRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/chat/completions" {
			t.Fatalf("request = %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("authorization header was not set")
		}
		var payload deepSeekRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.Model != "deepseek-chat" || payload.ResponseFormat.Type != "json_object" || len(payload.Messages) != 2 || !strings.Contains(payload.Messages[1].Content, "household") {
			t.Fatalf("unexpected request: %#v", payload)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"recommendations\":[]}"}}]}`))
	}))
	defer server.Close()

	model, err := NewDeepSeekModel(DeepSeekConfig{APIKey: "test-key", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("new model: %v", err)
	}
	completion, err := model.Complete(context.Background(), "household context")
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if string(completion) != `{"recommendations":[]}` {
		t.Fatalf("completion = %s", completion)
	}
}

func TestDeepSeekModelRejectsInvalidConfigurationAndResponses(t *testing.T) {
	if _, err := NewDeepSeekModel(DeepSeekConfig{}); err == nil {
		t.Fatalf("missing API key was accepted")
	}
	if _, err := NewDeepSeekModel(DeepSeekConfig{APIKey: "test", BaseURL: "://invalid"}); err == nil {
		t.Fatalf("invalid base URL was accepted")
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Error(writer, "secret provider detail", http.StatusUnauthorized)
	}))
	defer server.Close()
	model, err := NewDeepSeekModel(DeepSeekConfig{APIKey: "test", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("new model: %v", err)
	}
	_, err = model.Complete(context.Background(), "prompt")
	if err == nil || strings.Contains(err.Error(), "secret provider detail") || !strings.Contains(err.Error(), "401") {
		t.Fatalf("unexpected error: %v", err)
	}
}
