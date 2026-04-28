package app

import (
	"errors"
	"strings"
	"testing"
)

func TestCheckToolsRequiresWhisperCLIInRealMode(t *testing.T) {
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
