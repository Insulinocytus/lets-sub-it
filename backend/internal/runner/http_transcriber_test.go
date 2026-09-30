package runner

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestHTTPTranscriberWritesSourceVTTFromVerboseJSON(t *testing.T) {
	audioPath := writeTestAudio(t)
	sourcePath := filepath.Join(t.TempDir(), "nested", "source.vtt")

	var upload *multipart.Form
	var uploadAudio string
	var uploadContentLength int64
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/audio/transcriptions" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		uploadContentLength = r.ContentLength
		authorization = r.Header.Get("Authorization")
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			t.Errorf("ParseMultipartForm() error = %v", err)
			return
		}
		upload = r.MultipartForm
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Errorf("FormFile(file) error = %v", err)
			return
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			t.Errorf("ReadAll(file) error = %v", err)
			return
		}
		uploadAudio = string(data)
		writeJSON(t, w, map[string]any{
			"task":     "transcribe",
			"language": "ja",
			"text":     "こんにちは",
			"segments": []map[string]any{
				{"id": 0, "seek": 0, "start": 0.0, "end": 1.5, "text": "こんにちは", "tokens": []int{1}, "temperature": 0.0, "avg_logprob": -0.1, "compression_ratio": 0.5, "no_speech_prob": 0.01},
				{"id": 1, "seek": 0, "start": 1.5, "end": 3.25, "text": " 世界", "tokens": []int{2}, "temperature": 0.0, "avg_logprob": -0.1, "compression_ratio": 0.5, "no_speech_prob": 0.01},
			},
		})
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL+"/v1", "secret-key", time.Second, server.Client())
	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		AudioPath:  audioPath,
		SourcePath: sourcePath,
		Model:      "small",
		Language:   "ja",
	})
	if err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}

	assertMultipartField(t, upload.Value, "model", "small")
	assertMultipartField(t, upload.Value, "language", "ja")
	assertMultipartField(t, upload.Value, "response_format", "verbose_json")
	if uploadAudio != "fake-audio" {
		t.Fatalf("uploaded audio = %q, want fake-audio", uploadAudio)
	}
	if uploadContentLength != -1 {
		t.Fatalf("upload ContentLength = %d, want -1 for streaming upload", uploadContentLength)
	}
	if authorization != "Bearer secret-key" {
		t.Fatalf("Authorization = %q, want bearer API key", authorization)
	}

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("os.ReadFile(source) error = %v", err)
	}
	want := "WEBVTT\n\n00:00:00.000 --> 00:00:01.500\nこんにちは\n\n00:00:01.500 --> 00:00:03.250\n世界\n\n"
	if string(data) != want {
		t.Fatalf("source.vtt = %q, want %q", string(data), want)
	}
}

func TestHTTPTranscriberOmitsAuthorizationAndLanguageWhenUnset(t *testing.T) {
	audioPath := writeTestAudio(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Header["Authorization"]; ok {
			t.Errorf("Authorization header = %#v, want none", r.Header["Authorization"])
		}
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			t.Errorf("ParseMultipartForm() error = %v", err)
			return
		}
		if _, ok := r.MultipartForm.Value["language"]; ok {
			t.Errorf("multipart language = %#v, want omitted", r.MultipartForm.Value["language"])
		}
		writeJSON(t, w, map[string]any{
			"segments": []map[string]any{{"start": 0.0, "end": 1.0, "text": "hello"}},
		})
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL, "", time.Second, server.Client())
	if err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		AudioPath:  audioPath,
		SourcePath: filepath.Join(t.TempDir(), "source.vtt"),
		Model:      "small",
	}); err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}
}

func TestHTTPTranscriberRejectsResponseWithoutValidSegments(t *testing.T) {
	tests := []struct {
		name     string
		response map[string]any
		want     string
	}{
		{name: "missing segments", response: map[string]any{"text": "hello"}, want: "no segments"},
		{name: "empty segments", response: map[string]any{"segments": []map[string]any{}}, want: "no segments"},
		{name: "empty text", response: map[string]any{"segments": []map[string]any{{"start": 0.0, "end": 1.0, "text": "  "}}}, want: "text is empty"},
		{name: "missing start", response: map[string]any{"segments": []map[string]any{{"end": 1.0, "text": "hello"}}}, want: "start is required"},
		{name: "missing end", response: map[string]any{"segments": []map[string]any{{"start": 0.0, "text": "hello"}}}, want: "end is required"},
		{name: "end before start", response: map[string]any{"segments": []map[string]any{{"start": 2.0, "end": 1.0, "text": "hello"}}}, want: "must be after start"},
		{name: "negative start", response: map[string]any{"segments": []map[string]any{{"start": -1.0, "end": 1.0, "text": "hello"}}}, want: "negative"},
		{name: "sub-millisecond times", response: map[string]any{"segments": []map[string]any{{"start": 0.0001, "end": 0.0002, "text": "hello"}}}, want: "not positive after rounding"},
		{name: "non-monotonic start", response: map[string]any{"segments": []map[string]any{{"start": 1.0, "end": 2.0, "text": "hello"}, {"start": 0.5, "end": 1.5, "text": "world"}}}, want: "before previous"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audioPath := writeTestAudio(t)
			sourcePath := filepath.Join(t.TempDir(), "source.vtt")
			oldContent := "WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nold transcript\n"
			if err := os.WriteFile(sourcePath, []byte(oldContent), 0o644); err != nil {
				t.Fatalf("os.WriteFile(source) error = %v", err)
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeJSON(t, w, tt.response)
			}))
			t.Cleanup(server.Close)

			transcriber := NewHTTPTranscriber(server.URL, "", time.Second, server.Client())
			err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
				AudioPath:  audioPath,
				SourcePath: sourcePath,
				Model:      "small",
			})
			if err == nil {
				t.Fatal("Transcribe() error = nil, want invalid segment error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Transcribe() error = %v, want %q", err, tt.want)
			}

			data, readErr := os.ReadFile(sourcePath)
			if readErr != nil {
				t.Fatalf("os.ReadFile(source) error = %v", readErr)
			}
			if string(data) != oldContent {
				t.Fatalf("source.vtt = %q, want preserved old content", string(data))
			}
		})
	}
}

func TestHTTPTranscriberPreservesSourceOnServerError(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		wantErr  string
		wantLeak string
	}{
		{name: "docker whisper detail", status: http.StatusBadRequest, body: `{"detail":"Model small is not available"}`, wantErr: "Model small is not available"},
		{name: "openrouter error message", status: http.StatusBadGateway, body: `{"error":{"message":"upstream failure"}}`, wantErr: "upstream failure"},
		{name: "credentials not leaked", status: http.StatusUnauthorized, body: `{"error":{"message":"unauthorized secret-key"}}`, wantErr: "unauthorized", wantLeak: "secret-key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audioPath := writeTestAudio(t)
			sourcePath := filepath.Join(t.TempDir(), "source.vtt")
			oldContent := "WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nold transcript\n"
			if err := os.WriteFile(sourcePath, []byte(oldContent), 0o644); err != nil {
				t.Fatalf("os.WriteFile(source) error = %v", err)
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.String(), "secret-key") {
					t.Errorf("request URL = %q, must not contain credentials", r.URL.String())
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			t.Cleanup(server.Close)

			transcriber := NewHTTPTranscriber(server.URL, "secret-key", time.Second, server.Client())
			err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
				AudioPath:  audioPath,
				SourcePath: sourcePath,
				Model:      "small",
			})
			if err == nil {
				t.Fatal("Transcribe() error = nil, want status error")
			}
			if !strings.Contains(err.Error(), "status "+strconv.Itoa(tt.status)) || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Transcribe() error = %v, want status and %q", err, tt.wantErr)
			}
			if tt.wantLeak != "" && strings.Contains(err.Error(), tt.wantLeak) {
				t.Fatalf("Transcribe() error = %v, must not contain credentials", err)
			}

			data, readErr := os.ReadFile(sourcePath)
			if readErr != nil {
				t.Fatalf("os.ReadFile(source) error = %v", readErr)
			}
			if string(data) != oldContent {
				t.Fatalf("source.vtt = %q, want preserved old content", string(data))
			}
		})
	}
}

func TestHTTPTranscriberReturnsDecodeError(t *testing.T) {
	audioPath := writeTestAudio(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{"))
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL, "", time.Second, server.Client())
	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		AudioPath:  audioPath,
		SourcePath: filepath.Join(t.TempDir(), "source.vtt"),
		Model:      "small",
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want decode error")
	}
	if !strings.Contains(err.Error(), "decode transcription response") {
		t.Fatalf("Transcribe() error = %v, want decode error", err)
	}
}

func TestHTTPTranscriberReturnsContextCancellationError(t *testing.T) {
	audioPath := writeTestAudio(t)
	sourcePath := filepath.Join(t.TempDir(), "source.vtt")
	oldContent := "WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nold transcript\n"
	if err := os.WriteFile(sourcePath, []byte(oldContent), 0o644); err != nil {
		t.Fatalf("os.WriteFile(source) error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server must not be reached with a canceled context")
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL, "", time.Second, server.Client())
	err := transcriber.Transcribe(ctx, TranscriptionRequest{
		AudioPath:  audioPath,
		SourcePath: sourcePath,
		Model:      "small",
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want cancellation error")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("Transcribe() error = %v, want context canceled", err)
	}

	data, readErr := os.ReadFile(sourcePath)
	if readErr != nil {
		t.Fatalf("os.ReadFile(source) error = %v", readErr)
	}
	if string(data) != oldContent {
		t.Fatalf("source.vtt = %q, want preserved old content", string(data))
	}
}

func TestHTTPTranscriberFailsWhenAudioMissing(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "source.vtt")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL, "", time.Second, server.Client())
	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		AudioPath:  filepath.Join(t.TempDir(), "missing.mp3"),
		SourcePath: sourcePath,
		Model:      "small",
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want open audio error")
	}
	if !strings.Contains(err.Error(), "open audio file") {
		t.Fatalf("Transcribe() error = %v, want open audio file", err)
	}
	if _, statErr := os.Stat(sourcePath); !os.IsNotExist(statErr) {
		t.Fatalf("source.vtt stat error = %v, want not created", statErr)
	}
}

func assertMultipartField(t *testing.T, values map[string][]string, field string, want string) {
	t.Helper()
	got := values[field]
	if len(got) != 1 || got[0] != want {
		t.Fatalf("multipart field %s = %#v, want %q", field, got, want)
	}
}

func writeTestAudio(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "audio.mp3")
	if err := os.WriteFile(path, []byte("fake-audio"), 0o644); err != nil {
		t.Fatalf("os.WriteFile(audio) error = %v", err)
	}
	return path
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("Encode(%#v) error = %v", value, err)
	}
}
