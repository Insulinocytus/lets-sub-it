# 真实音频下载 — 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 backend mock runner 的 downloading 阶段替换为真实 `yt-dlp` 调用，下载 YouTube 音频并转码为 MP3 128kbps。

**Architecture:** 新增 `RealRunner` 实现 `Runner` 接口（下载走真实 yt-dlp，后续阶段复用 mock 逻辑），通过 `LSI_RUNNER_MODE` 环境变量切换。`downloader.go` 封装 `exec.CommandContext` 调用，使用 `var execCommand` 注入方式支持 FakeExec 测试。`app.go` 启动时在 real 模式下前置检查 yt-dlp/ffmpeg 可用性。

**Tech Stack:** Go 1.22, `os/exec`, `context`, `os` (LookPath), 现有 SQLite/GORM store

---

## 文件结构

```
backend/internal/
├── app/
│   ├── config.go          # 修改：新增 RunnerMode, DownloadTimeout 字段
│   ├── config_test.go      # 修改：新增配置测试
│   └── app.go              # 修改：按模式创建 runner + 启动检查
├── runner/
│   ├── runner.go           # 不变
│   ├── mock_runner.go      # 不变
│   ├── mock_runner_test.go # 不变
│   ├── vtt.go              # 不变
│   ├── downloader.go       # 新增：downloadAudio + execCommand 变量
│   ├── downloader_test.go  # 新增：FakeExec 测试 5 个场景
│   ├── real_runner.go      # 新增：RealRunner Start 方法
│   └── real_runner_test.go # 新增：RealRunner 集成测试 3 个场景
```

---

### Task 1: 扩展 Config 结构体

**Files:**
- Modify: `backend/internal/app/config.go`
- Modify: `backend/internal/app/config_test.go`

- [ ] **Step 1: 在 `config.go` 中新增字段和解析逻辑**

```go
package app

import (
	"os"
	"time"
)

type Config struct {
	Addr            string
	DBPath          string
	WorkDir         string
	RunnerMode      string
	DownloadTimeout time.Duration
}

func LoadConfig() Config {
	return Config{
		Addr:            envOrDefault("LSI_ADDR", "127.0.0.1:8080"),
		DBPath:          envOrDefault("LSI_DB_PATH", "./data/backend.sqlite3"),
		WorkDir:         envOrDefault("LSI_WORK_DIR", "./data/jobs"),
		RunnerMode:      envOrDefault("LSI_RUNNER_MODE", "mock"),
		DownloadTimeout: envDurationOrDefault("LSI_DOWNLOAD_TIMEOUT", 10*time.Minute),
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
```

- [ ] **Step 2: 添加配置测试**

文件：`backend/internal/app/config_test.go`

```go
package app

import (
	"os"
	"testing"
	"time"
)

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
```

- [ ] **Step 3: 运行测试验证**

Run: `cd backend && mise exec -- go test ./internal/app -v -run TestLoadConfig`
Expected: 5 个新测试全部 PASS

- [ ] **Step 4: 提交**

```bash
git add backend/internal/app/config.go backend/internal/app/config_test.go
git commit -m "feat(backend): add runner mode and download timeout config"
```

---

### Task 2: 创建 downloader.go

**Files:**
- Create: `backend/internal/runner/downloader.go`

- [ ] **Step 1: 创建 downloader.go**

```go
package runner

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
)

var execCommand = exec.CommandContext

func downloadAudio(ctx context.Context, workDir string, jobID string, youtubeURL string) (string, error) {
	audioPath := filepath.Join(workDir, jobID, "audio.mp3")
	args := []string{
		"-x",
		"--audio-format", "mp3",
		"--audio-quality", "128K",
		"-o", filepath.Join(workDir, jobID, "audio.%(ext)s"),
		youtubeURL,
	}
	cmd := execCommand(ctx, "yt-dlp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w\n%s", err, string(output))
	}
	return audioPath, nil
}
```

- [ ] **Step 2: 验证编译**

Run: `cd backend && mise exec -- go build ./internal/runner`
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add backend/internal/runner/downloader.go
git commit -m "feat(backend): add downloadAudio using yt-dlp"
```

---

### Task 3: 创建 downloader_test.go

**Files:**
- Create: `backend/internal/runner/downloader_test.go`

- [ ] **Step 1: 创建 downloader_test.go**

```go
package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDownloadAudioSuccess(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	tmpDir := t.TempDir()
	jobDir := filepath.Join(tmpDir, "job_1")
	if err := os.MkdirAll(jobDir, 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "mkdir -p "+jobDir+" && echo fake-audio-data > "+jobDir+"/audio.mp3")
	}

	audioPath, err := downloadAudio(context.Background(), tmpDir, "job_1", "https://www.youtube.com/watch?v=abc123")
	if err != nil {
		t.Fatalf("downloadAudio() error = %v", err)
	}
	if audioPath != filepath.Join(jobDir, "audio.mp3") {
		t.Fatalf("audioPath = %q, want %q", audioPath, filepath.Join(jobDir, "audio.mp3"))
	}

	data, readErr := os.ReadFile(audioPath)
	if readErr != nil {
		t.Fatalf("os.ReadFile(audio.mp3) error = %v", readErr)
	}
	if len(data) == 0 {
		t.Fatal("audio.mp3 is empty")
	}
}

func TestDownloadAudioVideoUnavailable(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	tmpDir := t.TempDir()
	jobDir := filepath.Join(tmpDir, "job_1")
	os.MkdirAll(jobDir, 0o755)

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "echo 'ERROR: Video unavailable' >&2 && exit 1")
	}

	_, err := downloadAudio(context.Background(), tmpDir, "job_1", "https://www.youtube.com/watch?v=deleted")
	if err == nil {
		t.Fatal("downloadAudio() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "Video unavailable") {
		t.Fatalf("error = %q, want containing 'Video unavailable'", err.Error())
	}
}

func TestDownloadAudioTimeout(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "job_1"), 0o755)

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sleep", "10")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := downloadAudio(ctx, tmpDir, "job_1", "https://www.youtube.com/watch?v=abc123")
	if err == nil {
		t.Fatal("downloadAudio() error = nil, want context deadline exceeded")
	}
}

func TestDownloadAudioNetworkError(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "job_1"), 0o755)

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "echo 'ERROR: Unable to download webpage: network error' >&2 && exit 1")
	}

	_, err := downloadAudio(context.Background(), tmpDir, "job_1", "https://www.youtube.com/watch?v=abc123")
	if err == nil {
		t.Fatal("downloadAudio() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "network error") {
		t.Fatalf("error = %q, want containing 'network error'", err.Error())
	}
}

func TestDownloadAudioYtDlpMissing(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "job_1"), 0o755)

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "yt-dlp-this-tool-does-not-exist-xyz")
	}

	_, err := downloadAudio(context.Background(), tmpDir, "job_1", "https://www.youtube.com/watch?v=abc123")
	if err == nil {
		t.Fatal("downloadAudio() error = nil, want exec error")
	}
}
```

- [ ] **Step 2: 运行测试验证全部 PASS**

Run: `cd backend && mise exec -- go test ./internal/runner -v -run TestDownloadAudio`
Expected: 5 个测试全部 PASS

- [ ] **Step 3: 提交**

```bash
git add backend/internal/runner/downloader_test.go
git commit -m "test(backend): add downloadAudio FakeExec tests"
```

---

### Task 4: 创建 real_runner.go

**Files:**
- Create: `backend/internal/runner/real_runner.go`

- [ ] **Step 1: 创建 real_runner.go**

```go
package runner

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"lets-sub-it-api/internal/store"
)

type RealRunner struct {
	store          Store
	downloadTimeout time.Duration
}

func NewRealRunner(store Store, downloadTimeout time.Duration) *RealRunner {
	return &RealRunner{store: store, downloadTimeout: downloadTimeout}
}

func (r *RealRunner) Start(ctx context.Context, job store.Job) error {
	// Stage 1: downloading (real yt-dlp)
	if err := r.set(job.ID, store.StatusDownloading, "正在下载音频...", ""); err != nil {
		return r.fail(job.ID, store.StatusDownloading, err)
	}
	if err := ctx.Err(); err != nil {
		return r.fail(job.ID, store.StatusDownloading, err)
	}

	downloadCtx, cancel := context.WithTimeout(ctx, r.downloadTimeout)
	defer cancel()

	if _, err := downloadAudio(downloadCtx, filepath.Dir(job.WorkingDir), job.ID, job.YoutubeURL); err != nil {
		return r.fail(job.ID, store.StatusDownloading, err)
	}

	// Stage 2-5: mock (copied from MockRunner)
	if err := r.set(job.ID, store.StatusTranscribing, "生成 mock source.vtt", ""); err != nil {
		return r.fail(job.ID, store.StatusTranscribing, err)
	}
	if err := os.MkdirAll(job.WorkingDir, 0o755); err != nil {
		return r.fail(job.ID, store.StatusTranscribing, err)
	}

	sourcePath := filepath.Join(job.WorkingDir, "source.vtt")
	translatedPath := filepath.Join(job.WorkingDir, "translated.vtt")
	bilingualPath := filepath.Join(job.WorkingDir, "bilingual.vtt")

	if err := os.WriteFile(sourcePath, []byte(mockSourceVTT), 0o644); err != nil {
		return r.fail(job.ID, store.StatusTranscribing, err)
	}

	for _, progress := range []string{"1/3 segments", "2/3 segments", "3/3 segments"} {
		if err := r.set(job.ID, store.StatusTranslating, progress, ""); err != nil {
			return r.fail(job.ID, store.StatusTranslating, err)
		}
	}

	if err := os.WriteFile(translatedPath, []byte(mockTranslatedVTT), 0o644); err != nil {
		return r.fail(job.ID, store.StatusTranslating, err)
	}

	if err := r.set(job.ID, store.StatusPackaging, "生成字幕资产", ""); err != nil {
		return r.fail(job.ID, store.StatusPackaging, err)
	}
	if err := os.WriteFile(bilingualPath, []byte(mockBilingualVTT), 0o644); err != nil {
		return r.fail(job.ID, store.StatusPackaging, err)
	}

	asset := store.SubtitleAsset{
		JobID:             job.ID,
		VideoID:           job.VideoID,
		TargetLanguage:    job.TargetLanguage,
		SourceLanguage:    job.SourceLanguage,
		SourceVTTPath:     sourcePath,
		TranslatedVTTPath: translatedPath,
		BilingualVTTPath:  bilingualPath,
	}
	if err := r.store.CreateSubtitleAsset(asset); err != nil {
		return r.fail(job.ID, store.StatusPackaging, err)
	}

	if err := r.set(job.ID, store.StatusCompleted, "处理完成", ""); err != nil {
		return r.fail(job.ID, store.StatusCompleted, err)
	}
	return nil
}

func (r *RealRunner) set(jobID string, status string, progressText string, errorMessage string) error {
	return r.store.UpdateJobStatus(jobID, status, status, progressText, errorMessage)
}

func (r *RealRunner) fail(jobID string, stage string, cause error) error {
	if updateErr := r.store.UpdateJobStatus(jobID, store.StatusFailed, stage, "处理失败", cause.Error()); updateErr != nil {
		return updateErr
	}
	return cause
}
```

- [ ] **Step 2: 验证编译**

Run: `cd backend && mise exec -- go build ./...`
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add backend/internal/runner/real_runner.go
git commit -m "feat(backend): add RealRunner with real yt-dlp download stage"
```

---

### Task 5: 创建 real_runner_test.go

**Files:**
- Create: `backend/internal/runner/real_runner_test.go`

- [ ] **Step 1: 创建 real_runner_test.go**

```go
package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lets-sub-it-api/internal/store"
)

func TestRealRunnerCompletesJob(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	testStore := openTestStore(t)
	jobDir := t.TempDir()
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh-CN", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "mkdir -p "+jobDir+" && echo fake-audio-data > "+jobDir+"/audio.mp3")
	}

	if err := NewRealRunner(testStore, 10*time.Minute).Start(context.Background(), job); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	updatedJob, err := testStore.FindJob("job_1")
	if err != nil {
		t.Fatalf("FindJob() error = %v", err)
	}
	if updatedJob.Status != store.StatusCompleted {
		t.Fatalf("Status = %q, want %q", updatedJob.Status, store.StatusCompleted)
	}

	audioPath := filepath.Join(jobDir, "audio.mp3")
	data, readErr := os.ReadFile(audioPath)
	if readErr != nil {
		t.Fatalf("os.ReadFile(audio.mp3) error = %v", readErr)
	}
	if len(data) == 0 {
		t.Fatal("audio.mp3 is empty")
	}

	asset, assetErr := testStore.FindSubtitleAsset("abc123", "zh-CN")
	if assetErr != nil {
		t.Fatalf("FindSubtitleAsset() error = %v", assetErr)
	}
	for _, filePath := range []string{asset.SourceVTTPath, asset.TranslatedVTTPath, asset.BilingualVTTPath} {
		content, readErr := os.ReadFile(filePath)
		if readErr != nil {
			t.Fatalf("os.ReadFile(%q) error = %v", filePath, readErr)
		}
		if !strings.HasPrefix(string(content), "WEBVTT") {
			t.Fatalf("%q content = %q, want WEBVTT prefix", filePath, string(content))
		}
	}
}

func TestRealRunnerDownloadFailed(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	testStore := openTestStore(t)
	jobDir := t.TempDir()
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=deleted", "ja", "zh-CN", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "echo 'ERROR: Video unavailable' >&2 && exit 1")
	}

	err := NewRealRunner(testStore, 10*time.Minute).Start(context.Background(), job)
	if err == nil {
		t.Fatal("Start() error = nil, want error")
	}

	updatedJob, findErr := testStore.FindJob("job_1")
	if findErr != nil {
		t.Fatalf("FindJob() error = %v", findErr)
	}
	if updatedJob.Status != store.StatusFailed {
		t.Fatalf("Status = %q, want %q", updatedJob.Status, store.StatusFailed)
	}
	if updatedJob.Stage != store.StatusDownloading {
		t.Fatalf("Stage = %q, want %q", updatedJob.Stage, store.StatusDownloading)
	}
	if updatedJob.ErrorMessage == nil || !strings.Contains(*updatedJob.ErrorMessage, "Video unavailable") {
		t.Fatalf("ErrorMessage = %v, want containing 'Video unavailable'", updatedJob.ErrorMessage)
	}
}

func TestRealRunnerMarksCanceledJobAsFailed(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	testStore := openTestStore(t)
	jobDir := t.TempDir()
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh-CN", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sleep", "10")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewRealRunner(testStore, 10*time.Minute).Start(ctx, job)
	if err == nil {
		t.Fatal("Start() error = nil, want context canceled")
	}

	updatedJob, findErr := testStore.FindJob("job_1")
	if findErr != nil {
		t.Fatalf("FindJob() error = %v", findErr)
	}
	if updatedJob.Status != store.StatusFailed {
		t.Fatalf("Status = %q, want %q", updatedJob.Status, store.StatusFailed)
	}
	if updatedJob.Stage != store.StatusDownloading {
		t.Fatalf("Stage = %q, want %q", updatedJob.Stage, store.StatusDownloading)
	}
}
```

- [ ] **Step 2: 运行测试验证全部 PASS**

Run: `cd backend && mise exec -- go test ./internal/runner -v -run TestRealRunner`
Expected: 3 个测试全部 PASS

- [ ] **Step 3: 提交**

```bash
git add backend/internal/runner/real_runner_test.go
git commit -m "test(backend): add RealRunner integration tests"
```

---

### Task 6: 修改 app.go — runner 选择与启动检查

**Files:**
- Modify: `backend/internal/app/app.go`

- [ ] **Step 1: 修改 app.go 按模式创建 runner，添加启动工具检查**

```go
package app

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"lets-sub-it-api/internal/api"
	"lets-sub-it-api/internal/runner"
	"lets-sub-it-api/internal/store"
)

func NewHTTPHandler(config Config) (http.Handler, error) {
	if err := os.MkdirAll(filepath.Dir(config.DBPath), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(config.WorkDir, 0o755); err != nil {
		return nil, err
	}

	database, err := store.Open(config.DBPath)
	if err != nil {
		return nil, err
	}
	if err := database.Migrate(); err != nil {
		return nil, err
	}

	var jobRunner runner.Runner
	switch config.RunnerMode {
	case "real":
		if err := checkTools(); err != nil {
			return nil, err
		}
		jobRunner = runner.NewRealRunner(database, config.DownloadTimeout)
	default:
		jobRunner = runner.NewMockRunner(database)
	}

	handler := api.NewHandler(database, jobRunner, config.WorkDir)
	return api.Routes(handler), nil
}

func checkTools() error {
	for _, tool := range []string{"yt-dlp", "ffmpeg"} {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("LSI_RUNNER_MODE=real requires %s to be installed and on PATH", tool)
		}
	}
	return nil
}
```

- [ ] **Step 2: 运行所有 backend 测试验证没有回归**

Run: `cd backend && mise exec -- go test ./...`
Expected: 全部测试 PASS

- [ ] **Step 3: 验证构建**

Run: `cd backend && mise exec -- go build ./...`
Expected: 编译成功

- [ ] **Step 4: 提交**

```bash
git add backend/internal/app/app.go
git commit -m "feat(backend): wire up runner selection and tool check"
```

---

### 验证清单

实现完成后执行：

```bash
cd backend && mise exec -- go test ./...
cd backend && mise exec -- go build ./...
```

预期：全部测试 PASS，构建成功。
