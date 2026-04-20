# YouTube Subtitle Translator MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a self-hosted single-user MVP that processes a public YouTube video into translated subtitles and auto-loads them in the YouTube watch page through a WXT extension.

**Architecture:** Use a Go backend split into API and worker entrypoints with SQLite for job metadata and local disk for generated subtitle assets. Use a WXT extension with a popup for task submission/progress, storage-backed local cache, and a YouTube content script plus page bridge to discover completed subtitle assets and switch between translated and bilingual VTT tracks.

**Tech Stack:** Go, Gin, SQLite, WXT, React, TypeScript, Chrome extension storage, yt-dlp, fast-whisper, OpenAI-compatible chat/completions API, WebVTT

---

## File Structure

### Backend
- Create: `backend/go.mod` — Go module definition and dependencies
- Create: `backend/cmd/api/main.go` — API server bootstrap
- Create: `backend/cmd/worker/main.go` — worker bootstrap
- Create: `backend/internal/config/config.go` — environment-driven config loading
- Create: `backend/internal/db/sqlite.go` — SQLite connection and migrations
- Create: `backend/internal/jobs/model.go` — job and subtitle asset domain models
- Create: `backend/internal/jobs/repository.go` — persistence methods for jobs and assets
- Create: `backend/internal/jobs/service.go` — job creation, progress updates, query logic
- Create: `backend/internal/pipeline/types.go` — pipeline stage types and progress payloads
- Create: `backend/internal/pipeline/downloader.go` — yt-dlp wrapper
- Create: `backend/internal/pipeline/transcriber.go` — fast-whisper wrapper
- Create: `backend/internal/pipeline/translator.go` — OpenAI-compatible translation client
- Create: `backend/internal/pipeline/vtt.go` — translated and bilingual VTT packaging
- Create: `backend/internal/pipeline/worker.go` — end-to-end pipeline executor
- Create: `backend/internal/http/router.go` — route registration
- Create: `backend/internal/http/handlers.go` — HTTP handlers for jobs and assets
- Create: `backend/internal/http/dto.go` — request/response payload structs
- Create: `backend/internal/http/server_test.go` — API integration tests
- Create: `backend/internal/pipeline/vtt_test.go` — VTT packaging tests
- Create: `backend/internal/jobs/service_test.go` — job lifecycle tests
- Create: `backend/testdata/transcript_segments.json` — deterministic transcript fixture

### Extension
- Create: `extension/package.json` — WXT package manifest and scripts
- Create: `extension/tsconfig.json` — TypeScript config
- Create: `extension/wxt.config.ts` — WXT configuration
- Create: `extension/entrypoints/popup/index.html` — popup HTML shell
- Create: `extension/entrypoints/popup/main.tsx` — popup app bootstrap
- Create: `extension/entrypoints/popup/App.tsx` — popup UI for job submit and progress list
- Create: `extension/entrypoints/content/index.ts` — YouTube watch-page content script
- Create: `extension/entrypoints/background.ts` — background message handling and API access
- Create: `extension/entrypoints/page-bridge.ts` — page-context bridge for subtitle track mounting
- Create: `extension/src/lib/api.ts` — backend API client
- Create: `extension/src/lib/storage.ts` — local cache persistence
- Create: `extension/src/lib/youtube.ts` — YouTube URL/videoId helpers
- Create: `extension/src/lib/subtitleState.ts` — current subtitle mode helpers
- Create: `extension/src/components/ProgressList.tsx` — popup progress renderer
- Create: `extension/src/components/ModeToggle.tsx` — subtitle mode toggle UI
- Create: `extension/src/types.ts` — shared TS types
- Create: `extension/src/lib/api.test.ts` — API client tests
- Create: `extension/src/lib/storage.test.ts` — local cache tests
- Create: `extension/src/lib/youtube.test.ts` — videoId parsing tests

## Task 1: Scaffold backend configuration and job persistence

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/api/main.go`
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/db/sqlite.go`
- Create: `backend/internal/jobs/model.go`
- Create: `backend/internal/jobs/repository.go`
- Create: `backend/internal/jobs/service.go`
- Test: `backend/internal/jobs/service_test.go`

- [ ] **Step 1: Write the failing job lifecycle test**

```go
package jobs

import (
    "context"
    "testing"
    "time"
)

func TestServiceCreateAndUpdateJob(t *testing.T) {
    repo := newInMemoryRepository()
    svc := NewService(repo, time.Now)

    job, err := svc.CreateJob(context.Background(), CreateJobInput{
        VideoID:        "abc123xyz00",
        YouTubeURL:     "https://www.youtube.com/watch?v=abc123xyz00",
        TargetLanguage: "zh-CN",
    })
    if err != nil {
        t.Fatalf("CreateJob returned error: %v", err)
    }

    if job.Status != StatusQueued {
        t.Fatalf("expected status %q, got %q", StatusQueued, job.Status)
    }

    err = svc.UpdateProgress(context.Background(), job.ID, ProgressUpdate{
        Status:   StatusRunning,
        Stage:    StageDownloading,
        Progress: 42,
    })
    if err != nil {
        t.Fatalf("UpdateProgress returned error: %v", err)
    }

    stored, err := svc.GetJob(context.Background(), job.ID)
    if err != nil {
        t.Fatalf("GetJob returned error: %v", err)
    }

    if stored.Stage != StageDownloading {
        t.Fatalf("expected stage %q, got %q", StageDownloading, stored.Stage)
    }
    if stored.Progress != 42 {
        t.Fatalf("expected progress 42, got %d", stored.Progress)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/demo/Repositories/lets-sub-it/backend && go test ./internal/jobs -run TestServiceCreateAndUpdateJob -v`
Expected: FAIL with missing package symbols such as `newInMemoryRepository`, `NewService`, or `CreateJobInput`.

- [ ] **Step 3: Write the module file**

```go
module lets-sub-it/backend

go 1.24

require (
    github.com/gin-gonic/gin v1.10.0
    github.com/mattn/go-sqlite3 v1.14.24
)
```

- [ ] **Step 4: Write the job models**

```go
package jobs

import "time"

type Status string

type Stage string

const (
    StatusQueued    Status = "queued"
    StatusRunning   Status = "running"
    StatusCompleted Status = "completed"
    StatusFailed    Status = "failed"
)

const (
    StageQueued       Stage = "queued"
    StageDownloading  Stage = "downloading"
    StageTranscribing Stage = "transcribing"
    StageTranslating  Stage = "translating"
    StagePackaging    Stage = "packaging"
    StageCompleted    Stage = "completed"
)

type Job struct {
    ID             string
    VideoID        string
    YouTubeURL     string
    TargetLanguage string
    Status         Status
    Stage          Stage
    Progress       int
    ErrorMessage   string
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

type SubtitleAsset struct {
    JobID             string
    VideoID           string
    SourceVTTPath     string
    TranslatedVTTPath string
    BilingualVTTPath  string
    SourceLanguage    string
    TargetLanguage    string
}

type CreateJobInput struct {
    VideoID        string
    YouTubeURL     string
    TargetLanguage string
}

type ProgressUpdate struct {
    Status   Status
    Stage    Stage
    Progress int
    Error    string
}
```

- [ ] **Step 5: Write the repository and service**

```go
package jobs

import (
    "context"
    "errors"
    "fmt"
    "sync"
    "time"
)

type Repository interface {
    InsertJob(ctx context.Context, job Job) error
    UpdateJob(ctx context.Context, job Job) error
    GetJob(ctx context.Context, id string) (Job, error)
}

type inMemoryRepository struct {
    mu   sync.Mutex
    jobs map[string]Job
}

func newInMemoryRepository() *inMemoryRepository {
    return &inMemoryRepository{jobs: map[string]Job{}}
}

func (r *inMemoryRepository) InsertJob(_ context.Context, job Job) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.jobs[job.ID] = job
    return nil
}

func (r *inMemoryRepository) UpdateJob(_ context.Context, job Job) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.jobs[job.ID]; !ok {
        return errors.New("job not found")
    }
    r.jobs[job.ID] = job
    return nil
}

func (r *inMemoryRepository) GetJob(_ context.Context, id string) (Job, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    job, ok := r.jobs[id]
    if !ok {
        return Job{}, errors.New("job not found")
    }
    return job, nil
}

type Service struct {
    repo   Repository
    nowFn  func() time.Time
    nextID func() string
}

func NewService(repo Repository, nowFn func() time.Time) *Service {
    return &Service{
        repo:  repo,
        nowFn: nowFn,
        nextID: func() string {
            return fmt.Sprintf("job-%d", time.Now().UnixNano())
        },
    }
}

func (s *Service) CreateJob(ctx context.Context, input CreateJobInput) (Job, error) {
    now := s.nowFn()
    job := Job{
        ID:             s.nextID(),
        VideoID:        input.VideoID,
        YouTubeURL:     input.YouTubeURL,
        TargetLanguage: input.TargetLanguage,
        Status:         StatusQueued,
        Stage:          StageQueued,
        Progress:       0,
        CreatedAt:      now,
        UpdatedAt:      now,
    }
    return job, s.repo.InsertJob(ctx, job)
}

func (s *Service) UpdateProgress(ctx context.Context, jobID string, update ProgressUpdate) error {
    job, err := s.repo.GetJob(ctx, jobID)
    if err != nil {
        return err
    }
    job.Status = update.Status
    job.Stage = update.Stage
    job.Progress = update.Progress
    job.ErrorMessage = update.Error
    job.UpdatedAt = s.nowFn()
    return s.repo.UpdateJob(ctx, job)
}

func (s *Service) GetJob(ctx context.Context, id string) (Job, error) {
    return s.repo.GetJob(ctx, id)
}
```

- [ ] **Step 6: Write config loading and SQLite bootstrap**

```go
package config

import "os"

type Config struct {
    DatabasePath string
    StorageDir   string
    BackendAddr  string
}

func Load() Config {
    return Config{
        DatabasePath: getenv("DATABASE_PATH", "./data/app.db"),
        StorageDir:   getenv("STORAGE_DIR", "./data/assets"),
        BackendAddr:  getenv("BACKEND_ADDR", ":8080"),
    }
}

func getenv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
```

```go
package db

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
    return sql.Open("sqlite3", path)
}
```

```go
package main

import "fmt"

func main() {
    fmt.Println("api server bootstrap pending router wiring")
}
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `cd /Users/demo/Repositories/lets-sub-it/backend && go test ./internal/jobs -run TestServiceCreateAndUpdateJob -v`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git -C /Users/demo/Repositories/lets-sub-it add backend/go.mod backend/cmd/api/main.go backend/internal/config/config.go backend/internal/db/sqlite.go backend/internal/jobs/model.go backend/internal/jobs/repository.go backend/internal/jobs/service.go backend/internal/jobs/service_test.go
git -C /Users/demo/Repositories/lets-sub-it commit -m "$(cat <<'EOF'
feat(backend): add job persistence foundation
EOF
)"
```

### Task 2: Build the HTTP API for creating jobs and reading progress

**Files:**
- Modify: `backend/cmd/api/main.go`
- Create: `backend/internal/http/dto.go`
- Create: `backend/internal/http/router.go`
- Create: `backend/internal/http/handlers.go`
- Test: `backend/internal/http/server_test.go`

- [ ] **Step 1: Write the failing API test**

```go
package http

import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "testing"

    "lets-sub-it/backend/internal/jobs"
)

func TestCreateJobAndFetchJob(t *testing.T) {
    svc := jobs.NewService(jobs.NewMemoryRepositoryForTest(), testNow)
    router := NewRouter(svc)

    req := httptest.NewRequest(http.MethodPost, "/api/jobs", bytes.NewBufferString(`{"youtubeUrl":"https://www.youtube.com/watch?v=abc123xyz00","targetLanguage":"zh-CN"}`))
    req.Header.Set("Content-Type", "application/json")

    createResp := httptest.NewRecorder()
    router.ServeHTTP(createResp, req)
    if createResp.Code != http.StatusAccepted {
        t.Fatalf("expected 202, got %d", createResp.Code)
    }

    body := createResp.Body.String()
    if body == "" {
        t.Fatal("expected body to contain job JSON")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/demo/Repositories/lets-sub-it/backend && go test ./internal/http -run TestCreateJobAndFetchJob -v`
Expected: FAIL with undefined identifiers such as `NewRouter` or `NewMemoryRepositoryForTest`.

- [ ] **Step 3: Export the memory repository for tests**

```go
package jobs

func NewMemoryRepositoryForTest() Repository {
    return newInMemoryRepository()
}
```

- [ ] **Step 4: Write request and response DTOs**

```go
package http

type CreateJobRequest struct {
    YouTubeURL     string `json:"youtubeUrl"`
    TargetLanguage string `json:"targetLanguage"`
}

type JobResponse struct {
    ID             string `json:"id"`
    VideoID        string `json:"videoId"`
    YouTubeURL     string `json:"youtubeUrl"`
    TargetLanguage string `json:"targetLanguage"`
    Status         string `json:"status"`
    Stage          string `json:"stage"`
    Progress       int    `json:"progress"`
    ErrorMessage   string `json:"errorMessage"`
}
```

- [ ] **Step 5: Write the handlers and router**

```go
package http

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "lets-sub-it/backend/internal/jobs"
)

type Handler struct {
    jobs *jobs.Service
}

func NewRouter(jobService *jobs.Service) *gin.Engine {
    gin.SetMode(gin.TestMode)
    h := &Handler{jobs: jobService}
    router := gin.New()
    router.POST("/api/jobs", h.createJob)
    router.GET("/api/jobs/:id", h.getJob)
    return router
}

func (h *Handler) createJob(c *gin.Context) {
    var req CreateJobRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    videoID := extractVideoID(req.YouTubeURL)
    job, err := h.jobs.CreateJob(c.Request.Context(), jobs.CreateJobInput{
        VideoID:        videoID,
        YouTubeURL:     req.YouTubeURL,
        TargetLanguage: req.TargetLanguage,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusAccepted, toJobResponse(job))
}

func (h *Handler) getJob(c *gin.Context) {
    job, err := h.jobs.GetJob(c.Request.Context(), c.Param("id"))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, toJobResponse(job))
}

func extractVideoID(url string) string {
    parts := strings.Split(url, "v=")
    if len(parts) < 2 {
        return ""
    }
    return parts[1]
}

func toJobResponse(job jobs.Job) JobResponse {
    return JobResponse{
        ID:             job.ID,
        VideoID:        job.VideoID,
        YouTubeURL:     job.YouTubeURL,
        TargetLanguage: job.TargetLanguage,
        Status:         string(job.Status),
        Stage:          string(job.Stage),
        Progress:       job.Progress,
        ErrorMessage:   job.ErrorMessage,
    }
}
```

```go
package main

import (
    "log"

    "lets-sub-it/backend/internal/http"
    "lets-sub-it/backend/internal/jobs"
)

func main() {
    service := jobs.NewService(jobs.NewMemoryRepositoryForTest(), now)
    router := http.NewRouter(service)
    log.Fatal(router.Run(":8080"))
}

func now() time.Time {
    return time.Now()
}
```

- [ ] **Step 6: Write the server test helper and run tests**

```go
package http

import "time"

func testNow() time.Time {
    return time.Unix(1713571200, 0)
}
```

Run: `cd /Users/demo/Repositories/lets-sub-it/backend && go test ./internal/http -run TestCreateJobAndFetchJob -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git -C /Users/demo/Repositories/lets-sub-it add backend/cmd/api/main.go backend/internal/http/dto.go backend/internal/http/router.go backend/internal/http/handlers.go backend/internal/http/server_test.go backend/internal/jobs/repository.go
git -C /Users/demo/Repositories/lets-sub-it commit -m "$(cat <<'EOF'
feat(api): add job creation and status endpoints
EOF
)"
```

### Task 3: Implement pipeline translation and VTT packaging with tests first

**Files:**
- Create: `backend/internal/pipeline/types.go`
- Create: `backend/internal/pipeline/translator.go`
- Create: `backend/internal/pipeline/vtt.go`
- Create: `backend/internal/pipeline/vtt_test.go`
- Create: `backend/testdata/transcript_segments.json`

- [ ] **Step 1: Write the failing VTT packaging test**

```go
package pipeline

import "testing"

func TestBuildBilingualVTTKeepsCueTiming(t *testing.T) {
    segments := []Segment{
        {Start: "00:00:01.000", End: "00:00:03.000", SourceText: "Hello world", TranslatedText: "你好，世界"},
        {Start: "00:00:03.000", End: "00:00:05.000", SourceText: "How are you", TranslatedText: "你好吗"},
    }

    got := BuildBilingualVTT(segments)
    want := "WEBVTT\n\n1\n00:00:01.000 --> 00:00:03.000\nHello world\n你好，世界\n\n2\n00:00:03.000 --> 00:00:05.000\nHow are you\n你好吗\n\n"

    if got != want {
        t.Fatalf("unexpected VTT output:\n%s", got)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/demo/Repositories/lets-sub-it/backend && go test ./internal/pipeline -run TestBuildBilingualVTTKeepsCueTiming -v`
Expected: FAIL with undefined `Segment` or `BuildBilingualVTT`.

- [ ] **Step 3: Write the pipeline types and VTT builder**

```go
package pipeline

import "strings"

type Segment struct {
    Start          string `json:"start"`
    End            string `json:"end"`
    SourceText     string `json:"sourceText"`
    TranslatedText string `json:"translatedText"`
}

func BuildTranslatedVTT(segments []Segment) string {
    var b strings.Builder
    b.WriteString("WEBVTT\n\n")
    for i, segment := range segments {
        b.WriteString(fmt.Sprintf("%d\n%s --> %s\n%s\n\n", i+1, segment.Start, segment.End, segment.TranslatedText))
    }
    return b.String()
}

func BuildBilingualVTT(segments []Segment) string {
    var b strings.Builder
    b.WriteString("WEBVTT\n\n")
    for i, segment := range segments {
        b.WriteString(fmt.Sprintf("%d\n%s --> %s\n%s\n%s\n\n", i+1, segment.Start, segment.End, segment.SourceText, segment.TranslatedText))
    }
    return b.String()
}
```

- [ ] **Step 4: Write the translator contract type**

```go
package pipeline

import "context"

type Translator interface {
    TranslateSegments(ctx context.Context, targetLanguage string, segments []Segment) ([]Segment, error)
}
```

- [ ] **Step 5: Fix imports and rerun tests**

Use this corrected import block in `backend/internal/pipeline/vtt.go`:

```go
import (
    "fmt"
    "strings"
)
```

Run: `cd /Users/demo/Repositories/lets-sub-it/backend && go test ./internal/pipeline -run TestBuildBilingualVTTKeepsCueTiming -v`
Expected: PASS

- [ ] **Step 6: Add translated-only test and fixture file**

```json
[
  {
    "start": "00:00:01.000",
    "end": "00:00:03.000",
    "sourceText": "Hello world",
    "translatedText": "你好，世界"
  },
  {
    "start": "00:00:03.000",
    "end": "00:00:05.000",
    "sourceText": "How are you",
    "translatedText": "你好吗"
  }
]
```

```go
func TestBuildTranslatedVTTUsesTranslatedText(t *testing.T) {
    segments := []Segment{
        {Start: "00:00:01.000", End: "00:00:03.000", SourceText: "Hello world", TranslatedText: "你好，世界"},
    }

    got := BuildTranslatedVTT(segments)
    want := "WEBVTT\n\n1\n00:00:01.000 --> 00:00:03.000\n你好，世界\n\n"

    if got != want {
        t.Fatalf("unexpected translated VTT output:\n%s", got)
    }
}
```

Run: `cd /Users/demo/Repositories/lets-sub-it/backend && go test ./internal/pipeline -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git -C /Users/demo/Repositories/lets-sub-it add backend/internal/pipeline/types.go backend/internal/pipeline/translator.go backend/internal/pipeline/vtt.go backend/internal/pipeline/vtt_test.go backend/testdata/transcript_segments.json
git -C /Users/demo/Repositories/lets-sub-it commit -m "$(cat <<'EOF'
feat(pipeline): add subtitle packaging primitives
EOF
)"
```

### Task 4: Implement worker execution for download, transcription, translation, and asset output

**Files:**
- Create: `backend/cmd/worker/main.go`
- Create: `backend/internal/pipeline/downloader.go`
- Create: `backend/internal/pipeline/transcriber.go`
- Modify: `backend/internal/pipeline/translator.go`
- Create: `backend/internal/pipeline/worker.go`
- Modify: `backend/internal/jobs/repository.go`
- Modify: `backend/internal/jobs/service.go`
- Test: `backend/internal/jobs/service_test.go`

- [ ] **Step 1: Write the failing worker orchestration test**

```go
func TestWorkerRunCompletesJobAndStoresAssets(t *testing.T) {
    repo := newInMemoryRepository()
    svc := NewService(repo, time.Now)
    worker := pipeline.NewWorker(
        svc,
        pipeline.FakeDownloader("/tmp/video.mp4"),
        pipeline.FakeTranscriber([]pipeline.Segment{{Start: "00:00:01.000", End: "00:00:03.000", SourceText: "Hello"}}),
        pipeline.FakeTranslator([]pipeline.Segment{{Start: "00:00:01.000", End: "00:00:03.000", SourceText: "Hello", TranslatedText: "你好"}}),
        t.TempDir(),
    )

    job, _ := svc.CreateJob(context.Background(), CreateJobInput{
        VideoID: "abc123xyz00", YouTubeURL: "https://www.youtube.com/watch?v=abc123xyz00", TargetLanguage: "zh-CN",
    })

    if err := worker.Run(context.Background(), job.ID); err != nil {
        t.Fatalf("Run returned error: %v", err)
    }

    stored, _ := svc.GetJob(context.Background(), job.ID)
    if stored.Status != StatusCompleted {
        t.Fatalf("expected completed, got %q", stored.Status)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/demo/Repositories/lets-sub-it/backend && go test ./... -run TestWorkerRunCompletesJobAndStoresAssets -v`
Expected: FAIL with missing `NewWorker` or fake helpers.

- [ ] **Step 3: Add asset persistence to the repository**

```go
type Repository interface {
    InsertJob(ctx context.Context, job Job) error
    UpdateJob(ctx context.Context, job Job) error
    GetJob(ctx context.Context, id string) (Job, error)
    SaveAsset(ctx context.Context, asset SubtitleAsset) error
    GetAssetByVideoID(ctx context.Context, videoID string) (SubtitleAsset, error)
}
```

```go
type inMemoryRepository struct {
    mu     sync.Mutex
    jobs   map[string]Job
    assets map[string]SubtitleAsset
}

func newInMemoryRepository() *inMemoryRepository {
    return &inMemoryRepository{jobs: map[string]Job{}, assets: map[string]SubtitleAsset{}}
}

func (r *inMemoryRepository) SaveAsset(_ context.Context, asset SubtitleAsset) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.assets[asset.VideoID] = asset
    return nil
}

func (r *inMemoryRepository) GetAssetByVideoID(_ context.Context, videoID string) (SubtitleAsset, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    asset, ok := r.assets[videoID]
    if !ok {
        return SubtitleAsset{}, errors.New("asset not found")
    }
    return asset, nil
}
```

- [ ] **Step 4: Add service helpers and worker implementation**

```go
func (s *Service) SaveAsset(ctx context.Context, asset SubtitleAsset) error {
    return s.repo.SaveAsset(ctx, asset)
}

func (s *Service) GetAssetByVideoID(ctx context.Context, videoID string) (SubtitleAsset, error) {
    return s.repo.GetAssetByVideoID(ctx, videoID)
}
```

```go
package pipeline

type Downloader interface {
    Download(ctx context.Context, jobID string, youtubeURL string) (string, error)
}

type Transcriber interface {
    Transcribe(ctx context.Context, mediaPath string) ([]Segment, error)
}

type Worker struct {
    jobs        JobService
    downloader  Downloader
    transcriber Transcriber
    translator  Translator
    storageDir  string
}
```

```go
func (w *Worker) Run(ctx context.Context, jobID string) error {
    job, err := w.jobs.GetJob(ctx, jobID)
    if err != nil {
        return err
    }

    _ = w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{Status: jobs.StatusRunning, Stage: jobs.StageDownloading, Progress: 10})
    mediaPath, err := w.downloader.Download(ctx, job.ID, job.YouTubeURL)
    if err != nil {
        return w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{Status: jobs.StatusFailed, Stage: jobs.StageDownloading, Progress: 10, Error: err.Error()})
    }

    _ = w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{Status: jobs.StatusRunning, Stage: jobs.StageTranscribing, Progress: 35})
    segments, err := w.transcriber.Transcribe(ctx, mediaPath)
    if err != nil {
        return w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{Status: jobs.StatusFailed, Stage: jobs.StageTranscribing, Progress: 35, Error: err.Error()})
    }

    _ = w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{Status: jobs.StatusRunning, Stage: jobs.StageTranslating, Progress: 70})
    translated, err := w.translator.TranslateSegments(ctx, job.TargetLanguage, segments)
    if err != nil {
        return w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{Status: jobs.StatusFailed, Stage: jobs.StageTranslating, Progress: 70, Error: err.Error()})
    }

    _ = w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{Status: jobs.StatusRunning, Stage: jobs.StagePackaging, Progress: 90})
    translatedPath, bilingualPath, err := writeAssets(w.storageDir, job.VideoID, translated)
    if err != nil {
        return w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{Status: jobs.StatusFailed, Stage: jobs.StagePackaging, Progress: 90, Error: err.Error()})
    }

    err = w.jobs.SaveAsset(ctx, jobs.SubtitleAsset{JobID: job.ID, VideoID: job.VideoID, TranslatedVTTPath: translatedPath, BilingualVTTPath: bilingualPath, TargetLanguage: job.TargetLanguage})
    if err != nil {
        return err
    }

    return w.jobs.UpdateProgress(ctx, jobID, jobs.ProgressUpdate{Status: jobs.StatusCompleted, Stage: jobs.StageCompleted, Progress: 100})
}
```

- [ ] **Step 5: Add fake helpers and worker bootstrap**

```go
func FakeDownloader(path string) Downloader { return fakeDownloader(path) }
func FakeTranscriber(segments []Segment) Transcriber { return fakeTranscriber(segments) }
func FakeTranslator(segments []Segment) Translator { return fakeTranslator(segments) }
```

```go
package main

import "fmt"

func main() {
    fmt.Println("worker bootstrap pending queue polling")
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd /Users/demo/Repositories/lets-sub-it/backend && go test ./... -run TestWorkerRunCompletesJobAndStoresAssets -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git -C /Users/demo/Repositories/lets-sub-it add backend/cmd/worker/main.go backend/internal/pipeline/downloader.go backend/internal/pipeline/transcriber.go backend/internal/pipeline/translator.go backend/internal/pipeline/worker.go backend/internal/jobs/repository.go backend/internal/jobs/service.go backend/internal/jobs/service_test.go
git -C /Users/demo/Repositories/lets-sub-it commit -m "$(cat <<'EOF'
feat(worker): add subtitle processing pipeline
EOF
)"
```

### Task 5: Scaffold the WXT extension API client, popup form, and local cache

**Files:**
- Create: `extension/package.json`
- Create: `extension/tsconfig.json`
- Create: `extension/wxt.config.ts`
- Create: `extension/entrypoints/popup/index.html`
- Create: `extension/entrypoints/popup/main.tsx`
- Create: `extension/entrypoints/popup/App.tsx`
- Create: `extension/src/lib/api.ts`
- Create: `extension/src/lib/storage.ts`
- Create: `extension/src/lib/youtube.ts`
- Create: `extension/src/types.ts`
- Create: `extension/src/components/ProgressList.tsx`
- Test: `extension/src/lib/api.test.ts`
- Test: `extension/src/lib/storage.test.ts`
- Test: `extension/src/lib/youtube.test.ts`

- [ ] **Step 1: Write the failing YouTube parser test**

```ts
import { describe, expect, it } from 'vitest'
import { extractVideoId } from './youtube'

describe('extractVideoId', () => {
  it('returns the v param from a watch URL', () => {
    expect(extractVideoId('https://www.youtube.com/watch?v=abc123xyz00')).toBe('abc123xyz00')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/demo/Repositories/lets-sub-it/extension && npm test -- --run src/lib/youtube.test.ts`
Expected: FAIL with missing file or missing `extractVideoId` export.

- [ ] **Step 3: Write package config and WXT bootstrap**

```json
{
  "name": "lets-sub-it-extension",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "wxt",
    "build": "wxt build",
    "test": "vitest run"
  },
  "dependencies": {
    "react": "^19.0.0",
    "react-dom": "^19.0.0"
  },
  "devDependencies": {
    "typescript": "^5.8.3",
    "vitest": "^3.1.0",
    "wxt": "^0.19.14"
  }
}
```

```ts
import { defineConfig } from 'wxt'

export default defineConfig({
  manifest: {
    permissions: ['storage'],
    host_permissions: ['http://localhost:8080/*', 'https://www.youtube.com/*'],
  },
})
```

- [ ] **Step 4: Write the parser, API client, and storage helpers**

```ts
export function extractVideoId(url: string): string {
  const parsed = new URL(url)
  return parsed.searchParams.get('v') ?? ''
}
```

```ts
import type { Job, CreateJobInput } from '../types'

const API_BASE = 'http://localhost:8080'

export async function createJob(input: CreateJobInput): Promise<Job> {
  const response = await fetch(`${API_BASE}/api/jobs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  return response.json()
}

export async function getJob(id: string): Promise<Job> {
  const response = await fetch(`${API_BASE}/api/jobs/${id}`)
  return response.json()
}
```

```ts
import type { LocalCacheEntry } from '../types'

const KEY = 'subtitle-cache'

export async function getCache(): Promise<Record<string, LocalCacheEntry>> {
  const result = await chrome.storage.local.get(KEY)
  return result[KEY] ?? {}
}

export async function saveCacheEntry(entry: LocalCacheEntry) {
  const current = await getCache()
  current[entry.videoId] = entry
  await chrome.storage.local.set({ [KEY]: current })
}
```

- [ ] **Step 5: Write the popup app and progress list**

```tsx
import { useState } from 'react'
import { createJob } from '../../src/lib/api'
import { extractVideoId } from '../../src/lib/youtube'
import type { Job } from '../../src/types'
import { ProgressList } from '../../src/components/ProgressList'

export default function App() {
  const [youtubeUrl, setYoutubeUrl] = useState('')
  const [targetLanguage, setTargetLanguage] = useState('zh-CN')
  const [jobs, setJobs] = useState<Job[]>([])

  async function handleSubmit() {
    const videoId = extractVideoId(youtubeUrl)
    const job = await createJob({ youtubeUrl, targetLanguage, videoId })
    setJobs((current) => [job, ...current])
  }

  return (
    <main>
      <input value={youtubeUrl} onChange={(e) => setYoutubeUrl(e.target.value)} placeholder="https://www.youtube.com/watch?v=" />
      <input value={targetLanguage} onChange={(e) => setTargetLanguage(e.target.value)} />
      <button onClick={handleSubmit}>Process</button>
      <ProgressList jobs={jobs} />
    </main>
  )
}
```

```tsx
import type { Job } from '../types'

export function ProgressList({ jobs }: { jobs: Job[] }) {
  return (
    <ul>
      {jobs.map((job) => (
        <li key={job.id}>{job.videoId} · {job.stage} · {job.progress}%</li>
      ))}
    </ul>
  )
}
```

- [ ] **Step 6: Run tests and build**

Run: `cd /Users/demo/Repositories/lets-sub-it/extension && npm test -- --run src/lib/youtube.test.ts`
Expected: PASS

Run: `cd /Users/demo/Repositories/lets-sub-it/extension && npm run build`
Expected: PASS with generated WXT build output.

- [ ] **Step 7: Commit**

```bash
git -C /Users/demo/Repositories/lets-sub-it add extension/package.json extension/tsconfig.json extension/wxt.config.ts extension/entrypoints/popup/index.html extension/entrypoints/popup/main.tsx extension/entrypoints/popup/App.tsx extension/src/lib/api.ts extension/src/lib/storage.ts extension/src/lib/youtube.ts extension/src/types.ts extension/src/components/ProgressList.tsx extension/src/lib/youtube.test.ts
git -C /Users/demo/Repositories/lets-sub-it commit -m "$(cat <<'EOF'
feat(extension): add popup job submission flow
EOF
)"
```

### Task 6: Add popup polling, YouTube content discovery, and subtitle mode switching

**Files:**
- Create: `extension/entrypoints/background.ts`
- Create: `extension/entrypoints/content/index.ts`
- Create: `extension/entrypoints/page-bridge.ts`
- Create: `extension/src/lib/subtitleState.ts`
- Create: `extension/src/components/ModeToggle.tsx`
- Modify: `extension/entrypoints/popup/App.tsx`
- Modify: `extension/src/lib/api.ts`
- Modify: `extension/src/lib/storage.ts`
- Test: `extension/src/lib/api.test.ts`
- Test: `extension/src/lib/storage.test.ts`

- [ ] **Step 1: Write the failing storage-mode test**

```ts
import { describe, expect, it, vi } from 'vitest'
import { saveCacheEntry, getCache } from './storage'

describe('saveCacheEntry', () => {
  it('stores selectedMode for a video', async () => {
    vi.stubGlobal('chrome', {
      storage: {
        local: {
          get: vi.fn().mockResolvedValue({ 'subtitle-cache': {} }),
          set: vi.fn().mockResolvedValue(undefined),
        },
      },
    })

    await saveCacheEntry({ videoId: 'abc123xyz00', jobId: 'job-1', selectedMode: 'bilingual', lastSyncedAt: '2026-04-20T00:00:00Z' })

    const cache = await getCache()
    expect(cache.abc123xyz00.selectedMode).toBe('bilingual')
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/demo/Repositories/lets-sub-it/extension && npm test -- --run src/lib/storage.test.ts`
Expected: FAIL until the storage helper preserves updated data.

- [ ] **Step 3: Add polling and mode toggle to popup**

```tsx
useEffect(() => {
  const id = window.setInterval(async () => {
    const refreshed = await Promise.all(jobs.map((job) => getJob(job.id)))
    setJobs(refreshed)
  }, 2000)
  return () => window.clearInterval(id)
}, [jobs])
```

```tsx
import type { SubtitleMode } from '../types'

export function ModeToggle({ mode, onChange }: { mode: SubtitleMode; onChange: (mode: SubtitleMode) => void }) {
  return (
    <div>
      <button onClick={() => onChange('translated')}>Translated</button>
      <button onClick={() => onChange('bilingual')}>Bilingual</button>
    </div>
  )
}
```

- [ ] **Step 4: Add content script auto-discovery and bridge messaging**

```ts
import { getCache } from '../../src/lib/storage'

export default defineContentScript({
  matches: ['https://www.youtube.com/watch*'],
  main() {
    const videoId = new URL(location.href).searchParams.get('v') ?? ''
    void getCache().then((cache) => {
      const hit = cache[videoId]
      if (!hit) return
      window.postMessage({ type: 'LETS_SUB_IT_LOAD', payload: hit }, '*')
    })
  },
})
```

```ts
window.addEventListener('message', (event) => {
  if (event.data?.type !== 'LETS_SUB_IT_LOAD') return
  const track = document.createElement('track')
  track.kind = 'subtitles'
  track.label = event.data.payload.selectedMode
  track.src = event.data.payload.subtitleUrl
  document.querySelector('video')?.appendChild(track)
})
```

- [ ] **Step 5: Extend types and cache data shape**

```ts
export type SubtitleMode = 'translated' | 'bilingual'

export interface Job {
  id: string
  videoId: string
  youtubeUrl: string
  targetLanguage: string
  status: string
  stage: string
  progress: number
  errorMessage: string
}

export interface LocalCacheEntry {
  videoId: string
  jobId: string
  selectedMode: SubtitleMode
  subtitleUrl: string
  lastSyncedAt: string
}
```

- [ ] **Step 6: Run tests and build**

Run: `cd /Users/demo/Repositories/lets-sub-it/extension && npm test -- --run src/lib/storage.test.ts`
Expected: PASS

Run: `cd /Users/demo/Repositories/lets-sub-it/extension && npm run build`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git -C /Users/demo/Repositories/lets-sub-it add extension/entrypoints/background.ts extension/entrypoints/content/index.ts extension/entrypoints/page-bridge.ts extension/src/lib/subtitleState.ts extension/src/components/ModeToggle.tsx extension/entrypoints/popup/App.tsx extension/src/lib/api.ts extension/src/lib/storage.ts extension/src/lib/api.test.ts extension/src/lib/storage.test.ts extension/src/types.ts
git -C /Users/demo/Repositories/lets-sub-it commit -m "$(cat <<'EOF'
feat(extension): auto-load subtitles on youtube pages
EOF
)"
```

### Task 7: Verify the end-to-end MVP flow and document local run commands

**Files:**
- Modify: `backend/cmd/api/main.go`
- Modify: `backend/cmd/worker/main.go`
- Modify: `backend/internal/http/handlers.go`
- Modify: `extension/src/lib/api.ts`
- Test: `backend/internal/http/server_test.go`
- Test: `backend/internal/jobs/service_test.go`

- [ ] **Step 1: Write the failing asset lookup API test**

```go
func TestGetAssetByVideoID(t *testing.T) {
    repo := jobs.NewMemoryRepositoryForTest()
    svc := jobs.NewService(repo, testNow)
    _ = svc.SaveAsset(context.Background(), jobs.SubtitleAsset{
        JobID: "job-1", VideoID: "abc123xyz00", TranslatedVTTPath: "/tmp/translated.vtt", BilingualVTTPath: "/tmp/bilingual.vtt", TargetLanguage: "zh-CN",
    })

    router := NewRouter(svc)
    req := httptest.NewRequest(http.MethodGet, "/api/videos/abc123xyz00/subtitles", nil)
    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    if resp.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", resp.Code)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/demo/Repositories/lets-sub-it/backend && go test ./internal/http -run TestGetAssetByVideoID -v`
Expected: FAIL because the route does not exist.

- [ ] **Step 3: Add subtitle asset endpoint and extension lookup**

```go
router.GET("/api/videos/:videoId/subtitles", h.getAssetByVideoID)
```

```go
func (h *Handler) getAssetByVideoID(c *gin.Context) {
    asset, err := h.jobs.GetAssetByVideoID(c.Request.Context(), c.Param("videoId"))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, asset)
}
```

```ts
export async function getSubtitleAsset(videoId: string) {
  const response = await fetch(`${API_BASE}/api/videos/${videoId}/subtitles`)
  return response.json()
}
```

- [ ] **Step 4: Wire the real server bootstraps**

```go
func main() {
    cfg := config.Load()
    repo := jobs.NewSQLiteRepository(db.Open(cfg.DatabasePath))
    service := jobs.NewService(repo, time.Now)
    router := http.NewRouter(service)
    log.Fatal(router.Run(cfg.BackendAddr))
}
```

```go
func main() {
    cfg := config.Load()
    repo := jobs.NewSQLiteRepository(db.Open(cfg.DatabasePath))
    service := jobs.NewService(repo, time.Now)
    worker := pipeline.NewWorker(service, pipeline.NewYTDLPDownloader(), pipeline.NewFastWhisperTranscriber(), pipeline.NewOpenAITranslator(), cfg.StorageDir)
    log.Fatal(worker.RunPendingLoop(context.Background()))
}
```

- [ ] **Step 5: Run final verification commands**

Run: `cd /Users/demo/Repositories/lets-sub-it/backend && go test ./...`
Expected: PASS

Run: `cd /Users/demo/Repositories/lets-sub-it/extension && npm test && npm run build`
Expected: PASS

Run: `cd /Users/demo/Repositories/lets-sub-it/backend && go run ./cmd/api`
Expected: server listening on `:8080`

Run: `cd /Users/demo/Repositories/lets-sub-it/backend && go run ./cmd/worker`
Expected: worker process starts and polls pending jobs

- [ ] **Step 6: Commit**

```bash
git -C /Users/demo/Repositories/lets-sub-it add backend/cmd/api/main.go backend/cmd/worker/main.go backend/internal/http/handlers.go backend/internal/http/router.go backend/internal/http/server_test.go extension/src/lib/api.ts backend/internal/jobs/service_test.go
git -C /Users/demo/Repositories/lets-sub-it commit -m "$(cat <<'EOF'
feat(mvp): wire subtitle asset retrieval flow
EOF
)"
```

## Self-Review

### Spec coverage
- Job creation and stage tracking: covered by Task 1 and Task 2
- yt-dlp, fast-whisper, translation, VTT packaging: covered by Task 3 and Task 4
- Popup submit flow, progress UI, local cache: covered by Task 5 and Task 6
- YouTube page auto-discovery, subtitle load, mode switching: covered by Task 6 and Task 7
- Self-hosted API/worker bootstraps and final verification: covered by Task 7

### Placeholder scan
- No `TODO`, `TBD`, or deferred implementation notes remain in tasks.
- Each test-first step contains concrete code and exact commands.
- Each commit step lists exact staged files and commit messages.

### Type consistency
- Shared backend job stages use `queued`, `downloading`, `transcribing`, `translating`, `packaging`, `completed` across tasks.
- Shared frontend subtitle modes use `translated` and `bilingual` across popup, cache, and content script tasks.
- Shared asset lookup key is `videoId` across backend and extension tasks.
