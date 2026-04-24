package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"lets-sub-it-api/internal/store"
)

func TestToJobResponseUsesExpectedJSONFields(t *testing.T) {
	job := store.Job{
		ID:             "job-1",
		VideoID:        "abc123",
		YoutubeURL:     "https://www.youtube.com/watch?v=abc123",
		SourceLanguage: "ja",
		TargetLanguage: "en",
		Status:         store.StatusQueued,
		Stage:          store.StatusQueued,
		ProgressText:   "等待处理",
		ErrorMessage:   nil,
		CreatedAt:      time.Date(2026, 4, 25, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
		UpdatedAt:      time.Date(2026, 4, 25, 11, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
	}

	responseBytes, err := json.Marshal(toJobResponse(job))
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var response map[string]any
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	for _, key := range []string{"videoId", "youtubeUrl", "sourceLanguage", "targetLanguage"} {
		if _, ok := response[key]; !ok {
			t.Fatalf("response missing key %q", key)
		}
	}
	if value, ok := response["errorMessage"]; !ok || value != nil {
		t.Fatalf("errorMessage = %#v, ok = %v", value, ok)
	}
}

func TestFormatTimeReturnsUTCRFC3339(t *testing.T) {
	value := time.Date(2026, 4, 25, 18, 30, 45, 0, time.FixedZone("JST", 9*60*60))

	if got, want := formatTime(value), "2026-04-25T09:30:45Z"; got != want {
		t.Fatalf("formatTime() = %q, want %q", got, want)
	}
}

func TestWriteErrorReturnsJSONEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()

	writeError(rr, http.StatusBadRequest, "invalid_request", "bad request")

	if got, want := rr.Header().Get("Content-Type"), "application/json; charset=utf-8"; got != want {
		t.Fatalf("Content-Type = %q, want %q", got, want)
	}
	if got, want := rr.Code, http.StatusBadRequest; got != want {
		t.Fatalf("status code = %d, want %d", got, want)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	errorValue, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("body error = %#v", body["error"])
	}
	if got, want := errorValue["code"], "invalid_request"; got != want {
		t.Fatalf("error.code = %#v, want %#v", got, want)
	}
	if got, want := errorValue["message"], "bad request"; got != want {
		t.Fatalf("error.message = %#v, want %#v", got, want)
	}
}
