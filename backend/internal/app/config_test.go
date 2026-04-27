package app

import (
	"os"
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
	os.Unsetenv("LSI_RUNNER_MODE")
	config := LoadConfig()
	if config.RunnerMode != "mock" {
		t.Fatalf("RunnerMode = %q, want %q", config.RunnerMode, "mock")
	}
}

func TestLoadConfigRunnerModeCustom(t *testing.T) {
	os.Setenv("LSI_RUNNER_MODE", "real")
	defer os.Unsetenv("LSI_RUNNER_MODE")
	config := LoadConfig()
	if config.RunnerMode != "real" {
		t.Fatalf("RunnerMode = %q, want %q", config.RunnerMode, "real")
	}
}

func TestLoadConfigDownloadTimeoutDefault(t *testing.T) {
	os.Unsetenv("LSI_DOWNLOAD_TIMEOUT")
	config := LoadConfig()
	if config.DownloadTimeout != 10*time.Minute {
		t.Fatalf("DownloadTimeout = %v, want %v", config.DownloadTimeout, 10*time.Minute)
	}
}

func TestLoadConfigDownloadTimeoutCustom(t *testing.T) {
	os.Setenv("LSI_DOWNLOAD_TIMEOUT", "5m")
	defer os.Unsetenv("LSI_DOWNLOAD_TIMEOUT")
	config := LoadConfig()
	if config.DownloadTimeout != 5*time.Minute {
		t.Fatalf("DownloadTimeout = %v, want %v", config.DownloadTimeout, 5*time.Minute)
	}
}

func TestLoadConfigDownloadTimeoutInvalid(t *testing.T) {
	os.Setenv("LSI_DOWNLOAD_TIMEOUT", "not-a-duration")
	defer os.Unsetenv("LSI_DOWNLOAD_TIMEOUT")
	config := LoadConfig()
	if config.DownloadTimeout != 10*time.Minute {
		t.Fatalf("DownloadTimeout = %v, want fallback %v", config.DownloadTimeout, 10*time.Minute)
	}
}
