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

func TestLoadConfigWhisperComputeTypeDefault(t *testing.T) {
	t.Setenv("LSI_WHISPER_COMPUTE_TYPE", "")
	config := LoadConfig()
	if config.WhisperComputeType != "default" {
		t.Fatalf("WhisperComputeType = %q, want %q", config.WhisperComputeType, "default")
	}
}

func TestLoadConfigWhisperComputeTypeCustom(t *testing.T) {
	t.Setenv("LSI_WHISPER_COMPUTE_TYPE", "int8")
	config := LoadConfig()
	if config.WhisperComputeType != "int8" {
		t.Fatalf("WhisperComputeType = %q, want %q", config.WhisperComputeType, "int8")
	}
}

func TestLoadConfigWhisperServiceDefaults(t *testing.T) {
	t.Setenv("LSI_WHISPER_BASE_URL", "")
	t.Setenv("LSI_WHISPER_TIMEOUT", "")
	t.Setenv("LSI_WHISPER_POLL_INTERVAL", "")

	config := LoadConfig()

	if config.WhisperBaseURL != "http://127.0.0.1:8081" {
		t.Fatalf("WhisperBaseURL = %q, want default", config.WhisperBaseURL)
	}
	if config.WhisperTimeout != 30*time.Minute {
		t.Fatalf("WhisperTimeout = %v, want %v", config.WhisperTimeout, 30*time.Minute)
	}
	if config.WhisperPollInterval != 2*time.Second {
		t.Fatalf("WhisperPollInterval = %v, want %v", config.WhisperPollInterval, 2*time.Second)
	}
}

func TestLoadConfigWhisperServiceCustomValues(t *testing.T) {
	t.Setenv("LSI_WHISPER_BASE_URL", "http://whisper:8081")
	t.Setenv("LSI_WHISPER_TIMEOUT", "45m")
	t.Setenv("LSI_WHISPER_POLL_INTERVAL", "500ms")

	config := LoadConfig()

	if config.WhisperBaseURL != "http://whisper:8081" {
		t.Fatalf("WhisperBaseURL = %q", config.WhisperBaseURL)
	}
	if config.WhisperTimeout != 45*time.Minute {
		t.Fatalf("WhisperTimeout = %v", config.WhisperTimeout)
	}
	if config.WhisperPollInterval != 500*time.Millisecond {
		t.Fatalf("WhisperPollInterval = %v", config.WhisperPollInterval)
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

func TestLoadConfigLogLevelDefault(t *testing.T) {
	t.Setenv("LSI_LOG_LEVEL", "")
	config := LoadConfig()
	if config.LogLevel != "info" {
		t.Fatalf("LogLevel = %q, want %q", config.LogLevel, "info")
	}
}

func TestLoadConfigLogLevelCustom(t *testing.T) {
	t.Setenv("LSI_LOG_LEVEL", "debug")
	config := LoadConfig()
	if config.LogLevel != "debug" {
		t.Fatalf("LogLevel = %q, want %q", config.LogLevel, "debug")
	}
}
