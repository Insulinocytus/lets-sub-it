<div align="center">

# Lets Sub It

自托管 YouTube 字幕生成与翻译

</div>

Lets Sub It 下载 YouTube 视频音频，通过本地 Whisper 服务转写字幕，再经由 OpenAI 兼容 LLM 翻译，最终在 Chrome 扩展中渲染到 YouTube 播放页。

```text
Chrome MV3 扩展  →  后端 HTTP API  →  yt-dlp + ffmpeg
                        │             →  Whisper HTTP 服务
                        │             →  OpenAI 兼容 LLM
                        ▼
                 SQLite + WebVTT 文件
```

## 功能特性

- 提交 YouTube 字幕任务并实时跟踪进度
- 使用 `faster-whisper` 本地转写，无需云端语音 API
- 通过任意 OpenAI 兼容 Chat Completions API 翻译字幕
- 生成 `source`（源语言）、`translated`（翻译）、`bilingual`（双语）三种 WebVTT 输出
- Chrome MV3 扩展直接在 YouTube 播放页渲染字幕
- SQLite 持久化，已完成的任务结果可复用
- Docker Compose 一键自托管部署

> [!NOTE]
> 当前 Chrome 扩展仅连接 `http://127.0.0.1:8080`，manifest 仅授予 `localhost` 和 `127.0.0.1` 的主机权限。

## 模块组成

| 路径 | 用途 | 技术栈 |
| --- | --- | --- |
| `backend/` | HTTP API、任务生命周期、SQLite 持久化、下载与编排 | Go 1.22, GORM, SQLite |
| `whisper/` | 本地转写 HTTP 服务，返回校验后的 WebVTT | Python 3.12, FastAPI, faster-whisper |
| `extension/` | Chrome MV3 扩展，提交任务并渲染字幕 | WXT, Vue 3, TypeScript, Tailwind CSS, shadcn-vue |

根目录 `Taskfile.yml` 是跨模块命令入口，工具版本由 `mise.toml` 固定。

## 前置依赖

- [mise](https://mise.jdx.dev/) — 管理固定版本工具链
- Docker 和 Docker Compose — 最简自托管方式
- Chrome 或 Chromium — 加载本地扩展

本地运行后端（不走 Docker）时，还需确保 `yt-dlp` 和 `ffmpeg` 在 `PATH` 中。

> [!TIP]
> 想最快跑通完整服务，优先使用 Docker。Docker Compose 会启动 backend 和 whisper 两个服务，并包含所需运行时依赖。

## 快速开始

### Docker 方式（推荐）

1. **配置环境**

   ```bash
   cp .env.example .env
   ```

   编辑 `.env`，至少设置：

   ```env
   LSI_LLM_API_KEY=sk-your-key-here
   LSI_LLM_MODEL=gpt-4.1-mini
   ```

2. **构建并启动服务**

   ```bash
   mise trust
   mise install
   task docker:build
   ```

   后端监听 `http://127.0.0.1:8080`。

3. **提交测试任务**

   ```bash
   task api:smoke
   ```

4. **加载 Chrome 扩展**

   ```bash
   task build:extension
   ```

   打开 `chrome://extensions`，开启开发者模式，加载 `extension/.output/chrome-mv3` 作为未打包扩展。打开 YouTube 视频页面，通过扩展弹窗提交字幕任务。

### 本地开发方式

```bash
mise trust
mise install
task setup
```

在独立终端分别启动服务：

```bash
task dev:whisper    # Whisper 服务 :8081
task dev:backend     # 后端 API :8080
task dev:extension   # 扩展开发服务器
```

## 常用命令

| 命令 | 说明 |
| --- | --- |
| `task setup` | 安装工具链及全部模块依赖 |
| `task dev:backend` | 本地运行 Go 后端 |
| `task dev:whisper` | 本地运行 Whisper HTTP 服务 |
| `task dev:extension` | 启动扩展开发服务器 |
| `task api:smoke` | 向本地后端提交测试任务 |
| `task test` | 运行全部模块测试 |
| `task check` | 运行全部测试 + 扩展类型检查 |
| `task build` | 构建全部模块 |
| `task docker:build` | 重新构建并启动 Docker 服务 |
| `task docker:logs` | 查看 Docker 服务日志 |
| `task docker:down` | 停止 Docker 服务 |

查看完整命令列表：

```bash
task --list
```

## API 接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/jobs` | 创建或复用字幕任务 |
| `GET` | `/jobs/{jobId}` | 查询任务状态 |
| `GET` | `/jobs/active?videoId=...&targetLanguage=...` | 查询指定视频和语言的最新任务 |
| `GET` | `/subtitle-assets?videoId=...&targetLanguage=...` | 列出已生成的字幕资产 |
| `GET` | `/subtitle-files/{jobId}/{mode}` | 下载 VTT 文件（`source`、`translated` 或 `bilingual`） |

创建任务示例：

```bash
curl -X POST "http://127.0.0.1:8080/jobs" \
  -H "Content-Type: application/json" \
  -d '{"youtubeUrl":"https://www.youtube.com/watch?v=dQw4w9WgXcQ","sourceLanguage":"en","targetLanguage":"zh"}'
```

任务状态流转：`queued` → `downloading` → `transcribing` → `translating` → `packaging` → `completed`（或 `failed`）。

## 配置

全部通过环境变量配置，Docker 从 `.env` 读取。

### 后端

| 变量 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `LSI_ADDR` | 否 | `127.0.0.1:8080` | 后端监听地址 |
| `LSI_DB_PATH` | 否 | `./data/backend.sqlite3` | SQLite 数据库路径 |
| `LSI_WORK_DIR` | 否 | `./data/jobs` | 任务工作目录 |
| `LSI_LOG_LEVEL` | 否 | `info` | 日志级别 |
| `LSI_DOWNLOAD_TIMEOUT` | 否 | `10m` | 音频下载超时 |
| `LSI_WHISPER_BASE_URL` | 否 | `http://127.0.0.1:8081` | Whisper 服务地址（Docker 内为 `http://whisper:8081`） |
| `LSI_WHISPER_MODEL` | 否 | `small` | faster-whisper 模型名 |
| `LSI_WHISPER_COMPUTE_TYPE` | 否 | `default` | faster-whisper 计算类型 |
| `LSI_WHISPER_TIMEOUT` | 否 | `30m` | Whisper 请求超时 |
| `LSI_WHISPER_POLL_INTERVAL` | 否 | `2s` | Whisper 任务轮询间隔 |
| `LSI_LLM_BASE_URL` | 否 | `https://api.openai.com` | OpenAI 兼容 API 地址 |
| `LSI_LLM_API_KEY` | 翻译必填 | 空 | LLM API 密钥 |
| `LSI_LLM_MODEL` | 翻译必填 | 空 | LLM 模型名称 |
| `LSI_LLM_TIMEOUT` | 否 | `2m` | LLM 请求超时 |

### Whisper 服务

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `LSI_WHISPER_WORK_DIR` | `/data/transcriptions` | 转写任务数据目录 |
| `HF_HOME` | 工具默认值 | Hugging Face 缓存目录（Docker 内为 `/huggingface`） |
| `HF_TOKEN` | 空 | 可选，用于提高 Hugging Face Hub 下载限额 |

### Docker

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `LSI_DOCKER_BIND_HOST` | — | **`.env` 中必填。** Docker 端口绑定主机地址（本地使用填 `127.0.0.1`） |

Docker Compose 使用三个持久卷：

| 卷名 | 用途 |
| --- | --- |
| `lsi-data` | SQLite 数据库和任务文件 |
| `lsi-hf-cache` | Hugging Face 模型缓存（重启后保留） |
| `lsi-whisper-data` | Whisper 服务转写工作目录 |

> [!WARNING]
> 请勿提交 `.env` 或真实密钥。本地配置请从 `.env.example` 开始。

## 测试与构建

```bash
task test             # 全部模块测试
task typecheck        # 扩展 TypeScript 类型检查
task check            # 测试 + 类型检查
task build            # 构建全部模块
```

按模块：

```bash
task test:backend     # Go 测试
task test:whisper     # Python pytest
task test:extension   # Vitest 测试
task build:extension  # Chrome MV3 输出
```

构建产物：

| 模块 | 输出 |
| --- | --- |
| Extension | `extension/.output/chrome-mv3` |
| Whisper 包 | `whisper/dist/` |
| Backend | 由 `backend/Dockerfile` 构建的 Docker 镜像 |

## 项目结构

```text
.
├── backend/             # Go HTTP API 与任务运行器
│   ├── cmd/               # 入口程序
│   └── internal/          # API、store、runner 包
├── whisper/              # Python Whisper HTTP 服务
│   └── src/whisper_cli/
│       ├── server.py      # FastAPI 端点
│       ├── transcribe.py  # 转写逻辑
│       └── vtt.py          # WebVTT 渲染辅助
├── extension/            # Chrome MV3 扩展
│   ├── entrypoints/       # WXT 入口（popup、background、content）
│   └── src/                # Vue 3 组件和工具函数
├── docker-compose.yml    # 后端 + Whisper 服务栈
├── Taskfile.yml          # 跨模块命令入口
├── mise.toml             # 固定工具版本
└── .env.example          # 本地配置模板
```

## 常见问题

**后端报错找不到工具** — 本地运行需确保 `yt-dlp` 和 `ffmpeg` 在 `PATH` 中。也可直接使用 Docker，镜像内已包含运行时依赖。

**后端无法连接 Whisper 服务** — 确认 Whisper 服务已启动，且 `LSI_WHISPER_BASE_URL` 指向正确地址。或使用 Docker Compose 同时启动两个服务。

**翻译环节失败** — 检查 `.env` 或当前 shell 中是否设置了 `LSI_LLM_API_KEY` 和 `LSI_LLM_MODEL`。

**扩展无法连接后端** — 扩展仅支持 `http://127.0.0.1` 或 `http://localhost` 来源。远程后端地址需同时修改 manifest 主机权限和 URL 校验。

**首次转写较慢** — `faster-whisper` 首次使用时会下载模型。Docker 模式下模型缓存在 `lsi-hf-cache` 卷中，后续任务可复用。