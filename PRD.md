# YouTube 字幕翻译 MVP PRD

## 1. 项目概述

### 要解决的问题

很多 YouTube 公开视频没有可用字幕。用户听不懂原语言时，内容基本看不下去。现有方案通常会卡在几个地方：部署麻烦、排障困难，或者很难直接在播放页里用起来。

### 方案

做一个开源、自托管的 MVP。用户通过 Chrome extension 提交 YouTube 公开视频链接，系统自动完成下载、转写、翻译和字幕打包，再把目标语言字幕加载到对应的播放页。

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

1. 作为一个看不懂视频原语言的用户，我希望提交 YouTube 公开视频链接并选择目标语言，这样我就能拿到可读字幕。
2. 作为一个等待任务完成的用户，我希望在 extension 里看到清晰的阶段状态，这样我知道系统正在处理什么。
3. 作为一个回到播放页继续观看的用户，我希望字幕能自动发现，并恢复到上次使用的模式，这样我可以直接接着看。
4. 作为一个自托管部署者，我希望系统只依赖单个 Go 服务、本地 CLI 工具和本地磁盘，这样部署和排障都更直接。

### 验收标准

#### 用户故事 1

- popup 支持输入单条 YouTube 公开视频 URL。
- popup 支持为每个 job 选择一个目标语言。
- `create job` 会基于 `videoId + targetLanguage` 复用已有结果或已有任务。
- 任务完成后会生成 `source.vtt`、`translated.vtt`、`bilingual.vtt` 三个字幕产物。

#### 用户故事 2

- extension 可以轮询任务状态，并展示当前阶段。
- `translating` 阶段可以展示 `done/total segments` 形式的进度。
- `transcribing` 阶段至少展示“处理中”和已耗时。
- 失败任务会展示失败阶段和错误信息。

#### 用户故事 3

- content script 可以识别当前 YouTube 页面的 `videoId`。
- extension 本地缓存以 `videoId + targetLanguage` 为键保存结果索引和上次模式。
- 播放页可以切换字幕开关，以及 `translated` 和 `bilingual` 两种模式。

#### 用户故事 4

- 服务端部署形态是单个 Go 服务。
- 长任务通过 Go 服务内部的 goroutine 调度执行。
- 转写能力通过本地 `fast-whisper` CLI 提供，Go 服务通过 `exec` 调用。
- 元数据使用 SQLite 持久化，字幕文件和中间产物保存在本地磁盘。

### 这次不做

- 多用户系统、登录鉴权、计费
- 批量任务、播放中实时生成
- 私有视频和受限视频支持
- 多 provider 深度抽象
- 字幕编辑、字幕分享、移动端支持

## 3. 外部能力与验证方式

### 依赖工具

- `yt-dlp`
  - 拉取视频元数据
  - 下载音频或视频媒体
  - 提供下载完成后的最终文件路径
- `fast-whisper` CLI
  - 接收本地音频路径
  - 输出 `source.vtt`
  - 通过进程存活状态、退出码和产物文件，让 runner 判断任务是否完成
- OpenAI-compatible LLM API
  - 接收逐段字幕文本
  - 返回逐段翻译结果
  - 保持 segment 一一对应，不改动时间轴
- `ffmpeg`
  - 在需要时抽取音频
  - 统一音频输入格式，减少转写波动

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

系统分成三个运行模块，执行链路里再拆成四类任务。

- 运行模块
  - `extension`
  - `Go api server with embedded runner`
  - `fast-whisper` local CLI
- 执行子任务
  - `extension`
  - `api`
  - `embedded runner`
  - `fast-whisper cli`

数据流如下：

1. 用户在 extension popup 中提交 `youtubeUrl` 和 `targetLanguage`。
2. extension 解析 `videoId`，调用 Go 服务的 `create job` 接口。
3. Go 服务读取 SQLite，基于 `videoId + targetLanguage` 复用已有结果或已有进行中任务。
4. 需要新任务时，Go 服务创建 job 记录，并由 embedded runner 在后台 goroutine 中执行。
5. runner 依次调用 `yt-dlp`、`ffmpeg`、`fast-whisper` CLI、翻译模块和 VTT 打包模块。
6. 各阶段状态和产物路径持续写回 SQLite 和工作目录。
7. extension 轮询 job 状态，任务完成后缓存结果索引。
8. content script 在 YouTube 播放页识别 `videoId`，结合用户上次选择的目标语言找到对应字幕资产。
9. 页面里的自定义字幕层渲染 `translated` 或 `bilingual` 字幕。

### 数据模型

#### Job

```text
Job {
  id
  videoId
  youtubeUrl
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

#### SubtitleAsset

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

#### LocalCacheEntry

```text
LocalCacheEntry {
  videoId
  targetLanguage
  jobId
  selectedMode
  lastSyncedAt
}
```

### 状态定义

- `queued`: job 已创建，等待 embedded runner 执行
- `downloading`: `yt-dlp` 或媒体准备过程进行中
- `transcribing`: `fast-whisper` CLI 运行中
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
  - `fast-whisper` CLI 子进程退出码成功
  - `source.vtt` 文件存在
  - `source.vtt` 可解析且 cue 数量大于 0
- 翻译阶段完成
  - 所有 source segments 都产出对应目标语言文本
  - segment 数量与 source 一致
- 打包阶段完成
  - `translated.vtt` 和 `bilingual.vtt` 文件存在
  - 两个文件都可解析，且 cue 数量与 `source.vtt` 一致

### 集成点

- Extension -> Go API
  - `POST /jobs`
  - `GET /jobs/:id`
  - `GET /subtitle-assets?videoId=...&targetLanguage=...`
  - `GET /subtitle-files/:jobId/:mode`
- Go API -> SQLite
  - 存储 job 状态、字幕资产索引、错误信息、工作目录位置
- Embedded runner -> Local CLI tools
  - 通过 `exec` 调用 `yt-dlp`
  - 通过 `exec` 调用 `ffmpeg`
  - 通过 `exec` 调用 `fast-whisper` CLI
- Go API -> OpenAI-compatible API
  - 逐段发送文本
  - 接收逐段翻译结果

### 安全与隐私

- 只处理 YouTube 公开视频。
- API key 保存在服务端配置里，extension 不直接持有翻译 provider 密钥。
- 服务端本地磁盘保存中间音频、转写文件和最终字幕文件，部署者负责主机访问控制。
- 日志里避免输出完整 API key 和敏感请求头。

## 5. 风险与迭代

### 迭代节奏

- MVP
  - 单个 Go 服务承载 API 和后台执行
  - 单条 URL 提交
  - 单个目标语言
  - `translated` / `bilingual` 两种模式
  - 本地缓存 + 服务端持久化
- v1.1
  - 更好的任务历史和结果管理
  - 更可读的错误提示和日志
  - 更稳的页面内快捷切换控件
- v2.0
  - 更强的翻译质量策略
  - 更多字幕显示细节控制
  - 更完整的排障和配置工具

### 技术风险

- YouTube 页面结构变化会影响 content script 和字幕层挂载。
- 长视频会拉长下载、转写、翻译时间，也会放大失败恢复成本。
- `fast-whisper` CLI 的运行时依赖和模型准备会影响部署体验。
- 翻译服务一旦破坏 segment 对齐，字幕时间轴也会跟着出问题。
- 外部命令执行失败时，状态推进、错误展示和重试策略要保持一致。

## 6. 任务拆分

### 子任务 1：extension

目标：提供任务入口、进度展示、本地缓存和播放页字幕体验。

范围：

- popup URL 输入
- 目标语言选择
- job 创建和状态轮询
- 本地缓存 `videoId + targetLanguage`
- content script 页面识别
- 播放页自定义字幕层
- 字幕开关和模式切换

完成标准：

- 可以提交 job，并稳定展示阶段状态
- 可以在播放页自动发现字幕并显示
- 可以恢复上次目标语言和字幕模式

### 子任务 2：api

目标：提供 job 生命周期、字幕资产索引和文件访问接口。

范围：

- SQLite schema
- job create/query API
- subtitle asset query API
- subtitle file serving
- 配置加载
- 错误模型

完成标准：

- API 可以创建或复用 job
- API 可以返回准确的 job 状态和字幕资产信息
- API 可以稳定提供最终字幕文件访问

### 子任务 3：embedded runner

目标：在单个 Go 服务内部执行长任务，并推进状态机。

范围：

- 后台 goroutine 调度
- `queued` -> `completed` 状态推进
- 工作目录管理
- 命令执行和结果回写
- 启动恢复逻辑
- 阶段级失败处理

完成标准：

- 新建 job 后可以自动进入后台执行
- 重启服务后可以根据工作目录和产物状态恢复任务
- 失败阶段和错误信息可以稳定落库

### 子任务 4：fast-whisper cli

目标：提供一个可由 embedded runner 直接 `exec` 调用的本地转写 CLI。

范围：

- 输入音频路径和输出路径
- 调用 `fast-whisper`
- 输出 `source.vtt`
- 统一退出码
- 支持 runner 做完成判定

完成标准：

- CLI 可以在本地稳定生成 `source.vtt`
- 成功时退出码和产物一致
- 失败时返回明确错误，并让 runner 能判定阶段失败
