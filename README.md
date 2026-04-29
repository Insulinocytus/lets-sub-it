# Lets Sub It

<!-- prettier-ignore -->
<div align="center">

**自托管的 YouTube 字幕生成与翻译工具**

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go&logoColor=white)
![Python](https://img.shields.io/badge/Python-3.12-3776AB?style=flat-square&logo=python&logoColor=white)
![Node.js](https://img.shields.io/badge/Node.js-22-3C873A?style=flat-square&logo=node.js&logoColor=white)
![Chrome MV3](https://img.shields.io/badge/Chrome-MV3-4285F4?style=flat-square&logo=googlechrome&logoColor=white)
![SQLite](https://img.shields.io/badge/SQLite-local-003B57?style=flat-square&logo=sqlite&logoColor=white)

[概览](#概览) • [快速开始](#快速开始) • [真实模式](#真实下载与转写模式) • [架构](#架构) • [API](#api) • [开发](#开发与验证)

</div>

Lets Sub It 是一个本地优先的字幕工作台：提交 YouTube 公开视频链接，后端创建字幕任务，下载音频、转写、翻译并打包字幕，Chrome extension 再在 YouTube 播放页加载 `translated` 或 `bilingual` 字幕。

> [!IMPORTANT]
> 项目仍处于 MVP 阶段。默认 backend 使用 mock runner，不访问 YouTube、Whisper 模型或 LLM。设置 `LSI_RUNNER_MODE=real` 后，目前只有下载和源语言转写阶段会真实执行；翻译和最终 `translated.vtt` / `bilingual.vtt` 仍是 mock 结果。

## 概览

当前仓库由三个可独立运行和测试的模块组成：

| 模块 | 技术栈 | 当前能力 |
| --- | --- | --- |
| `backend/` | Go 1.22, SQLite, GORM | HTTP API、SQLite 持久化、job 复用、mock runner、可选真实下载与转写、VTT 文件服务 |
| `whisper/` | Python 3.12, `faster-whisper`, `uv` | 本地音频转 WebVTT、CLI JSON 输出、退出码契约和离线测试 |
| `extension/` | WXT, Vue, TypeScript, Vitest | Chrome MV3 popup、background API 网关、storage 缓存、YouTube 播放页字幕层 |

主要链路：

1. 在 extension popup 中提交 YouTube URL、源语言和目标语言。
2. background service worker 调用本机 backend。
3. backend 创建或复用 `videoId + targetLanguage` 对应的 job。
4. runner 推进 `queued -> downloading -> transcribing -> translating -> packaging -> completed`。
5. 播放页 content script 查询字幕资产并渲染字幕层。

## 功能

- **本机自托管**：面向单用户本地部署，SQLite 数据库、job 工作目录和字幕文件都保存在本机。
- **任务复用**：同一个 `videoId + targetLanguage` 会复用已完成结果或进行中的 job。
- **清晰状态流**：extension 可展示 `queued`、`downloading`、`transcribing`、`translating`、`packaging`、`completed` 和 `failed`。
- **字幕资产服务**：backend 只通过 `/subtitle-files/:jobId/:mode` 暴露 VTT，不向前端返回本地绝对路径。
- **播放页集成**：YouTube watch 页面支持 `translated` 和 `bilingual` 两种字幕模式。
- **可替换转写 CLI**：Go backend 通过 PATH 调用本仓库的 `whisper-cli`，CLI 输出经过校验的 WebVTT。

## 快速开始

### 1. 准备工具链

项目使用根目录 `mise.toml` 固定工具版本：

| 工具 | 版本 |
| --- | --- |
| Go | `1.22` |
| Python | `3.12` |
| Node.js | `22` |
| uv | `latest` |

```bash
mise install
```

> [!TIP]
> 如果当前 shell 没有自动激活 `mise`，请使用 `mise exec --` 前缀运行项目命令。下面示例都采用这种写法。

### 2. 启动 backend mock API

```bash
cd backend
mise exec -- go mod download
LSI_ADDR=127.0.0.1:8080 mise exec -- go run ./cmd/server
```

另开一个终端创建字幕任务：

```bash
curl -X POST "http://127.0.0.1:8080/jobs" \
  -H "Content-Type: application/json" \
  -d '{
    "youtubeUrl": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "sourceLanguage": "en",
    "targetLanguage": "zh"
  }'
```

mock runner 会推进完整状态流，并在 `LSI_WORK_DIR` 下生成 `source.vtt`、`translated.vtt` 和 `bilingual.vtt`。

### 3. 启动 Chrome extension

```bash
cd extension
mise exec -- npm install
mise exec -- npm run dev
```

在 Chrome extension developer mode 中加载 WXT 输出目录：

```text
extension/.output/chrome-mv3
```

popup 默认连接 `http://127.0.0.1:8080`。当前只允许带端口的本机 HTTP origin，例如 `http://localhost:8080` 或 `http://127.0.0.1:8080`。

### 4. 单独运行 Whisper CLI

```bash
cd whisper
mise exec -- uv sync --dev
mise exec -- uv run whisper-cli \
  --input /path/to/audio.mp3 \
  --output /tmp/source.vtt \
  --model small \
  --language ja
```

成功时 stdout 输出 JSON，`--output` 写入合法 WebVTT：

```json
{
  "output": "/tmp/source.vtt",
  "language": "ja",
  "duration_seconds": 123.45,
  "segments": 42
}
```

## 真实下载与转写模式

`LSI_RUNNER_MODE=real` 会让 backend 在 `downloading` 阶段调用 `yt-dlp`，在 `transcribing` 阶段调用 PATH 中的 `whisper-cli`。

```bash
cd whisper
mise exec -- uv sync --dev
```

```bash
cd backend
PATH="$PWD/../whisper/.venv/bin:$PATH" \
LSI_RUNNER_MODE=real \
LSI_DOWNLOAD_TIMEOUT=10m \
LSI_WHISPER_MODEL=small \
LSI_ADDR=127.0.0.1:8080 \
mise exec -- go run ./cmd/server
```

> [!NOTE]
> real runner 还要求本机安装 `yt-dlp` 和 `ffmpeg`。这个模式目前会生成真实 `audio.mp3` 和真实 `source.vtt`；翻译阶段和最终双语字幕仍未接入真实 LLM。

## 架构

```mermaid
flowchart LR
  Popup["extension popup"] -->|"runtime message"| Background["background service worker"]
  Content["YouTube content script"] -->|"runtime message"| Background
  Background -->|"HTTP"| API["Go API server"]
  API --> Store[("SQLite")]
  API --> Runner["runner<br/>mock by default"]
  Runner --> Files["source.vtt<br/>translated.vtt<br/>bilingual.vtt"]
  API -->|"text/vtt"| Background
  Content -->|"Shadow DOM"| YouTube["YouTube watch page"]

  Runner -. "real mode" .-> Ytdlp["yt-dlp + ffmpeg<br/>audio.mp3"]
  Runner -. "real mode" .-> Whisper["whisper-cli<br/>source.vtt"]
  Runner -. "planned" .-> LLM["OpenAI-compatible<br/>translation"]
```

状态流转：

```text
queued -> downloading -> transcribing -> translating -> packaging -> completed
```

失败时状态为 `failed`，响应中的 `errorMessage` 会记录错误摘要。

## API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/jobs` | 创建或复用字幕生成 job |
| `GET` | `/jobs/:id` | 查询 job 状态 |
| `GET` | `/subtitle-assets?videoId=...&targetLanguage=...` | 查询已完成字幕资产 |
| `GET` | `/subtitle-files/:jobId/:mode` | 读取 VTT 文件，`mode` 为 `source`、`translated` 或 `bilingual` |

主要配置：

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `LSI_ADDR` | `127.0.0.1:8080` | HTTP 监听地址 |
| `LSI_DB_PATH` | `./data/backend.sqlite3` | SQLite 数据库路径 |
| `LSI_WORK_DIR` | `./data/jobs` | job 工作目录根路径 |
| `LSI_RUNNER_MODE` | `mock` | runner 模式：`mock` 或 `real` |
| `LSI_DOWNLOAD_TIMEOUT` | `10m` | `real` 模式下单次下载超时 |
| `LSI_WHISPER_MODEL` | `small` | `real` 模式下传给 `whisper-cli --model` 的模型名 |

## Whisper CLI 契约

`whisper-cli` 输入本地音频文件，输出合法 WebVTT。

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `--input` | 是 | 本地音频文件路径 |
| `--output` | 是 | 输出 `.vtt` 路径，不能与输入路径相同 |
| `--model` | 是 | `faster-whisper` 模型名，例如 `small` |
| `--language` | 是 | 转写语言代码，例如 `ja`、`en` |

退出码：

| 退出码 | 含义 |
| --- | --- |
| `0` | 成功 |
| `2` | 输入校验失败，例如输入文件不存在、模型名或语言无效 |
| `3` | 转写失败 |
| `4` | 输出校验失败，例如无法生成合法 VTT |

## 仓库结构

```text
.
├── backend/                 # Go API server
│   ├── cmd/server/          # HTTP server 入口
│   └── internal/            # api、store、runner、app
├── docs/                    # PRD、规格说明和实施计划
├── extension/               # Chrome MV3 extension
│   ├── entrypoints/         # popup、background、content script
│   └── src/                 # API、storage、subtitle、YouTube 集成和 UI
├── whisper/                 # Python faster-whisper CLI
│   ├── src/whisper_cli/     # CLI、转写适配和 VTT 处理
│   └── tests/               # pytest 单元测试
└── mise.toml                # 本地工具链版本
```

## 开发与验证

安装依赖：

```bash
cd backend && mise exec -- go mod download
cd ../whisper && mise exec -- uv sync --dev
cd ../extension && mise exec -- npm install
```

运行测试：

```bash
cd backend && mise exec -- go test ./...
cd ../whisper && mise exec -- uv run pytest
cd ../extension && mise exec -- npm run test
```

构建验证：

```bash
cd backend && mise exec -- go build ./...
cd ../whisper && mise exec -- uv build
cd ../extension && mise exec -- npm run build
```

## 当前限制

- 只面向单用户本地自托管，没有登录、鉴权、多用户或公网部署能力。
- 只处理 YouTube 公开视频，不支持私有视频、登录态或 cookie 导入。
- extension 第一版只支持 `en` 和 `zh`，且 `sourceLanguage` 与 `targetLanguage` 不能相同。
- extension 只允许 `localhost` / `127.0.0.1` backend URL，不支持任意远程主机。
- OpenAI-compatible LLM 翻译链路尚未实现，真实 `translated.vtt` 和 `bilingual.vtt` 仍是后续工作。

## 路线图

- [x] `whisper-cli` 本地转写命令、WebVTT 渲染和退出码契约
- [x] Go API、SQLite、job 复用、mock runner 和字幕文件服务
- [x] Chrome extension 任务提交、状态轮询、字幕缓存和播放页字幕层
- [x] 可选真实下载阶段，调用 `yt-dlp` 和 `ffmpeg` 产出 `audio.mp3`
- [x] backend 调用真实 `whisper-cli` 产出 `source.vtt`
- [ ] OpenAI-compatible LLM 翻译链路
- [ ] 基于真实转写和翻译结果生成 `translated.vtt` / `bilingual.vtt`
- [ ] 更完整的任务历史、错误恢复和配置体验

## 排障

- `LSI_RUNNER_MODE=real` 启动失败：确认 `yt-dlp`、`ffmpeg` 和 `whisper-cli` 都能在当前 shell 的 `PATH` 中找到。
- extension 无法连接 backend：确认 backend URL 是带端口的本机 HTTP origin，不要带路径、查询参数或 hash。
- Whisper CLI 首次运行较慢：真实转写可能触发模型下载，单元测试不会下载模型或依赖 GPU。
- 字幕文件返回 404：确认 job 已完成，且请求的 mode 是 `source`、`translated` 或 `bilingual`。

## 相关文档

- [PRD](docs/PRD.md)
- [Backend README](backend/README.md)
- [Whisper README](whisper/README.md)
- [Extension README](extension/README.md)
- [Whisper CLI 设计说明](docs/superpowers/specs/2026-04-23-whisper-cli-design.md)
- [Backend Mock MVP 设计](docs/superpowers/specs/2026-04-24-backend-mock-mvp-design.md)
- [Extension MVP 设计](docs/superpowers/specs/2026-04-25-extension-mvp-design.md)
- [真实音频下载设计](docs/superpowers/specs/2026-04-27-real-audio-download-design.md)
