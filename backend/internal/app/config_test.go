package app

import (
	"testing"
	"time"
)

func TestLoadConfigUsesDefaults(t *testing.T) {
	t.Setenv("LSI_ADDR", "")
	t.Setenv("LSI_DB_PATH", "")
	t.Setenv("LSI_WORK_DIR", "")

	config := LoadConfig()

	if config.Addr != "127.0.0.1:8080" {
		t.Fatalf("Addr = %q", config.Addr)
	}
	if config.DBPath != "./data/backend.sqlite3" {
		t.Fatalf("DBPath = %q", config.DBPath)
	}
	if config.WorkDir != "./data/jobs" {
		t.Fatalf("WorkDir = %q", config.WorkDir)
	}
}

func TestLoadConfigReadsEnvironment(t *testing.T) {
	t.Setenv("LSI_ADDR", "127.0.0.1:9090")
	t.Setenv("LSI_DB_PATH", "/tmp/test.sqlite3")
	t.Setenv("LSI_WORK_DIR", "/tmp/jobs")

	config := LoadConfig()

	if config.Addr != "127.0.0.1:9090" {
		t.Fatalf("Addr = %q", config.Addr)
	}
	if config.DBPath != "/tmp/test.sqlite3" {
		t.Fatalf("DBPath = %q", config.DBPath)
	}
	if config.WorkDir != "/tmp/jobs" {
		t.Fatalf("WorkDir = %q", config.WorkDir)
	}
}

func TestLoadConfigRunnerModeDefault(t *testing.T) {
	t.Setenv("LSI_RUNNER_MODE", "")
	config := LoadConfig()
	if config.RunnerMode != "mock" {
		t.Fatalf("RunnerMode = %q, want %q", config.RunnerMode, "mock")
	}
}

func TestLoadConfigRunnerModeCustom(t *testing.T) {
	t.Setenv("LSI_RUNNER_MODE", "real")
	config := LoadConfig()
	if config.RunnerMode != "real" {
		t.Fatalf("RunnerMode = %q, want %q", config.RunnerMode, "real")
	}
}

func TestLoadConfigDownloadTimeoutDefault(t *testing.T) {
	t.Setenv("LSI_DOWNLOAD_TIMEOUT", "")
	config := LoadConfig()
	if config.DownloadTimeout != 10*time.Minute {
		t.Fatalf("DownloadTimeout = %v, want %v", config.DownloadTimeout, 10*time.Minute)
	}
}

func TestLoadConfigDownloadTimeoutCustom(t *testing.T) {
	t.Setenv("LSI_DOWNLOAD_TIMEOUT", "5m")
	config := LoadConfig()
	if config.DownloadTimeout != 5*time.Minute {
		t.Fatalf("DownloadTimeout = %v, want %v", config.DownloadTimeout, 5*time.Minute)
	}
}

func TestLoadConfigDownloadTimeoutInvalid(t *testing.T) {
	t.Setenv("LSI_DOWNLOAD_TIMEOUT", "not-a-duration")
	config := LoadConfig()
	if config.DownloadTimeout != 10*time.Minute {
		t.Fatalf("DownloadTimeout = %v, want fallback %v", config.DownloadTimeout, 10*time.Minute)
	}
}

func TestLoadConfigWhisperModelDefault(t *testing.T) {
	t.Setenv("LSI_WHISPER_MODEL", "")
	config := LoadConfig()
	if config.WhisperModel != "small" {
		t.Fatalf("WhisperModel = %q, want %q", config.WhisperModel, "small")
	}
}

func TestLoadConfigWhisperModelCustom(t *testing.T) {
	t.Setenv("LSI_WHISPER_MODEL", "medium")
	config := LoadConfig()
	if config.WhisperModel != "medium" {
		t.Fatalf("WhisperModel = %q, want %q", config.WhisperModel, "medium")
	}
}

func TestLoadConfigLLMDefaults(t *testing.T) {
	t.Setenv("LSI_LLM_BASE_URL", "")
	t.Setenv("LSI_LLM_API_KEY", "")
	t.Setenv("LSI_LLM_MODEL", "")
	t.Setenv("LSI_LLM_TIMEOUT", "")

	config := LoadConfig()

	if config.LLMBaseURL != "https://api.openai.com" {
		t.Fatalf("LLMBaseURL = %q, want default", config.LLMBaseURL)
	}
	if config.LLMAPIKey != "" {
		t.Fatalf("LLMAPIKey = %q, want empty", config.LLMAPIKey)
	}
	if config.LLMModel != "" {
		t.Fatalf("LLMModel = %q, want empty", config.LLMModel)
	}
	if config.LLMTimeout != 2*time.Minute {
		t.Fatalf("LLMTimeout = %v, want %v", config.LLMTimeout, 2*time.Minute)
	}
}

func TestLoadConfigLLMCustomValues(t *testing.T) {
	t.Setenv("LSI_LLM_BASE_URL", "http://127.0.0.1:11434")
	t.Setenv("LSI_LLM_API_KEY", "test-key")
	t.Setenv("LSI_LLM_MODEL", "gpt-4.1-mini")
	t.Setenv("LSI_LLM_TIMEOUT", "30s")

	config := LoadConfig()

	if config.LLMBaseURL != "http://127.0.0.1:11434" {
		t.Fatalf("LLMBaseURL = %q", config.LLMBaseURL)
	}
	if config.LLMAPIKey != "test-key" {
		t.Fatalf("LLMAPIKey = %q", config.LLMAPIKey)
	}
	if config.LLMModel != "gpt-4.1-mini" {
		t.Fatalf("LLMModel = %q", config.LLMModel)
	}
	if config.LLMTimeout != 30*time.Second {
		t.Fatalf("LLMTimeout = %v", config.LLMTimeout)
	}
}
