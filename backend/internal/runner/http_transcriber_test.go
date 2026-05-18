package runner

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHTTPTranscriberUploadsAudioPollsDownloadsVTT(t *testing.T) {
	audioPath := filepath.Join(t.TempDir(), "audio.mp3")
	if err := os.WriteFile(audioPath, []byte("fake-audio"), 0o644); err != nil {
		t.Fatalf("os.WriteFile(audio) error = %v", err)
	}
	sourcePath := filepath.Join(t.TempDir(), "nested", "source.vtt")

	var upload multipart.Form
	var uploadAudio string
	var uploadContentLength int64
	deleteCalls := 0
	statusCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/transcriptions":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %q, want POST", r.Method)
			}
			uploadContentLength = r.ContentLength
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				t.Fatalf("ParseMultipartForm() error = %v", err)
			}
			upload = *r.MultipartForm
			file, _, err := r.FormFile("audio")
			if err != nil {
				t.Fatalf("FormFile(audio) error = %v", err)
			}
			defer file.Close()
			data, err := io.ReadAll(file)
			if err != nil {
				t.Fatalf("ReadAll(audio) error = %v", err)
			}
			uploadAudio = string(data)
			writeTranscriptionJSON(t, w, map[string]string{"id": "tx_123", "status": "queued"})
		case "/transcriptions/tx_123":
			if r.Method == http.MethodDelete {
				deleteCalls++
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if r.Method != http.MethodGet {
				t.Fatalf("method = %q, want GET or DELETE", r.Method)
			}
			statusCalls++
			if statusCalls == 1 {
				writeTranscriptionJSON(t, w, map[string]string{"id": "tx_123", "status": "running", "progressText": "正在转写音频"})
				return
			}
			writeTranscriptionJSON(t, w, map[string]string{"id": "tx_123", "status": "completed", "progressText": "转写完成"})
		case "/transcriptions/tx_123/vtt":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %q, want GET", r.Method)
			}
			_, _ = w.Write([]byte("WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nhello\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	var progress []string
	transcriber := NewHTTPTranscriber(server.URL, time.Second, time.Millisecond, server.Client())
	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		JobID:       "job_1",
		AudioPath:   audioPath,
		SourcePath:  sourcePath,
		Model:       "small",
		ComputeType: "int8",
		Language:    "ja",
		OnProgress: func(text string) error {
			progress = append(progress, text)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}

	assertMultipartField(t, upload.Value, "model", "small")
	assertMultipartField(t, upload.Value, "computeType", "int8")
	assertMultipartField(t, upload.Value, "language", "ja")
	assertMultipartField(t, upload.Value, "jobId", "job_1")
	if uploadAudio != "fake-audio" {
		t.Fatalf("uploaded audio = %q, want fake-audio", uploadAudio)
	}
	if uploadContentLength != -1 {
		t.Fatalf("upload ContentLength = %d, want -1 for streaming upload", uploadContentLength)
	}
	if statusCalls != 2 {
		t.Fatalf("statusCalls = %d, want 2", statusCalls)
	}
	if deleteCalls != 1 {
		t.Fatalf("deleteCalls = %d, want 1", deleteCalls)
	}
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("os.ReadFile(source) error = %v", err)
	}
	if !strings.Contains(string(data), "hello") {
		t.Fatalf("source VTT = %q, want downloaded content", string(data))
	}
	if len(progress) == 0 || progress[len(progress)-1] != "转写完成" {
		t.Fatalf("progress = %#v, want 转写完成 callback", progress)
	}
	if containsString(progress, "completed") {
		t.Fatalf("progress = %#v, must use progressText instead of status", progress)
	}
	if !containsString(progress, "正在转写音频") {
		t.Fatalf("progress = %#v, want running progressText", progress)
	}
}

func TestHTTPTranscriberDeletesRemoteTaskWhenContextCanceled(t *testing.T) {
	audioPath := writeTestAudio(t)
	ctx, cancel := context.WithCancel(context.Background())
	deleteCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/transcriptions":
			writeTranscriptionJSON(t, w, map[string]string{"id": "tx_cancel", "status": "queued", "progressText": "等待转写"})
		case "/transcriptions/tx_cancel":
			if r.Method == http.MethodDelete {
				deleteCalls++
				w.WriteHeader(http.StatusNoContent)
				return
			}
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL, time.Second, 50*time.Millisecond, server.Client())
	err := transcriber.Transcribe(ctx, TranscriptionRequest{
		JobID:      "job_1",
		AudioPath:  audioPath,
		SourcePath: filepath.Join(t.TempDir(), "source.vtt"),
		OnProgress: func(text string) error {
			cancel()
			return nil
		},
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want cancellation error")
	}
	if deleteCalls != 1 {
		t.Fatalf("deleteCalls = %d, want 1", deleteCalls)
	}
}

func TestHTTPTranscriberIgnoresCleanupErrorAfterSuccessfulDownload(t *testing.T) {
	audioPath := writeTestAudio(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/transcriptions":
			writeTranscriptionJSON(t, w, map[string]string{"id": "tx_cleanup", "status": "completed"})
		case "/transcriptions/tx_cleanup/vtt":
			_, _ = w.Write([]byte("WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nhello\n"))
		case "/transcriptions/tx_cleanup":
			if r.Method == http.MethodDelete {
				http.Error(w, `{"error":{"message":"cleanup unavailable"}}`, http.StatusServiceUnavailable)
				return
			}
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL, time.Second, time.Millisecond, server.Client())
	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		JobID:      "job_1",
		AudioPath:  audioPath,
		SourcePath: filepath.Join(t.TempDir(), "source.vtt"),
	})
	if err != nil {
		t.Fatalf("Transcribe() error = %v, want nil after successful download", err)
	}
}

func TestNewHTTPTranscriberDefaultsNonPositivePollInterval(t *testing.T) {
	transcriber := NewHTTPTranscriber("http://example.test", time.Second, 0, nil)
	if transcriber.pollInterval != 2*time.Second {
		t.Fatalf("pollInterval = %v, want 2s", transcriber.pollInterval)
	}
}

func TestHTTPTranscriberReturnsNon2xxStatusError(t *testing.T) {
	audioPath := writeTestAudio(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"service unavailable"}}`, http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL, time.Second, time.Millisecond, server.Client())
	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		JobID:      "job_1",
		AudioPath:  audioPath,
		SourcePath: filepath.Join(t.TempDir(), "source.vtt"),
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want status error")
	}
	if !strings.Contains(err.Error(), "status 503") || !strings.Contains(err.Error(), "service unavailable") {
		t.Fatalf("Transcribe() error = %v, want status and response message", err)
	}
}

func TestHTTPTranscriberReturnsMalformedJSONError(t *testing.T) {
	audioPath := writeTestAudio(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{"))
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL, time.Second, time.Millisecond, server.Client())
	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		JobID:      "job_1",
		AudioPath:  audioPath,
		SourcePath: filepath.Join(t.TempDir(), "source.vtt"),
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want decode error")
	}
	if !strings.Contains(err.Error(), "decode transcription response") {
		t.Fatalf("Transcribe() error = %v, want decode error", err)
	}
}

func TestHTTPTranscriberRequiresResponseIDAndStatus(t *testing.T) {
	tests := []struct {
		name          string
		transcription map[string]string
		want          string
	}{
		{name: "missing id", transcription: map[string]string{"status": "queued"}, want: "transcription response transcription.id is required"},
		{name: "missing status", transcription: map[string]string{"id": "tr_1"}, want: "transcription response transcription.status is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audioPath := writeTestAudio(t)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeTranscriptionJSON(t, w, tt.transcription)
			}))
			t.Cleanup(server.Close)

			transcriber := NewHTTPTranscriber(server.URL, time.Second, time.Millisecond, server.Client())
			err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
				JobID:      "job_1",
				AudioPath:  audioPath,
				SourcePath: filepath.Join(t.TempDir(), "source.vtt"),
			})
			if err == nil {
				t.Fatal("Transcribe() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Transcribe() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestHTTPTranscriberReturnsProgressCallbackError(t *testing.T) {
	audioPath := writeTestAudio(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeTranscriptionJSON(t, w, map[string]string{"id": "tr_1", "status": "queued", "progressText": "queued"})
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL, time.Second, time.Millisecond, server.Client())
	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		JobID:      "job_1",
		AudioPath:  audioPath,
		SourcePath: filepath.Join(t.TempDir(), "source.vtt"),
		OnProgress: func(text string) error {
			return errors.New("progress unavailable")
		},
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want progress error")
	}
	if !strings.Contains(err.Error(), "progress unavailable") {
		t.Fatalf("Transcribe() error = %v, want progress error", err)
	}
}

func TestHTTPTranscriberReturnsFailedStatusErrorMessage(t *testing.T) {
	audioPath := writeTestAudio(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/transcriptions":
			writeTranscriptionJSON(t, w, map[string]string{"id": "tx_failed", "status": "queued"})
		case "/transcriptions/tx_failed":
			writeTranscriptionJSON(t, w, map[string]string{"id": "tx_failed", "status": "failed", "errorMessage": "model download error"})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL, time.Second, time.Millisecond, server.Client())
	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		JobID:      "job_1",
		AudioPath:  audioPath,
		SourcePath: filepath.Join(t.TempDir(), "source.vtt"),
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "model download error") {
		t.Fatalf("Transcribe() error = %v, want model download error", err)
	}
}

func TestHTTPTranscriberReturnsInvalidStatusError(t *testing.T) {
	audioPath := writeTestAudio(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/transcriptions":
			writeTranscriptionJSON(t, w, map[string]string{"id": "tr_1", "status": "not-real", "progressText": "???"})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL, time.Second, time.Millisecond, server.Client())
	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		JobID:      "job_1",
		AudioPath:  audioPath,
		SourcePath: filepath.Join(t.TempDir(), "source.vtt"),
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want invalid status error")
	}
	if !strings.Contains(err.Error(), `invalid transcription status "not-real"`) {
		t.Fatalf("Transcribe() error = %v, want invalid status error", err)
	}
}

func TestHTTPTranscriberFailsOnEmptyVTT(t *testing.T) {
	audioPath := writeTestAudio(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/transcriptions":
			writeTranscriptionJSON(t, w, map[string]string{"id": "tx_empty", "status": "completed", "progressText": "转写完成"})
		case "/transcriptions/tx_empty/vtt":
			_, _ = w.Write([]byte(""))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL, time.Second, time.Millisecond, server.Client())
	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		JobID:      "job_1",
		AudioPath:  audioPath,
		SourcePath: filepath.Join(t.TempDir(), "source.vtt"),
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "empty source.vtt") {
		t.Fatalf("Transcribe() error = %v, want empty source.vtt", err)
	}
}

func TestHTTPTranscriberFailsOnInvalidVTTPrefix(t *testing.T) {
	audioPath := writeTestAudio(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/transcriptions":
			writeTranscriptionJSON(t, w, map[string]string{"id": "tx_invalid", "status": "completed", "progressText": "转写完成"})
		case "/transcriptions/tx_invalid/vtt":
			_, _ = w.Write([]byte("not webvtt"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL, time.Second, time.Millisecond, server.Client())
	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		JobID:      "job_1",
		AudioPath:  audioPath,
		SourcePath: filepath.Join(t.TempDir(), "source.vtt"),
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want invalid VTT error")
	}
	if !strings.Contains(err.Error(), "source.vtt must start with WEBVTT") {
		t.Fatalf("Transcribe() error = %v, want invalid VTT error", err)
	}
}

func TestHTTPTranscriberPreservesExistingSourceOnInvalidVTT(t *testing.T) {
	audioPath := writeTestAudio(t)
	sourcePath := filepath.Join(t.TempDir(), "source.vtt")
	oldContent := "WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nold transcript\n"
	if err := os.WriteFile(sourcePath, []byte(oldContent), 0o644); err != nil {
		t.Fatalf("os.WriteFile(source) error = %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/transcriptions":
			writeTranscriptionJSON(t, w, map[string]string{"id": "tx_invalid", "status": "completed", "progressText": "转写完成"})
		case "/transcriptions/tx_invalid/vtt":
			_, _ = w.Write([]byte("not webvtt"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	transcriber := NewHTTPTranscriber(server.URL, time.Second, time.Millisecond, server.Client())
	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		JobID:      "job_1",
		AudioPath:  audioPath,
		SourcePath: sourcePath,
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want invalid VTT error")
	}

	data, readErr := os.ReadFile(sourcePath)
	if readErr != nil {
		t.Fatalf("os.ReadFile(source) error = %v", readErr)
	}
	if string(data) != oldContent {
		t.Fatalf("source.vtt = %q, want preserved old content", string(data))
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

func writeTranscriptionJSON(t *testing.T, w http.ResponseWriter, transcription map[string]string) {
	t.Helper()
	writeJSON(t, w, map[string]any{"transcription": transcription})
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
