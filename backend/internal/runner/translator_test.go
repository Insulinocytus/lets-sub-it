package runner

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestChatTranslatorSendsContextWindowAndReturnsTranslations(t *testing.T) {
	cues := makeTranslatorTestCues(25)
	var requests []chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %q, want /v1/chat/completions", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want Bearer test-key", got)
		}

		var req chatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		requests = append(requests, req)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"translation\":\"译文\"}"}}]}`))
	}))
	t.Cleanup(server.Close)

	translator := NewChatTranslator(server.URL, "test-key", "test-model", time.Second, server.Client())
	translations, err := translator.Translate(context.Background(), cues, "en", "zh")
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}
	if len(translations) != len(cues) {
		t.Fatalf("len(translations) = %d, want %d", len(translations), len(cues))
	}
	if len(requests) != len(cues) {
		t.Fatalf("requests = %d, want %d", len(requests), len(cues))
	}

	content := requests[12].Messages[len(requests[12].Messages)-1].Content
	for _, want := range []string{`"index":2`, `"index":22`, `"isTarget":true`, `"index":12`} {
		if !strings.Contains(content, want) {
			t.Fatalf("request content = %s, want containing %s", content, want)
		}
	}
}

func TestChatTranslatorTruncatesContextAtEdges(t *testing.T) {
	cues := makeTranslatorTestCues(3)
	var firstRequest chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(firstRequest.Messages) == 0 {
			firstRequest = req
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"translation\":\"译文\"}"}}]}`))
	}))
	t.Cleanup(server.Close)

	translator := NewChatTranslator(server.URL, "", "test-model", time.Second, server.Client())
	if _, err := translator.Translate(context.Background(), cues, "en", "zh"); err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	content := firstRequest.Messages[len(firstRequest.Messages)-1].Content
	for _, want := range []string{`"index":0`, `"index":2`} {
		if !strings.Contains(content, want) {
			t.Fatalf("request content = %s, want containing %s", content, want)
		}
	}
	if strings.Contains(content, `"index":-1`) {
		t.Fatalf("request content = %s, want no negative index", content)
	}
}

func TestChatTranslatorPromptRequiresTranslationField(t *testing.T) {
	var request chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"translation\":\"译文\"}"}}]}`))
	}))
	t.Cleanup(server.Close)

	translator := NewChatTranslator(server.URL, "", "test-model", time.Second, server.Client())
	if _, err := translator.Translate(context.Background(), makeTranslatorTestCues(1), "en", "zh"); err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	if len(request.Messages) == 0 {
		t.Fatal("request messages are empty")
	}
	systemContent := request.Messages[0].Content
	if !strings.Contains(systemContent, `"translation"`) {
		t.Fatalf("system prompt = %q, want explicit translation field name", systemContent)
	}
}

func TestChatTranslatorRequiresModel(t *testing.T) {
	translator := NewChatTranslator("http://example.test", "test-key", "", time.Second, nil)

	_, err := translator.Translate(context.Background(), makeTranslatorTestCues(1), "en", "zh")
	if err == nil {
		t.Fatal("Translate() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "LSI_LLM_MODEL is required") {
		t.Fatalf("Translate() error = %v, want LSI_LLM_MODEL is required", err)
	}
}

func TestChatTranslatorFailsOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "secret-key upstream failed", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	translator := NewChatTranslator(server.URL, "secret-key", "test-model", time.Second, server.Client())
	_, err := translator.Translate(context.Background(), makeTranslatorTestCues(1), "en", "zh")
	if err == nil {
		t.Fatal("Translate() error = nil, want error")
	}
	if strings.Contains(err.Error(), "secret-key") {
		t.Fatalf("Translate() error leaks api key: %v", err)
	}
}

func TestChatTranslatorFailsWhenTranslationMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{}"}}]}`))
	}))
	t.Cleanup(server.Close)

	translator := NewChatTranslator(server.URL, "test-key", "test-model", time.Second, server.Client())
	_, err := translator.Translate(context.Background(), makeTranslatorTestCues(1), "en", "zh")
	if err == nil {
		t.Fatal("Translate() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "translation") {
		t.Fatalf("Translate() error = %v, want mentioning translation", err)
	}
}

func makeTranslatorTestCues(count int) []Cue {
	cues := make([]Cue, count)
	for i := range cues {
		cues[i] = Cue{
			TimeLine:  "00:00:00.000 --> 00:00:01.000",
			TextLines: []string{"cue text"},
		}
	}
	return cues
}
