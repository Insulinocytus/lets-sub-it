# OpenAI-compatible 翻译链路实施计划

> **给 agentic workers：** REQUIRED SUB-SKILL: 使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans` 按任务逐步实现本计划。步骤使用 checkbox（`- [ ]`）语法跟踪。

**目标:** 在 `LSI_RUNNER_MODE=real` 下接入 Chat Completions 兼容翻译链路，用真实 `source.vtt` 生成真实 `translated.vtt` 和 `bilingual.vtt`。

**架构:** `mock` runner 保持不变。`real` runner 在 `whisper-cli` 生成 `source.vtt` 后解析 cue，调用可注入的 `Translator`；生产环境使用 Chat Completions translator，测试使用 fake translator。Chat translator 逐条翻译 cue，每次请求带目标 cue 前后各 10 条上下文，但只接受目标 cue 的单条 `translation`。

**技术栈:** Go 1.22、标准库 `net/http`、`httptest`、SQLite/GORM 现有 store、现有 backend 测试工具。

---

## 文件结构

- 修改 `backend/internal/app/config.go`：新增 LLM 配置字段和环境变量读取。
- 修改 `backend/internal/app/config_test.go`：覆盖新增默认值和环境变量读取。
- 新增 `backend/internal/runner/vtt_cue.go`：解析、渲染 WebVTT cue。
- 新增 `backend/internal/runner/vtt_cue_test.go`：覆盖 cue 解析和渲染。
- 新增 `backend/internal/runner/translator.go`：定义 `Translator` 接口、Chat Completions translator、请求和响应结构。
- 新增 `backend/internal/runner/translator_test.go`：覆盖 HTTP 请求、上下文窗口和错误处理。
- 修改 `backend/internal/runner/real_runner.go`：接入 translator，移除 real runner 内的 mock 翻译写入。
- 修改 `backend/internal/runner/real_runner_test.go`：使用 fake translator 验证真实翻译输出和失败阶段。
- 修改 `backend/internal/app/app.go`：生产环境构造 Chat translator 并传给 real runner。
- 修改 `README.md`：更新 real runner 说明、环境变量表和架构图。
- 修改 `backend/README.md`：更新运行示例和 runner boundary。

## Task 1: LLM 配置

**文件:**
- 修改: `backend/internal/app/config.go`
- 修改: `backend/internal/app/config_test.go`

- [ ] **步骤 1: 写失败测试**

在 `backend/internal/app/config_test.go` 末尾添加：

```go
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
```

- [ ] **步骤 2: 运行测试确认失败**

运行:

```bash
cd backend
mise exec -- go test ./internal/app
```

预期: FAIL，提示 `Config` 没有 `LLMBaseURL`、`LLMAPIKey`、`LLMModel` 或 `LLMTimeout` 字段。

- [ ] **步骤 3: 写最小实现**

更新 `backend/internal/app/config.go`：

```go
type Config struct {
	Addr            string
	DBPath          string
	WorkDir         string
	RunnerMode      string
	DownloadTimeout time.Duration
	WhisperModel    string
	LLMBaseURL      string
	LLMAPIKey       string
	LLMModel        string
	LLMTimeout      time.Duration
}
```

在 `LoadConfig()` 返回值中追加：

```go
LLMBaseURL: envOrDefault("LSI_LLM_BASE_URL", "https://api.openai.com"),
LLMAPIKey:  os.Getenv("LSI_LLM_API_KEY"),
LLMModel:   os.Getenv("LSI_LLM_MODEL"),
LLMTimeout: envDurationOrDefault("LSI_LLM_TIMEOUT", 2*time.Minute),
```

- [ ] **步骤 4: 运行测试确认通过**

运行:

```bash
cd backend
mise exec -- go test ./internal/app
```

预期: PASS。

- [ ] **步骤 5: 提交**

```bash
git add backend/internal/app/config.go backend/internal/app/config_test.go
git commit -m "feat(backend): add llm configuration"
```

## Task 2: VTT cue 解析和渲染

**文件:**
- 新增: `backend/internal/runner/vtt_cue.go`
- 新增: `backend/internal/runner/vtt_cue_test.go`

- [ ] **步骤 1: 写失败测试**

创建 `backend/internal/runner/vtt_cue_test.go`：

```go
package runner

import (
	"strings"
	"testing"
)

func TestParseWebVTTCuesParsesMultipleCuesAndMultilineText(t *testing.T) {
	input := "WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nline one\nline two\n\n00:00:01.000 --> 00:00:02.000\nsecond\n"

	cues, err := parseWebVTTCues(input)
	if err != nil {
		t.Fatalf("parseWebVTTCues() error = %v", err)
	}

	if len(cues) != 2 {
		t.Fatalf("len(cues) = %d, want 2", len(cues))
	}
	if cues[0].TimeLine != "00:00:00.000 --> 00:00:01.000" {
		t.Fatalf("TimeLine = %q", cues[0].TimeLine)
	}
	if got := cueText(cues[0]); got != "line one\nline two" {
		t.Fatalf("cueText = %q", got)
	}
}

func TestParseWebVTTCuesRejectsInvalidHeader(t *testing.T) {
	_, err := parseWebVTTCues("not vtt\n\n00:00:00.000 --> 00:00:01.000\ntext\n")
	if err == nil {
		t.Fatal("parseWebVTTCues() error = nil, want error")
	}
}

func TestRenderTranslatedVTTUsesSourceTimeline(t *testing.T) {
	cues := []Cue{
		{TimeLine: "00:00:00.000 --> 00:00:01.000", TextLines: []string{"hello"}},
		{TimeLine: "00:00:01.000 --> 00:00:02.000", TextLines: []string{"world"}},
	}

	got, err := renderTranslatedVTT(cues, []string{"你好", "世界"})
	if err != nil {
		t.Fatalf("renderTranslatedVTT() error = %v", err)
	}

	if !strings.Contains(got, "00:00:00.000 --> 00:00:01.000\n你好") {
		t.Fatalf("translated VTT = %q", got)
	}
	if strings.Contains(got, "hello") {
		t.Fatalf("translated VTT = %q, should not include source text", got)
	}
}

func TestRenderBilingualVTTCombinesSourceAndTranslation(t *testing.T) {
	cues := []Cue{
		{TimeLine: "00:00:00.000 --> 00:00:01.000", TextLines: []string{"hello"}},
	}

	got, err := renderBilingualVTT(cues, []string{"你好"})
	if err != nil {
		t.Fatalf("renderBilingualVTT() error = %v", err)
	}

	want := "00:00:00.000 --> 00:00:01.000\nhello\n你好"
	if !strings.Contains(got, want) {
		t.Fatalf("bilingual VTT = %q, want containing %q", got, want)
	}
}

func TestRenderTranslatedVTTRejectsMismatchedCounts(t *testing.T) {
	_, err := renderTranslatedVTT([]Cue{{TimeLine: "00:00:00.000 --> 00:00:01.000", TextLines: []string{"hello"}}}, nil)
	if err == nil {
		t.Fatal("renderTranslatedVTT() error = nil, want error")
	}
}
```

- [ ] **步骤 2: 运行测试确认失败**

运行:

```bash
cd backend
mise exec -- go test ./internal/runner
```

预期: FAIL，提示 `Cue`、`parseWebVTTCues`、`cueText`、`renderTranslatedVTT` 或 `renderBilingualVTT` 未定义。

- [ ] **步骤 3: 写最小实现**

创建 `backend/internal/runner/vtt_cue.go`：

```go
package runner

import (
	"fmt"
	"strings"
)

type Cue struct {
	TimeLine  string
	TextLines []string
}

func parseWebVTTCues(content string) ([]Cue, error) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return nil, fmt.Errorf("empty vtt")
	}
	blocks := strings.Split(normalized, "\n\n")
	if len(blocks) == 0 || strings.TrimSpace(blocks[0]) != "WEBVTT" {
		return nil, fmt.Errorf("vtt must start with WEBVTT")
	}

	var cues []Cue
	for _, block := range blocks[1:] {
		lines := nonEmptyLines(block)
		if len(lines) < 2 {
			return nil, fmt.Errorf("invalid cue block")
		}
		if !strings.Contains(lines[0], "-->") {
			return nil, fmt.Errorf("invalid cue timeline %q", lines[0])
		}
		cues = append(cues, Cue{TimeLine: lines[0], TextLines: lines[1:]})
	}
	if len(cues) == 0 {
		return nil, fmt.Errorf("vtt has no cues")
	}
	return cues, nil
}

func nonEmptyLines(block string) []string {
	rawLines := strings.Split(strings.TrimSpace(block), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func cueText(cue Cue) string {
	return strings.Join(cue.TextLines, "\n")
}

func renderTranslatedVTT(cues []Cue, translations []string) (string, error) {
	if len(cues) != len(translations) {
		return "", fmt.Errorf("translation count %d does not match cue count %d", len(translations), len(cues))
	}
	var b strings.Builder
	b.WriteString("WEBVTT\n\n")
	for i, cue := range cues {
		b.WriteString(cue.TimeLine)
		b.WriteString("\n")
		b.WriteString(strings.TrimSpace(translations[i]))
		b.WriteString("\n\n")
	}
	return b.String(), nil
}

func renderBilingualVTT(cues []Cue, translations []string) (string, error) {
	if len(cues) != len(translations) {
		return "", fmt.Errorf("translation count %d does not match cue count %d", len(translations), len(cues))
	}
	var b strings.Builder
	b.WriteString("WEBVTT\n\n")
	for i, cue := range cues {
		b.WriteString(cue.TimeLine)
		b.WriteString("\n")
		b.WriteString(cueText(cue))
		b.WriteString("\n")
		b.WriteString(strings.TrimSpace(translations[i]))
		b.WriteString("\n\n")
	}
	return b.String(), nil
}
```

- [ ] **步骤 4: 运行测试确认通过**

运行:

```bash
cd backend
mise exec -- go test ./internal/runner
```

预期: PASS。

- [ ] **步骤 5: 提交**

```bash
git add backend/internal/runner/vtt_cue.go backend/internal/runner/vtt_cue_test.go
git commit -m "feat(backend): add vtt cue rendering"
```

## Task 3: Chat Completions translator

**文件:**
- 新增: `backend/internal/runner/translator.go`
- 新增: `backend/internal/runner/translator_test.go`

- [ ] **步骤 1: 写失败测试**

创建 `backend/internal/runner/translator_test.go`：

```go
package runner

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestChatTranslatorSendsContextWindowAndReturnsTranslations(t *testing.T) {
	var requests []chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %q, want /v1/chat/completions", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q", got)
		}
		var request chatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode request error = %v", err)
		}
		requests = append(requests, request)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"translation\":\"译文\"}"}}]}`))
	}))
	defer server.Close()

	cues := make([]Cue, 25)
	for i := range cues {
		cues[i] = Cue{TimeLine: "00:00:00.000 --> 00:00:01.000", TextLines: []string{string(rune('a' + i%26))}}
	}

	translator := NewChatTranslator(server.URL, "test-key", "test-model", time.Second, server.Client())
	translations, err := translator.Translate(context.Background(), cues, "ja", "zh")
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	if len(translations) != len(cues) {
		t.Fatalf("len(translations) = %d, want %d", len(translations), len(cues))
	}
	if len(requests) != len(cues) {
		t.Fatalf("requests = %d, want %d", len(requests), len(cues))
	}
	if requests[12].Model != "test-model" {
		t.Fatalf("Model = %q", requests[12].Model)
	}
	userContent := requests[12].Messages[1].Content
	if !strings.Contains(userContent, `"index":2`) || !strings.Contains(userContent, `"index":22`) {
		t.Fatalf("user content = %s, want context window 2..22", userContent)
	}
	if !strings.Contains(userContent, `"isTarget":true`) || !strings.Contains(userContent, `"index":12`) {
		t.Fatalf("user content = %s, want target index 12", userContent)
	}
}

func TestChatTranslatorTruncatesContextAtEdges(t *testing.T) {
	var firstRequest chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if firstRequest.Model == "" {
			if err := json.NewDecoder(r.Body).Decode(&firstRequest); err != nil {
				t.Fatalf("Decode request error = %v", err)
			}
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"translation\":\"译文\"}"}}]}`))
	}))
	defer server.Close()

	cues := make([]Cue, 3)
	for i := range cues {
		cues[i] = Cue{TimeLine: "00:00:00.000 --> 00:00:01.000", TextLines: []string{"text"}}
	}

	translator := NewChatTranslator(server.URL, "", "test-model", time.Second, server.Client())
	if _, err := translator.Translate(context.Background(), cues, "ja", "zh"); err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	userContent := firstRequest.Messages[1].Content
	if strings.Contains(userContent, `"index":-1`) || !strings.Contains(userContent, `"index":0`) || !strings.Contains(userContent, `"index":2`) {
		t.Fatalf("user content = %s, want clipped context 0..2", userContent)
	}
}

func TestChatTranslatorRequiresModel(t *testing.T) {
	translator := NewChatTranslator("https://api.openai.com", "", "", time.Second, http.DefaultClient)
	_, err := translator.Translate(context.Background(), []Cue{{TimeLine: "00:00:00.000 --> 00:00:01.000", TextLines: []string{"text"}}}, "ja", "zh")
	if err == nil {
		t.Fatal("Translate() error = nil, want model error")
	}
}

func TestChatTranslatorFailsOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad key", http.StatusUnauthorized)
	}))
	defer server.Close()

	translator := NewChatTranslator(server.URL, "secret-key", "test-model", time.Second, server.Client())
	_, err := translator.Translate(context.Background(), []Cue{{TimeLine: "00:00:00.000 --> 00:00:01.000", TextLines: []string{"text"}}}, "ja", "zh")
	if err == nil || strings.Contains(err.Error(), "secret-key") {
		t.Fatalf("Translate() error = %v, want non-secret error", err)
	}
}

func TestChatTranslatorFailsWhenTranslationMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"notTranslation\":\"x\"}"}}]}`))
	}))
	defer server.Close()

	translator := NewChatTranslator(server.URL, "", "test-model", time.Second, server.Client())
	_, err := translator.Translate(context.Background(), []Cue{{TimeLine: "00:00:00.000 --> 00:00:01.000", TextLines: []string{"text"}}}, "ja", "zh")
	if err == nil {
		t.Fatal("Translate() error = nil, want missing translation error")
	}
}
```

- [ ] **步骤 2: 运行测试确认失败**

运行:

```bash
cd backend
mise exec -- go test ./internal/runner
```

预期: FAIL，提示 `chatCompletionRequest`、`NewChatTranslator` 或 `Translate` 未定义。

- [ ] **步骤 3: 写最小实现**

创建 `backend/internal/runner/translator.go`：

```go
package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const translationContextRadius = 10

type Translator interface {
	Translate(ctx context.Context, cues []Cue, sourceLanguage string, targetLanguage string) ([]string, error)
}

type ChatTranslator struct {
	baseURL string
	apiKey  string
	model   string
	timeout time.Duration
	client  *http.Client
}

func NewChatTranslator(baseURL string, apiKey string, model string, timeout time.Duration, client *http.Client) *ChatTranslator {
	if client == nil {
		client = http.DefaultClient
	}
	return &ChatTranslator{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		timeout: timeout,
		client:  client,
	}
}

func (t *ChatTranslator) Translate(ctx context.Context, cues []Cue, sourceLanguage string, targetLanguage string) ([]string, error) {
	if strings.TrimSpace(t.model) == "" {
		return nil, fmt.Errorf("LSI_LLM_MODEL is required")
	}
	translations := make([]string, 0, len(cues))
	for i := range cues {
		translation, err := t.translateOne(ctx, cues, i, sourceLanguage, targetLanguage)
		if err != nil {
			return nil, err
		}
		translations = append(translations, translation)
	}
	if len(translations) != len(cues) {
		return nil, fmt.Errorf("translation count %d does not match cue count %d", len(translations), len(cues))
	}
	return translations, nil
}

type chatCompletionRequest struct {
	Model          string               `json:"model"`
	Messages       []chatMessage        `json:"messages"`
	ResponseFormat chatResponseFormat   `json:"response_format"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponseFormat struct {
	Type string `json:"type"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type translationContent struct {
	Translation string `json:"translation"`
}

type translationPrompt struct {
	SourceLanguage string             `json:"sourceLanguage"`
	TargetLanguage string             `json:"targetLanguage"`
	Target         translationCueItem `json:"target"`
	Context        []translationCueItem `json:"context"`
}

type translationCueItem struct {
	Index    int    `json:"index"`
	Text     string `json:"text"`
	IsTarget bool   `json:"isTarget"`
}

func (t *ChatTranslator) translateOne(ctx context.Context, cues []Cue, index int, sourceLanguage string, targetLanguage string) (string, error) {
	promptBytes, err := json.Marshal(buildTranslationPrompt(cues, index, sourceLanguage, targetLanguage))
	if err != nil {
		return "", fmt.Errorf("build translation prompt: %w", err)
	}
	requestBody := chatCompletionRequest{
		Model: t.model,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: "Translate only the target subtitle cue. Use the surrounding context to resolve pronouns, tone, and terminology. Return JSON only, with a single string field named translation.",
			},
			{
				Role:    "user",
				Content: string(promptBytes),
			},
		},
		ResponseFormat: chatResponseFormat{Type: "json_object"},
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("encode chat completion request: %w", err)
	}

	requestCtx := ctx
	cancel := func() {}
	if t.timeout > 0 {
		requestCtx, cancel = context.WithTimeout(ctx, t.timeout)
	}
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, t.baseURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create chat completion request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if t.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+t.apiKey)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("chat completion request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return "", fmt.Errorf("read chat completion response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("chat completion returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var decoded chatCompletionResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return "", fmt.Errorf("decode chat completion response: %w", err)
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("chat completion response missing message content")
	}

	var content translationContent
	if err := json.Unmarshal([]byte(decoded.Choices[0].Message.Content), &content); err != nil {
		return "", fmt.Errorf("decode translation content: %w", err)
	}
	if strings.TrimSpace(content.Translation) == "" {
		return "", fmt.Errorf("translation content missing translation")
	}
	return content.Translation, nil
}

func buildTranslationPrompt(cues []Cue, targetIndex int, sourceLanguage string, targetLanguage string) translationPrompt {
	start := targetIndex - translationContextRadius
	if start < 0 {
		start = 0
	}
	end := targetIndex + translationContextRadius
	if end >= len(cues) {
		end = len(cues) - 1
	}
	contextItems := make([]translationCueItem, 0, end-start+1)
	for i := start; i <= end; i++ {
		contextItems = append(contextItems, translationCueItem{
			Index:    i,
			Text:     cueText(cues[i]),
			IsTarget: i == targetIndex,
		})
	}
	target := translationCueItem{
		Index:    targetIndex,
		Text:     cueText(cues[targetIndex]),
		IsTarget: true,
	}
	return translationPrompt{
		SourceLanguage: sourceLanguage,
		TargetLanguage: targetLanguage,
		Target:         target,
		Context:        contextItems,
	}
}
```

运行 `gofmt`：

```bash
cd backend
mise exec -- gofmt -w internal/runner/translator.go internal/runner/translator_test.go
```

- [ ] **步骤 4: 运行测试确认通过**

运行:

```bash
cd backend
mise exec -- go test ./internal/runner
```

预期: PASS。

- [ ] **步骤 5: 提交**

```bash
git add backend/internal/runner/translator.go backend/internal/runner/translator_test.go
git commit -m "feat(backend): add chat completions translator"
```

## Task 4: RealRunner 接入真实翻译

**文件:**
- 修改: `backend/internal/runner/real_runner.go`
- 修改: `backend/internal/runner/real_runner_test.go`
- 修改: `backend/internal/app/app.go`

- [ ] **步骤 1: 写失败测试**

修改 `backend/internal/runner/real_runner_test.go`：

1. 在 `TestRealRunnerCompletesJob` 中把构造改为：

```go
translator := fakeTranslator{translations: []string{"translated one"}}
if err := NewRealRunner(testStore, 10*time.Minute, "tiny", translator).Start(context.Background(), job); err != nil {
	t.Fatalf("Start() error = %v", err)
}
```

2. 在该测试读取 asset 后添加：

```go
translatedContent, readTranslatedErr := os.ReadFile(asset.TranslatedVTTPath)
if readTranslatedErr != nil {
	t.Fatalf("os.ReadFile(translated.vtt) error = %v", readTranslatedErr)
}
if !strings.Contains(string(translatedContent), "translated one") {
	t.Fatalf("translated.vtt content = %q, want translator output", string(translatedContent))
}
if strings.Contains(string(translatedContent), "mock 翻译") {
	t.Fatalf("translated.vtt content = %q, should not contain mock translation", string(translatedContent))
}

bilingualContent, readBilingualErr := os.ReadFile(asset.BilingualVTTPath)
if readBilingualErr != nil {
	t.Fatalf("os.ReadFile(bilingual.vtt) error = %v", readBilingualErr)
}
if !strings.Contains(string(bilingualContent), "real transcript\ntranslated one") {
	t.Fatalf("bilingual.vtt content = %q, want source and translation", string(bilingualContent))
}
```

3. 新增失败测试：

```go
func TestRealRunnerTranslationFailed(t *testing.T) {
	origExec := execCommand
	t.Cleanup(func() { execCommand = origExec })

	testStore := openTestStore(t)
	workDir := t.TempDir()
	jobDir := filepath.Join(workDir, "job_1")
	job := store.NewJob("job_1", "abc123", "https://www.youtube.com/watch?v=abc123", "ja", "zh", jobDir)

	if err := testStore.CreateJob(job); err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		switch name {
		case "yt-dlp":
			return exec.CommandContext(ctx, "sh", "-c", "mkdir -p \"$1\" && printf fake-audio-data > \"$1/audio.mp3\"", "sh", jobDir)
		case "whisper-cli":
			outputPath := argValue(t, args, "--output")
			return exec.CommandContext(ctx, "sh", "-c", "printf 'WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nreal transcript\n' > \"$1\"", "sh", outputPath)
		default:
			return exec.CommandContext(ctx, "sh", "-c", "echo unexpected command >&2; exit 127")
		}
	}

	err := NewRealRunner(testStore, 10*time.Minute, "small", fakeTranslator{err: errors.New("translation unavailable")}).Start(context.Background(), job)
	if err == nil {
		t.Fatal("Start() error = nil, want translation error")
	}

	updatedJob, findErr := testStore.FindJob("job_1")
	if findErr != nil {
		t.Fatalf("FindJob() error = %v", findErr)
	}
	if updatedJob.Status != store.StatusFailed {
		t.Fatalf("Status = %q, want %q", updatedJob.Status, store.StatusFailed)
	}
	if updatedJob.Stage != store.StatusTranslating {
		t.Fatalf("Stage = %q, want %q", updatedJob.Stage, store.StatusTranslating)
	}
	if updatedJob.ErrorMessage == nil || !strings.Contains(*updatedJob.ErrorMessage, "translation unavailable") {
		t.Fatalf("ErrorMessage = %v, want translation unavailable", updatedJob.ErrorMessage)
	}
}
```

4. 在文件末尾添加 fake translator：

```go
type fakeTranslator struct {
	translations []string
	err          error
}

func (t fakeTranslator) Translate(ctx context.Context, cues []Cue, sourceLanguage string, targetLanguage string) ([]string, error) {
	if t.err != nil {
		return nil, t.err
	}
	if len(t.translations) > 0 {
		return t.translations, nil
	}
	translations := make([]string, len(cues))
	for i := range cues {
		translations[i] = "translated"
	}
	return translations, nil
}
```

5. 为 `real_runner_test.go` import 添加 `errors`。
6. 把该文件内其他 `NewRealRunner(testStore, 10*time.Minute, "...")` 调用都改为传入 `fakeTranslator{}`，保持下载失败、取消和转写失败测试聚焦原本阶段。

- [ ] **步骤 2: 运行测试确认失败**

运行:

```bash
cd backend
mise exec -- go test ./internal/runner
```

预期: FAIL，提示 `NewRealRunner` 参数数量不匹配或 real runner 仍写 mock 翻译。

- [ ] **步骤 3: 写最小实现**

修改 `backend/internal/runner/real_runner.go`：

```go
type RealRunner struct {
	store           Store
	downloadTimeout time.Duration
	whisperModel    string
	translator      Translator
}

func NewRealRunner(store Store, downloadTimeout time.Duration, whisperModel string, translator Translator) *RealRunner {
	return &RealRunner{store: store, downloadTimeout: downloadTimeout, whisperModel: whisperModel, translator: translator}
}
```

把 mock 翻译写入段替换为：

```go
sourceContent, err := os.ReadFile(sourcePath)
if err != nil {
	return r.fail(job.ID, store.StatusTranslating, err)
}
cues, err := parseWebVTTCues(string(sourceContent))
if err != nil {
	return r.fail(job.ID, store.StatusTranslating, err)
}
translations, err := r.translator.Translate(ctx, cues, job.SourceLanguage, job.TargetLanguage)
if err != nil {
	return r.fail(job.ID, store.StatusTranslating, err)
}
translatedVTT, err := renderTranslatedVTT(cues, translations)
if err != nil {
	return r.fail(job.ID, store.StatusTranslating, err)
}
if err := os.WriteFile(translatedPath, []byte(translatedVTT), 0o644); err != nil {
	return r.fail(job.ID, store.StatusTranslating, err)
}
```

把 bilingual 写入改为：

```go
bilingualVTT, err := renderBilingualVTT(cues, translations)
if err != nil {
	return r.fail(job.ID, store.StatusPackaging, err)
}
if err := os.WriteFile(bilingualPath, []byte(bilingualVTT), 0o644); err != nil {
	return r.fail(job.ID, store.StatusPackaging, err)
}
```

修改 `backend/internal/app/app.go` 的 real runner 构造：

```go
translator := runner.NewChatTranslator(config.LLMBaseURL, config.LLMAPIKey, config.LLMModel, config.LLMTimeout, http.DefaultClient)
jobRunner = runner.NewRealRunner(database, config.DownloadTimeout, config.WhisperModel, translator)
```

运行 `gofmt`：

```bash
cd backend
mise exec -- gofmt -w internal/runner/real_runner.go internal/runner/real_runner_test.go internal/app/app.go
```

- [ ] **步骤 4: 运行测试确认通过**

运行:

```bash
cd backend
mise exec -- go test ./internal/runner ./internal/app
```

预期: PASS。

- [ ] **步骤 5: 提交**

```bash
git add backend/internal/runner/real_runner.go backend/internal/runner/real_runner_test.go backend/internal/app/app.go
git commit -m "feat(backend): translate real runner subtitles"
```

## Task 5: 文档、全量测试和构建验证

**文件:**
- 修改: `README.md`
- 修改: `backend/README.md`

- [ ] **步骤 1: 更新 README**

修改 `README.md`：

- 把 “翻译阶段和最终双语字幕仍未接入真实 LLM” 改为说明 real runner 会调用 Chat Completions 兼容 LLM。
- real runner 示例追加：

```bash
LSI_LLM_BASE_URL=https://api.openai.com \
LSI_LLM_API_KEY="$OPENAI_API_KEY" \
LSI_LLM_MODEL=gpt-4.1-mini \
LSI_LLM_TIMEOUT=2m \
```

- 架构图中把 `Runner -. "planned" .-> LLM` 改为 `Runner -. "real mode" .-> LLM`。
- 配置表新增 `LSI_LLM_BASE_URL`、`LSI_LLM_API_KEY`、`LSI_LLM_MODEL`、`LSI_LLM_TIMEOUT`。
- 路线图勾选 `OpenAI-compatible LLM 翻译链路` 和 `基于真实转写和翻译结果生成 translated.vtt / bilingual.vtt`。

- [ ] **步骤 2: 更新 backend README**

修改 `backend/README.md`：

- real runner 示例追加 LLM 环境变量。
- runner boundary 改为说明 real runner 会生成真实 `audio.mp3`、`source.vtt`、`translated.vtt` 和 `bilingual.vtt`。
- 说明 `LSI_LLM_API_KEY` 只在 backend 读取，extension 不保存 provider key。

- [ ] **步骤 3: 运行全量 backend 测试**

运行:

```bash
cd backend
mise exec -- go test ./...
```

预期: PASS。

- [ ] **步骤 4: 运行 backend 构建**

运行:

```bash
cd backend
mise exec -- go build ./...
```

预期: PASS，无编译错误。

- [ ] **步骤 5: 检查工作树**

运行:

```bash
git status --short
```

预期: 只包含本功能相关文件。

- [ ] **步骤 6: 提交**

```bash
git add README.md backend/README.md
git commit -m "docs: document real translation runner"
```

## 自检清单

- [ ] `mock` runner 没有调用 translator 或 LLM。
- [ ] `real` runner 的 `translated.vtt` 不再使用 `mockTranslatedVTT`。
- [ ] `real` runner 的 `bilingual.vtt` 使用真实 source cue 和 translator 输出。
- [ ] Chat translator 对每条 cue 单独请求，并携带前后各 10 条以内的上下文。
- [ ] Chat translator 只接受单条 `translation` 字段。
- [ ] 错误消息不包含 `LSI_LLM_API_KEY`。
- [ ] `cd backend && mise exec -- go test ./...` 通过。
- [ ] `cd backend && mise exec -- go build ./...` 通过。
