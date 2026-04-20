# YouTube 字幕翻译产品设计

## 1. 目标与范围

### 目标
构建一个开源、自托管的 MVP 产品，帮助个人用户为 YouTube public 视频生成目标语言字幕，并通过 Chrome extension 挂载到对应播放页中使用。

### MVP 成功标准
用户可以提交一条 public YouTube 视频链接，系统稳定完成以下全链路：

1. 后端接收任务并创建 job
2. 使用 `yt-dlp` 下载视频媒体
3. 使用 `fast-whisper` 生成原文字幕
4. 使用 OpenAI-compatible LLM 服务将字幕翻译为目标语言
5. 生成最终字幕文件
6. extension 展示各阶段进度
7. 用户打开对应 YouTube 视频页后，extension 自动发现并挂载字幕
8. 用户可以切换字幕开关，以及在 `translated` 和 `bilingual` 两种模式之间切换

### 明确范围
MVP 只包含以下能力：

- 支持 YouTube public 视频
- 支持单用户、自托管部署
- 支持手动提交单条 URL
- 支持单一目标语言
- 支持完整离线批处理
- 支持服务端持久化 + extension 本地缓存
- 支持任务进度展示
- 支持播放页自动发现并挂载字幕
- 支持 `translated` 和 `bilingual` 两种显示模式

### 暂不纳入 MVP

- 多用户系统
- 登录鉴权
- 计费
- 批量任务
- 私有或受限可见性视频
- 实时边看边生成
- 多 provider 深度抽象
- 字幕编辑
- 字幕分享
- 移动端支持

## 2. 总体架构

系统拆分为 4 个模块：

### 2.1 WXT extension UI
负责用户交互与本地状态管理。

职责：
- 输入 YouTube URL
- 选择目标语言
- 创建处理任务
- 展示任务进度与结果状态
- 缓存最近任务与已完成字幕元数据
- 提供字幕模式切换与重新加载入口

建议形态：
- MVP 优先使用 popup 作为任务入口
- 后续可扩展 sidepanel 作为任务中心

### 2.2 Content script / page integration
负责在 YouTube 播放页识别目标视频，并将字幕接入播放器体验。

职责：
- 识别当前页面 `videoId`
- 查询 extension 本地缓存
- 未命中时向后端查询已完成结果
- 将字幕资源挂载到当前视频播放器
- 暴露轻量页面级快捷切换入口

### 2.3 Go backend API
负责 job 管理、状态持久化、字幕资产管理与 worker 调度。

职责：
- 创建任务
- 查询任务状态与进度
- 提供字幕资源下载或读取接口
- 管理失败信息与阶段日志
- 调度长任务执行

### 2.4 Worker pipeline
负责耗时处理流程。

职责：
- 下载媒体
- 语音转写
- 字幕翻译
- 字幕打包
- 进度写回
- 错误记录

### 架构边界
- extension 负责交互、展示、缓存、挂载
- backend 负责 API、持久化、任务编排
- worker 负责长任务执行

这个边界适合自托管 MVP，部署简单，排障路径清晰。

## 3. 核心流程

### 3.1 任务创建与处理流程

1. 用户在 extension 中输入 `youtubeUrl`
2. extension 解析 `videoId`
3. 用户选择目标语言
4. extension 调用后端 `create job` 接口
5. 后端创建 `jobId`
6. worker 顺序执行：
   - `yt-dlp` 下载视频或音频
   - `fast-whisper` 生成原文字幕
   - 使用 OpenAI-compatible 接口逐段翻译字幕
   - 生成 `source.vtt`、`translated.vtt`、`bilingual.vtt`
7. backend 持续更新 job 状态和阶段进度
8. extension 轮询任务状态并更新 UI
9. 任务完成后，extension 缓存结果元数据
10. 用户打开对应 YouTube 页面时，content script 自动发现并挂载字幕

### 3.2 播放页自动发现流程

1. content script 识别当前 YouTube 页面 `videoId`
2. 查询 extension 本地缓存是否存在已完成字幕
3. 本地未命中时查询后端
4. 如果存在字幕资产，则加载上次使用的模式
5. 将对应 VTT 资源挂载到播放器
6. 用户可切换字幕开关与显示模式

### 3.3 失败与恢复流程

- 如果 job 尚未完成，播放页显示“处理中”状态
- 如果后端不可达，提示服务不可用
- 如果字幕资产缺失，提示结果不可用
- 如果播放器挂载失败，允许用户点击 `Retry load subtitle`
- 失败信息同时展示在 popup 中，作为排障入口

## 4. 数据模型

### 4.1 Job

```text
Job {
  id
  videoId
  youtubeUrl
  targetLanguage
  status
  stage
  progress
  errorMessage
  createdAt
  updatedAt
}
```

### 4.2 SubtitleAsset

```text
SubtitleAsset {
  jobId
  videoId
  sourceVttPath
  translatedVttPath
  bilingualVttPath
  sourceLanguage
  targetLanguage
}
```

### 4.3 LocalCacheEntry

```text
LocalCacheEntry {
  videoId
  jobId
  selectedMode
  lastSyncedAt
}
```

### 4.4 状态枚举

建议 job 状态包含：

- `queued`
- `downloading`
- `transcribing`
- `translating`
- `packaging`
- `completed`
- `failed`

### 4.5 进度表达
MVP 使用“阶段进度 + 粗粒度百分比”。

示例：
- downloading: `42%`
- transcribing: `3/10 chunks`
- translating: `18/120 segments`

这个方案实现成本低，也足够支撑用户感知系统正在工作。

## 5. 字幕处理策略

### 5.1 转写策略
MVP 统一走 `fast-whisper` 音频转写链路。

理由：
- 保持字幕时间轴口径一致
- 流程可控
- 避免依赖视频已有字幕质量波动

### 5.2 翻译策略
MVP 支持 OpenAI-compatible 接口，由用户自行配置：

- `base URL`
- `model`
- `API key`

后端只维护一个统一兼容层，负责发起翻译请求。

### 5.3 分段原则
翻译时按转写结果逐段处理，并保持一一对应。

要求：
- 每个 segment 单独翻译
- 保持 cue 的起止时间不变
- 禁止翻译过程主动合并段落
- 禁止翻译过程主动拆分段落

这个约束用于保证时间轴稳定，降低字幕错位风险。

### 5.4 产物格式
统一产出 WebVTT：

- `source.vtt`
- `translated.vtt`
- `bilingual.vtt`

双语字幕格式建议：
- 第一行原文
- 第二行译文

这样最适合 MVP，并方便在扩展端切换模式。

## 6. 播放页挂载设计

### 6.1 挂载目标
第一版目标是尽量贴近原生字幕轨道体验。

设计原则：
- 优先将字幕作为标准字幕资源接入播放器体验
- extension 负责选择当前轨道资源
- 页面逻辑集中在桥接脚本中

### 6.2 挂载流程

1. content script 获取当前 `videoId`
2. 找到匹配的 VTT 资源
3. 注入桥接脚本到页面上下文
4. 桥接脚本完成播放器上下文中的字幕接入
5. extension 控制当前显示模式：
   - `translated`
   - `bilingual`

### 6.3 控件设计
MVP 只保留两个核心控件：

- 字幕开关
- 显示模式切换：`translated` / `bilingual`

交互入口：
- popup 中提供完整控制入口
- 播放页右上角可选提供轻量浮层快捷切换入口

MVP 不直接深度改写 YouTube 原生控制条，优先保证兼容性与维护性。

### 6.4 自动恢复
当用户再次打开同一视频时：
- 优先读取本地缓存
- 自动应用上次选择的显示模式
- 本地没有再回源后端

## 7. 存储与部署设计

### 7.1 存储策略
采用“双端保留”：

- 服务端持久化完整任务与字幕产物
- extension 本地缓存任务索引与最近使用状态

这样可以同时满足：
- 重复观看时快速命中
- 服务端结果可重复读取
- extension 本地体验更顺滑

### 7.2 服务端持久化建议
MVP 建议：
- 元数据存 SQLite
- 字幕文件与中间产物存本地磁盘

理由：
- 自托管部署简单
- 依赖少
- 排障容易

### 7.3 进程形态
建议拆成两个进程：

1. `api server`
2. `worker`

这样更清晰：
- API 负责响应用户请求
- worker 负责长任务执行

## 8. 风险与约束

### 8.1 YouTube 页面结构变化
风险：页面结构和播放器上下文可能变化，影响 content script 与挂载逻辑。

应对方式：
- 将页面适配逻辑集中到独立模块
- 减少 DOM 耦合点
- 保持挂载逻辑可替换

### 8.2 字幕轨道接入兼容性
风险：播放器对外部字幕轨道的接入能力可能受限。

应对方式：
- 第一版先验证标准 VTT 接入路径
- 在设计上保留后续降级到自定义渲染层的空间

### 8.3 长视频处理耗时与成本
风险：长视频带来更高的下载、转写与翻译成本。

应对方式：
- MVP 限制最大视频时长为 60 分钟
- 后端在任务创建时校验时长限制

### 8.4 翻译结果破坏时间轴
风险：翻译服务输出改变段落结构，导致字幕错位。

应对方式：
- 固定逐段翻译策略
- 使用结构化 prompt 明确要求保留 segment 对齐
- 翻译后进行 segment 数量校验

## 9. 测试策略

### 9.1 后端测试
覆盖以下内容：
- job lifecycle
- 状态推进
- 各 pipeline stage 的进度写回
- 字幕打包输出
- translation contract 校验

### 9.2 Extension 测试
覆盖以下内容：
- popup 提交任务
- 进度刷新
- 错误展示
- content script 的 `videoId` 识别
- 本地缓存命中与后端回源
- 模式切换与默认模式恢复

### 9.3 集成验证
准备一个公开英文视频作为基准样例，验证：

1. 从 URL 提交到任务完成的完整链路
2. 播放页自动发现字幕
3. `translated` 与 `bilingual` 切换
4. 刷新页面后自动恢复字幕

## 10. 推荐实现顺序

建议后续实现计划按以下顺序展开：

1. 定义 backend job API 与基础数据模型
2. 实现 worker 生命周期与状态推进
3. 接入 `yt-dlp`
4. 接入 `fast-whisper`
5. 接入 OpenAI-compatible 翻译
6. 生成 VTT 产物
7. 完成 extension popup 任务入口与进度展示
8. 完成 content script 自动发现与加载
9. 完成播放页字幕切换
10. 做端到端验证

## 11. 结论

这个设计适合开源、自托管、个人使用优先的 MVP。

它的核心路径是：
- 用 Go backend + worker 稳定处理长任务
- 用 WXT extension 提供任务入口、进度反馈和播放页集成
- 用服务端持久化 + 本地缓存兼顾稳定性与使用体验
- 用标准 WebVTT 作为字幕交付格式

第一版最重要的目标是稳定跑通一条 public YouTube 视频的完整处理与挂载链路。