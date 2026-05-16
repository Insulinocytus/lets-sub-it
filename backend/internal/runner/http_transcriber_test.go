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
	statusCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/transcriptions":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %q, want POST", r.Method)
			}
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
			if r.Method != http.MethodGet {
				t.Fatalf("method = %q, want GET", r.Method)
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
	if statusCalls != 2 {
		t.Fatalf("statusCalls = %d, want 2", statusCalls)
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
