# Chrome 扩展 YouTube 字幕显示改造设计

## 背景

当前 Chrome 扩展的 YouTube 字幕 UI 由 `extension/entrypoints/youtube.content.ts` 挂载到 `body`，`extension/src/content/YoutubeOverlay.vue` 使用 `fixed inset-x-0 bottom-7` 定位。结果是字幕显示在整个页面底部，而不是 YouTube 视频播放器内部。当前字幕字体大小也没有用户可配置入口。

本设计参考 `mengxi-ream/read-frog` 的 YouTube 字幕实现方式：字幕层挂载到 `#movie_player.html5-video-player` 内部，使用播放器容器内的绝对定位层渲染；入口按钮插入 YouTube 控制栏，而不是悬浮在页面底部。当前项目继续使用 Vue、WXT、现有消息协议和存储边界，不引入 read-frog 的 React/Jotai/adapter 全量架构。

## 目标

- 字幕文本显示在 YouTube 视频播放器内部，而不是整个页面底部。
- YouTube 播放器控制栏中新增一个 Lets Sub It 字幕开关按钮。
- 播放器控制栏按钮只负责打开或关闭当前页面字幕显示。
- popup 改为两个 tab：生成字幕、字幕设置。
- 字幕设置 tab 支持手动输入字体大小，单位为 px，默认 `20`，不设置上下限。
- 字幕设置 tab 支持选择全局默认字幕模式：翻译 only 或双语。
- popup 中修改字幕模式后，当前 YouTube watch 页立即应用；后续视频也沿用该全局默认模式。

## 非目标

- 不实现播放器内设置面板。
- 不实现字幕拖拽位置、字体族、颜色、透明度或更多样式配置。
- 不引入新的前端框架、状态库或 read-frog 的完整字幕架构。
- 不改变后端 API、VTT 生成逻辑、字幕文件服务或后台网络边界。

## 现状与参考

当前项目：

- `youtube.content.ts` 使用 `createShadowRootUi`，`anchor: 'body'`，因此 UI 根节点脱离播放器。
- `YoutubeOverlay.vue` 同时承担字幕加载、模式切换、状态展示、播放器浮动控制条和字幕文本显示。
- `Settings` 只包含 backend URL 和语言设置。
- 字幕模式当前保存在每个字幕资源缓存条目的 `selectedMode` 中。

read-frog 的关键参考点：

- 等待 YouTube 播放器容器 `#movie_player.html5-video-player`。
- 如果播放器容器是 `position: static`，设置为 `position: relative`。
- 在播放器容器中追加 shadow host，host 样式为绝对定位覆盖播放器。
- UI 层设置 `pointer-events: none`，只让按钮或可交互区域重新开启 pointer events。
- 控制按钮插入 YouTube 控制栏区域，而不是显示在页面底部。

## 架构设计

### 播放器内字幕层

内容脚本负责把字幕 UI 挂载到 YouTube 播放器容器：

1. 优先查找 `#movie_player.html5-video-player`。
2. 如果找不到，不抛错，等待 YouTube SPA 导航或 DOM 重建后重试。
3. 挂载前检查播放器容器的 computed `position`；如果是 `static`，临时设置为 `relative`。
4. 在播放器容器内挂载 shadow host，host 覆盖播放器区域：`position: absolute; inset: 0; pointer-events: none; z-index: 9999`。
5. `YoutubeOverlay.vue` 不再使用 `fixed` 定位，不再渲染旧的底部浮动控制条，只渲染字幕文本。
6. 字幕文本在播放器内部底部居中显示，并预留底部空间避免被 YouTube 控制栏遮挡。

YouTube SPA 导航或播放器 DOM 重建时，内容脚本需要重新检查播放器容器。若发现 host 挂在旧容器上，应卸载旧 Vue app 并移除旧 host，再挂载到新的播放器容器，避免重复字幕层。

### 控制栏开关按钮

内容脚本同时负责向 YouTube 控制栏插入一个开关按钮：

1. 查找 `#movie_player .ytp-right-controls`。
2. 插入一个 Lets Sub It 字幕开关按钮。
3. 按钮只切换当前页面的 `enabled` 状态，不改变字幕模式、不展示生成状态、不触发后端请求。
4. 按钮使用 `aria-pressed` 表示开关状态。
5. 按钮 click、pointerdown、mousedown、dblclick 事件应阻止冒泡，避免影响 YouTube 播放器快捷交互。
6. 如果控制栏不存在，保留字幕层功能并在导航或 DOM 变化后重试插入。

播放器内不再显示“字幕开 / 翻译 / 双语 / 状态”的浮动控制条。生成进度、错误和任务状态继续主要由 popup 展示。

### Popup 双 tab

`extension/entrypoints/popup/App.vue` 改为两个 tab：

- `生成字幕`：保留当前表单和任务状态，包括 backend URL、YouTube URL、源语言、目标语言、生成按钮、任务轮询和完成通知。
- `字幕设置`：放消费侧设置，包括字幕字体大小和默认字幕模式。

字体大小设置：

- 输入框手动输入数字。
- 单位固定为 px。
- 默认值为 `20`。
- 不设置上下限。
- 保存时只要求能解析为正数；非法值在 popup 中显示校验错误，不写入 storage。
- content script 应用时使用 `${subtitleFontSizePx}px` 作为字幕文本 `font-size`。

模式设置：

- 选项为 `translated` 和 `bilingual`，文案分别是“翻译 only”和“双语”。
- 作为全局默认模式保存到设置中。
- 保存后立即通知当前 YouTube watch 页应用。

### Settings 存储

扩展 `Settings` 类型：

- `subtitleFontSizePx: number`
- `subtitleMode: SubtitleMode`

默认设置：

- `subtitleFontSizePx: 20`
- `subtitleMode: 'translated'`

`getSettings()` 返回时应合并默认值，保证旧 storage 中缺少新字段时仍能得到完整设置。`updateSettings()` 应继续校验 backend URL 和语言差异，并新增字体大小正数校验。

### 消息流

popup 保存字幕设置后：

1. 调用现有 `settings:update` 持久化 `subtitleFontSizePx` 或 `subtitleMode`。
2. 查询当前活动 tab。
3. 如果是 YouTube watch 页，发送 `lets-sub-it:settings-updated` 消息，携带更新后的完整 settings。

content script 收到设置更新后：

1. 更新本地字体大小并立即影响当前字幕文本样式。
2. 如果 `subtitleMode` 改变，并且当前视频存在字幕资源，调用现有 `subtitle:update-mode` 写回该视频和目标语言的缓存条目。
3. 重新请求对应模式的 VTT 并刷新当前 active cue。
4. 如果当前视频没有字幕资源，只更新全局设置，不报错。

内容脚本初始加载字幕时应读取 settings 中的 `subtitleMode` 作为优先模式，并与当前资源缓存的 `selectedMode` 保持一致。这样后续视频会沿用 popup 中设置的默认模式。

## 错误处理

- 找不到播放器容器、控制栏或 video 元素时不抛错，等待后续导航或 DOM 变化重试。
- VTT 加载失败时隐藏字幕文本，保留内部状态供测试和调试。
- 字体大小输入非法时只在 popup 中显示错误，不写入 storage，不通知 content script。
- popup 通知当前 tab 失败不影响设置持久化。
- 模式立即应用失败时保留当前字幕显示和已保存的全局默认设置；当前视频缓存写回失败时显示或记录可读错误，但不回滚全局设置。

## 测试计划

- `extension/src/storage/settings.test.ts`：覆盖默认 `subtitleFontSizePx: 20`、默认 `subtitleMode: 'translated'`、旧存储字段合并、正数校验和无上限输入。
- `extension/src/content/YoutubeOverlay.test.ts`：覆盖字幕文本使用设置中的 px 字体大小；收到 `lets-sub-it:settings-updated` 后立即更新样式；模式更新后重新请求对应 VTT；播放器内 UI 不再渲染旧浮动模式按钮。
- `extension/entrypoints/popup/App.test.ts`：覆盖两个 tab 的显示；字幕设置 tab 可保存字体大小和模式；保存后调用 `settings:update` 并通知当前 YouTube tab。
- 必要时新增内容脚本挂载辅助函数测试，覆盖播放器容器查找、host 去重和控制栏按钮去重。

验证命令：

```bash
cd extension && mise exec -- npm run test
cd extension && mise exec -- npm run typecheck
```

## 影响范围

主要影响文件预计包括：

- `extension/entrypoints/youtube.content.ts`
- `extension/src/content/YoutubeOverlay.vue`
- `extension/src/content/YoutubeOverlay.test.ts`
- `extension/src/storage/settings.ts`
- `extension/src/storage/settings.test.ts`
- `extension/src/api/messages.ts`
- `extension/entrypoints/popup/App.vue`
- `extension/entrypoints/popup/App.test.ts`

不需要修改后端、whisper CLI、后端 API、字幕文件格式或 Chrome manifest 权限。
