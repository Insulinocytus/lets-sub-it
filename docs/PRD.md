# YouTube 字幕翻译 PRD

## 1. 项目概述

### 要解决的问题

很多 YouTube 公开视频没有可用字幕。用户听不懂原语言时，内容基本看不下去。现有方案通常会卡在几个地方：部署麻烦、排障困难，或者很难直接在播放页里用起来。

### 方案

做一个开源、自托管的工具。用户通过 Chrome extension 提交 YouTube 公开视频链接，系统自动完成下载、转写、翻译和字幕打包，再把目标语言字幕加载到对应的播放页。

### 成功标准

- 用户提交单条 YouTube URL 后，系统可以稳定走完创建任务、下载、转写、翻译、打包、播放页加载这条链路。
- Extension 可以展示 `queued`、`downloading`、`transcribing`、`translating`、`packaging`、`completed`、`failed` 七个阶段。
- 同一个 `videoId + targetLanguage` 会复用已有完成结果，或者复用已有进行中任务。
- 用户重新打开同一个 YouTube 视频页面后，extension 可以自动恢复上次使用的目标语言和字幕模式。
- 任务完成后稳定产出 `translated.vtt` 和 `bilingual.vtt`，并能在播放页切换使用。

## 2. 用户与使用体验

### 目标用户

- 单用户自托管使用者：愿意自己配置本地服务、OpenAI-compatible API 和模型参数，希望整个字幕生成过程可控、可调试。
- 跨语言视频消费者：面对没有字幕的公开视频，希望尽快拿到能看懂的目标语言字幕。

### 用户故事

1. 作为一个看不懂视频原语言的用户，我希望提交 YouTube 公开视频链接并选择源语言和目标语言，这样我就能拿到可读字幕。
2. 作为一个等待任务完成的用户，我希望在 extension 里看到清晰的阶段状态，这样我知道系统正在处理什么。
3. 作为一个回到播放页继续观看的用户，我希望字幕能自动发现，并恢复到上次使用的模式，这样我可以直接接着看。
4. 作为一个自托管部署者，我希望系统只依赖单个 Go 服务、本地 CLI 工具和本地磁盘，这样部署和排障都更直接。

### 验收标准

#### 用户故事 1

- popup 支持输入单条 YouTube 公开视频 URL（`youtube.com/watch?v=` 格式）。
- popup 支持选择源语言（`en`）和目标语言（`zh`），且两者不能相同。
- popup 支持配置 backend URL，默认值 `http://127.0.0.1:8080`，只允许 `http://localhost:<port>` 或 `http://127.0.0.1:<port>`。
- `create job` 会基于 `videoId + targetLanguage` 复用已有结果或已有任务，返回 `reused: true/false` 标识。
- 任务完成后会生成 `source.vtt`、`translated.vtt`、`bilingual.vtt` 三个字幕产物。

#### 用户故事 2

- extension 可以轮询任务状态，并展示当前阶段的中文标签。
- MockRunner 的 `translating` 阶段会展示 `done/total segments` 形式的进度文本。
- `transcribing` 阶段展示"处理中"（RealRunner）或模拟进度（MockRunner）。
- 失败任务会展示失败阶段和 `errorMessage`。

#### 用户故事 3

- content script 可以识别当前 YouTube 页面（`youtube.com/watch?v=`）的 `videoId`。
- extension 本地缓存以 `backendBaseUrl + videoId + targetLanguage` 为键保存结果索引和上次模式。
- 播放页可以切换字幕开关，以及 `translated` 和 `bilingual` 两种模式。
- 字幕切换失败时自动回退到上一个有效模式。

#### 用户故事 4

- 服务端部署形态是单个 Go 服务。
- 长任务通过 Go 服务内部的 goroutine 调度执行。
- 转写能力通过本地 `whisper-cli`（`faster-whisper` CLI 包装）提供，Go 服务通过 `exec` 调用。
- 元数据使用 SQLite 持久化，字幕文件和中间产物保存在本地磁盘。

### 这次不做

- 多用户系统、登录鉴权、计费
- 批量任务、播放中实时生成
- 私有视频和受限视频支持
- 多 provider 深度抽象
- 字幕编辑、字幕分享、移动端支持
- `youtu.be` 短链接支持（backend 已实现，extension popup 暂不支持）

## 3. 外部能力与验证方式

### 依赖工具

- `yt-dlp`
  - 拉取视频元数据
  - 下载音频并转码为 MP3（内部调用 `ffmpeg`）
  - 提供下载完成后的最终文件路径
- `whisper-cli`（本仓库 `whisper/` 模块）
  - 接收本地音频路径和输出路径
  - 调用 `faster-whisper` 生成 `source.vtt`
  - 成功时在 stdout 输出 JSON 摘要，失败时通过退出码指示阶段
  - 退出码契约：`0` 成功，`2` 输入校验失败，`3` 转写失败，`4` 输出校验失败
- OpenAI-compatible LLM API
  - 接收逐段字幕文本（含上下文窗口）
  - 返回逐段翻译结果
  - 保持 segment 一一对应，不改动时间轴
- `ffmpeg`
  - 由 `yt-dlp` 内部调用来抽取和转码音频
  - Go 服务不直接调用 `ffmpeg`，但要求它在 `PATH` 上

### 验证方式

- 链路验证
  - 选一个固定的公开视频做端到端基准样例。
  - 验证从 URL 提交到播放页加载字幕的完整路径。
- 转写结果验证
  - `source.vtt` 可以被解析，cue 数量大于 0。
  - 时间轴单调递增，字幕文件格式合法。
- 翻译结果验证
  - `translated.vtt` 的 cue 数量与 `source.vtt` 一致。
  - `bilingual.vtt` 的 cue 数量与 `source.vtt` 一致。
  - 每个 cue 保持原始起止时间不变。
- 播放页验证
  - 页面刷新后可以恢复字幕状态。
  - `translated` 和 `bilingual` 两种模式都能切换并正常显示。

## 4. 技术方案

### 整体架构

系统分成三个运行模块：

- `extension` — Chrome MV3 extension（WXT + Vue + TypeScript）
- `Go api server with embedded runner` — HTTP API、SQLite 持久化、后台任务执行
- `whisper-cli` — 本地 `faster-whisper` CLI 包装（Python）

数据流：

1. 用户在 extension popup 中提交 `youtubeUrl`、`sourceLanguage` 和 `targetLanguage`。
2. extension 通过 background service worker 调用 Go 服务的 `POST /jobs` 接口。
3. Go 服务读取 SQLite，基于 `videoId + targetLanguage` 复用已有结果或已有进行中任务。
4. 需要新任务时，Go 服务创建 job 记录，并由 embedded runner 在后台 goroutine 中执行。
5. runner 依次调用 `yt-dlp`（下载音频）、`whisper-cli`（生成 `source.vtt`）、Chat Completions 兼容 LLM（逐段翻译）和 VTT 打包模块。
6. 各阶段状态和产物路径持续写回 SQLite 和工作目录。
7. extension 轮询 `GET /jobs/:id`，任务完成后通过 `GET /subtitle-assets` 缓存结果索引。
8. content script 在 YouTube 播放页识别 `videoId`，结合用户上次选择的目标语言找到对应字幕资产。
9. 页面里的自定义字幕层渲染 `translated` 或 `bilingual` 字幕。

### 数据模型

#### Job

```text
Job {
  id              string     // 主键
  videoId         string     // YouTube 视频 ID（索引）
  youtubeUrl      string     // 原始提交 URL
  sourceLanguage  string     // 源语言（如 "en"）
  targetLanguage  string     // 目标语言（如 "zh"）（索引）
  status          string     // 当前状态（索引）
  stage           string     // 当前阶段
  progressText    string     // 阶段进度文本
  errorMessage    *string    // 失败时的错误摘要（可空）
  attempt         int        // 预留重试字段（当前未使用）
  workingDir      string     // 本地工作目录
  createdAt       time.Time
  updatedAt       time.Time  // （索引）
}
```

任务复用键：`videoId + targetLanguage`。

#### SubtitleAsset

```text
SubtitleAsset {
  jobId               string     // 主键（关联 Job）
  videoId             string     // （索引）
  targetLanguage      string     // （索引）
  sourceLanguage      string
  sourceVttPath       string     // 本地文件路径
  translatedVttPath   string
  bilingualVttPath    string
  createdAt           time.Time
}
```

前端通过 `GET /subtitle-files/:jobId/:mode` 获取 VTT 内容，不直接使用本地路径。

#### LocalCacheEntry

本地缓存存储在 extension storage 中：

```text
SubtitleAssetCacheEntry {
  backendBaseUrl    string     // backend 地址（缓存键组成部分）
  videoId           string
  targetLanguage    string
  jobId             string
  sourceLanguage    string
  files             { source, translated, bilingual 相对 URL }
  createdAt         string
}

VideoPreference {
  backendBaseUrl    string
  videoId           string
  targetLanguage    string
  selectedMode      string     // "translated" 或 "bilingual"
  lastSyncedAt      string
}
```

缓存键包含 `backendBaseUrl`，支持多 backend 场景。

### 状态定义

- `queued`: job 已创建，等待 embedded runner 执行
- `downloading`: `yt-dlp` 音频下载进行中
- `transcribing`: `whisper-cli` 运行中
- `translating`: 逐段翻译进行中
- `packaging`: 正在生成最终字幕产物
- `completed`: 字幕产物已落盘并通过校验
- `failed`: 当前任务以错误结束，同时记录失败阶段和错误信息

### 各阶段完成条件

- 下载阶段完成
  - 下载子进程退出码成功
  - 工作目录中存在目标媒体文件
  - 服务端可以读取最终文件路径
- 转写阶段完成
  - `whisper-cli` 子进程退出码为 `0`
  - `source.vtt` 文件存在
  - `source.vtt` 可解析且 cue 数量大于 0
- 翻译阶段完成
  - 所有 source cues 都产出对应目标语言文本
  - 每个 cue 的翻译请求携带最多前后各 10 条上下文 cue
  - LLM 返回 JSON 格式 `{"translation": "..."}`，cue 数量与 source 一致
- 打包阶段完成
  - `translated.vtt` 和 `bilingual.vtt` 文件存在
  - 两个文件都可解析，且 cue 数量与 `source.vtt` 一致

### API 接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/jobs` | 创建或复用字幕生成 job |
| `GET` | `/jobs/:id` | 查询 job 状态和进度 |
| `GET` | `/subtitle-assets?videoId=...&targetLanguage=...` | 查询已完成字幕资产 |
| `GET` | `/subtitle-files/:jobId/:mode` | 读取 VTT 文件；`mode` 为 `source`、`translated` 或 `bilingual` |

创建 job 请求体：

```json
{
  "youtubeUrl": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
  "sourceLanguage": "en",
  "targetLanguage": "zh"
}
```

创建 job 响应体：

```json
{
  "job": { "id": "...", "status": "queued", "stage": "queued", "progressText": "", ... },
  "reused": true
}
```

### 集成点

- Extension -> Go API
  - popup 通过 background service worker 发送 `job:create` 消息
  - background service worker 调用 `POST /jobs`、`GET /jobs/:id`、`GET /subtitle-assets`、`GET /subtitle-files/:jobId/:mode`
  - content script 通过 background 获取字幕内容，不直接访问 Go API
- Go API -> SQLite
  - 存储 job 状态、字幕资产索引、错误信息、工作目录位置
- Embedded runner -> Local CLI tools
  - 通过 `exec` 调用 `yt-dlp`（含音频抽取）
  - 通过 `exec` 调用 `whisper-cli`
  - 通过 HTTP 调用 Chat Completions 兼容 LLM
- 文件服务
  - `GET /subtitle-files/:jobId/:mode` 从工作目录读取 VTT 文件
  - 文件服务限制在 job 工作目录内，防止路径穿越和符号链接逃逸

### 安全与隐私

- 只处理 YouTube 公开视频。
- API key 保存在服务端配置（`LSI_LLM_API_KEY`），extension 不直接持有翻译 provider 密钥。
- 服务端本地磁盘保存中间音频、转写文件和最终字幕文件，部署者负责主机访问控制。
- 日志里避免输出完整 API key 和敏感请求头。
- Extension manifest host permissions 只允许 `http://127.0.0.1:*/*` 和 `http://localhost:*/*`。
- CORS 只允许带显式端口的 localhost HTTP origin。

### Backend 配置

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `LSI_ADDR` | `127.0.0.1:8080` | HTTP 监听地址 |
| `LSI_DB_PATH` | `./data/backend.sqlite3` | SQLite 数据库路径 |
| `LSI_WORK_DIR` | `./data/jobs` | job 工作目录根路径 |
| `LSI_RUNNER_MODE` | `mock` | runner 模式：`mock` 或 `real` |
| `LSI_DOWNLOAD_TIMEOUT` | `10m` | `real` 模式下单次下载超时 |
| `LSI_WHISPER_MODEL` | `small` | `real` 模式下传给 `whisper-cli --model` 的模型名 |
| `LSI_LLM_BASE_URL` | `https://api.openai.com` | OpenAI-compatible API origin |
| `LSI_LLM_API_KEY` | 空 | OpenAI 默认 endpoint 必填；Bearer token，仅 backend 读取 |
| `LSI_LLM_MODEL` | 空 | `real` 模式下翻译必填的模型名 |
| `LSI_LLM_TIMEOUT` | `2m` | 单条 cue 翻译请求超时 |

### Whisper CLI 契约

`whisper-cli` 接收本地音频文件，输出合法 WebVTT。

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `--input` | 是 | 本地音频文件路径 |
| `--output` | 是 | 输出 `.vtt` 路径，不能与输入路径相同 |
| `--model` | 是 | `faster-whisper` 模型名，如 `small` |
| `--language` | 是 | 语言代码，如 `ja`、`en` |

退出码：`0` 成功，`2` 输入校验失败，`3` 转写失败，`4` 输出校验失败。

成功时 stdout 输出 JSON：

```json
{
  "output": "/tmp/source.vtt",
  "language": "ja",
  "duration_seconds": 123.45,
  "segments": 42
}
```

### Extension 契约

- 技术栈：WXT + Vue + TypeScript + Vite + Vitest + shadcn-vue，npm 包管理。
- 入口文件：`background.ts`、`youtube.content.ts`、`popup/`。
- Background service worker 是唯一 HTTP API 网关；popup 和 content script 通过 `runtime.sendMessage` 通信。
- 支持语言：`en`（源语言）和 `zh`（目标语言），`sourceLanguage` ≠ `targetLanguage`。
- 播放页字幕模式：`translated` 和 `bilingual`（backend 仍保留 `source` 文件服务）。
- Backend URL 必须是带端口的 localhost HTTP origin。
- 消息协议：`settings:get`、`settings:update`、`job:create`、`job:get`、`subtitle:resolve`、`subtitle:fetch-file`、`subtitle:update-mode`。

## 5. 风险与迭代

### 迭代节奏

- MVP（当前）
  - 单个 Go 服务承载 API 和后台执行
  - 单条 URL 提交
  - 源语言和目标语言选择（`en` → `zh`）
  - `translated` / `bilingual` 两种字幕模式
  - 本地缓存 + 服务端持久化
  - MockRunner 完整模拟状态推进；RealRunner 支持真实下载、转写、翻译和打包
- v1.1
  - RealRunner 翻译阶段逐段进度更新（当前只有 MockRunner 展示 `done/total segments`）
  - 服务重启后恢复进行中任务（当前重启会丢失非终态任务）
  - 更可读的错误提示和日志
  - 更稳的页面内快捷切换控件
- v2.0
  - 更多源语言和目标语言支持
  - 更强的翻译质量策略（批量翻译、术语一致性）
  - 任务历史管理和结果复用 UI
  - 更完整的排障和配置工具
  - `youtu.be` 短链接支持（extension 端）

### 技术风险

- YouTube 页面结构变化会影响 content script 和字幕层挂载。
- 长视频会拉长下载、转写、翻译时间，也会放大失败恢复成本。
- `whisper-cli` 的运行时依赖（`faster-whisper`、模型文件）会影响部署体验。
- 翻译服务一旦破坏 segment 对齐，字幕时间轴也会跟着出问题。
- 外部命令执行失败时，状态推进、错误展示和重试策略要保持一致。

## 6. 任务拆分

### 子任务 1：extension — 已完成

目标：提供任务入口、进度展示、本地缓存和播放页字幕体验。

已完成范围：

- popup URL 输入（`youtube.com/watch?v=` 格式）和 backend URL 配置
- 源语言和目标语言选择
- job 创建（含 `sourceLanguage`）和状态轮询
- 本地缓存 `backendBaseUrl + videoId + targetLanguage`
- content script 页面识别和 SPA 导航监听
- 播放页自定义字幕层（Shadow DOM）
- 字幕开关和 `translated` / `bilingual` 模式切换
- 模式切换失败自动回退
- background service worker 作为唯一 HTTP 网关
- 服务 worker 重启后恢复任务监控

### 子任务 2：api — 已完成

目标：提供 job 生命周期、字幕资产索引和文件访问接口。

已完成范围：

- SQLite schema（Job + SubtitleAsset）
- `POST /jobs` 创建或复用 job，返回 `reused` 标识
- `GET /jobs/:id` 查询状态和进度
- `GET /subtitle-assets` 查询已完成字幕资产
- `GET /subtitle-files/:jobId/:mode` 安全提供 VTT 文件
- 任务复用（`videoId + targetLanguage` 去重）
- CORS：只允许带显式端口的 localhost HTTP origin

### 子任务 3：embedded runner — 已完成（MVP）

目标：在单个 Go 服务内部执行长任务，并推进状态机。

已完成范围：

- 后台 goroutine 调度
- `queued` → `completed` 状态推进
- 工作目录管理
- 命令执行和结果回写
- MockRunner（模拟全阶段，含翻译进度文本）
- RealRunner（真实下载、转写、逐段翻译含上下文窗口、VTT 打包）
- 阶段级失败处理和 `errorMessage` 记录

MVP 未完成：

- 服务重启后恢复进行中任务（重启会丢失非终态 job 状态）
- `attempt` 字段预留但当前未使用（无重试逻辑）
- RealRunner 翻译阶段未实时更新 `done/total segments` 进度

### 子任务 4：whisper-cli — 已完成

目标：提供一个可由 embedded runner 直接 `exec` 调用的本地转写 CLI。

已完成范围：

- 输入音频路径、输出路径、模型名和语言代码
- 调用 `faster-whisper` 生成 `source.vtt`
- stdout 输出 JSON 摘要（`output`、`language`、`duration_seconds`、`segments`）
- 统一退出码：`0` 成功、`2` 输入校验失败、`3` 转写失败、`4` 输出校验失败
- WebVTT 校验：非空 cue、正时间戳、单调递增起始时间、文本无非法内容
- 单元测试使用 fake model，不下载模型或依赖 GPU