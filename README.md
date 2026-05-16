# Lets Sub It

Self-hosted YouTube subtitle generation and translation helper.

Lets Sub It 是一个面向本地自托管的 YouTube 字幕工具：后端下载视频音频、调用本地 Whisper 转写字幕、通过 OpenAI-compatible LLM 翻译字幕，Chrome 扩展负责提交任务并在 YouTube 播放页显示字幕。

## Features

- 提交 YouTube 字幕生成任务，并跟踪任务状态。
- 使用 `yt-dlp` 下载音频，使用本地 `faster-whisper` 生成 WebVTT 字幕。
- 使用 OpenAI-compatible Chat Completions API 翻译字幕。
- 生成 `source`、`translated`、`bilingual` 三种 WebVTT 字幕资产。
- Chrome Manifest V3 扩展可在 YouTube watch 页面渲染已完成字幕。
- SQLite 持久化任务、字幕资产和可复用结果。
- 支持本地开发和 Docker 自托管运行。

> [!NOTE]
> 当前扩展默认只连接 `http://127.0.0.1:8080`，并且 manifest 只授予 `localhost` 和 `127.0.0.1` 后端访问权限。

## Architecture

```text
Chrome extension  ->  backend HTTP API  ->  yt-dlp
        |                    |              whisper-cli
        |                    |              OpenAI-compatible LLM
        |                    v
        +----------  SQLite + WebVTT subtitle files
```

项目由三个主要模块组成：

| Path | Purpose | Stack |
| --- | --- | --- |
| `backend/` | HTTP API、任务生命周期、SQLite 持久化、下载、转写和翻译编排 | Go 1.22 |
| `whisper/` | 本地 `whisper-cli`，把音频转写成经过校验的 WebVTT | Python 3.12, faster-whisper |
| `extension/` | Chrome MV3 扩展，提交任务并在 YouTube 播放页显示字幕 | WXT, Vue 3, TypeScript, Tailwind CSS |

根目录 `Taskfile.yml` 是跨模块命令入口。工具版本由 `mise.toml` 固定。

## Prerequisites

- [mise](https://mise.jdx.dev/) 用于安装固定版本工具链。
- Docker 和 Docker Compose，用于最简单的后端自托管运行方式。
- Chrome 或 Chromium，用于加载本地扩展。

本地直接运行后端时，还需要系统中可用的：

- `yt-dlp`
- `ffmpeg`
- `whisper-cli`，可通过 `task deps:whisper` 安装到 `whisper/.venv/bin`

> [!TIP]
> 如果只是想先跑通后端，优先使用 Docker。Docker 镜像会打包 Go 后端、Python `whisper-cli`、`yt-dlp` 和 `ffmpeg`。

## Quick Start

### 1. Install tools and dependencies

```bash
task setup
```

这会安装固定版本工具链，并安装 backend、whisper 和 extension 的依赖。

### 2. Configure the backend

```bash
cp .env.example .env
```

编辑 `.env`，至少设置真实的 LLM 配置：

```env
LSI_LLM_API_KEY=sk-your-key-here
LSI_LLM_MODEL=gpt-4.1-mini
```

### 3. Start the backend with Docker

```bash
task docker:build
```

后端默认监听：

```text
http://127.0.0.1:8080
```

### 4. Submit a smoke-test job

```bash
task api:smoke
```

或者手动调用：

```bash
curl -X POST "http://127.0.0.1:8080/jobs" \
  -H "Content-Type: application/json" \
  -d '{"youtubeUrl":"https://www.youtube.com/watch?v=dQw4w9WgXcQ","sourceLanguage":"en","targetLanguage":"zh"}'
```

### 5. Build and load the Chrome extension

```bash
task build:extension
```

然后在 Chrome 中打开 `chrome://extensions`：

1. 开启 Developer mode。
2. 选择 Load unpacked。
3. 加载 `extension/.output/chrome-mv3`。
4. 打开 YouTube watch 页面，通过扩展 popup 提交字幕任务。

## Development

### Backend

```bash
task dev:backend
```

默认值：

| Variable | Default |
| --- | --- |
| `LSI_ADDR` | `127.0.0.1:8080` |
| `LSI_DOWNLOAD_TIMEOUT` | `10m` |
| `LSI_WHISPER_MODEL` | `small` |
| `LSI_LLM_BASE_URL` | `https://api.openai.com` |
| `LSI_LLM_TIMEOUT` | `2m` |

`task dev:backend` 会确保 whisper 依赖存在，并把 `whisper/.venv/bin` 加入 `PATH`，让 Go 后端可以调用 `whisper-cli`。

### Extension

```bash
task dev:extension
```

WXT 会生成开发用扩展输出。扩展 popup 默认连接 `http://127.0.0.1:8080`。

### Docker

```bash
task docker:up       # start backend
task docker:logs     # follow logs
task docker:down     # stop backend
task docker:build    # rebuild and start backend
```

Docker Compose 使用两个持久卷：

| Volume | Purpose |
| --- | --- |
| `lsi-data` | SQLite 数据库和任务文件 |
| `lsi-hf-cache` | Hugging Face 模型缓存 |

## Configuration

后端配置来自环境变量。Docker 运行时从 `.env` 读取。

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `LSI_DOCKER_BIND_HOST` | Docker only | `127.0.0.1` in `.env.example` | Docker 端口绑定主机 |
| `LSI_ADDR` | No | `127.0.0.1:8080` | 后端监听地址 |
| `LSI_DB_PATH` | No | `./data/backend.sqlite3` | SQLite 数据库路径 |
| `LSI_WORK_DIR` | No | `./data/jobs` | 任务工作目录 |
| `LSI_LOG_LEVEL` | No | `info` | 日志级别 |
| `LSI_DOWNLOAD_TIMEOUT` | No | `10m` | 音频下载超时 |
| `LSI_WHISPER_MODEL` | No | `small` | faster-whisper 模型名 |
| `LSI_WHISPER_COMPUTE_TYPE` | No | `default` | faster-whisper compute type |
| `LSI_LLM_BASE_URL` | No | `https://api.openai.com` | OpenAI-compatible API base URL |
| `LSI_LLM_API_KEY` | Yes for translation | empty | LLM API key |
| `LSI_LLM_MODEL` | Yes for translation | empty | LLM model name |
| `LSI_LLM_TIMEOUT` | No | `2m` | LLM 请求超时 |
| `HF_TOKEN` | No | empty | 可选，用于提高 Hugging Face Hub 下载限额 |

> [!WARNING]
> `.env` 是本地密钥文件，已被 `.gitignore` 忽略。不要提交真实 API key。

## API Overview

主要后端接口：

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/jobs` | 创建或复用字幕任务 |
| `GET` | `/jobs/{jobId}` | 查询任务状态 |
| `GET` | `/jobs/active?videoId=...&targetLanguage=...` | 查询某个视频和目标语言的最新任务 |
| `GET` | `/subtitle-assets?videoId=...&targetLanguage=...` | 查询已生成字幕资产 |
| `GET` | `/subtitle-files/{jobId}/{mode}` | 下载 VTT 文件，`mode` 为 `source`、`translated` 或 `bilingual` |

创建任务请求示例：

```json
{
  "youtubeUrl": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
  "sourceLanguage": "en",
  "targetLanguage": "zh"
}
```

任务状态包括：`queued`、`downloading`、`transcribing`、`translating`、`packaging`、`completed`、`failed`。

## Commands

| Command | Description |
| --- | --- |
| `task setup` | 安装工具链和全部模块依赖 |
| `task dev:backend` | 本地启动 Go 后端 |
| `task dev:extension` | 启动扩展开发服务器 |
| `task api:smoke` | 向本地后端提交一个测试任务 |
| `task test` | 运行 backend、whisper、extension 测试 |
| `task typecheck` | 运行扩展 TypeScript 类型检查 |
| `task check` | 运行全部测试和类型检查 |
| `task build` | 构建全部模块 |
| `task docker:build` | 构建并启动 Docker 后端 |
| `task docker:logs` | 查看 Docker 后端日志 |

查看完整命令列表：

```bash
task --list
```

## Testing and Build

```bash
task test          # all tests
task typecheck     # extension typecheck
task check         # tests + typecheck
task build         # backend + whisper + extension build
```

按模块运行：

```bash
task test:backend
task test:whisper
task test:extension
task build:backend
task build:whisper
task build:extension
```

构建产物：

| Module | Output |
| --- | --- |
| Extension | `extension/.output/chrome-mv3` |
| Whisper package | `whisper/dist` |
| Backend Docker image | 由 `backend/Dockerfile` 构建 |

## Project Structure

```text
.
├── backend/             # Go HTTP API and job runner
├── whisper/             # Python whisper-cli package
├── extension/           # Chrome MV3 extension
├── docs/superpowers/    # Design specs and implementation plans
├── docker-compose.yml   # Docker backend runtime
├── Taskfile.yml         # Cross-module command interface
├── mise.toml            # Pinned tool versions
└── .env.example         # Local Docker/backend environment template
```

## Troubleshooting

### Backend fails with missing tool errors

本地运行后端时需要 `yt-dlp`、`ffmpeg` 和 `whisper-cli` 都在 `PATH` 中。可以改用 Docker，或安装系统依赖后再运行：

```bash
task deps:whisper
task dev:backend
```

### Translation fails during a job

确认 `.env` 或当前 shell 中设置了：

```env
LSI_LLM_API_KEY=...
LSI_LLM_MODEL=...
```

### Extension cannot connect to backend

确认 backend URL 是带端口的本机 HTTP origin，例如：

```text
http://127.0.0.1:8080
```

当前扩展不支持任意远程主机 URL。

### First transcription is slow

首次处理视频时，`faster-whisper` 需要下载模型。Docker 模式会把 Hugging Face 缓存保存在 `lsi-hf-cache` 卷中，后续任务会复用缓存。
