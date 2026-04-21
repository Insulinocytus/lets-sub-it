package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Translator returns a translated copy of segments for the requested language.
//
// Contract:
// - the returned slice must have the same length and order as the input slice
// - Start, End, and SourceText must be preserved for each segment
// - TranslatedText must contain the translated cue text for each segment
// - implementations must not mutate the provided slice or its elements in place
type Translator interface {
	TranslateSegments(ctx context.Context, targetLanguage string, segments []Segment) ([]Segment, error)
}

type fakeTranslator struct {
	segments []Segment
}

func FakeTranslator(segments []Segment) Translator {
	cloned := make([]Segment, len(segments))
	copy(cloned, segments)
	return fakeTranslator{segments: cloned}
}

func (f fakeTranslator) TranslateSegments(_ context.Context, _ string, _ []Segment) ([]Segment, error) {
	cloned := make([]Segment, len(f.segments))
	copy(cloned, f.segments)
	return cloned, nil
}

type openAITranslator struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

func NewOpenAITranslator() Translator {
	return openAITranslator{
		baseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")), "/"),
		apiKey:  strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		model:   strings.TrimSpace(os.Getenv("OPENAI_MODEL")),
		client:  http.DefaultClient,
	}
}

func (t openAITranslator) TranslateSegments(ctx context.Context, targetLanguage string, segments []Segment) ([]Segment, error) {
	if t.baseURL == "" || t.apiKey == "" || t.model == "" {
		return nil, errors.New("OPENAI_BASE_URL, OPENAI_API_KEY, and OPENAI_MODEL are required for translation")
	}

	result := make([]Segment, len(segments))
	for i, segment := range segments {
		translatedText, err := t.translateText(ctx, targetLanguage, segment.SourceText)
		if err != nil {
			return nil, err
		}

		result[i] = Segment{
			Start:          segment.Start,
			End:            segment.End,
			SourceText:     segment.SourceText,
			TranslatedText: translatedText,
		}
	}

	return result, nil
}

func (t openAITranslator) translateText(ctx context.Context, targetLanguage string, sourceText string) (string, error) {
	payload := map[string]any{
		"model": t.model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Translate subtitle text to the requested target language. Return only the translated text.",
			},
			{
				"role": "user",
				"content": fmt.Sprintf(
					"Target language: %s\nText:\n%s",
					targetLanguage,
					sourceText,
				),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var responseBody struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		if responseBody.Error != nil && responseBody.Error.Message != "" {
			return "", errors.New(responseBody.Error.Message)
		}

		return "", fmt.Errorf("translation request failed with status %d", resp.StatusCode)
	}
	if len(responseBody.Choices) == 0 {
		return "", errors.New("translation response did not include choices")
	}

	return strings.TrimSpace(responseBody.Choices[0].Message.Content), nil
}
