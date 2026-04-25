# lets-sub-it-api

## Setup

先准备 Go 工具链和依赖：

```bash
mise install
cd backend
mise exec -- go mod download
```

## Run

启动本地 mock API server：

```bash
LSI_ADDR=127.0.0.1:8080 mise exec -- go run ./cmd/server
```

## API quick check

可以用下面的请求快速确认 `POST /jobs` 正常工作。请求体需要包含 `youtubeUrl`、`sourceLanguage` 和 `targetLanguage`：

```bash
curl -X POST "http://127.0.0.1:8080/jobs" \
  -H "Content-Type: application/json" \
  -d '{
    "youtubeUrl": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "sourceLanguage": "ja",
    "targetLanguage": "zh-Hans"
  }'
```

接口会创建一个新 job；如果同一个视频和目标语言已有可复用 job，会返回已复用结果。

## Test

运行后端测试：

```bash
mise exec -- go test ./...
```

## Mock boundary

当前 backend 已经实现了真实 HTTP API、SQLite 持久化、job 复用、mock 状态推进和 VTT 字幕文件服务，方便前端和播放页联调。

这一阶段不会调用真实的 `yt-dlp`、`ffmpeg`、`whisper-cli` 或 LLM；这些能力仍然保留为后续接入真实外部工具的边界。
