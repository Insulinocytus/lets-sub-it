package app

import (
	"os"
	"time"
)

type Config struct {
	Addr               string
	DBPath             string
	WorkDir            string
	RunnerMode         string
	DownloadTimeout    time.Duration
	WhisperModel       string
	WhisperComputeType string
	LLMBaseURL         string
	LLMAPIKey          string
	LLMModel           string
	LLMTimeout         time.Duration
}

func LoadConfig() Config {
	return Config{
		Addr:               envOrDefault("LSI_ADDR", "127.0.0.1:8080"),
		DBPath:             envOrDefault("LSI_DB_PATH", "./data/backend.sqlite3"),
		WorkDir:            envOrDefault("LSI_WORK_DIR", "./data/jobs"),
		RunnerMode:         envOrDefault("LSI_RUNNER_MODE", "mock"),
		DownloadTimeout:    envDurationOrDefault("LSI_DOWNLOAD_TIMEOUT", 10*time.Minute),
		WhisperModel:       envOrDefault("LSI_WHISPER_MODEL", "small"),
		WhisperComputeType: envOrDefault("LSI_WHISPER_COMPUTE_TYPE", "default"),
		LLMBaseURL:         envOrDefault("LSI_LLM_BASE_URL", "https://api.openai.com"),
		LLMAPIKey:          os.Getenv("LSI_LLM_API_KEY"),
		LLMModel:           os.Getenv("LSI_LLM_MODEL"),
		LLMTimeout:         envDurationOrDefault("LSI_LLM_TIMEOUT", 2*time.Minute),
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}
