# 本地开发

本文说明如何在本机启动 Lets Sub It 的后端、Whisper CLI 和 Chrome 扩展开发环境。

## 前置条件

- 在仓库根目录运行 `mise install`，安装固定版本的 Go、Python、Node.js 和 `uv`。
- 确认 `yt-dlp` 和 `ffmpeg` 已安装，并且在后端运行时的 `PATH` 中可用。
- 配置 LLM 相关环境变量。OpenAI 默认 endpoint 至少需要 `LSI_LLM_API_KEY` 和 `LSI_LLM_MODEL`；如使用其他 OpenAI-compatible 服务，还需要确认 `LSI_LLM_BASE_URL`。

## 安装依赖

```bash
cd backend && mise exec -- go mod download
cd ../whisper && mise exec -- uv sync --dev
cd ../extension && mise exec -- npm install
```

## 启动后端

先确保 `whisper-cli` 的虚拟环境已经通过上一步创建，然后从 `backend/` 目录启动后端：

```bash
cd backend
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

## API smoke test

后端启动后，可以用一个公开 YouTube URL 创建任务：

```bash
curl -X POST "http://127.0.0.1:8080/jobs" \
  -H "Content-Type: application/json" \
  -d '{"youtubeUrl":"https://www.youtube.com/watch?v=dQw4w9WgXcQ","sourceLanguage":"en","targetLanguage":"zh"}'
```

## 启动扩展开发服务器

```bash
cd extension && mise exec -- npm run dev
```

在 Chrome 扩展开发者模式中加载构建目录：

```text
extension/.output/chrome-mv3
```
