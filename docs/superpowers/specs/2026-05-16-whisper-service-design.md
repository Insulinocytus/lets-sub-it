# Whisper 微服务化设计

> 日期: 2026-05-16
> 状态: Draft

## 背景

当前后端在 `transcribing` 阶段通过 `exec whisper-cli` 调用 `whisper/` Python CLI。这个方式让 Go 后端镜像必须同时包含 Python、faster-whisper、Whisper 模型缓存和 CLI 可执行文件，也让转写阶段只能通过进程退出码和输出文件判断结果。

本设计把 Whisper 转写能力拆成独立 HTTP 微服务。后端继续负责 YouTube 音频下载、业务 job 状态、翻译、字幕打包和字幕文件服务；Whisper 服务只负责接收音频、转写、暴露转写任务状态和返回 `source.vtt`。

这份设计取代早期 `2026-04-23-whisper-cli-design.md` 中的 CLI 集成方向。`whisper-cli` 不再作为后端依赖保留。

## 目标

- 移除后端对 `whisper-cli`、Python venv 和 faster-whisper 运行时的直接依赖。
- 把 Whisper 转写能力拆成独立服务，后端通过 HTTP 调用。
- 使用异步转写任务模型，让后端可以通过 HTTP 轮询同步进度文本。
- 后端继续下载音频，Whisper 服务只接收后端上传的音频文件。
- 保持现有外部业务 API 基本不变，extension 仍通过后端查询 job 进度和字幕资产。
- Docker Compose 首版支持同机双服务部署。

## 非目标

- 不在首版引入 WebSocket。
- 不让 Whisper 服务下载 YouTube 视频或理解 YouTube URL。
- 不在首版支持跨机器部署鉴权、限流和多租户隔离。
- 不在首版给 backend `jobs` 表新增百分比字段。
- 不保留后端 CLI fallback。
- 不实现 GPU 专用编排；用户可以后续自行扩展 compose 配置。

## 方案选择

选中方案: **HTTP 上传 + Whisper 异步任务服务 + 后端轮询**。

流程是后端先用现有 `yt-dlp` 下载音频，再通过 `multipart/form-data` 上传给 Whisper 服务。Whisper 服务立即返回 `transcription.id`，后台执行转写。后端定时轮询状态接口，完成后下载 VTT 内容写入现有 `source.vtt`，再继续翻译和打包。

排除的方案:

- 同步 HTTP 转写: 接口少，但长视频容易超时，也无法自然表达转写进度。
- 共享目录路径: 性能更好，但要求 backend 和 whisper 共享文件系统，服务边界不够清晰。
- WebSocket 推送: 实时性更好，但连接管理、重连和错误恢复复杂度超过首版需要。

## 架构边界

### backend 服务

后端继续拥有业务流程和持久化状态:

- 接收 extension 和用户提交的 job。
- 查询和复用已有 job。
- 下载 YouTube 音频到 job 工作目录。
- 调用 Whisper HTTP API 执行转写。
- 将完成的 VTT 写入 `source.vtt`。
- 调用 LLM 翻译字幕。
- 生成 translated 和 bilingual VTT。
- 在 SQLite 中保存 job 状态和字幕资产路径。

后端不再:

- 检查 `whisper-cli` 是否在 `PATH`。
- `exec` 调用 Python CLI。
- 在 backend 镜像中安装 faster-whisper 或 Whisper Python package。

### whisper 服务

Whisper 服务只拥有转写执行能力:

- 接收后端上传的音频文件。
- 根据请求中的 `model`、`computeType` 和 `language` 调用 faster-whisper。
- 渲染并校验 WebVTT。
- 暴露转写任务状态。
- 在任务完成后返回 VTT 内容。

Whisper 服务不负责:

- YouTube 下载。
- backend job 去重或业务状态。
- 翻译字幕。
- 生成双语字幕。
- 服务外部用户认证。

## Whisper HTTP API

### 错误格式

所有 JSON 错误响应使用统一格式:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "audio is required"
  }
}
```

常用错误码:

| HTTP 状态码 | code | 语义 |
|---|---|---|
| `400` | `invalid_request` | 请求字段缺失或非法 |
| `404` | `not_found` | 转写任务不存在 |
| `409` | `not_ready` | 任务未完成，不能读取 VTT |
| `413` | `payload_too_large` | 音频文件超过服务限制 |
| `500` | `internal_error` | Whisper 服务内部错误 |

### `GET /healthz`

用于 Docker healthcheck 和人工排障。

成功响应 `200 OK`:

```json
{
  "status": "ok"
}
```

### `POST /transcriptions`

创建异步转写任务。

请求类型: `multipart/form-data`

字段:

| 字段 | 必填 | 说明 |
|---|---|---|
| `audio` | 是 | 音频文件，后端首版上传 `audio.mp3` |
| `model` | 是 | Whisper 模型名，例如 `small` |
| `computeType` | 否 | faster-whisper compute type，默认 `default` |
| `language` | 是 | 源语言代码，例如 `en`、`ja`、`zh` |
| `jobId` | 否 | backend job ID，仅用于日志关联 |

成功响应 `202 Accepted`:

```json
{
  "transcription": {
    "id": "tr_01h...",
    "status": "queued",
    "progress": 0,
    "progressText": "等待转写",
    "language": null,
    "durationSeconds": null,
    "segments": null,
    "errorMessage": null,
    "createdAt": "2026-05-16T00:00:00Z",
    "updatedAt": "2026-05-16T00:00:00Z"
  }
}
```

### `GET /transcriptions/{id}`

查询转写任务状态。

任务状态枚举:

- `queued`
- `running`
- `completed`
- `failed`

运行中响应 `200 OK`:

```json
{
  "transcription": {
    "id": "tr_01h...",
    "status": "running",
    "progress": 10,
    "progressText": "正在转写音频",
    "language": null,
    "durationSeconds": null,
    "segments": null,
    "errorMessage": null,
    "createdAt": "2026-05-16T00:00:00Z",
    "updatedAt": "2026-05-16T00:00:10Z"
  }
}
```

完成响应 `200 OK`:

```json
{
  "transcription": {
    "id": "tr_01h...",
    "status": "completed",
    "progress": 100,
    "progressText": "转写完成",
    "language": "en",
    "durationSeconds": 123.4,
    "segments": 42,
    "errorMessage": null,
    "createdAt": "2026-05-16T00:00:00Z",
    "updatedAt": "2026-05-16T00:02:00Z"
  }
}
```

失败响应仍使用 `200 OK`，状态体表达失败:

```json
{
  "transcription": {
    "id": "tr_01h...",
    "status": "failed",
    "progress": 100,
    "progressText": "转写失败",
    "language": null,
    "durationSeconds": null,
    "segments": null,
    "errorMessage": "model download error",
    "createdAt": "2026-05-16T00:00:00Z",
    "updatedAt": "2026-05-16T00:00:30Z"
  }
}
```

### `GET /transcriptions/{id}/vtt`

读取完成任务的 WebVTT 内容。

完成后响应 `200 OK`:

```text
Content-Type: text/vtt; charset=utf-8

WEBVTT

00:00:00.000 --> 00:00:01.000
hello
```

如果任务不存在，返回 `404`。如果任务还未完成或已经失败，返回 `409`。

## 任务执行模型

首版使用 Whisper 服务进程内任务表和本地工作目录，不引入独立数据库或消息队列。

任务数据包括:

- `id`
- `status`
- `progress`
- `progressText`
- `language`
- `durationSeconds`
- `segments`
- `errorMessage`
- `audioPath`
- `vttPath`
- `createdAt`
- `updatedAt`

上传的音频和生成的 VTT 存放在 `LSI_WHISPER_WORK_DIR`，默认 `/data/transcriptions`。

首版并发策略固定为单 worker 串行转写。原因是 faster-whisper 对 CPU、GPU 和内存占用较高，默认并发转写容易让自托管环境不稳定。后续如果需要并行，可以引入 `LSI_WHISPER_MAX_CONCURRENCY`，但不作为首版目标。

服务重启会丢失进程内任务状态。这个行为在首版可接受，因为 backend SQLite 仍是业务 job 的真相来源。后端轮询失败或 VTT 读取失败时会将业务 job 标记为 `failed`，用户可以重新提交。

## 进度语义

Faster Whisper 首版不提供可靠百分比进度。`progress` 字段仅表达粗粒度状态，方便未来扩展:

- `queued`: `0`
- `running`: `10`
- `completed`: `100`
- `failed`: `100`

后端首版只使用 `progressText` 更新现有 `jobs.progress_text`，不新增数据库字段保存 `progress`。如果 extension 之后需要百分比，可以在后续设计中给 backend job 响应和数据库增加百分比字段。

## 后端集成设计

### 配置

后端新增配置:

| 变量 | 默认值 | 说明 |
|---|---|---|
| `LSI_WHISPER_BASE_URL` | `http://127.0.0.1:8081` | Whisper 服务地址；Docker Compose 中设为 `http://whisper:8081` |
| `LSI_WHISPER_TIMEOUT` | `30m` | 单次转写总等待上限 |
| `LSI_WHISPER_POLL_INTERVAL` | `2s` | 轮询转写状态间隔 |

后端继续保留并转发这些配置给 Whisper 服务:

| 变量 | 说明 |
|---|---|
| `LSI_WHISPER_MODEL` | Whisper 模型名 |
| `LSI_WHISPER_COMPUTE_TYPE` | faster-whisper compute type |

### Runner 流程

`RealRunner` 的 `transcribing` 阶段改为:

1. 确保 job 工作目录存在。
2. 设置 job 状态为 `transcribing`，进度文本为 `正在提交音频到 Whisper 服务`。
3. 调用 `POST /transcriptions` 上传 `audio.mp3`。
4. 获取 `transcription.id`。
5. 按 `LSI_WHISPER_POLL_INTERVAL` 轮询 `GET /transcriptions/{id}`。
6. 当状态为 `queued` 或 `running` 时，用 `progressText` 更新 backend job。
7. 当状态为 `completed` 时，请求 `GET /transcriptions/{id}/vtt`。
8. 将返回内容写入 `source.vtt` 并验证非空。
9. 继续现有翻译和打包流程。
10. 当状态为 `failed`、接口返回不可恢复错误、上下文取消或超过 `LSI_WHISPER_TIMEOUT` 时，将 backend job 标记为 `failed`，stage 保持 `transcribing`。

### Whisper client

后端新增 HTTP client 封装，职责包括:

- 构造 multipart 上传请求。
- 解析 Whisper 服务 JSON 响应。
- 验证第三方服务响应形状和关键字段。
- 将 Whisper API 错误转换成 Go error。
- 下载 VTT 文本。

`RealRunner` 只依赖一个转写接口，例如 `Transcriber`，避免业务流程直接耦合 HTTP 细节。

## Docker 和本地开发

### Docker Compose

Compose 拆成两个服务:

- `backend`: Go API、yt-dlp、ffmpeg、SQLite 数据和字幕资产。
- `whisper`: Python HTTP 服务、faster-whisper、模型缓存和转写临时文件。

卷职责:

| volume | 使用方 | 说明 |
|---|---|---|
| `lsi-data` | backend | SQLite、job 工作目录和字幕资产 |
| `lsi-hf-cache` | whisper | Hugging Face / Whisper 模型缓存 |
| `lsi-whisper-data` | whisper | 转写上传文件和生成的 VTT 临时文件 |

`backend` 通过 Docker 内网调用 `http://whisper:8081`。首版不把 whisper 端口映射到宿主机。

### Dockerfile

`backend/Dockerfile` 不再包含 Python whisper 构建阶段，运行镜像只安装后端运行所需的 `yt-dlp` 和 `ffmpeg`。

新增 `whisper/Dockerfile`:

- 基于 `python:3.12-slim-bookworm`。
- 安装 `uv`。
- 安装 `whisper/` 项目依赖。
- 安装 `ffmpeg`。
- 设置 `HF_HOME=/huggingface`。
- 设置 `LSI_WHISPER_WORK_DIR=/data/transcriptions`。
- 暴露 `8081`。
- 通过 Uvicorn 启动 HTTP 服务。

### Taskfile

更新 root `Taskfile.yml`:

- `dev:whisper`: 启动 Whisper HTTP 服务。
- `dev:backend`: 不再把 `whisper/.venv/bin` 放到 `PATH`。
- `deps:whisper`: 安装服务依赖。
- `test:whisper`: 运行服务 API 和核心转写测试。
- `build:whisper`: 构建 Python package，或至少校验服务包可安装。

## 迁移计划

1. 在 `whisper/` 中新增 HTTP 服务入口和测试。
2. 删除 `whisper-cli` console script 和 CLI 参数测试。
3. 保留并复用 `transcribe.py` 与 `vtt.py` 的核心逻辑。
4. 在后端新增 Whisper HTTP client 和 transcriber 接口。
5. 修改 `RealRunner`，将 CLI 调用替换为 HTTP 上传、轮询和 VTT 下载。
6. 更新后端配置、启动检查和测试。
7. 拆分 Dockerfile 和 Docker Compose。
8. 更新 `.env.example`、`Taskfile.yml` 和相关文档。

## 测试策略

### Whisper Python 测试

- `POST /transcriptions` 缺少音频、模型或语言时返回 `400`。
- 成功创建任务时返回 `202` 和稳定响应结构。
- 任务从 `queued` 进入 `running`，完成后进入 `completed`。
- 转写失败时任务进入 `failed`，并返回 `errorMessage`。
- 未完成任务请求 VTT 返回 `409`。
- 完成任务请求 VTT 返回 `text/vtt`。
- `render_vtt` 的现有格式校验继续覆盖。
- `transcribe_audio` 的模型和语言校验继续覆盖。

### Backend Go 测试

- `checkTools()` 不再要求 `whisper-cli`。
- 成功链路中，后端向模拟 Whisper 服务上传音频、轮询完成、下载 VTT，并继续翻译和打包。
- Whisper 服务返回 `failed` 时，backend job 失败在 `transcribing` 阶段。
- Whisper 服务返回非法 JSON 或缺少关键字段时，backend job 失败在 `transcribing` 阶段。
- Whisper VTT 为空时，backend job 失败在 `transcribing` 阶段。
- 轮询超时或 context cancellation 时，backend job 失败在 `transcribing` 阶段。

### 验证命令

实现阶段至少运行:

```bash
task test:whisper
task test:backend
```

Docker 变更完成后运行:

```bash
task docker:build
```

## 风险和取舍

- HTTP 上传会复制一次音频文件。这个成本换来 backend 和 whisper 文件系统解耦，符合首版服务边界。
- 进程内任务表在服务重启时会丢失。首版接受这个限制，因为 backend 可以将业务 job 标记为失败并允许用户重试。
- 不保留 CLI fallback 会让迁移更彻底，但也要求 Docker 和本地开发命令同步更新。
- 轮询不是实时推送，但实现简单，足以支撑现有 extension 的 job 查询模型。
- 不新增百分比字段避免数据库迁移，进度先通过现有 `progressText` 表达。

## 成功标准

- 后端启动和运行不再依赖 `whisper-cli`。
- Docker Compose 启动后包含独立 `backend` 和 `whisper` 服务。
- 提交字幕 job 后，backend 能下载音频、调用 Whisper 服务、生成 `source.vtt`、翻译并打包字幕资产。
- extension 通过现有 backend job API 仍能看到转写、翻译和完成状态。
- `task test:backend` 和 `task test:whisper` 通过。
