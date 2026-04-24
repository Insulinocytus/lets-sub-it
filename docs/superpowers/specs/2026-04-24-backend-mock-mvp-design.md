# Backend Mock MVP 设计

## 背景

本文定义 `backend/` 第一阶段的 MVP 设计。当前项目已经完成 `whisper/` 下的本地转写 CLI，下一步需要让前端可以并行开发和联调，因此 backend 第一阶段优先稳定 API、SQLite、任务状态机和字幕文件访问契约。

本阶段采用“真实 API + 真实持久化 + mock runner”的方式：服务端对前端暴露的接口和数据结构尽量按最终形态设计，但下载、音频处理、转写和翻译等外部能力暂时用 mock 实现代替。

## 目标

- 提供一个可以本地启动的 Go HTTP 服务。
- 提供稳定的 job 创建、查询、复用和字幕资产查询接口。
- 使用 SQLite 持久化 job 状态和 subtitle asset 索引。
- 使用本地工作目录保存 `.vtt` 字幕文件，并通过 HTTP 提供文件访问。
- 用 mock runner 模拟完整阶段流转，让 extension 可以先基于 backend 联调。
- 明确真实实现和 mock 实现的边界，避免前后端对当前能力产生误解。

## 当前阶段真实实现范围

以下能力在本阶段需要真实实现，并作为后续接入真实 runner 的稳定基础：

- Go module 与 backend 项目结构。
- Go module 名称固定为 `lets-sub-it-api`。
- HTTP server 启动和路由注册。
- JSON 请求解析和响应输出。
- YouTube URL 中 `videoId` 的基础解析与校验。
- SQLite schema 初始化。
- `jobs` 表的创建、查询、更新。
- `subtitle_assets` 表的创建、查询、写入。
- 基于 `videoId + targetLanguage` 的复用逻辑：
  - 已有 `completed` job 时直接复用完成结果。
  - 已有非终态 job 时复用进行中任务。
  - 已有 `failed` job 不复用，创建新 job。
- 后台 goroutine 调度 mock runner。
- job 状态推进和错误信息落库。
- 每个 job 的本地工作目录管理。
- mock `.vtt` 文件写入磁盘。
- `source`、`translated`、`bilingual` 三种字幕文件的 HTTP 访问。
- 面向前端的稳定错误响应格式。
- 允许 `localhost` 和 `127.0.0.1` 任意端口访问的本地开发 CORS。
- 针对 handler、store、runner 状态机和文件服务的 Go 测试。

## 当前阶段 mock 实现范围

以下能力在本阶段不接真实外部工具，只提供可预测的 mock 行为：

- `yt-dlp` 下载阶段：不访问 YouTube，不下载媒体，只把 job 推进到 `downloading` 并等待一个很短的模拟时间。
- `ffmpeg` 音频处理阶段：不执行音频抽取或转码，只作为下载后的内部模拟步骤。
- `whisper-cli` 转写阶段：不调用 `whisper-cli`，只把 job 推进到 `transcribing`，然后写出合法的 mock `source.vtt`。
- LLM 翻译阶段：不请求 OpenAI-compatible API，只基于 mock source cue 生成 mock 翻译文本。
- 翻译进度：使用固定 mock cue 数量，推进类似 `1/3 segments`、`2/3 segments`、`3/3 segments` 的进度文本。
- 双语字幕打包：第一阶段会真实写出 `bilingual.vtt` 文件，但输入来自 mock source 和 mock translated cue。
- 失败重试：第一阶段不做自动重试，只保留 `failed` 状态和错误字段。
- 启动恢复：第一阶段可以先不恢复中断中的 job；如果服务重启后存在非终态 job，先保持原状态，后续真实 runner 阶段再完善恢复策略。

## 非目标

- 不做真实 YouTube 下载。
- 不做真实 `ffmpeg` 调用。
- 不做真实 `whisper-cli` 调用。
- 不做真实 LLM 翻译调用。
- 不做多用户、鉴权、计费或权限模型。
- 不做批量任务。
- 不做复杂队列、并发限制和分布式 worker。
- 不做字幕编辑、分享或管理后台。

## 项目结构

```text
backend/
  .gitignore
  go.mod
  cmd/
    server/
      main.go
  internal/
    api/
      handler.go
      routes.go
      response.go
      youtube.go
    app/
      app.go
      config.go
    runner/
      mock_runner.go
      runner.go
      vtt.go
    store/
      sqlite.go
      migrations.go
      models.go
  测试文件与被测包放在同目录，命名为 *_test.go
```

第一阶段保持单一 server binary，入口使用 `cmd/server/main.go`，避免在 MVP 阶段引入过深目录。`go.mod` 的 module 名称固定为 `lets-sub-it-api`；未来如果拆出 worker，可以使用独立 module 或 binary 名 `lets-sub-it-worker`。

## 配置

第一阶段配置保持最小，通过环境变量读取：

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `LSI_ADDR` | `127.0.0.1:8080` | HTTP 监听地址 |
| `LSI_DB_PATH` | `./data/backend.sqlite3` | SQLite 数据库路径 |
| `LSI_WORK_DIR` | `./data/jobs` | job 工作目录根路径 |

本阶段不读取 LLM API key、Whisper 模型名、`yt-dlp` 路径或 `ffmpeg` 路径，因为对应能力仍是 mock。

## Go 技术选型

第一阶段优先快速开发和清晰验证：

- SQLite driver 使用 `gorm.io/driver/sqlite`，CGO 不是限制；该驱动默认基于 `github.com/mattn/go-sqlite3`。
- ORM 使用 `gorm.io/gorm`，减少手写 CRUD 和 row scan 模板代码。
- migration 第一阶段使用 GORM `AutoMigrate` 初始化 `jobs` 和 `subtitle_assets`。
- 如果后续 schema 变复杂或需要严格版本化迁移，再引入 `goose` 管理显式 SQL migration。
- job id 使用应用层生成的 UUIDv7 字符串，SQLite 以 `TEXT` 存储，不依赖 SQLite 原生 UUID 类型。

选择 GORM 的原因是当前目标是尽快跑通 mock backend，让前端可以并行联调。`sqlc + goose` 更适合 schema 已稳定、SQL 需要强类型生成和精细控制的阶段，可以作为后续替换或补强方向。

## 数据模型

### Job

```text
Job {
  id
  videoId
  youtubeUrl
  sourceLanguage
  targetLanguage
  status
  stage
  progressText
  errorMessage
  attempt
  workingDir
  createdAt
  updatedAt
}
```

字段说明：

- `id`：服务端生成的 job id，建议格式为 `job_` 加随机字符串。
- `videoId`：从 YouTube URL 解析出的 video id。
- `youtubeUrl`：用户提交的原始 URL。
- `sourceLanguage`：前端传入的源语言；本项目不做视频语言自动识别。
- `targetLanguage`：目标语言，例如 `zh-CN`、`en`、`ja`。
- `status`：前端展示和轮询使用的主状态。
- `stage`：当前执行阶段；第一阶段与 `status` 保持一致。
- `progressText`：给前端直接展示的短文本。
- `errorMessage`：失败时的错误摘要；非失败状态为 `null`。
- `attempt`：当前尝试次数，第一阶段创建时为 `1`。
- `workingDir`：本地 job 工作目录，默认不在 API 响应里暴露。
- `createdAt` / `updatedAt`：RFC3339 UTC 时间字符串。

### SubtitleAsset

```text
SubtitleAsset {
  jobId
  videoId
  targetLanguage
  sourceVttPath
  translatedVttPath
  bilingualVttPath
  sourceLanguage
  createdAt
}
```

字段说明：

- `jobId`：关联的 completed job。
- `videoId`：视频 id。
- `targetLanguage`：目标语言。
- `sourceVttPath` / `translatedVttPath` / `bilingualVttPath`：服务端本地文件路径，不直接暴露给前端。
- `sourceLanguage`：来自创建 job 请求，不由 backend 自动识别。
- `createdAt`：资产创建时间。

## 状态定义

| 状态 | 含义 | 是否终态 |
| --- | --- | --- |
| `queued` | job 已创建，等待后台 runner 执行 | 否 |
| `downloading` | 模拟下载或媒体准备阶段 | 否 |
| `transcribing` | 模拟转写阶段 | 否 |
| `translating` | 模拟逐段翻译阶段 | 否 |
| `packaging` | 写入最终字幕文件和资产索引 | 否 |
| `completed` | 字幕产物已落盘并可访问 | 是 |
| `failed` | 当前任务失败 | 是 |

mock runner 的正常状态顺序固定为：

```text
queued -> downloading -> transcribing -> translating -> packaging -> completed
```

## HTTP API 契约

所有 JSON 响应使用 `application/json; charset=utf-8`。时间字段使用 RFC3339 UTC 字符串。

本地开发 CORS 规则：

- 允许 `Origin` 为 `http://localhost:<任意端口>`。
- 允许 `Origin` 为 `http://127.0.0.1:<任意端口>`。
- 允许方法：`GET`、`POST`、`OPTIONS`。
- 允许请求头：`Content-Type`。
- 其他来源默认不返回 CORS allow header。

### 创建或复用 job

```http
POST /jobs
Content-Type: application/json
```

请求体：

```json
{
  "youtubeUrl": "https://www.youtube.com/watch?v=abc123",
  "sourceLanguage": "ja",
  "targetLanguage": "zh-CN"
}
```

成功响应：`201 Created` 或 `200 OK`。

- 创建新 job 时返回 `201 Created`，`reused` 为 `false`。
- 复用已有 job 时返回 `200 OK`，`reused` 为 `true`。

```json
{
  "job": {
    "id": "job_01hwxyz123",
    "videoId": "abc123",
    "youtubeUrl": "https://www.youtube.com/watch?v=abc123",
    "sourceLanguage": "ja",
    "targetLanguage": "zh-CN",
    "status": "queued",
    "stage": "queued",
    "progressText": "等待处理",
    "errorMessage": null,
    "createdAt": "2026-04-24T03:00:00Z",
    "updatedAt": "2026-04-24T03:00:00Z"
  },
  "reused": false
}
```

校验失败响应：`400 Bad Request`。

```json
{
  "error": {
    "code": "invalid_request",
    "message": "youtubeUrl, sourceLanguage, and targetLanguage are required"
  }
}
```

第一阶段支持的 YouTube URL 形式：

- `https://www.youtube.com/watch?v=abc123`
- `https://youtube.com/watch?v=abc123`
- `https://youtu.be/abc123`

### 查询 job

```http
GET /jobs/{jobId}
```

成功响应：`200 OK`。

```json
{
  "job": {
    "id": "job_01hwxyz123",
    "videoId": "abc123",
    "youtubeUrl": "https://www.youtube.com/watch?v=abc123",
    "sourceLanguage": "ja",
    "targetLanguage": "zh-CN",
    "status": "translating",
    "stage": "translating",
    "progressText": "2/3 segments",
    "errorMessage": null,
    "createdAt": "2026-04-24T03:00:00Z",
    "updatedAt": "2026-04-24T03:00:06Z"
  }
}
```

不存在响应：`404 Not Found`。

```json
{
  "error": {
    "code": "not_found",
    "message": "job not found"
  }
}
```

### 查询字幕资产

```http
GET /subtitle-assets?videoId=abc123&targetLanguage=zh-CN
```

未完成或不存在时返回 `200 OK`，`asset` 为 `null`。

```json
{
  "asset": null
}
```

完成后返回：

```json
{
  "asset": {
    "jobId": "job_01hwxyz123",
    "videoId": "abc123",
    "targetLanguage": "zh-CN",
    "sourceLanguage": "ja",
    "files": {
      "source": "/subtitle-files/job_01hwxyz123/source",
      "translated": "/subtitle-files/job_01hwxyz123/translated",
      "bilingual": "/subtitle-files/job_01hwxyz123/bilingual"
    },
    "createdAt": "2026-04-24T03:00:10Z"
  }
}
```

参数缺失响应：`400 Bad Request`。

```json
{
  "error": {
    "code": "invalid_request",
    "message": "videoId and targetLanguage are required"
  }
}
```

### 获取字幕文件

```http
GET /subtitle-files/{jobId}/{mode}
```

`mode` 支持：

- `source`
- `translated`
- `bilingual`

成功响应：`200 OK`。

响应头：

```http
Content-Type: text/vtt; charset=utf-8
```

响应体示例：

```text
WEBVTT

00:00:00.000 --> 00:00:02.000
这是 mock source 字幕第一行。

00:00:02.000 --> 00:00:04.000
这是 mock source 字幕第二行。
```

不存在响应：`404 Not Found`。

非法 mode 响应：`400 Bad Request`。

```json
{
  "error": {
    "code": "invalid_mode",
    "message": "mode must be source, translated, or bilingual"
  }
}
```

## 前端并行开发说明

前端可以把本阶段 backend 当作可运行 mock server 使用，并先完成以下链路：

1. popup 提交 `youtubeUrl`、`sourceLanguage` 和 `targetLanguage`。
2. 调用 `POST /jobs` 创建或复用 job。
3. 使用 `GET /jobs/{jobId}` 轮询并展示状态。
4. 在 `status` 为 `completed` 后调用 `GET /subtitle-assets`。
5. 缓存 `videoId + targetLanguage` 对应的 `jobId` 和文件 URL。
6. 播放页按用户选择加载 `translated` 或 `bilingual` VTT。
7. 刷新页面后根据本地缓存恢复目标语言和字幕模式。

前端需要注意：

- 当前字幕内容是 mock，不代表最终翻译质量。
- 当前 backend 不自动识别源语言，前端必须在创建 job 时传入 `sourceLanguage`。
- 当前 `progressText` 可以直接展示，但不要解析成强类型业务逻辑。
- 当前 job 完成速度会很快，前端轮询间隔可以先使用 500ms 到 1000ms。
- 当前没有鉴权，默认本机自托管使用。

## Runner 设计

runner 通过接口隔离，便于后续替换为真实实现：

```go
type Runner interface {
    Start(ctx context.Context, job Job) error
}
```

第一阶段实现 `MockRunner`：

1. 更新 job 为 `downloading`，`progressText` 为 `准备 mock 媒体`。
2. 更新 job 为 `transcribing`，`progressText` 为 `生成 mock source.vtt`。
3. 写入 `source.vtt`。
4. 更新 job 为 `translating`，逐步写入 `1/3 segments`、`2/3 segments`、`3/3 segments`。
5. 写入 `translated.vtt`。
6. 更新 job 为 `packaging`，`progressText` 为 `生成字幕资产`。
7. 写入 `bilingual.vtt`。
8. 写入 `subtitle_assets` 记录。
9. 更新 job 为 `completed`，`progressText` 为 `处理完成`。

如果 mock runner 任一步失败，runner 将 job 更新为：

```text
status = failed
stage = 当前失败阶段
errorMessage = 简短错误信息
```

## SQLite schema

第一阶段 schema 建议如下：

```sql
CREATE TABLE IF NOT EXISTS jobs (
  id TEXT PRIMARY KEY,
  video_id TEXT NOT NULL,
  youtube_url TEXT NOT NULL,
  source_language TEXT NOT NULL,
  target_language TEXT NOT NULL,
  status TEXT NOT NULL,
  stage TEXT NOT NULL,
  progress_text TEXT NOT NULL,
  error_message TEXT,
  attempt INTEGER NOT NULL,
  working_dir TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_jobs_lookup
ON jobs(video_id, target_language, status, updated_at);

CREATE TABLE IF NOT EXISTS subtitle_assets (
  job_id TEXT PRIMARY KEY,
  video_id TEXT NOT NULL,
  target_language TEXT NOT NULL,
  source_vtt_path TEXT NOT NULL,
  translated_vtt_path TEXT NOT NULL,
  bilingual_vtt_path TEXT NOT NULL,
  source_language TEXT NOT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY(job_id) REFERENCES jobs(id)
);

CREATE INDEX IF NOT EXISTS idx_subtitle_assets_lookup
ON subtitle_assets(video_id, target_language);
```

## 错误响应模型

所有 JSON 错误使用统一结构：

```json
{
  "error": {
    "code": "invalid_request",
    "message": "human readable message"
  }
}
```

第一阶段错误码：

| code | HTTP 状态 | 含义 |
| --- | --- | --- |
| `invalid_request` | 400 | 请求体或查询参数不合法 |
| `invalid_youtube_url` | 400 | 无法解析支持的 YouTube URL |
| `invalid_mode` | 400 | 字幕文件 mode 不合法 |
| `not_found` | 404 | job 或字幕文件不存在 |
| `internal_error` | 500 | 服务端内部错误 |

## 测试策略

第一阶段优先覆盖可被前端依赖的契约：

- `POST /jobs` 创建新 job。
- `POST /jobs` 复用进行中 job。
- `POST /jobs` 复用 completed job。
- `POST /jobs` 不复用 failed job。
- `GET /jobs/{jobId}` 返回状态。
- `GET /subtitle-assets` 在完成前返回 `asset: null`。
- `GET /subtitle-assets` 在完成后返回三个文件 URL。
- `GET /subtitle-files/{jobId}/{mode}` 返回 `text/vtt`。
- YouTube URL 解析支持 `watch?v=` 和 `youtu.be`。
- mock runner 正常推进到 `completed` 并写入资产。
- mock runner 文件写入失败时进入 `failed`。

## 后续替换路径

本阶段完成后，真实能力按以下顺序接入：

1. 用真实 `yt-dlp` 替换 mock downloading 阶段。
2. 在需要时加入真实 `ffmpeg` 音频处理。
3. 用 `exec` 调用 `whisper-cli` 替换 mock transcribing 阶段。
4. 解析真实 `source.vtt`，接入 OpenAI-compatible LLM 翻译。
5. 用真实 source 和 translated cue 生成最终 bilingual 字幕。
6. 增加启动恢复、失败重试、阶段级日志和更细进度。

替换过程中应保持 HTTP API 响应结构尽量不变，避免阻塞 extension 开发。

## 已确认决策

- server 入口使用 `cmd/server/main.go`。
- Go module 名称使用 `lets-sub-it-api`。
- job id 使用 UUIDv7 字符串，由应用层生成，SQLite 以 `TEXT` 保存。
- 创建 job 时必须由前端传入 `sourceLanguage`，backend 不做视频语言自动识别。
- SQLite 开发栈使用 GORM + `gorm.io/driver/sqlite`，第一阶段用 `AutoMigrate`。
- 本地开发 CORS 允许 `localhost` 和 `127.0.0.1` 的任意端口。
