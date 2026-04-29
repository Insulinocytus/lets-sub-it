package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const translationContextRadius = 10

type Translator interface {
	Translate(ctx context.Context, cues []Cue, sourceLanguage string, targetLanguage string) ([]string, error)
}

type ChatTranslator struct {
	baseURL string
	apiKey  string
	model   string
	timeout time.Duration
	client  *http.Client
}

func NewChatTranslator(baseURL, apiKey, model string, timeout time.Duration, client *http.Client) *ChatTranslator {
	if client == nil {
		client = http.DefaultClient
	}
	return &ChatTranslator{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		timeout: timeout,
		client:  client,
	}
}

func (t *ChatTranslator) Translate(ctx context.Context, cues []Cue, sourceLanguage string, targetLanguage string) ([]string, error) {
	if t.model == "" {
		return nil, fmt.Errorf("LSI_LLM_MODEL is required")
	}

	translations := make([]string, 0, len(cues))
	for i := range cues {
		translation, err := t.translateOne(ctx, cues, i, sourceLanguage, targetLanguage)
		if err != nil {
			return nil, err
		}
		translations = append(translations, translation)
	}
	if len(translations) != len(cues) {
		return nil, fmt.Errorf("translation count %d does not match cue count %d", len(translations), len(cues))
	}
	return translations, nil
}

func (t *ChatTranslator) translateOne(ctx context.Context, cues []Cue, index int, sourceLanguage string, targetLanguage string) (string, error) {
	content, err := json.Marshal(newTranslationPrompt(cues, index, sourceLanguage, targetLanguage))
	if err != nil {
		return "", fmt.Errorf("marshal translation prompt: %w", err)
	}

	body, err := json.Marshal(chatCompletionRequest{
		Model: t.model,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: `Translate the target subtitle cue. Return exactly one JSON object in this schema: {"translation":"<translated target cue>"}. Do not include any keys other than "translation".`,
			},
			{Role: "user", Content: string(content)},
		},
		ResponseFormat: chatResponseFormat{Type: "json_object"},
	})
	if err != nil {
		return "", fmt.Errorf("marshal chat completion request: %w", err)
	}

	requestCtx := ctx
	var cancel context.CancelFunc
	if t.timeout > 0 {
		requestCtx, cancel = context.WithTimeout(ctx, t.timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, t.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create chat completion request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if t.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+t.apiKey)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send chat completion request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		summary, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("chat completion request failed with status %d: %s", resp.StatusCode, t.redactAPIKey(strings.TrimSpace(string(summary))))
	}

	var chatResp chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("decode chat completion response: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("chat completion response has no choices")
	}

	var translationResp translationResponse
	if err := json.Unmarshal([]byte(chatResp.Choices[0].Message.Content), &translationResp); err != nil {
		return "", fmt.Errorf("decode translation response: %w", err)
	}
	if strings.TrimSpace(translationResp.Translation) == "" {
		return "", fmt.Errorf("translation is required")
	}
	return translationResp.Translation, nil
}

func (t *ChatTranslator) redactAPIKey(message string) string {
	if t.apiKey == "" {
		return message
	}
	return strings.ReplaceAll(message, t.apiKey, "[redacted]")
}

func newTranslationPrompt(cues []Cue, index int, sourceLanguage string, targetLanguage string) translationPrompt {
	start := max(0, index-translationContextRadius)
	end := min(len(cues)-1, index+translationContextRadius)
	contextItems := make([]translationCue, 0, end-start+1)
	for i := start; i <= end; i++ {
		contextItems = append(contextItems, translationCue{
			Index:    i,
			Text:     cueText(cues[i]),
			IsTarget: i == index,
		})
	}
	return translationPrompt{
		SourceLanguage: sourceLanguage,
		TargetLanguage: targetLanguage,
		Target: translationCue{
			Index:    index,
			Text:     cueText(cues[index]),
			IsTarget: true,
		},
		Context: contextItems,
	}
}

type translationPrompt struct {
	SourceLanguage string           `json:"sourceLanguage"`
	TargetLanguage string           `json:"targetLanguage"`
	Target         translationCue   `json:"target"`
	Context        []translationCue `json:"context"`
}

type translationCue struct {
	Index    int    `json:"index"`
	Text     string `json:"text"`
	IsTarget bool   `json:"isTarget"`
}

type chatCompletionRequest struct {
	Model          string             `json:"model"`
	Messages       []chatMessage      `json:"messages"`
	ResponseFormat chatResponseFormat `json:"response_format"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponseFormat struct {
	Type string `json:"type"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

type translationResponse struct {
	Translation string `json:"translation"`
}
