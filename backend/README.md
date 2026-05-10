# lets-sub-it-api

## Setup

先准备 Go 工具链和依赖：

```bash
mise install
cd backend
mise exec -- go mod download
```

## Run

启动本地 API server 前，先同步本仓库的 `whisper-cli` 依赖：

```bash
cd ../whisper
mise exec -- uv sync --dev
cd ../backend
```

还需要确认本机已安装 `yt-dlp` 和 `ffmpeg`；否则后端启动会失败。

```bash
PATH="$PWD/../whisper/.venv/bin:$PATH" \
LSI_DOWNLOAD_TIMEOUT=10m \
LSI_WHISPER_MODEL=small \
LSI_WHISPER_COMPUTE_TYPE=int8 \
LSI_LLM_BASE_URL=https://api.openai.com \
LSI_LLM_API_KEY="$OPENAI_API_KEY" \
LSI_LLM_MODEL=gpt-4.1-mini \
LSI_LLM_TIMEOUT=2m \
LSI_LOG_LEVEL=info \
LSI_ADDR=127.0.0.1:8080 \
mise exec -- go run ./cmd/server
```

## Logs

backend 输出 JSON 结构化日志。`LSI_LOG_LEVEL` 默认是 `info`，可设为 `debug`、`info`、`warn` 或 `error`；非法值会回退到 `info`。

默认 `info` 级别会记录 server 启停、HTTP 请求方法/路径/状态/耗时、job 创建或复用、job 阶段开始/完成/失败，以及非预期 store 查询错误。job 链路日志会携带 `job_id`、`video_id`、语言、阶段或耗时等字段，覆盖从创建 job、下载、转写、翻译、打包到完成或失败的链路。

`debug` 会额外输出外部命令调用、字幕 cue 数、逐条翻译请求成功诊断和数据库查询耗时。数据库慢查询和查询错误分别使用 `warn` 和 `error`。日志不会记录完整 LLM 请求体、字幕正文、provider key 或 HTTP query string。

## API quick check

可以用下面的请求快速确认 `POST /jobs` 正常工作。请求体需要包含 `youtubeUrl`、`sourceLanguage` 和 `targetLanguage`：

```bash
curl -X POST "http://127.0.0.1:8080/jobs" \
  -H "Content-Type: application/json" \
  -d '{
    "youtubeUrl": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "sourceLanguage": "ja",
    "targetLanguage": "zh"
  }'
```

接口会创建一个新 job；如果同一个视频和目标语言已有可复用 job，会返回已复用结果。

popup 重新打开时可使用 `GET /jobs/active?videoId=...&targetLanguage=...` 查询指定视频和目标语言最近的 job，用于恢复任务进度展示。

## Test

运行后端测试：

```bash
mise exec -- go test ./...
```

## Runner boundary

当前 backend 已经实现了真实 HTTP API、SQLite 持久化、job 复用、真实下载、转写、翻译和 VTT 字幕文件服务。

backend 会调用 `yt-dlp` 产出 `audio.mp3`，调用 PATH 中本仓库的 Python `whisper-cli` 产出真实 `source.vtt`，再调用 Chat Completions 兼容 LLM 生成真实 `translated.vtt` 和 `bilingual.vtt`。

`LSI_LLM_API_KEY` 对 OpenAI 默认 endpoint 必填；本地无鉴权兼容服务可留空。该值只由 backend 读取，extension 不保存 provider key，也不直接调用翻译 provider。

当前 LLM 翻译链路会对临时上游错误和格式错误响应重试；尚未实现并发控制或成本统计。
