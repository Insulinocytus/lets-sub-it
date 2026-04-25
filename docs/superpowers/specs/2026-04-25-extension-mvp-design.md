# Chrome Extension MVP 设计

## 背景

本文定义 `extension/` Chrome extension MVP 设计。当前项目已经完成 `whisper/`（本地转写 CLI）和 `backend/`（mock API server），extension 是系统中面向用户的最后一环。

本阶段的目标是提供一个可运行的 Chrome extension，包含三个运行上下文：

- **Popup**：URL 提交、目标语言选择、状态轮询、结果展示
- **Background service worker**：API 代理、缓存管理、消息中转
- **Content script**：YouTube 页面识别、字幕叠加渲染

## 技术选型

| 决策 | 选择 |
|------|------|
| Extension 框架 | **WXT** (Chrome MV3) |
| UI 框架 | **Vue 3 Composition API** |
| CSS | **Tailwind CSS + shadcn-vue** |
| 包管理 | **npm** |
| 测试 | **Vitest** |
| 状态管理 | composables + WXT Storage |

## 架构方案

采用**单页 popup + 背景脚本代理**方案：

- background service worker 统一管理 API 调用和缓存
- popup 通过 `runtime.sendMessage` 与 background 通信
- content script 通过 `storage.onChanged` 感知状态变化，不主动发 message

```
┌─────────────────────────────────────────────────────────┐
│  Background (service worker)                            │
│  ┌──────────┐    ┌───────────┐    ┌──────────────────┐  │
│  │ useApi()  │───▶│ useCache()│───▶│ wxt/storage     │  │
│  └────┬─────┘    └───────────┘    └──────────────────┘  │
│       │                                                   │
│    fetch backend                                         │
└───────┼───────────────────────────────────────────────────┘
        │ runtime.sendMessage
        │
┌───────▼──────────┐    ┌──────────────────────┐
│ Popup            │    │ Content Script       │
│ useJobPolling()  │    │ storage.onChanged ◀──┼── storage
│                  │    │ subtitle overlay      │
└──────────────────┘    └──────────────────────┘
```

## 项目结构

```
extension/
  package.json
  wxt.config.ts                   # srcDir: 'src'
  tsconfig.json                   # paths: { "@/*": ["./src/*"] }
  tailwind.config.ts
  postcss.config.js
  components.json                 # shadcn-vue aliases → src/
  src/
    entrypoints/
      popup/
        index.html
        main.ts
        App.vue
        components/
          JobForm.vue             # URL 输入 + 语言选择 + 提交
          JobStatus.vue           # 轮询状态展示
          JobResult.vue           # 完成结果 + 字幕模式切换
      background.ts               # Service worker: API 代理 + 缓存
      content.ts                  # YouTube 注入：字幕叠加层
    components/
      ui/                         # shadcn-vue 组件 (button, dialog 等)
    composables/
      useApi.ts                   # 封装 sendMessage → background
      useCache.ts                 # 封装 storage 读写
      useJobPolling.ts            # 轮询逻辑
    lib/
      utils.ts                    # cn() 工具函数
    styles/
      globals.css                 # Tailwind + shadcn-vue CSS 变量
    types/
      index.ts
```

## Popup 组件树

Popup 单页面，按 `status` 切换三个互斥视图：

```
App.vue
├── JobForm.vue          # status = null：提交表单
├── JobStatus.vue        # queued ~ packaging：进度展示
└── JobResult.vue        # completed / failed：结果或重试
```

## 数据模型

### LocalCacheEntry

```typescript
interface LocalCacheEntry {
  videoId: string
  targetLanguage: string
  jobId: string
  selectedMode: 'translated' | 'bilingual'
  lastSyncedAt: string   // ISO 8601
}
```

### UserPreferences

```typescript
interface UserPreferences {
  videoId: string
  targetLanguage: string | null  // null = 未设置
  selectedMode: 'translated' | 'bilingual'
}
```

### Storage Key 设计

| Key | 类型 | 用途 |
|-----|------|------|
| `cache:{videoId}:{targetLanguage}` | `LocalCacheEntry` | job 结果缓存 |
| `prefs:{videoId}` | `UserPreferences` | 上次选择 |

## Content Script 字幕叠加

采用 shadow DOM 隔离 YouTube 页面样式。字幕数据直接拉取 backend 提供的 VTT 文件（`translated.vtt` 或 `bilingual.vtt`），不做拼合。

```
content.ts
  │
  ├── 1. 检查页面: youtube.com/watch*
  ├── 2. 等待视频播放器 (MutationObserver, 最多 3 次重试)
  ├── 3. 挂载 shadow DOM 容器
  ├── 4. 监听 storage.onChanged
  │     ├── subtitleData[videoId]  → 更新 VTT URL
  │     └── subtitleMode[videoId]  → 切换 translated/bilingual
  ├── 5. 获取 VTT 并解析 → 内存 cue 列表
  ├── 6. video timeupdate → cue 匹配 → 渲染
  └── 7. SPA 导航检测 → 重新挂载
```

**字幕模式切换逻辑**：用户选 `translated` 则拉取 `translated.vtt`，选 `bilingual` 则拉取 `bilingual.vtt`。后端已提供完整的 `bilingual.vtt`（每个 cue 包含源语言和翻译两行文本），前端不做文本合并。

## 缓存读取流程

```
用户打开 youtube.com/watch?v=xxx

  content.ts
    ├── 读 prefs:{videoId}
    │   ├── null → 不自动显示字幕
    │   └── 有值 → 读 cache:{videoId}:{targetLanguage}
    │       ├── 完成 → 拉 VTT → 挂载字幕
    │       └── 未完成 → 静默等待
    │
    └── 即使无缓存，也挂载 shadow DOM 容器
        等待 storage 变化通知
```

Popup 在 job 完成后写入 `cache:` 和 `prefs:`。Content script 只读不写 storage。

## 错误处理

### Popup

| 场景 | 行为 |
|------|------|
| Backend 不可达 | 按钮 loading 超时，显示"无法连接服务器"与重试按钮 |
| URL 无效 | 前端实时校验格式，无效时按钮 disabled |
| Job failed | 展示失败阶段 + errorMessage，提供"重新提交" |
| Popup 关闭后重开 | 通过 background 缓存恢复上次视图 |

### Content Script

| 场景 | 行为 |
|------|------|
| DOM 结构变化 | 重试 3 次 (1s 间隔)，失败则静默 |
| VTT 加载失败 | 不显示字幕，等待下次 storage 变化 |
| SPA 换视频 | 检测 history 变化 → 重新检测 videoId → 读取缓存 |
| 页面关闭 | 缓存已持久化，下次打开恢复 |

### Background

| 场景 | 行为 |
|------|------|
| API 5xx | 重试 1 次，仍失败则返回错误 |
| Service worker 回收 | 从 storage 恢复缓存 |
| 重复请求 | in-flight map 去重 |

## 非目标

- 不做多页面 popup（MVP 只需要单页）
- 不做 Playwright E2E 测试（MVP 只写 Vitest 单元测试）
- 不做字幕编辑、历史管理、设置页面
- 不直接持有翻译 provider 密钥
- 不支持非 YouTube 平台

## 后续方向

- 增加 Playwright E2E 测试覆盖 extension 加载和 content script 注入
- 支持更多 YouTube 页面变体（短剧、直播）
- 弹出页增加任务历史列表
- 设置页面允许配置后端地址
