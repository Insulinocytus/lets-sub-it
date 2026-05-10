# 字幕播放器内嵌与字体大小设置 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将字幕从页面底部 `fixed` 定位改为嵌入 YouTube 播放器内部，支持字体大小手动 px 输入，并将 popup 拆分为双 Tab 布局。

**Architecture:** 手动管理 Shadow DOM 挂载到 `#movie_player`；在 `.ytp-right-controls` 注入纯 DOM 开关按钮；popup 用 reka-ui Tabs 拆分为"字幕生成"和"字幕设置"两个 Tab；设置通过 `ExtensionMessage` 经 background 同步到 content script。

**Tech Stack:** Vue 3, WXT, TypeScript, reka-ui, Tailwind CSS v4, Vitest + fakeBrowser

---

### Task 1: 扩展 Settings 类型和存储

**Files:**
- Modify: `extension/src/api/messages.ts:45-49`
- Modify: `extension/src/storage/settings.ts:9-13`

- [ ] **Step 1: 在 messages.ts 中给 Settings 添加 subtitleFontSize**

```typescript
export type Settings = {
  backendBaseUrl: string
  sourceLanguage: LanguageCode
  targetLanguage: LanguageCode
  subtitleFontSize: number
}
```

- [ ] **Step 2: 更新 DEFAULT_SETTINGS 添加默认值 20**

在 `extension/src/storage/settings.ts`:

```typescript
export const DEFAULT_SETTINGS: Settings = Object.freeze({
  backendBaseUrl: 'http://127.0.0.1:8080',
  sourceLanguage: 'en',
  targetLanguage: 'zh',
  subtitleFontSize: 20,
})
```

- [ ] **Step 3: 运行 settings 测试确认兼容**

Run: `cd extension && mise exec -- npx vitest run src/storage/settings.test.ts`

Expected: All tests PASS (existing tests use partial updates via `updateSettings`, old entries won't include `subtitleFontSize` but spread will pick up the new default implicitly if needed; test assertions check specific fields, new field won't break them).

---

### Task 2: 更新测试中的 DEFAULT_SETTINGS 引用

**Files:**
- Modify: `extension/src/api/message-handler.test.ts`

- [ ] **Step 1: 添加 subtitleFontSize 到任何直接构造 Settings 的断言**

Search for `DEFAULT_SETTINGS` references in test assertions. Since `DEFAULT_SETTINGS` is now `{ backendBaseUrl, sourceLanguage, targetLanguage, subtitleFontSize: 20 }`, test at line 203 (`expect(getSettings()).resolves.toEqual(DEFAULT_SETTINGS)`) needs updating:

```typescript
it('rejects invalid backend URLs without persisting settings', async () => {
  const result = await handleExtensionMessage({
    type: 'settings:update',
    payload: {
      backendBaseUrl: 'http://localhost:8080/api',
    },
  })

  expect(result).toEqual({
    ok: false,
    error: {
      code: 'invalid_backend_url',
      message: 'backendBaseUrl must be a localhost or 127.0.0.1 origin',
    },
  })
  await expect(getSettings()).resolves.toEqual(expect.objectContaining({
    backendBaseUrl: 'http://127.0.0.1:8080',
    sourceLanguage: 'en',
    targetLanguage: 'zh',
  }))
})
```

Also update the test at line 62 of `settings.test.ts`:
```typescript
it('rejects invalid backend ports without persisting settings', async () => {
  await expect(
    updateSettings({ backendBaseUrl: 'http://localhost:65536' }),
  ).rejects.toMatchObject({
    code: 'invalid_backend_url',
    message: 'backendBaseUrl must be a localhost or 127.0.0.1 origin',
  })
  await expect(getSettings()).resolves.toEqual(
    expect.objectContaining({
      backendBaseUrl: 'http://127.0.0.1:8080',
      sourceLanguage: 'en',
      targetLanguage: 'zh',
    }),
  )
})
```

- [ ] **Step 2: 运行全部测试确认通过**

Run: `cd extension && mise exec -- npx vitest run`

---

### Task 3: 新增 settings:update-subtitle 和 subtitle:settings-changed 消息处理

**Files:**
- Modify: `extension/src/api/messages.ts:57-100`
- Modify: `extension/src/api/message-handler.ts:1-156`

- [ ] **Step 1: 在 ExtensionMessage 联合类型中新增消息**

```typescript
export type ExtensionMessage =
  | { type: 'settings:get' }
  | { type: 'settings:update'; payload: Partial<Settings> }
  | { type: 'job:create'; payload: CreateJobInput }
  | { type: 'job:get'; payload: { jobId: string } }
  | {
      type: 'job:active'
      payload: { videoId: string; targetLanguage: LanguageCode }
    }
  | { type: 'subtitle:resolve'; payload: { videoId: string } }
  | {
      type: 'subtitle:fetch-file'
      payload: { jobId: string; mode: SubtitleMode }
    }
  | {
      type: 'subtitle:update-mode'
      payload: { videoId: string; targetLanguage: LanguageCode; mode: SubtitleMode }
    }
  | {
      type: 'settings:update-subtitle'
      payload: { fontSize?: number; mode?: SubtitleMode }
    }
```

- [ ] **Step 2: 在 message-handler.ts switch 中新增 case，处理 settings:update-subtitle**

在 `case 'subtitle:update-mode':` 之后，`}` 结束之前插入：

```typescript
      case 'settings:update-subtitle': {
        const payload = message.payload
        const patch: Partial<Settings> = {}
        if (payload.fontSize !== undefined) {
          patch.subtitleFontSize = payload.fontSize
        }
        const updated = await updateSettings(patch)
        if (payload.mode !== undefined) {
          const settings = await getSettings()
          const tabs = await browser.tabs.query({
            url: ['https://www.youtube.com/watch*'],
          })
          await Promise.allSettled(
            tabs.map(async (tab) => {
              if (!tab.id) return
              await browser.tabs.sendMessage(tab.id, {
                type: 'subtitle:settings-changed',
                payload: {
                  mode: payload.mode,
                },
              })
            }),
          )
        }
        await notifySubtitleSettingsChanged(payload)
        return ok(updated)
      }
```

Wait - actually, looking at the background architecture, `handleExtensionMessage` is called from `background.ts`. The `notifySubtitleSettingsChanged` function needs to be in `job-monitor.ts` or directly in the message handler. Let me revise - we need a simpler approach.

The background handler should:
1. Store the partial settings update
2. Notify all YouTube tabs about the change so content scripts can apply it

Since `message-handler.ts` already imports from `@/storage/settings`, it can directly update. And we can inline the tab notification.

Actually, let me reconsider. The `browser` import should already be available in the test via `fakeBrowser`. Let me check - the message-handler.ts doesn't currently import `browser`. Let me add the import and keep the notification inline.

Let me rewrite this step more carefully.

In `message-handler.ts`:

Add import at top:
```typescript
import { browser } from 'wxt/browser'
```

Add new case in switch:
```typescript
      case 'settings:update-subtitle': {
        const patch: Partial<Settings> = {}
        if (message.payload.fontSize !== undefined) {
          patch.subtitleFontSize = message.payload.fontSize
        }
        const updated = await updateSettings(patch)
        const tabs = await browser.tabs.query({
          url: ['https://www.youtube.com/watch*'],
        })
        await Promise.allSettled(
          tabs.map(async (tab) => {
            if (!tab.id) return
            await browser.tabs.sendMessage(tab.id, {
              type: 'subtitle:settings-changed',
              payload: message.payload,
            })
          }),
        )
        return ok(updated)
      }
```

- [ ] **Step 3: 运行 message-handler 测试确认通过**

Run: `cd extension && mise exec -- npx vitest run src/api/message-handler.test.ts`

---

### Task 4: 创建播放器控制条按钮模块

**Files:**
- Create: `extension/src/content/player-button.ts`
- Test: `extension/src/content/player-button.test.ts`

- [ ] **Step 1: 编写测试**

Create `extension/src/content/player-button.test.ts`:

```typescript
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { injectPlayerButton, removePlayerButton } from './player-button'

describe('player-button', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
  })

  it('injects a button into ytp-right-controls', () => {
    const controls = document.createElement('div')
    controls.className = 'ytp-right-controls'
    document.body.appendChild(controls)

    injectPlayerButton(() => {})

    const button = controls.querySelector('button[data-lsi-subtitle-toggle]')
    expect(button).not.toBeNull()
    expect(button?.textContent).toBe('LS')
  })

  it('does not inject a second button when one already exists', () => {
    const existing = document.createElement('button')
    existing.setAttribute('data-lsi-subtitle-toggle', '')
    const controls = document.createElement('div')
    controls.className = 'ytp-right-controls'
    controls.appendChild(existing)
    document.body.appendChild(controls)

    injectPlayerButton(() => {})

    expect(controls.querySelectorAll('button[data-lsi-subtitle-toggle]').length).toBe(1)
  })

  it('calls the toggle callback on click', () => {
    const controls = document.createElement('div')
    controls.className = 'ytp-right-controls'
    document.body.appendChild(controls)

    const toggle = vi.fn()
    injectPlayerButton(toggle)

    const button = controls.querySelector('button[data-lsi-subtitle-toggle]')!
    button.click()
    expect(toggle).toHaveBeenCalledOnce()
  })

  it('removePlayerButton removes the injected button', () => {
    const controls = document.createElement('div')
    controls.className = 'ytp-right-controls'
    document.body.appendChild(controls)

    injectPlayerButton(() => {})
    expect(controls.querySelector('button[data-lsi-subtitle-toggle]')).not.toBeNull()

    removePlayerButton()
    expect(document.body.querySelector('button[data-lsi-subtitle-toggle]')).toBeNull()
  })

  it('injectPlayerButton is a no-op when controls bar is missing', () => {
    expect(() => injectPlayerButton(() => {})).not.toThrow()
  })
})
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd extension && mise exec -- npx vitest run src/content/player-button.test.ts`

Expected: FAIL — module not found

- [ ] **Step 3: 实现 player-button.ts**

```typescript
const BUTTON_ATTR = 'data-lsi-subtitle-toggle'

export function injectPlayerButton(onToggle: () => void): void {
  const controls = document.querySelector('.ytp-right-controls')
  if (!controls) return

  const existing = controls.querySelector(`button[${BUTTON_ATTR}]`)
  if (existing) return

  const button = document.createElement('button')
  button.setAttribute(BUTTON_ATTR, '')
  button.className = 'ytp-button'
  button.title = 'Lets Sub It'
  button.setAttribute('aria-label', 'Lets Sub It 字幕')
  button.textContent = 'LS'
  button.style.cssText = `
    width: auto;
    min-width: 40px;
    font-size: 13px;
    font-weight: 600;
    color: #fff;
    text-shadow: 0 0 2px rgba(0,0,0,0.5);
  `
  button.addEventListener('click', (e) => {
    e.stopPropagation()
    e.preventDefault()
    onToggle()
  })

  controls.appendChild(button)
}

export function removePlayerButton(): void {
  const button = document.querySelector(`button[${BUTTON_ATTR}]`)
  button?.remove()
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd extension && mise exec -- npx vitest run src/content/player-button.test.ts`

Expected: PASS

---

### Task 5: 重构 content script 入口

**Files:**
- Modify: `extension/entrypoints/youtube.content.ts:1-26`

- [ ] **Step 1: 将 youtube.content.ts 改为手动 Shadow DOM 管理**

```typescript
import '@/style.css'
import { createApp } from 'vue'
import { browser } from 'wxt/browser'
import YoutubeOverlay from '@/content/YoutubeOverlay.vue'
import { injectPlayerButton, removePlayerButton } from '@/content/player-button'
import { getSettings } from '@/storage/settings'

export default defineContentScript({
  matches: ['https://www.youtube.com/watch*'],
  cssInjectionMode: 'ui',
  async main(ctx) {
    let app: ReturnType<typeof createApp> | null = null
    let shadowHost: HTMLDivElement | null = null

    function mountOverlay() {
      const playerContainer = document.querySelector('#movie_player')
      if (!playerContainer) {
        setTimeout(mountOverlay, 500)
        return
      }

      const existing = document.getElementById('lsi-subtitle-host')
      if (existing) return

      shadowHost = document.createElement('div')
      shadowHost.id = 'lsi-subtitle-host'
      shadowHost.style.cssText = `
        position: absolute;
        inset: 0;
        pointer-events: none;
        z-index: 20;
        overflow: visible;
      `
      const shadowRoot = shadowHost.attachShadow({ mode: 'open' })

      const container = document.createElement('div')
      container.style.cssText = `
        position: absolute;
        inset: 0;
        pointer-events: none;
        overflow: visible;
      `
      shadowRoot.appendChild(container)

      app = createApp(YoutubeOverlay)
      app.mount(container)

      playerContainer.appendChild(shadowHost)
    }

    function unmountOverlay() {
      app?.unmount()
      app = null
      shadowHost?.remove()
      shadowHost = null
      removePlayerButton()
    }

    function toggleSubtitle() {
      if (!app) return
      const instance = app._instance
      if (!instance?.exposed?.toggleEnabled) return
      instance.exposed.toggleEnabled()
    }

    mountOverlay()

    const settings = await getSettings()
    injectPlayerButton(toggleSubtitle)

    window.addEventListener('yt-navigate-finish', () => {
      unmountOverlay()
      mountOverlay()
      injectPlayerButton(toggleSubtitle)
    })

    ctx.onInvalidated(() => {
      unmountOverlay()
    })
  },
})
```

- [ ] **Step 2: 运行 typecheck 检查类型**

Run: `cd extension && mise exec -- npm run typecheck`

Expected: May fail if `toggleEnabled` is not exposed. Fix in Task 6.

---

### Task 6: 重构 YoutubeOverlay.vue

**Files:**
- Modify: `extension/src/content/YoutubeOverlay.vue:1-432`
- Modify: `extension/src/content/YoutubeOverlay.test.ts:1-384`

- [ ] **Step 1: 修改 YoutubeOverlay.vue — 移除控制栏，新增 fontSize 和设置监听**

将整个 `<script setup>` 部分替换为：

```typescript
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { browser } from 'wxt/browser'
import {
  type SubtitleMode,
  sendExtensionMessage,
} from '@/api/messages'
import type { SubtitleAssetCacheEntry } from '@/storage/subtitle-cache'
import { findActiveCue } from '@/subtitles/active-cue'
import { parseVtt, type VttCue } from '@/subtitles/vtt'
import {
  getCurrentVideoId,
  watchVideoIdChanges,
} from '@/youtube/page-watch'
import { getSettings } from '@/storage/settings'

const enabled = ref(true)
const status = ref('查找字幕')
const currentVideoId = ref<string | null>(null)
const currentAsset = ref<SubtitleAssetCacheEntry | null>(null)
const selectedMode = ref<SubtitleMode>('translated')
const cues = ref<VttCue[]>([])
const activeText = ref('')
const fontSize = ref(20)

let removeVideoListeners: (() => void) | null = null
let removeVideoIdWatch: (() => void) | null = null
let removeRuntimeListener: (() => void) | null = null
let isMounted = false
let requestToken = 0

const hasSubtitle = computed(() => cues.value.length > 0)
const isWatchPage = computed(() => currentVideoId.value !== null)

const subtitleStyle = computed(() => ({
  fontSize: `${fontSize.value}px`,
}))

function toggleEnabled() {
  enabled.value = !enabled.value
}

defineExpose({ toggleEnabled, enabled })

onMounted(async () => {
  isMounted = true

  try {
    const settings = await getSettings()
    fontSize.value = settings.subtitleFontSize ?? 20
  } catch {
    // keep default
  }

  const videoId = getCurrentVideoId()
  currentVideoId.value = videoId
  void loadForVideo(videoId)

  removeVideoIdWatch = watchVideoIdChanges((nextVideoId) => {
    currentVideoId.value = nextVideoId
    void loadForVideo(nextVideoId)
  })

  const handleRuntimeMessage = (message: unknown) => {
    if (isSubtitleUpdatedMessage(message)) {
      if (message.videoId !== currentVideoId.value) return
      void loadForVideo(currentVideoId.value)
      return
    }
    if (isSettingsChangedMessage(message)) {
      if (message.payload.fontSize !== undefined) {
        fontSize.value = message.payload.fontSize
      }
      if (message.payload.mode !== undefined) {
        void changeMode(message.payload.mode)
      }
      return
    }
  }
  browser.runtime.onMessage.addListener(handleRuntimeMessage)
  removeRuntimeListener = () => {
    browser.runtime.onMessage.removeListener(handleRuntimeMessage)
  }
})

onUnmounted(() => {
  isMounted = false
  requestToken += 1
  cleanupVideoListeners()
  removeVideoIdWatch?.()
  removeVideoIdWatch = null
  removeRuntimeListener?.()
  removeRuntimeListener = null
})

// ... 其余函数保持完全不变 (loadForVideo, loadVtt, changeMode, bindVideo, etc.)
// (line 72-377 of original file)
```

保持以下函数完全不变：`loadForVideo`, `loadVtt`, `changeMode`, `bindVideo`, `canUpdate`, `cleanupVideoListeners`, `resetLoadedSubtitles`, `restoreDisplayedSubtitles`, `handleModeClick`, `readableError`, `isSubtitleUpdatedMessage`。

新增函数：

```typescript
type SubtitleSettingsChangedMessage = {
  type: 'subtitle:settings-changed'
  payload: { fontSize?: number; mode?: SubtitleMode }
}

function isSettingsChangedMessage(message: unknown): message is SubtitleSettingsChangedMessage {
  if (typeof message !== 'object' || message === null) return false
  const candidate = message as Partial<SubtitleSettingsChangedMessage>
  return candidate.type === 'subtitle:settings-changed' && typeof candidate.payload === 'object'
}
```

- [ ] **Step 2: 修改 template — 移除控制栏，简化字幕显示**

```html
<template>
  <div v-if="isWatchPage && enabled && hasSubtitle && activeText" class="flex h-full items-end justify-center pb-[60px]">
    <div
      class="pointer-events-auto max-w-[90%] whitespace-pre-line rounded-md bg-black/78 px-4 py-2 text-center font-semibold leading-snug text-white shadow-lg [text-shadow:0_1px_2px_rgb(0_0_0/0.85)]"
      :style="subtitleStyle"
    >
      {{ activeText }}
    </div>
  </div>
</template>
```

说明：`pb-[60px]` 是预估的 YouTube 控制栏高度，确保字幕不遮挡进度条。整体 `h-full` 配合 `items-end` 将字幕放在播放器底部。

- [ ] **Step 3: 更新测试文件**

`extension/src/content/YoutubeOverlay.test.ts`:
- 移除对按钮组件 `Button` 和 `Badge` 的 mock（因为控制栏已移除）
- 移除 `getButtonByText` 辅助函数
- 更新测试：模式切换不再通过点击按钮触发，改为直接调用 `changeMode`
- 添加 `getSettings` mock 返回默认设置

关键改动：测试不再查找 button 的 `variant` 属性（控制栏已移除），改用检查 `activeText` 确认字幕加载成功。

由于测试改动较大，具体如下：

```typescript
const {
  sendExtensionMessage,
  getCurrentVideoId,
  watchVideoIdChanges,
  addRuntimeListener,
  removeRuntimeListener,
  getSettings,
} = vi.hoisted(() => ({
  sendExtensionMessage: vi.fn(),
  getCurrentVideoId: vi.fn(),
  watchVideoIdChanges: vi.fn(),
  addRuntimeListener: vi.fn(),
  removeRuntimeListener: vi.fn(),
  getSettings: vi.fn(),
}))

vi.mock('@/storage/settings', () => ({
  getSettings,
}))

vi.mock('wxt/browser', () => ({
  browser: {
    runtime: {
      onMessage: {
        addListener: addRuntimeListener,
        removeListener: removeRuntimeListener,
      },
    },
  },
}))

vi.mock('@/api/messages', () => ({
  sendExtensionMessage,
}))

vi.mock('@/youtube/page-watch', () => ({
  getCurrentVideoId,
  watchVideoIdChanges,
}))
```

Setup:
```typescript
getSettings.mockResolvedValue({ subtitleFontSize: 20 })
```

测试改用 `wrapper.vm` 访问 `changeMode`：
```typescript
it('writes previous mode back to storage when rollback succeeds', async () => {
  sendExtensionMessage
    .mockResolvedValueOnce({ ok: true, data: asset })
    .mockResolvedValueOnce({ ok: true, data: validVtt })
    .mockResolvedValueOnce({
      ok: true,
      data: { ...asset, selectedMode: 'bilingual' as const },
    })
    .mockResolvedValueOnce({
      ok: false,
      error: { code: 'subtitle_file_missing', message: '字幕文件不存在' },
    })
    .mockResolvedValueOnce({
      ok: true,
      data: { ...asset, selectedMode: 'translated' as const },
    })
    .mockResolvedValueOnce({ ok: true, data: validVtt })

  const wrapper = mount(YoutubeOverlay)
  await flushPromises()

  await (wrapper.vm as any).changeMode('bilingual')
  await flushPromises()

  expect(wrapper.text()).toContain('hello')
  expect(wrapper.text()).toContain('字幕已加载')
})
```

同样更新其他 4 个测试（将 `getButtonByText(wrapper, '双语').trigger('click')` 替换为 `(wrapper.vm as any).changeMode('bilingual')`）。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd extension && mise exec -- npx vitest run src/content/YoutubeOverlay.test.ts`

Expected: PASS

---

### Task 7: 重构 Popup App.vue 为双 Tab 布局

**Files:**
- Modify: `extension/entrypoints/popup/App.vue:1-421`

- [ ] **Step 1: 引入 reka-ui Tabs，将现有 UI 放入 Tab 1**

在 `<script setup>` 中添加：
```typescript
import { TabsContent, TabsList, TabsRoot, TabsTrigger } from 'reka-ui'
```

新增 refs：
```typescript
const subtitleFontSize = ref(20)
const subtitleMode = ref<'translated' | 'bilingual'>('translated')
```

在 `onMounted` 初始化后加载设置：
```typescript
try {
  const settingsResult = await sendExtensionMessage<Settings>({ type: 'settings:get' })
  if (settingsResult.ok) {
    subtitleFontSize.value = settingsResult.data.subtitleFontSize ?? 20
  }
} catch {}
```

已有 `onMounted` 内已经调用了 `sendExtensionMessage<Settings>({ type: 'settings:get' })`，直接在 `.then` 中提取 `subtitleFontSize` 即可。

新增函数：
```typescript
async function applySubtitleSettings() {
  try {
    await sendExtensionMessage({
      type: 'settings:update-subtitle',
      payload: {
        fontSize: subtitleFontSize.value,
        mode: subtitleMode.value,
      },
    })
  } catch {
    errorMessage.value = '设置同步失败'
  }
}
```

Template 重构：

```html
<template>
  <main class="w-[380px] bg-background p-4 text-foreground">
    <TabsRoot default-value="generation" class="space-y-4">
      <TabsList class="flex gap-1 rounded-lg bg-muted p-1">
        <TabsTrigger
          value="generation"
          class="flex-1 rounded-md px-3 py-1.5 text-xs font-medium data-[state=active]:bg-background data-[state=active]:shadow-sm"
        >
          字幕生成
        </TabsTrigger>
        <TabsTrigger
          value="settings"
          class="flex-1 rounded-md px-3 py-1.5 text-xs font-medium data-[state=active]:bg-background data-[state=active]:shadow-sm"
        >
          字幕设置
        </TabsTrigger>
      </TabsList>

      <TabsContent value="generation">
        <Card class="gap-4 rounded-lg py-4 shadow-none">
          <!-- 原有的 CardHeader / CardContent / form 保持不变 -->
        </Card>
      </TabsContent>

      <TabsContent value="settings">
        <Card class="gap-4 rounded-lg py-4 shadow-none">
          <CardHeader class="gap-1 px-4">
            <CardTitle class="text-base">字幕设置</CardTitle>
            <CardDescription class="text-xs">
              调整字幕显示样式和模式。
            </CardDescription>
          </CardHeader>

          <CardContent class="space-y-4 px-4">
            <label class="grid gap-1.5 text-xs font-medium">
              字体大小 (px)
              <Input
                v-model.number="subtitleFontSize"
                class="h-8 text-xs"
                type="number"
                :min="8"
                :max="72"
                placeholder="20"
              />
            </label>

            <div class="grid gap-1.5 text-xs font-medium">
              <span>字幕模式</span>
              <div class="flex gap-2">
                <Button
                  size="sm"
                  class="h-8 flex-1 text-xs"
                  :variant="subtitleMode === 'translated' ? 'secondary' : 'outline'"
                  @click="subtitleMode = 'translated'"
                >
                  翻译
                </Button>
                <Button
                  size="sm"
                  class="h-8 flex-1 text-xs"
                  :variant="subtitleMode === 'bilingual' ? 'secondary' : 'outline'"
                  @click="subtitleMode = 'bilingual'"
                >
                  双语
                </Button>
              </div>
            </div>

            <Button class="h-8 w-full text-xs" @click="applySubtitleSettings">
              应用设置
            </Button>
          </CardContent>
        </Card>
      </TabsContent>
    </TabsRoot>
  </main>
</template>
```

- [ ] **Step 2: 运行 typecheck 确认类型正确**

Run: `cd extension && mise exec -- npm run typecheck`

---

### Task 8: 运行全部测试和构建验证

- [ ] **Step 1: 运行全部测试**

Run: `cd extension && mise exec -- npm run test`

Expected: All tests PASS

- [ ] **Step 2: 运行 typecheck**

Run: `cd extension && mise exec -- npm run typecheck`

Expected: No errors

- [ ] **Step 3: 运行 build**

Run: `cd extension && mise exec -- npm run build`

Expected: Build succeeds, output in `extension/.output/chrome-mv3`
