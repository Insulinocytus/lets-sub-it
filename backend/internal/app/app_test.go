package app

import (
	"errors"
	"strings"
	"testing"

	"lets-sub-it-api/internal/store"
)

func TestCheckToolsDoesNotRequireWhisperCLI(t *testing.T) {
	origLookPath := lookPath
	t.Cleanup(func() { lookPath = origLookPath })

	lookPath = func(tool string) (string, error) {
		if tool == "whisper-cli" {
			return "", errors.New("missing")
		}
		return "/usr/bin/" + tool, nil
	}

	err := checkTools()
	if err != nil {
		t.Fatalf("checkTools() error = %v, want nil", err)
	}
}

func TestNewHTTPHandlerRequiresToolsByDefault(t *testing.T) {
	origLookPath := lookPath
	t.Cleanup(func() { lookPath = origLookPath })

	lookPath = func(tool string) (string, error) {
		if tool == "yt-dlp" {
			return "", errors.New("missing")
		}
		return "/usr/bin/" + tool, nil
	}

	_, err := NewHTTPHandler(Config{
		DBPath:             t.TempDir() + "/test.sqlite3",
		WorkDir:            t.TempDir(),
		DownloadTimeout:    0,
		WhisperModel:       "small",
		WhisperComputeType: "default",
		WhisperBaseURL:     "http://127.0.0.1:8081",
		LLMBaseURL:         "https://api.openai.com",
	})
	if err == nil {
		t.Fatal("NewHTTPHandler() error = nil, want missing yt-dlp error")
	}
	if !strings.Contains(err.Error(), "yt-dlp") {
		t.Fatalf("NewHTTPHandler() error = %q, want containing yt-dlp", err.Error())
	}
}

func TestNewHTTPHandlerMarksInterruptedJobsFailed(t *testing.T) {
	origLookPath := lookPath
	t.Cleanup(func() { lookPath = origLookPath })

	lookPath = func(tool string) (string, error) {
		return "/usr/bin/" + tool, nil
	}

	dbPath := t.TempDir() + "/test.sqlite3"
	testStore, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := testStore.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "en", "zh", t.TempDir()+"/job_1")
	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}
	if err := testStore.UpdateJobStatus(job.ID, store.StatusTranslating, store.StatusTranslating, "翻译字幕...", ""); err != nil {
		t.Fatalf("UpdateJobStatus() error = %v", err)
	}

	_, err = NewHTTPHandler(Config{
		DBPath:             dbPath,
		WorkDir:            t.TempDir(),
		DownloadTimeout:    0,
		WhisperModel:       "small",
		WhisperComputeType: "default",
		WhisperBaseURL:     "http://127.0.0.1:8081",
		LLMBaseURL:         "https://api.openai.com",
	})
	if err != nil {
		t.Fatalf("NewHTTPHandler() error = %v", err)
	}

	testStore, err = store.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	updated, err := testStore.FindJob(job.ID)
	if err != nil {
		t.Fatalf("FindJob() error = %v", err)
	}
	if updated.Status != store.StatusFailed {
		t.Fatalf("Status = %q, want %q", updated.Status, store.StatusFailed)
	}
	if updated.Stage != store.StatusTranslating {
		t.Fatalf("Stage = %q, want %q", updated.Stage, store.StatusTranslating)
	}
	if updated.ErrorMessage == nil || !strings.Contains(*updated.ErrorMessage, "重启中断") {
		t.Fatalf("ErrorMessage = %v, want 重启中断", updated.ErrorMessage)
	}
}
