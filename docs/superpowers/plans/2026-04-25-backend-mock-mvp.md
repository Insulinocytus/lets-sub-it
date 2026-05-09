# Backend Mock MVP 实施计划

> **给 agentic workers：** 必须使用子技能：推荐使用 `superpowers:subagent-driven-development`，也可以使用 `superpowers:executing-plans`，逐任务执行本计划。步骤使用 checkbox（`- [ ]`）语法跟踪。

**目标：** 在 `backend/` 下实现一个可本地运行的 Go mock backend，提供真实 HTTP API、SQLite 持久化、job 复用、mock runner 状态推进和字幕文件服务。

**架构：** `cmd/server/main.go` 只负责读取配置、初始化 GORM SQLite、创建 app 并启动 HTTP server。`internal/api` 负责请求/响应、路由、CORS 和 YouTube URL 解析；`internal/store` 负责 GORM model、migration 和查询；`internal/runner` 负责 mock 状态机和 VTT 文件生成。第一阶段外部工具全部 mock，但 API 和 SQLite 契约按前端可联调形态实现。

**技术栈：** Go, net/http, GORM, `gorm.io/driver/sqlite`, UUIDv7, SQLite, pytest 不涉及 backend，backend 测试使用 `go test`。

---

## 文件结构

- Modify: `mise.toml`
  - 在根工具链里增加 Go，保证 `mise exec -- go` 可用。
- Create: `backend/go.mod`
  - Go module 名称固定为 `lets-sub-it-api`。
  - 依赖 `gorm.io/gorm`、`gorm.io/driver/sqlite`、`github.com/google/uuid`。
- Create: `backend/cmd/server/main.go`
  - server binary 入口；加载配置、打开 SQLite、执行 migration、创建 app、启动 HTTP server。
- Create: `backend/internal/app/config.go`
  - 从 `LSI_ADDR`、`LSI_DB_PATH`、`LSI_WORK_DIR` 读取配置并提供默认值。
- Create: `backend/internal/app/app.go`
  - 组装 `store.Store`、`runner.MockRunner`、`api.Handler` 和 `http.Handler`。
- Create: `backend/internal/store/models.go`
  - 定义 `Job`、`SubtitleAsset`、状态常量和 GORM column tag。
- Create: `backend/internal/store/sqlite.go`
  - 封装 GORM DB 和 job/asset 查询写入方法。
- Create: `backend/internal/store/migrations.go`
  - 执行 `AutoMigrate`。
- Create: `backend/internal/api/response.go`
  - JSON 成功/错误响应、公共 response DTO。
- Create: `backend/internal/api/youtube.go`
  - 解析 `watch?v=` 与 `youtu.be` URL。
- Create: `backend/internal/api/routes.go`
  - 注册路由、基础 CORS middleware、method dispatch。
- Create: `backend/internal/api/handler.go`
  - 实现 `POST /jobs`、`GET /jobs/{id}`、`GET /subtitle-assets`、`GET /subtitle-files/{jobId}/{mode}`。
- Create: `backend/internal/runner/runner.go`
  - 定义 `Runner` 接口和 mock runner 依赖接口。
- Create: `backend/internal/runner/vtt.go`
  - 生成固定 mock `source.vtt`、`translated.vtt`、`bilingual.vtt`。
- Create: `backend/internal/runner/mock_runner.go`
  - 后台推进状态、写 VTT、创建 asset、失败时落库。
- Test: `backend/internal/app/config_test.go`
  - 覆盖默认配置与环境变量覆盖。
- Test: `backend/internal/api/youtube_test.go`
  - 覆盖支持/不支持的 YouTube URL。
- Test: `backend/internal/api/routes_test.go`
  - 覆盖 CORS 本地来源白名单。
- Test: `backend/internal/store/sqlite_test.go`
  - 覆盖 migration、job 创建查询、复用查询、asset 查询。
- Test: `backend/internal/runner/mock_runner_test.go`
  - 覆盖 mock runner 完成路径和失败路径。
- Test: `backend/internal/api/handler_test.go`
  - 覆盖前端依赖的 HTTP API 契约。

## Task 1: Scaffold Go Module And Config

**Files:**
- Modify: `mise.toml`
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/internal/app/config.go`
- Test: `backend/internal/app/config_test.go`

- [ ] **Step 1: 写配置测试**

Create `backend/internal/app/config_test.go`:

```go
package app

import "testing"

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
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
cd backend
mise exec -- go test ./internal/app
```

Expected:

```text
FAIL，因为 `go.mod` 或 `internal/app/config.go` 还不存在。
```

- [ ] **Step 3: 更新工具链、创建 Go module 和配置实现**

Modify `mise.toml`:

```toml
[tools]
python = "3.12"
uv = "latest"
go = "1.22"
```

Create `backend/go.mod`:

```go
module lets-sub-it-api

go 1.22

require (
	github.com/google/uuid v1.6.0
	gorm.io/driver/sqlite v1.5.7
	gorm.io/gorm v1.25.12
)
```

Create `backend/internal/app/config.go`:

```go
package app

import "os"

type Config struct {
	Addr    string
	DBPath  string
	WorkDir string
}

func LoadConfig() Config {
	return Config{
		Addr:    envOrDefault("LSI_ADDR", "127.0.0.1:8080"),
		DBPath:  envOrDefault("LSI_DB_PATH", "./data/backend.sqlite3"),
		WorkDir: envOrDefault("LSI_WORK_DIR", "./data/jobs"),
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
```

Create `backend/cmd/server/main.go`:

```go
package main

import (
	"log"

	"lets-sub-it-api/internal/app"
)

func main() {
	config := app.LoadConfig()
	log.Printf("starting lets-sub-it-api on %s", config.Addr)
}
```

- [ ] **Step 4: 同步依赖并运行配置测试**

Run:

```bash
cd backend
mise exec -- go mod tidy
mise exec -- go test ./internal/app
```

Expected:

```text
ok  	lets-sub-it-api/internal/app
```

- [ ] **Step 5: 提交 scaffold**

Run:

```bash
git add mise.toml backend/go.mod backend/go.sum backend/cmd/server/main.go backend/internal/app/config.go backend/internal/app/config_test.go
git commit -m "build(backend): scaffold Go API service"
```

## Task 2: Add Store Models And Migration

**Files:**
- Create: `backend/internal/store/models.go`
- Create: `backend/internal/store/migrations.go`
- Create: `backend/internal/store/sqlite.go`
- Test: `backend/internal/store/sqlite_test.go`

- [ ] **Step 1: 写 migration 和 CRUD 测试**

Create `backend/internal/store/sqlite_test.go`:

```go
package store

import (
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(filepath.Join(t.TempDir(), "test.sqlite3"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := store.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	return store
}

func TestStoreCreatesAndFindsJob(t *testing.T) {
	store := openTestStore(t)
	job := NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh-CN", "/tmp/job_1")

	if err := store.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	found, err := store.FindJob("job_1")
	if err != nil {
		t.Fatalf("FindJob() error = %v", err)
	}
	if found.VideoID != "abc123" || found.SourceLanguage != "ja" || found.TargetLanguage != "zh-CN" {
		t.Fatalf("found job = %+v", found)
	}
}

func TestStoreFindsReusableJob(t *testing.T) {
	store := openTestStore(t)
	job := NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh-CN", "/tmp/job_1")
	if err := store.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	found, err := store.FindReusableJob("abc123", "zh-CN")
	if err != nil {
		t.Fatalf("FindReusableJob() error = %v", err)
	}
	if found.ID != "job_1" {
		t.Fatalf("reused job id = %q", found.ID)
	}

	if err := store.UpdateJobStatus("job_1", StatusFailed, StatusTranscribing, "失败", "boom"); err != nil {
		t.Fatalf("UpdateJobStatus() error = %v", err)
	}
	if _, err := store.FindReusableJob("abc123", "zh-CN"); err == nil {
		t.Fatal("FindReusableJob() expected not found for failed job")
	}
}

func TestStoreCreatesAndFindsSubtitleAsset(t *testing.T) {
	store := openTestStore(t)
	job := NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh-CN", "/tmp/job_1")
	if err := store.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}
	asset := SubtitleAsset{
		JobID:             "job_1",
		VideoID:           "abc123",
		TargetLanguage:    "zh-CN",
		SourceLanguage:    "ja",
		SourceVTTPath:     "/tmp/job_1/source.vtt",
		TranslatedVTTPath: "/tmp/job_1/translated.vtt",
		BilingualVTTPath:  "/tmp/job_1/bilingual.vtt",
	}

	if err := store.CreateSubtitleAsset(asset); err != nil {
		t.Fatalf("CreateSubtitleAsset() error = %v", err)
	}
	found, err := store.FindSubtitleAsset("abc123", "zh-CN")
	if err != nil {
		t.Fatalf("FindSubtitleAsset() error = %v", err)
	}
	if found.JobID != "job_1" || found.SourceLanguage != "ja" {
		t.Fatalf("asset = %+v", found)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
cd backend
mise exec -- go test ./internal/store
```

Expected:

```text
FAIL，因为 `store` package 还不完整。
```

- [ ] **Step 3: 实现 model、migration 和 store 方法**

Create `backend/internal/store/models.go`:

```go
package store

import "time"

const (
	StatusQueued       = "queued"
	StatusDownloading  = "downloading"
	StatusTranscribing = "transcribing"
	StatusTranslating  = "translating"
	StatusPackaging    = "packaging"
	StatusCompleted    = "completed"
	StatusFailed       = "failed"
)

type Job struct {
	ID             string  `gorm:"primaryKey;column:id"`
	VideoID        string  `gorm:"column:video_id;index:idx_jobs_lookup"`
	YoutubeURL     string  `gorm:"column:youtube_url"`
	SourceLanguage string  `gorm:"column:source_language"`
	TargetLanguage string  `gorm:"column:target_language;index:idx_jobs_lookup"`
	Status         string  `gorm:"column:status;index:idx_jobs_lookup"`
	Stage          string  `gorm:"column:stage"`
	ProgressText   string  `gorm:"column:progress_text"`
	ErrorMessage   *string `gorm:"column:error_message"`
	Attempt        int     `gorm:"column:attempt"`
	WorkingDir     string  `gorm:"column:working_dir"`
	CreatedAt      time.Time
	UpdatedAt      time.Time `gorm:"index:idx_jobs_lookup"`
}

type SubtitleAsset struct {
	JobID             string `gorm:"primaryKey;column:job_id"`
	VideoID           string `gorm:"column:video_id;index:idx_subtitle_assets_lookup"`
	TargetLanguage    string `gorm:"column:target_language;index:idx_subtitle_assets_lookup"`
	SourceVTTPath     string `gorm:"column:source_vtt_path"`
	TranslatedVTTPath string `gorm:"column:translated_vtt_path"`
	BilingualVTTPath  string `gorm:"column:bilingual_vtt_path"`
	SourceLanguage    string `gorm:"column:source_language"`
	CreatedAt         time.Time
}

func NewJob(id string, videoID string, youtubeURL string, sourceLanguage string, targetLanguage string, workingDir string) Job {
	return Job{
		ID:             id,
		VideoID:        videoID,
		YoutubeURL:     youtubeURL,
		SourceLanguage: sourceLanguage,
		TargetLanguage: targetLanguage,
		Status:         StatusQueued,
		Stage:          StatusQueued,
		ProgressText:   "等待处理",
		Attempt:        1,
		WorkingDir:     workingDir,
	}
}
```

Create `backend/internal/store/sqlite.go`:

```go
package store

import (
	"errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	db *gorm.DB
}

func Open(path string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) CreateJob(job Job) error {
	return s.db.Create(&job).Error
}

func (s *Store) FindJob(id string) (Job, error) {
	var job Job
	if err := s.db.First(&job, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Job{}, ErrNotFound
		}
		return Job{}, err
	}
	return job, nil
}

func (s *Store) FindReusableJob(videoID string, targetLanguage string) (Job, error) {
	var job Job
	err := s.db.Where("video_id = ? AND target_language = ? AND status <> ?", videoID, targetLanguage, StatusFailed).
		Order("updated_at DESC").
		First(&job).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Job{}, ErrNotFound
		}
		return Job{}, err
	}
	return job, nil
}

func (s *Store) UpdateJobStatus(id string, status string, stage string, progressText string, errorMessage string) error {
	updates := map[string]any{
		"status":        status,
		"stage":         stage,
		"progress_text": progressText,
	}
	if errorMessage != "" {
		updates["error_message"] = errorMessage
	} else {
		updates["error_message"] = nil
	}
	result := s.db.Model(&Job{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) CreateSubtitleAsset(asset SubtitleAsset) error {
	return s.db.Create(&asset).Error
}

func (s *Store) FindSubtitleAsset(videoID string, targetLanguage string) (SubtitleAsset, error) {
	var asset SubtitleAsset
	if err := s.db.First(&asset, "video_id = ? AND target_language = ?", videoID, targetLanguage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SubtitleAsset{}, ErrNotFound
		}
		return SubtitleAsset{}, err
	}
	return asset, nil
}
```

Create `backend/internal/store/migrations.go`:

```go
package store

func (s *Store) Migrate() error {
	return s.db.AutoMigrate(&Job{}, &SubtitleAsset{})
}
```

- [ ] **Step 4: 运行 store 测试**

Run:

```bash
cd backend
mise exec -- go test ./internal/store
```

Expected:

```text
ok  	lets-sub-it-api/internal/store
```

- [ ] **Step 5: 提交 store**

Run:

```bash
git add backend/internal/store backend/go.mod backend/go.sum
git commit -m "feat(backend): add SQLite store models"
```

## Task 3: Add YouTube Parsing And JSON Responses

**Files:**
- Create: `backend/internal/api/youtube.go`
- Create: `backend/internal/api/response.go`
- Test: `backend/internal/api/youtube_test.go`

- [ ] **Step 1: 写 YouTube URL 解析测试**

Create `backend/internal/api/youtube_test.go`:

```go
package api

import "testing"

func TestParseVideoIDSupportsWatchURL(t *testing.T) {
	videoID, err := ParseVideoID("https://www.youtube.com/watch?v=abc123")
	if err != nil {
		t.Fatalf("ParseVideoID() error = %v", err)
	}
	if videoID != "abc123" {
		t.Fatalf("videoID = %q", videoID)
	}
}

func TestParseVideoIDSupportsShortURL(t *testing.T) {
	videoID, err := ParseVideoID("https://youtu.be/abc123")
	if err != nil {
		t.Fatalf("ParseVideoID() error = %v", err)
	}
	if videoID != "abc123" {
		t.Fatalf("videoID = %q", videoID)
	}
}

func TestParseVideoIDRejectsUnsupportedURL(t *testing.T) {
	if _, err := ParseVideoID("https://example.com/watch?v=abc123"); err == nil {
		t.Fatal("ParseVideoID() expected error")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
cd backend
mise exec -- go test ./internal/api -run TestParseVideoID
```

Expected:

```text
FAIL，因为 `ParseVideoID` 还未定义。
```

- [ ] **Step 3: 实现 YouTube URL 解析和响应 helper**

Create `backend/internal/api/youtube.go`:

```go
package api

import (
	"errors"
	"net/url"
	"strings"
)

var ErrInvalidYouTubeURL = errors.New("invalid youtube url")

func ParseVideoID(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidYouTubeURL
	}
	host := strings.ToLower(parsed.Host)
	if host == "www.youtube.com" || host == "youtube.com" {
		videoID := parsed.Query().Get("v")
		if videoID == "" {
			return "", ErrInvalidYouTubeURL
		}
		return videoID, nil
	}
	if host == "youtu.be" {
		videoID := strings.Trim(parsed.Path, "/")
		if videoID == "" || strings.Contains(videoID, "/") {
			return "", ErrInvalidYouTubeURL
		}
		return videoID, nil
	}
	return "", ErrInvalidYouTubeURL
}
```

Create `backend/internal/api/response.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"lets-sub-it-api/internal/store"
)

type errorBody struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type jobResponse struct {
	ID             string  `json:"id"`
	VideoID        string  `json:"videoId"`
	YoutubeURL     string  `json:"youtubeUrl"`
	SourceLanguage string  `json:"sourceLanguage"`
	TargetLanguage string  `json:"targetLanguage"`
	Status         string  `json:"status"`
	Stage          string  `json:"stage"`
	ProgressText   string  `json:"progressText"`
	ErrorMessage   *string `json:"errorMessage"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, errorBody{Error: apiError{Code: code, Message: message}})
}

func toJobResponse(job store.Job) jobResponse {
	return jobResponse{
		ID:             job.ID,
		VideoID:        job.VideoID,
		YoutubeURL:     job.YoutubeURL,
		SourceLanguage: job.SourceLanguage,
		TargetLanguage: job.TargetLanguage,
		Status:         job.Status,
		Stage:          job.Stage,
		ProgressText:   job.ProgressText,
		ErrorMessage:   job.ErrorMessage,
		CreatedAt:      formatTime(job.CreatedAt),
		UpdatedAt:      formatTime(job.UpdatedAt),
	}
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}
```

- [ ] **Step 4: 运行 api 基础测试**

Run:

```bash
cd backend
mise exec -- go test ./internal/api -run TestParseVideoID
```

Expected:

```text
ok  	lets-sub-it-api/internal/api
```

- [ ] **Step 5: 提交 api 基础 helper**

Run:

```bash
git add backend/internal/api/youtube.go backend/internal/api/youtube_test.go backend/internal/api/response.go
git commit -m "feat(backend): add API parsing helpers"
```

## Task 4: Add Mock VTT And Runner

**Files:**
- Create: `backend/internal/runner/runner.go`
- Create: `backend/internal/runner/vtt.go`
- Create: `backend/internal/runner/mock_runner.go`
- Test: `backend/internal/runner/mock_runner_test.go`

- [ ] **Step 1: 写 mock runner 测试**

Create `backend/internal/runner/mock_runner_test.go`:

```go
package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"lets-sub-it-api/internal/store"
)

func TestMockRunnerCompletesJobAndWritesAssets(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "test.sqlite3")
	jobDir := filepath.Join(t.TempDir(), "job_1")
	testStore, err := store.Open(storePath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := testStore.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh-CN", jobDir)
	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	runner := NewMockRunner(testStore)
	if err := runner.Start(context.Background(), job); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	finished, err := testStore.FindJob("job_1")
	if err != nil {
		t.Fatalf("FindJob() error = %v", err)
	}
	if finished.Status != store.StatusCompleted {
		t.Fatalf("status = %q", finished.Status)
	}
	asset, err := testStore.FindSubtitleAsset("abc123", "zh-CN")
	if err != nil {
		t.Fatalf("FindSubtitleAsset() error = %v", err)
	}
	for _, path := range []string{asset.SourceVTTPath, asset.TranslatedVTTPath, asset.BilingualVTTPath} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", path, err)
		}
		if len(content) == 0 || string(content[:6]) != "WEBVTT" {
			t.Fatalf("invalid VTT at %q", path)
		}
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
cd backend
mise exec -- go test ./internal/runner
```

Expected:

```text
FAIL，因为 `NewMockRunner` 还未定义。
```

- [ ] **Step 3: 实现 runner 和 VTT 生成**

Create `backend/internal/runner/runner.go`:

```go
package runner

import (
	"context"

	"lets-sub-it-api/internal/store"
)

type Runner interface {
	Start(ctx context.Context, job store.Job) error
}

type Store interface {
	UpdateJobStatus(id string, status string, stage string, progressText string, errorMessage string) error
	CreateSubtitleAsset(asset store.SubtitleAsset) error
}
```

Create `backend/internal/runner/vtt.go`:

```go
package runner

const mockSourceVTT = `WEBVTT

00:00:00.000 --> 00:00:02.000
これは mock source 字幕の一行目です。

00:00:02.000 --> 00:00:04.000
これは mock source 字幕の二行目です。

00:00:04.000 --> 00:00:06.000
これは mock source 字幕の三行目です。
`

const mockTranslatedVTT = `WEBVTT

00:00:00.000 --> 00:00:02.000
这是 mock 翻译字幕第一行。

00:00:02.000 --> 00:00:04.000
这是 mock 翻译字幕第二行。

00:00:04.000 --> 00:00:06.000
这是 mock 翻译字幕第三行。
`

const mockBilingualVTT = `WEBVTT

00:00:00.000 --> 00:00:02.000
これは mock source 字幕の一行目です。
这是 mock 翻译字幕第一行。

00:00:02.000 --> 00:00:04.000
これは mock source 字幕の二行目です。
这是 mock 翻译字幕第二行。

00:00:04.000 --> 00:00:06.000
これは mock source 字幕の三行目です。
这是 mock 翻译字幕第三行。
`
```

Create `backend/internal/runner/mock_runner.go`:

```go
package runner

import (
	"context"
	"os"
	"path/filepath"

	"lets-sub-it-api/internal/store"
)

type MockRunner struct {
	store Store
}

func NewMockRunner(store Store) *MockRunner {
	return &MockRunner{store: store}
}

func (r *MockRunner) Start(ctx context.Context, job store.Job) error {
	if err := r.set(job.ID, store.StatusDownloading, "准备 mock 媒体", ""); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return r.fail(job.ID, store.StatusDownloading, err)
	}
	if err := r.set(job.ID, store.StatusTranscribing, "生成 mock source.vtt", ""); err != nil {
		return err
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
			return err
		}
	}
	if err := os.WriteFile(translatedPath, []byte(mockTranslatedVTT), 0o644); err != nil {
		return r.fail(job.ID, store.StatusTranslating, err)
	}
	if err := r.set(job.ID, store.StatusPackaging, "生成字幕资产", ""); err != nil {
		return err
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
	return r.set(job.ID, store.StatusCompleted, "处理完成", "")
}

func (r *MockRunner) set(jobID string, status string, progressText string, errorMessage string) error {
	return r.store.UpdateJobStatus(jobID, status, status, progressText, errorMessage)
}

func (r *MockRunner) fail(jobID string, stage string, cause error) error {
	_ = r.store.UpdateJobStatus(jobID, store.StatusFailed, stage, "处理失败", cause.Error())
	return cause
}
```

- [ ] **Step 4: 运行 runner 测试**

Run:

```bash
cd backend
mise exec -- go test ./internal/runner
```

Expected:

```text
ok  	lets-sub-it-api/internal/runner
```

- [ ] **Step 5: 提交 runner**

Run:

```bash
git add backend/internal/runner backend/internal/store
git commit -m "feat(backend): add mock job runner"
```

## Task 5: Add HTTP Routes And CORS

**Files:**
- Create: `backend/internal/api/routes.go`
- Test: `backend/internal/api/routes_test.go`

- [ ] **Step 1: 写 CORS 测试**

Create `backend/internal/api/routes_test.go`:

```go
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddlewareAllowsLocalhostAnyPort(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodOptions, "/jobs", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("allow origin = %q", response.Header().Get("Access-Control-Allow-Origin"))
	}
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestCORSMiddlewareAllows127AnyPort(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	request := httptest.NewRequest(http.MethodGet, "/jobs/job_1", nil)
	request.Header.Set("Origin", "http://127.0.0.1:3000")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Header().Get("Access-Control-Allow-Origin") != "http://127.0.0.1:3000" {
		t.Fatalf("allow origin = %q", response.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddlewareRejectsOtherOrigins(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	request := httptest.NewRequest(http.MethodGet, "/jobs/job_1", nil)
	request.Header.Set("Origin", "https://example.com")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("allow origin = %q", response.Header().Get("Access-Control-Allow-Origin"))
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
cd backend
mise exec -- go test ./internal/api -run TestCORSMiddleware
```

Expected:

```text
FAIL，因为 `withCORS` 还未定义。
```

- [ ] **Step 3: 实现路由和 CORS middleware**

Create `backend/internal/api/routes.go`:

```go
package api

import (
	"net/http"
	"net/url"
)

func Routes(handler *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/jobs", handler.handleJobs)
	mux.HandleFunc("/jobs/", handler.handleJobByID)
	mux.HandleFunc("/subtitle-assets", handler.handleSubtitleAssets)
	mux.HandleFunc("/subtitle-files/", handler.handleSubtitleFile)
	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isAllowedLocalOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isAllowedLocalOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" && (parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1") && parsed.Port() != ""
}
```

- [ ] **Step 4: 运行 CORS 测试**

Run:

```bash
cd backend
mise exec -- go test ./internal/api -run TestCORSMiddleware
```

Expected:

```text
ok  	lets-sub-it-api/internal/api
```

- [ ] **Step 5: 提交 routes**

Run:

```bash
git add backend/internal/api/routes.go backend/internal/api/routes_test.go
git commit -m "feat(backend): add API routes and CORS"
```

## Task 6: Implement HTTP Handlers

**Files:**
- Create: `backend/internal/api/handler.go`
- Test: `backend/internal/api/handler_test.go`

- [ ] **Step 1: 写 HTTP API 契约测试**

Create `backend/internal/api/handler_test.go`:

```go
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"lets-sub-it-api/internal/runner"
	"lets-sub-it-api/internal/store"
)

type syncRunner struct {
	runner *runner.MockRunner
}

func (r syncRunner) Start(ctx context.Context, job store.Job) error {
	return r.runner.Start(ctx, job)
}

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	testStore, err := store.Open(filepath.Join(t.TempDir(), "test.sqlite3"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := testStore.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	handler := NewHandler(testStore, syncRunner{runner: runner.NewMockRunner(testStore)}, t.TempDir())
	return Routes(handler)
}

func TestPostJobsCreatesJobAndCompletesWithMockRunner(t *testing.T) {
	server := newTestServer(t)
	body := bytes.NewBufferString(`{"youtubeUrl":"https://www.youtube.com/watch?v=abc123","sourceLanguage":"ja","targetLanguage":"zh-CN"}`)
	request := httptest.NewRequest(http.MethodPost, "/jobs", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d body = %s", response.Code, response.Body.String())
	}
	var payload struct {
		Job    jobResponse `json:"job"`
		Reused bool        `json:"reused"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if payload.Reused || payload.Job.VideoID != "abc123" || payload.Job.SourceLanguage != "ja" {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestPostJobsRejectsMissingSourceLanguage(t *testing.T) {
	server := newTestServer(t)
	body := bytes.NewBufferString(`{"youtubeUrl":"https://www.youtube.com/watch?v=abc123","targetLanguage":"zh-CN"}`)
	request := httptest.NewRequest(http.MethodPost, "/jobs", body)
	response := httptest.NewRecorder()

	server.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestSubtitleAssetReturnsAssetAfterCompletion(t *testing.T) {
	server := newTestServer(t)
	body := bytes.NewBufferString(`{"youtubeUrl":"https://youtu.be/abc123","sourceLanguage":"ja","targetLanguage":"zh-CN"}`)
	createRequest := httptest.NewRequest(http.MethodPost, "/jobs", body)
	createResponse := httptest.NewRecorder()
	server.ServeHTTP(createResponse, createRequest)
	var createPayload struct {
		Job jobResponse `json:"job"`
	}
	if err := json.Unmarshal(createResponse.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("Unmarshal create response error = %v", err)
	}

	assetResponse := waitForAsset(t, server, createPayload.Job.ID)
	if !bytes.Contains(assetResponse.Body.Bytes(), []byte("/subtitle-files/"+createPayload.Job.ID+"/translated")) {
		t.Fatalf("asset body = %s", assetResponse.Body.String())
	}
}

func waitForAsset(t *testing.T, server http.Handler, jobID string) *httptest.ResponseRecorder {
	t.Helper()
	var last *httptest.ResponseRecorder
	for range 20 {
		assetRequest := httptest.NewRequest(http.MethodGet, "/subtitle-assets?videoId=abc123&targetLanguage=zh-CN", nil)
		assetResponse := httptest.NewRecorder()
		server.ServeHTTP(assetResponse, assetRequest)
		last = assetResponse
		if assetResponse.Code == http.StatusOK && bytes.Contains(assetResponse.Body.Bytes(), []byte("/subtitle-files/"+jobID+"/translated")) {
			return assetResponse
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("asset never became available, last response = %s", last.Body.String())
	return nil
}
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
cd backend
mise exec -- go test ./internal/api -run 'TestPostJobs|TestSubtitleAsset'
```

Expected:

```text
FAIL，因为 `NewHandler` 和 handler 方法还未定义。
```

- [ ] **Step 3: 实现 handler**

Create `backend/internal/api/handler.go` with the responsibilities below:

```go
package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"lets-sub-it-api/internal/store"
)

type Store interface {
	CreateJob(job store.Job) error
	FindJob(id string) (store.Job, error)
	FindReusableJob(videoID string, targetLanguage string) (store.Job, error)
	FindSubtitleAsset(videoID string, targetLanguage string) (store.SubtitleAsset, error)
}

type AssetStore interface {
	Store
}

type Runner interface {
	Start(ctx context.Context, job store.Job) error
}

type Handler struct {
	store   Store
	runner  Runner
	workDir string
}

func NewHandler(store Store, runner Runner, workDir string) *Handler {
	return &Handler{store: store, runner: runner, workDir: workDir}
}

type createJobRequest struct {
	YoutubeURL     string `json:"youtubeUrl"`
	SourceLanguage string `json:"sourceLanguage"`
	TargetLanguage string `json:"targetLanguage"`
}

func (h *Handler) handleJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	var request createJobRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}
	if request.YoutubeURL == "" || request.SourceLanguage == "" || request.TargetLanguage == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "youtubeUrl, sourceLanguage, and targetLanguage are required")
		return
	}
	videoID, err := ParseVideoID(request.YoutubeURL)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_youtube_url", "unsupported YouTube URL")
		return
	}
	if reusable, err := h.store.FindReusableJob(videoID, request.TargetLanguage); err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"job": toJobResponse(reusable), "reused": true})
		return
	} else if !errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query reusable job")
		return
	}
	jobID := "job_" + strings.ReplaceAll(uuid.Must(uuid.NewV7()).String(), "-", "")
	job := store.NewJob(jobID, videoID, request.YoutubeURL, request.SourceLanguage, request.TargetLanguage, filepath.Join(h.workDir, jobID))
	if err := h.store.CreateJob(job); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create job")
		return
	}
	createdJob, err := h.store.FindJob(job.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query created job")
		return
	}
	go func() { _ = h.runner.Start(context.Background(), createdJob) }()
	writeJSON(w, http.StatusCreated, map[string]any{"job": toJobResponse(createdJob), "reused": false})
}

func (h *Handler) handleJobByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	jobID := strings.TrimPrefix(r.URL.Path, "/jobs/")
	job, err := h.store.FindJob(jobID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "job not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query job")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"job": toJobResponse(job)})
}

func (h *Handler) handleSubtitleAssets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	videoID := r.URL.Query().Get("videoId")
	targetLanguage := r.URL.Query().Get("targetLanguage")
	if videoID == "" || targetLanguage == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "videoId and targetLanguage are required")
		return
	}
	asset, err := h.store.FindSubtitleAsset(videoID, targetLanguage)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusOK, map[string]any{"asset": nil})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query subtitle asset")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"asset": toAssetResponse(asset)})
}

func (h *Handler) handleSubtitleFile(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "not_found", "subtitle file handler is added in Task 7")
}
```

Also extend `backend/internal/api/response.go` with asset response helpers:

```go
type assetResponse struct {
	JobID          string            `json:"jobId"`
	VideoID        string            `json:"videoId"`
	TargetLanguage string            `json:"targetLanguage"`
	SourceLanguage string            `json:"sourceLanguage"`
	Files          map[string]string `json:"files"`
	CreatedAt      string            `json:"createdAt"`
}

func toAssetResponse(asset store.SubtitleAsset) assetResponse {
	return assetResponse{
		JobID:          asset.JobID,
		VideoID:        asset.VideoID,
		TargetLanguage: asset.TargetLanguage,
		SourceLanguage: asset.SourceLanguage,
		Files: map[string]string{
			"source":     "/subtitle-files/" + asset.JobID + "/source",
			"translated": "/subtitle-files/" + asset.JobID + "/translated",
			"bilingual":  "/subtitle-files/" + asset.JobID + "/bilingual",
		},
		CreatedAt: formatTime(asset.CreatedAt),
	}
}
```

- [ ] **Step 4: 运行 handler 测试，记录文件服务测试暂时失败**

Run:

```bash
cd backend
mise exec -- go test ./internal/api -run 'TestPostJobs|TestSubtitleAsset'
```

Expected:

```text
TestPostJobsCreatesJobAndCompletesWithMockRunner passes.
TestPostJobsRejectsMissingSourceLanguage passes.
TestSubtitleAssetReturnsAssetAfterCompletion passes after mock runner creates the asset.
```

- [ ] **Step 5: 提交 handler 创建和资产查询**

Run:

```bash
git add backend/internal/api/handler.go backend/internal/api/handler_test.go backend/internal/api/response.go
git commit -m "feat(backend): add job API handlers"
```

## Task 7: Add Subtitle File Serving

**Files:**
- Modify: `backend/internal/store/sqlite.go`
- Modify: `backend/internal/api/handler.go`
- Test: `backend/internal/api/handler_test.go`

- [ ] **Step 1: 给 store 和 handler 增加文件服务测试**

Append to `backend/internal/store/sqlite_test.go`:

```go
func TestStoreFindsSubtitleAssetByJobID(t *testing.T) {
	store := openTestStore(t)
	job := NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh-CN", "/tmp/job_1")
	if err := store.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}
	asset := SubtitleAsset{JobID: "job_1", VideoID: "abc123", TargetLanguage: "zh-CN", SourceLanguage: "ja", SourceVTTPath: "source.vtt", TranslatedVTTPath: "translated.vtt", BilingualVTTPath: "bilingual.vtt"}
	if err := store.CreateSubtitleAsset(asset); err != nil {
		t.Fatalf("CreateSubtitleAsset() error = %v", err)
	}

	found, err := store.FindSubtitleAssetByJobID("job_1")
	if err != nil {
		t.Fatalf("FindSubtitleAssetByJobID() error = %v", err)
	}
	if found.JobID != "job_1" {
		t.Fatalf("asset = %+v", found)
	}
}
```

Append to `backend/internal/api/handler_test.go`:

```go
func TestSubtitleFileServing(t *testing.T) {
	server := newTestServer(t)
	body := bytes.NewBufferString(`{"youtubeUrl":"https://youtu.be/abc123","sourceLanguage":"ja","targetLanguage":"zh-CN"}`)
	createRequest := httptest.NewRequest(http.MethodPost, "/jobs", body)
	createResponse := httptest.NewRecorder()
	server.ServeHTTP(createResponse, createRequest)
	var createPayload struct {
		Job jobResponse `json:"job"`
	}
	if err := json.Unmarshal(createResponse.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("Unmarshal create response error = %v", err)
	}
	waitForAsset(t, server, createPayload.Job.ID)

	fileRequest := httptest.NewRequest(http.MethodGet, "/subtitle-files/"+createPayload.Job.ID+"/translated", nil)
	fileResponse := httptest.NewRecorder()
	server.ServeHTTP(fileResponse, fileRequest)
	if fileResponse.Code != http.StatusOK {
		t.Fatalf("file status = %d", fileResponse.Code)
	}
	if fileResponse.Header().Get("Content-Type") != "text/vtt; charset=utf-8" {
		t.Fatalf("content-type = %q", fileResponse.Header().Get("Content-Type"))
	}
	if !bytes.HasPrefix(fileResponse.Body.Bytes(), []byte("WEBVTT")) {
		t.Fatalf("file body = %s", fileResponse.Body.String())
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
cd backend
mise exec -- go test ./internal/store -run TestStoreFindsSubtitleAssetByJobID
mise exec -- go test ./internal/api -run TestSubtitleFileServing
```

Expected:

```text
FAIL，因为 `FindSubtitleAssetByJobID` 还未定义，并且字幕文件服务仍是 stub。
```

- [ ] **Step 3: 实现按 job 查询 asset 和文件服务**

Add to `backend/internal/store/sqlite.go`:

```go
func (s *Store) FindSubtitleAssetByJobID(jobID string) (SubtitleAsset, error) {
	var asset SubtitleAsset
	if err := s.db.First(&asset, "job_id = ?", jobID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SubtitleAsset{}, ErrNotFound
		}
		return SubtitleAsset{}, err
	}
	return asset, nil
}
```

Update `backend/internal/api/handler.go` store interface:

```go
type Store interface {
	CreateJob(job store.Job) error
	FindJob(id string) (store.Job, error)
	FindReusableJob(videoID string, targetLanguage string) (store.Job, error)
	FindSubtitleAsset(videoID string, targetLanguage string) (store.SubtitleAsset, error)
	FindSubtitleAssetByJobID(jobID string) (store.SubtitleAsset, error)
}
```

Replace `handleSubtitleFile` in `backend/internal/api/handler.go`:

```go
func (h *Handler) handleSubtitleFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/subtitle-files/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		writeError(w, http.StatusNotFound, "not_found", "subtitle file not found")
		return
	}
	jobID := parts[0]
	mode := parts[1]
	asset, err := h.store.FindSubtitleAssetByJobID(jobID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "subtitle file not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query subtitle file")
		return
	}
	path := ""
	switch mode {
	case "source":
		path = asset.SourceVTTPath
	case "translated":
		path = asset.TranslatedVTTPath
	case "bilingual":
		path = asset.BilingualVTTPath
	default:
		writeError(w, http.StatusBadRequest, "invalid_mode", "mode must be source, translated, or bilingual")
		return
	}
	w.Header().Set("Content-Type", "text/vtt; charset=utf-8")
	http.ServeFile(w, r, path)
}
```

- [ ] **Step 4: 运行 store 与 API 文件服务测试**

Run:

```bash
cd backend
mise exec -- go test ./internal/store -run TestStoreFindsSubtitleAssetByJobID
mise exec -- go test ./internal/api -run TestSubtitleFileServing
```

Expected:

```text
ok  	lets-sub-it-api/internal/store
ok  	lets-sub-it-api/internal/api
```

- [ ] **Step 5: 提交文件服务**

Run:

```bash
git add backend/internal/store/sqlite.go backend/internal/store/sqlite_test.go backend/internal/api/handler.go backend/internal/api/handler_test.go
git commit -m "feat(backend): serve subtitle files"
```

## Task 8: Wire App And Server

**Files:**
- Create: `backend/internal/app/app.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: 实现 app 组装**

Create `backend/internal/app/app.go`:

```go
package app

import (
	"net/http"
	"os"
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
	mockRunner := runner.NewMockRunner(database)
	handler := api.NewHandler(database, mockRunner, config.WorkDir)
	return api.Routes(handler), nil
}
```

Update `backend/cmd/server/main.go`:

```go
package main

import (
	"log"
	"net/http"

	"lets-sub-it-api/internal/app"
)

func main() {
	config := app.LoadConfig()
	handler, err := app.NewHTTPHandler(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("starting lets-sub-it-api on %s", config.Addr)
	if err := http.ListenAndServe(config.Addr, handler); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 2: 运行全量 backend 测试**

Run:

```bash
cd backend
mise exec -- go test ./...
```

Expected:

```text
ok  	lets-sub-it-api/internal/api
ok  	lets-sub-it-api/internal/app
ok  	lets-sub-it-api/internal/runner
ok  	lets-sub-it-api/internal/store
```

- [ ] **Step 3: 运行 server 编译检查**

Run:

```bash
cd backend
mise exec -- go build ./cmd/server
```

Expected:

```text
命令以退出码 0 结束，并且不产生需要提交的构建产物。
```

- [ ] **Step 4: 提交 app wiring**

Run:

```bash
git add backend/cmd/server/main.go backend/internal/app/app.go
git commit -m "feat(backend): wire mock API server"
```

## Task 9: Add Backend README And Final Verification

**Files:**
- Create: `backend/README.md`
- Modify: `README.md`

- [ ] **Step 1: 新增 backend README**

Create `backend/README.md`:

```markdown
# lets-sub-it-api

Lets Sub It MVP 的 Go mock backend。

## Setup

```bash
mise install
cd backend
mise exec -- go mod download
```

## Run

```bash
cd backend
LSI_ADDR=127.0.0.1:8080 mise exec -- go run ./cmd/server
```

## API quick check

```bash
curl -sS -X POST http://127.0.0.1:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{"youtubeUrl":"https://www.youtube.com/watch?v=abc123","sourceLanguage":"ja","targetLanguage":"zh-CN"}'
```

## Test

```bash
cd backend
mise exec -- go test ./...
```

## Mock boundary

当前 backend 已实现真实 HTTP API、SQLite 持久化、job 复用、mock 状态推进和 VTT 文件服务，但还不会调用 `yt-dlp`、`ffmpeg`、`whisper-cli` 或任何 LLM API。
```

- [ ] **Step 2: 更新根 README 路线图状态**

Modify `README.md` roadmap lines so backend mock API is explicit:

```markdown
- [ ] Go API server 与 embedded runner
- [ ] Mock backend API、SQLite、状态机与字幕文件服务
- [ ] `yt-dlp` / `ffmpeg` 下载和音频处理链路
```

If implementation is complete when editing README, mark the mock backend line as checked:

```markdown
- [x] Mock backend API、SQLite、状态机与字幕文件服务
```

- [ ] **Step 3: 运行最终验证**

Run:

```bash
cd backend
mise exec -- go test ./...
mise exec -- go build ./cmd/server
cd ..
git status --short
```

Expected:

```text
go test ./... exits 0.
go build ./cmd/server exits 0.
git status only shows intentional backend and README changes.
```

- [ ] **Step 4: 提交文档与最终验证状态**

Run:

```bash
git add backend/README.md README.md
git commit -m "docs(backend): document mock API server"
```

## 自检记录

- Spec 覆盖：本计划覆盖根工具链、Go module、`cmd/server/main.go`、GORM SQLite、UUIDv7、`sourceLanguage`、job 复用、mock runner、VTT 文件服务、CORS、前端 API 契约和测试。
- Mock 边界：Task 4 明确只生成 mock VTT，不调用 `yt-dlp`、`ffmpeg`、`whisper-cli` 或 LLM。
- API 契约：Task 6 和 Task 7 覆盖 `POST /jobs`、`GET /jobs/{id}`、`GET /subtitle-assets`、`GET /subtitle-files/{jobId}/{mode}`。
- 红旗词扫描：未发现未决占位内容。
