# 字幕播放器内嵌与字体大小设置

## 目标

1. 将字幕从页面底部 `fixed` 定位改为嵌入 YouTube 视频播放器（`#movie_player`）内部，与 YouTube 原生字幕显示方式一致。
2. 支持用户手动输入 px 值自定义字幕字体大小。
3. 将扩展 popup 重构为双 Tab 布局：字幕生成（现有功能）和字幕设置（新增）。
4. 在 YouTube 控制条（`.ytp-right-controls`）中注入一个字幕开关按钮。

## 架构

### Content Script 改动

#### DOM 挂载方式

**现状：** 使用 WXT `createShadowRootUi`，`position: 'inline'`，`anchor: 'body'`。整个 UI 以 `fixed` 定位在页面底部。

**改为：** 手动管理 Shadow DOM：
- 等待 `#movie_player` 出现在 DOM 中。
- 在 `#movie_player` 内创建 Shadow DOM host `div`。
- Host 样式：`position: absolute; inset: 0; pointer-events: none; z-index: 20; overflow: visible`。
- 在 Shadow DOM 内挂载 Vue 应用（`YoutubeOverlay.vue`）。

原因：WXT 的 `createShadowRootUi` 只支持 `body` 作为 anchor，无法直接挂载到 `#movie_player`。

#### 播放器按钮注入

新文件：`extension/src/content/player-button.ts`

- 等待 `.ytp-right-controls` 出现。
- 创建纯 DOM 按钮元素（不走 Shadow DOM，直接继承 YouTube 样式）。
- 点击后切换字幕显示/隐藏，通过给 Shadow DOM 内的 Vue 实例发送消息实现。
- SPA 导航时重新注入。
- 按钮自然融入 YouTube 控制条的 flex 布局。

#### YoutubeOverlay.vue

模板改动：
- 移除控制栏（字幕开/关、翻译/双语按钮、状态 badge）。
- 字幕文本容器去掉 `fixed inset-x-0 bottom-7`，改为普通 div，样式由 ref 动态控制。
- 字幕文本 `font-size` 绑定 ref，响应 storage 变更。

脚本改动：
- 新增 `fontSize` ref，初始化时从 storage 读取，监听 `subtitle:settings-changed` 消息更新。
- 移除模式切换按钮逻辑（移到 popup）。

### Popup 改动

#### App.vue 重构

使用 shadcn-vue `Tabs` 组件拆分为两个 Tab：
- **Tab 1「字幕生成」**：现有 UI（URL 输入、语言选择、提交按钮、进度跟踪）。
- **Tab 2「字幕设置」**：字体大小输入（number，px，默认 20，范围 8-72）、模式切换（翻译/双语）。

Tab 2 修改后通过 `settings:update-subtitle` 消息经 background 同步到 content script。

### Storage & Settings

扩展 `Settings` 类型（`messages.ts`）：

```typescript
export type Settings = {
  backendBaseUrl: string
  sourceLanguage: LanguageCode
  targetLanguage: LanguageCode
  subtitleFontSize: number       // 新增，默认 20
}
```

同步更新 `DEFAULT_SETTINGS` 和 `updateSettings`。

### Messages

新增消息类型：

```typescript
| { type: 'subtitle:settings-changed'; payload: { fontSize?: number; mode?: SubtitleMode } }
| { type: 'settings:update-subtitle'; payload: { fontSize?: number; mode?: SubtitleMode } }
```

- `settings:update-subtitle`：Popup → Background，存储设置并转发到 content scripts。
- `subtitle:settings-changed`：Background → Content Script，通知 YoutubeOverlay 应用新设置。

### 文件清单

| 文件 | 变更 |
|---|---|
| `extension/entrypoints/youtube.content.ts` | 替换 `createShadowRootUi` 为手动 Shadow DOM 挂载到 `#movie_player`；调用 `injectPlayerButton` |
| `extension/src/content/YoutubeOverlay.vue` | 移除控制栏模板；新增 `fontSize` ref；监听 settings 消息；字幕文本动态 style 绑定 |
| `extension/src/content/player-button.ts` | **新增** — 向 `.ytp-right-controls` 注入字幕开关按钮 |
| `extension/entrypoints/popup/App.vue` | 拆分为双 Tab 布局 |
| `extension/src/storage/settings.ts` | 新增 `subtitleFontSize` 字段，默认 20，范围 8-72 |
| `extension/src/api/messages.ts` | `Settings` 类型增加 `subtitleFontSize`；新增两种消息类型定义 |
| `extension/entrypoints/background.ts` | 处理 `settings:update-subtitle` 消息；转发到 content scripts |

### 数据流

```
Popup Tab 2                  Background                       Content Script
    │                            │                                 │
    │ settings:update-subtitle ─>│                                 │
    │                            │── 写入 storage                   │
    │                            │── tab query youtube 标签页       │
    │                            │── subtitle:settings-changed ────>│
    │                            │                                 │── 更新 fontSize ref
    │                            │                                 │── 重新应用样式
```

### 边界情况

- SPA 导航：重新注入播放器按钮，重新挂载 Shadow DOM。
- 全屏：字幕容器正确跟随播放器大小变化（因为使用 `absolute` 定位在 `#movie_player` 内）。
- 字体大小边界：输入框限制最小值 8px，最大值 72px。
- 播放器按钮可见性：仅在 watch 页面（`youtube.com/watch`）显示。

## 不做

- 字幕垂直拖拽位置调整。
- 隐藏 YouTube 原生字幕。
- 字幕颜色 / 背景不透明度自定义。
- YouTube 播放器内的设置面板。
