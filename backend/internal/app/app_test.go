package app

import (
	"errors"
	"strings"
	"testing"
)

func TestCheckToolsRequiresWhisperCLI(t *testing.T) {
	origLookPath := lookPath
	t.Cleanup(func() { lookPath = origLookPath })

	lookPath = func(tool string) (string, error) {
		if tool == "whisper-cli" {
			return "", errors.New("missing")
		}
		return "/usr/bin/" + tool, nil
	}

	err := checkTools()
	if err == nil {
		t.Fatal("checkTools() error = nil, want missing whisper-cli error")
	}
	if !strings.Contains(err.Error(), "whisper-cli") {
		t.Fatalf("checkTools() error = %q, want containing whisper-cli", err.Error())
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
		LLMBaseURL:         "https://api.openai.com",
	})
	if err == nil {
		t.Fatal("NewHTTPHandler() error = nil, want missing yt-dlp error")
	}
	if !strings.Contains(err.Error(), "yt-dlp") {
		t.Fatalf("NewHTTPHandler() error = %q, want containing yt-dlp", err.Error())
	}
}
