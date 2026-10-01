<div align="center">

# Lets Sub It

**自托管 YouTube 字幕生成与翻译**

[English](README.md) | **简体中文**

</div>

Lets Sub It 从 YouTube 视频下载音频，用本地 Whisper 或兼容 OpenAI 的 STT 服务转写，再通过 OpenAI 兼容接口翻译。Chrome 扩展负责提交任务，并在视频播放页显示源语言、译文或双语字幕。

## 能做什么

- 用内置的 `hwdsl2/whisper-server`（faster-whisper）本地转写，生成 `source.vtt`、`translated.vtt`、`bilingual.vtt`。
- 在 YouTube 页面跟随播放时间显示字幕；弹窗关闭后，后台仍会跟踪任务完成状态。
- 用 SQLite 保存任务和字幕资产，复用已完成的结果。
- 通过 Docker Compose 运行 Go 后端与 Whisper 转写容器的组合。

> [!IMPORTANT]
> 扩展目前只能连接**同一台机器上**带端口的 `http://127.0.0.1` 或 `http://localhost` 后端。Docker 默认也只绑定本机地址；将容器端口暴露到局域网，不会让另一台机器上的扩展获得远程连接能力。

## 快速开始

需要 Docker（含 Compose）、Chrome/Chromium，以及可用的 OpenAI 兼容 Chat Completions API。以下命令在仓库根目录运行。

1. 复制 `.env.example` 为 `.env`。将示例值替换为真实的 `LSI_LLM_API_KEY`、`LSI_LLM_MODEL`；把 `LSI_LOCAL_STT_API_KEY` 改成自己的随机密钥（本地 whisper 容器与后端共用）；若使用其他翻译服务，再设置 `LSI_LLM_BASE_URL`。保留 `LSI_DOCKER_BIND_HOST=127.0.0.1`，即可仅在本机访问后端。
2. 启动服务：

   ```bash
   docker compose up -d --build
   ```

   后端位于 `http://127.0.0.1:8080`。运行 `docker compose ps` 查看状态；whisper 容器首次启动时会下载模型（`small` 约 465 MB），需要一些时间。

3. 安装扩展：在 GitHub Actions 中打开 **Extension Build** 工作流的一次运行，下载并解压 `lets-sub-it-extension-chrome-<运行编号>` 产物；或者安装 [mise](https://mise.jdx.dev/) 后在仓库根目录构建：

   ```bash
   mise trust
   mise install
   task deps:extension
   task build:extension
   ```

   在 `chrome://extensions` 开启开发者模式，选择“加载已解压的扩展程序”，加载 `extension/.output/chrome-mv3`（或下载产物中包含 `manifest.json` 的目录）。
4. 打开 YouTube 视频，点击扩展图标，在弹窗中选择语言并提交任务。任务完成后可切换源语言、译文和双语字幕。

> [!NOTE]
> 提交任务会实际下载视频音频、运行本地模型并请求翻译 API；后者可能产生费用。`.env` 包含密钥，请勿提交。

停止服务：`docker compose down`；SQLite 数据、字幕文件和模型缓存保存在 Compose 卷中。

## 本地开发

使用 `mise.toml` 中的 Go 1.22、Node 22（本地转写由 Docker 提供）。依次运行 `mise trust`、`mise install`、`task setup`，再在一个终端启动转写容器和后端，另一个终端启动扩展：

```bash
export LSI_STT_API_KEY='替换为自己生成的本地密钥'
docker run -d --rm -p 127.0.0.1:9000:9000 -v whisper-data:/var/lib/whisper \
  -e WHISPER_MODEL=small -e WHISPER_API_KEY="$LSI_STT_API_KEY" hwdsl2/whisper-server
LSI_STT_API_KEY="$LSI_STT_API_KEY" task dev:backend   # http://127.0.0.1:8080
# 另一个终端：task dev:extension
```

本地后端还需要 `yt-dlp` 和 `ffmpeg` 位于 `PATH`；`task dev:backend` **不会**自动启动 whisper，它默认连接 `http://127.0.0.1:9000/v1`（Compose 内为 `http://whisper:9000/v1`）。本地运行时需在后端进程环境中提供 `LSI_LLM_API_KEY` 和 `LSI_LLM_MODEL`，仅复制 `.env` 不会为本地进程加载配置。Taskfile 的开发命令使用 POSIX 风格的 shell 环境变量赋值。

## 使用 HTTP API

扩展通常会处理提交、轮询和字幕加载；也可以直接向本地后端提交任务：

```bash
curl -X POST http://127.0.0.1:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{"youtubeUrl":"https://www.youtube.com/watch?v=dQw4w9WgXcQ","sourceLanguage":"en","targetLanguage":"zh"}'
```

响应的 `job.id` 可用于查询进度；将下方 `JOB_ID` 替换为实际值，状态返回 `completed` 后再下载 VTT：

```bash
curl http://127.0.0.1:8080/jobs/JOB_ID
curl http://127.0.0.1:8080/subtitle-files/JOB_ID/bilingual
```

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `POST` | `/jobs` | 创建或复用任务，返回 `job` 与 `reused` |
| `GET` | `/jobs/{jobId}` | 查询状态；`failed` 时包含错误信息 |
| `GET` | `/jobs/active?videoId=...&targetLanguage=...` | 查询该视频和语言的最近任务 |
| `GET` | `/subtitle-assets?videoId=...&targetLanguage=...` | 查询生成的字幕资产 |
| `GET` | `/subtitle-files/{jobId}/{mode}` | 下载 `source`、`translated` 或 `bilingual` VTT |
| `DELETE` | `/subtitle-results?videoId=...&targetLanguage=...` | 删除该视频和语言的字幕结果，返回 `204`（无可删内容时同样返回） |

任务依次进入 `queued` → `downloading` → `transcribing` → `translating` → `packaging` → `completed`；任一阶段出错则进入 `failed`。

删除是逻辑删除：任务记录保留在数据库中作为历史，但不再参与复用、不再出现在任何查询里；工作目录（音频与 VTT）会被删掉。删除时仍在运行的任务会跑完当前步骤，在下一次更新状态时丢弃产物。之后再提交同一视频和语言会重新跑流水线。

`task api:smoke` 对运行中的后端做一次真实的端到端验证：先删除测试视频（默认 19 秒的 `jNQXAC9IVRw`，可用 `LSI_SMOKE_VIDEO_ID` 覆盖；后端地址用 `LSI_BACKEND_URL`）的 `en → zh` 字幕结果，再提交任务并等待完成，打印每个阶段和三份 VTT 的开头；失败或 10 分钟未完成时退出码非 0。它会调用真实的 YouTube、STT 和 LLM，并清掉该视频已有的字幕。

## 架构与配置

```text
Chrome 扩展（弹窗 / 后台 / YouTube 内容脚本）
             │ 提交任务、轮询状态、拉取 VTT
             ▼
Go HTTP API ── SQLite（任务、字幕资产）
      │
      ├── yt-dlp / ffmpeg 下载音频
      ├── Whisper HTTP 服务转写（docker-whisper）→ 本地生成 source.vtt
      └── OpenAI 兼容翻译 API → translated.vtt / bilingual.vtt
```

| 目录 | 职责 |
| --- | --- |
| [`backend/`](backend/) | Go API、SQLite 状态、下载/转写/翻译任务编排、WebVTT 生成 |
| [`extension/`](extension/) | WXT/Vue Chrome MV3 扩展、任务监控和 YouTube 字幕渲染 |

Go 后端的任务不依赖原始 HTTP 连接；服务重启后，未完成的任务会标记为失败，需要重新提交。转写是一次同步的 `POST /audio/transcriptions` 调用（`response_format=verbose_json`），后端要求返回有效的 `segments`（每段含 `start`、`end`、`text`）才会写出 `source.vtt`；没有分块、没有队列、也不回退到其他服务。扩展弹窗关闭后，后台会用浏览器 alarms 继续监测任务。

常用设置（完整列表及 Docker 默认值见 [`.env.example`](.env.example)；非 Docker 后端默认值见 [`backend/internal/app/config.go`](backend/internal/app/config.go)）：

| 变量 | 用途 |
| --- | --- |
| `LSI_LLM_API_KEY`、`LSI_LLM_MODEL` | 翻译服务凭据与模型，实际翻译必需 |
| `LSI_LLM_BASE_URL` | OpenAI 兼容服务地址，默认 `https://api.openai.com` |
| `LSI_LOCAL_STT_API_KEY` | Compose 内 whisper 容器的密钥；后端未单独设置 `LSI_STT_API_KEY` 时也回退使用它 |
| `WHISPER_MODEL`、`WHISPER_COMPUTE_TYPE` | 本地 whisper 容器的模型与计算类型（默认 `small` / `int8`） |
| `LSI_STT_BASE_URL`、`LSI_STT_API_KEY`、`LSI_STT_MODEL`、`LSI_STT_TIMEOUT` | 后端转写服务配置；留空即用 Compose 内的 whisper，可改为任意 OpenAI 兼容 STT 服务 |
| `LSI_DOCKER_BIND_HOST` | Compose 后端端口绑定地址；示例为 `127.0.0.1` |
| `LSI_DB_PATH`、`LSI_WORK_DIR` | 后端数据库与任务文件路径 |

转写配置与 LLM 配置相互独立：不设置 `LSI_STT_*` 时后端使用内置 whisper 容器；要改用外部服务（如 OpenRouter），在 `.env` 设置非空的 `LSI_STT_BASE_URL`、`LSI_STT_API_KEY`，可再设置 `LSI_STT_MODEL`。外部密钥只注入后端容器，不会传给本地 whisper 容器，也不影响 `LSI_LLM_*` 翻译配置。

## 测试与构建

根目录 [`Taskfile.yml`](Taskfile.yml) 定义跨模块命令；若 `task` 不在 `PATH`，可用 `mise exec -- task <命令>`。

| 命令 | 用途 |
| --- | --- |
| `task check` | 全部测试 + 扩展类型检查 |
| `task test:backend` / `task test:extension` | 单独运行 Go / Vitest |
| `task typecheck` | 扩展 TypeScript/Vue 类型检查 |
| `task build` | 构建全部模块 |
| `task build:extension` | 输出 `extension/.output/chrome-mv3` |
| `task --list` | 查看其他任务，包括 Docker 启停与日志 |

Go 测试与源码并置；扩展测试与组件并置。自动化测试使用替身，不需要真实模型、YouTube 下载或 LLM 密钥。

## 常见问题

- **后端无法启动：** 本地安装 `yt-dlp` 和 `ffmpeg` 并加入 `PATH`，或使用 Docker 镜像。
- **转写未开始：** 确认 whisper 容器已通过健康检查（`docker compose ps`）；本地运行时检查 `LSI_STT_BASE_URL` 是否指向运行中的服务。首次启动的模型下载可能较慢。
- **翻译失败：** 检查运行进程的 `LSI_LLM_API_KEY`、`LSI_LLM_MODEL` 和兼容接口地址。
- **扩展连接失败：** 后端 URL 必须是带端口的本机 HTTP 地址；当前 manifest 权限和客户端校验均不支持远程主机。
