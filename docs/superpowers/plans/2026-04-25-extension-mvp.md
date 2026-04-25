# Extension MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `extension/` 下实现 Chrome MV3 extension MVP，打通 popup 创建 job、background 访问 mock backend、storage 缓存字幕资产，以及 YouTube watch 页面字幕显示链路。

**Architecture:** WXT 负责 Manifest V3、entrypoints 和 Vite 构建；background service worker 是唯一 HTTP API 网关；popup 只处理任务提交与状态展示；YouTube content script 只处理页面识别、Shadow DOM UI 和字幕渲染。共享业务逻辑放在 `src/api`、`src/storage`、`src/subtitles`、`src/youtube`，优先通过 Vitest 覆盖。

**Tech Stack:** WXT, Chrome Manifest V3, Vue 3, TypeScript, Vite, Vitest, shadcn-vue, Tailwind CSS v4, npm, Node 22 via mise

---

## 文件结构

- Modify: `mise.toml`
  - 增加 Node 22 工具链。
- Modify: `README.md`
  - 在项目说明中补充 extension 已有 MVP 工程和本地验证命令。
- Modify: `docs/superpowers/specs/2026-04-25-extension-mvp-design.md`
  - 保留本计划发现的 `activeTab` 权限修正。
- Create: `extension/package.json`
  - npm scripts 和 WXT/Vue/Vitest 依赖入口。
- Create: `extension/package-lock.json`
  - 通过 `npm install` 生成。
- Create: `extension/wxt.config.ts`
  - WXT、Vue、Tailwind、manifest permissions 和 host permissions。
- Create: `extension/tsconfig.json`
  - TypeScript、WXT、Vitest 和 `@/*` alias。
- Create: `extension/vitest.config.ts`
  - WXT Vitest plugin 和 jsdom 测试环境。
- Create: `extension/components.json`
  - shadcn-vue 配置。
- Create: `extension/src/style.css`
  - Tailwind v4 和 shadcn-vue CSS variables。
- Create: `extension/src/lib/utils.ts`
  - shadcn-vue `cn` helper。
- Create: `extension/entrypoints/background.ts`
  - MV3 background message listener。
- Create: `extension/entrypoints/youtube.content.ts`
  - YouTube watch content script 和 Shadow DOM Vue UI 挂载。
- Create: `extension/entrypoints/popup/index.html`
  - popup HTML entrypoint。
- Create: `extension/entrypoints/popup/main.ts`
  - popup Vue mount。
- Create: `extension/entrypoints/popup/App.vue`
  - popup 表单和 job 状态 UI。
- Create: `extension/src/api/messages.ts`
  - extension 消息类型、结果类型、语言和模式类型、runtime message wrapper。
- Create: `extension/src/api/backend-client.ts`
  - Go backend HTTP client。
- Create: `extension/src/api/message-handler.ts`
  - background 消息处理和 storage/API 协调。
- Create: `extension/src/storage/settings.ts`
  - backend URL 和语言设置。
- Create: `extension/src/storage/subtitle-cache.ts`
  - 视频偏好和字幕资产缓存。
- Create: `extension/src/subtitles/vtt.ts`
  - WebVTT 解析。
- Create: `extension/src/subtitles/active-cue.ts`
  - 当前播放时间 cue 选择。
- Create: `extension/src/youtube/video-id.ts`
  - YouTube watch URL 中 `videoId` 解析。
- Create: `extension/src/youtube/page-watch.ts`
  - 当前页面 videoId 读取和 YouTube SPA 导航监听。
- Create: `extension/src/popup/form-validation.ts`
  - popup 创建 job 前的本地校验。
- Create: `extension/src/content/YoutubeOverlay.vue`
  - 播放页控制条和字幕层。
- Create: `extension/README.md`
  - extension 本地开发、测试、构建和联调说明。
- Test: `extension/src/youtube/video-id.test.ts`
- Test: `extension/src/storage/settings.test.ts`
- Test: `extension/src/storage/subtitle-cache.test.ts`
- Test: `extension/src/api/backend-client.test.ts`
- Test: `extension/src/api/message-handler.test.ts`
- Test: `extension/src/subtitles/vtt.test.ts`
- Test: `extension/src/subtitles/active-cue.test.ts`
- Test: `extension/src/popup/form-validation.test.ts`
- Test: `extension/src/youtube/page-watch.test.ts`

## Task 1: Scaffold WXT, Vue, Vitest, Tailwind, shadcn-vue

**Files:**
- Modify: `mise.toml`
- Create: `extension/package.json`
- Create: `extension/wxt.config.ts`
- Create: `extension/tsconfig.json`
- Create: `extension/vitest.config.ts`
- Create: `extension/components.json`
- Create: `extension/src/style.css`
- Create: `extension/src/lib/utils.ts`
- Create: `extension/entrypoints/background.ts`
- Create: `extension/entrypoints/youtube.content.ts`
- Create: `extension/entrypoints/popup/index.html`
- Create: `extension/entrypoints/popup/main.ts`
- Create: `extension/entrypoints/popup/App.vue`
- Generated: `extension/package-lock.json`
- Generated: `extension/src/components/ui/**`

- [ ] **Step 1: 更新根工具链**

Modify `mise.toml`:

```toml
[tools]
python = "3.12"
uv = "latest"
go = "1.22"
node = "22"
```

- [ ] **Step 2: 创建 npm package scaffold**

Create `extension/package.json`:

```json
{
  "name": "lets-sub-it-extension",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "wxt",
    "build": "wxt build --browser chrome",
    "test": "vitest run",
    "typecheck": "vue-tsc --noEmit",
    "postinstall": "wxt prepare"
  }
}
```

- [ ] **Step 3: 安装 extension 依赖**

Run:

```bash
cd extension
mise exec -- npm install vue class-variance-authority clsx tailwind-merge lucide-vue-next reka-ui
mise exec -- npm install -D wxt typescript vue-tsc vitest @vitejs/plugin-vue @vue/test-utils jsdom @types/node tailwindcss @tailwindcss/vite shadcn-vue
```

Expected:

```text
package-lock.json 被生成，npm 输出 0 个 audit 级别的安装阻断错误。
```

- [ ] **Step 4: 创建 WXT、TypeScript、Vitest 和 shadcn-vue 配置**

Create `extension/wxt.config.ts`:

```ts
import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'wxt'

export default defineConfig({
  manifest: {
    name: 'Lets Sub It',
    description: 'Self-hosted YouTube subtitle generation and translation helper.',
    version: '0.1.0',
    permissions: ['storage', 'activeTab'],
    host_permissions: ['http://127.0.0.1:*/*', 'http://localhost:*/*'],
  },
  vite: () => ({
    plugins: [vue(), tailwindcss()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
  }),
})
```

Create `extension/tsconfig.json`:

```json
{
  "extends": "./.wxt/tsconfig.json",
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    },
    "types": ["vitest/globals", "vitest/jsdom"]
  }
}
```

Create `extension/vitest.config.ts`:

```ts
import { defineConfig } from 'vitest/config'
import { WxtVitest } from 'wxt/testing/vitest-plugin'

export default defineConfig({
  plugins: [WxtVitest()],
  test: {
    environment: 'jsdom',
    globals: true,
    clearMocks: true,
    restoreMocks: true,
  },
})
```

Create `extension/components.json`:

```json
{
  "$schema": "https://shadcn-vue.com/schema.json",
  "style": "new-york",
  "typescript": true,
  "tsConfigPath": "tsconfig.json",
  "tailwind": {
    "config": "",
    "css": "src/style.css",
    "baseColor": "neutral",
    "cssVariables": true,
    "prefix": ""
  },
  "aliases": {
    "components": "@/components",
    "composables": "@/composables",
    "utils": "@/lib/utils",
    "ui": "@/components/ui",
    "lib": "@/lib"
  },
  "iconLibrary": "lucide"
}
```

- [ ] **Step 5: 创建样式和 shadcn-vue utility**

Create `extension/src/style.css`:

```css
@import "tailwindcss";

:root {
  --background: 0 0% 100%;
  --foreground: 240 10% 3.9%;
  --card: 0 0% 100%;
  --card-foreground: 240 10% 3.9%;
  --popover: 0 0% 100%;
  --popover-foreground: 240 10% 3.9%;
  --primary: 240 5.9% 10%;
  --primary-foreground: 0 0% 98%;
  --secondary: 240 4.8% 95.9%;
  --secondary-foreground: 240 5.9% 10%;
  --muted: 240 4.8% 95.9%;
  --muted-foreground: 240 3.8% 46.1%;
  --accent: 240 4.8% 95.9%;
  --accent-foreground: 240 5.9% 10%;
  --destructive: 0 84.2% 60.2%;
  --destructive-foreground: 0 0% 98%;
  --border: 240 5.9% 90%;
  --input: 240 5.9% 90%;
  --ring: 240 10% 3.9%;
  --radius: 0.5rem;
}

* {
  border-color: hsl(var(--border));
}

body {
  background: hsl(var(--background));
  color: hsl(var(--foreground));
  font-family:
    Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI",
    sans-serif;
}
```

Create `extension/src/lib/utils.ts`:

```ts
import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

- [ ] **Step 6: 创建最小 entrypoints**

Create `extension/entrypoints/background.ts`:

```ts
export default defineBackground(() => {
  console.info('Lets Sub It background loaded')
})
```

Create `extension/entrypoints/youtube.content.ts`:

```ts
export default defineContentScript({
  matches: ['https://www.youtube.com/watch*'],
  main() {
    console.info('Lets Sub It YouTube content script loaded')
  },
})
```

Create `extension/entrypoints/popup/index.html`:

```html
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Lets Sub It</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="./main.ts"></script>
  </body>
</html>
```

Create `extension/entrypoints/popup/main.ts`:

```ts
import '@/style.css'
import { createApp } from 'vue'
import App from './App.vue'

createApp(App).mount('#app')
```

Create `extension/entrypoints/popup/App.vue`:

```vue
<template>
  <main class="w-[360px] p-4">
    <h1 class="text-base font-semibold">Lets Sub It</h1>
    <p class="mt-2 text-sm text-muted-foreground">Extension scaffold is ready.</p>
  </main>
</template>
```

- [ ] **Step 7: 生成 shadcn-vue 组件源码**

Run:

```bash
cd extension
mise exec -- npx shadcn-vue@latest add button input select card alert badge separator
```

Expected:

```text
src/components/ui/button
src/components/ui/input
src/components/ui/select
src/components/ui/card
src/components/ui/alert
src/components/ui/badge
src/components/ui/separator
```

- [ ] **Step 8: 验证 scaffold 可构建和测试命令可运行**

Run:

```bash
cd extension
mise exec -- npm run test
mise exec -- npm run build
```

Expected:

```text
Vitest exits 0 with no test files or an empty test suite message accepted by the current Vitest version.
WXT builds Chrome output under .output/chrome-mv3.
```

If Vitest exits non-zero because no tests exist, add `extension/src/scaffold.test.ts`:

```ts
import { describe, expect, it } from 'vitest'

describe('scaffold', () => {
  it('runs the test harness', () => {
    expect(true).toBe(true)
  })
})
```

Run again:

```bash
cd extension
mise exec -- npm run test
mise exec -- npm run build
```

Expected:

```text
1 test passes and WXT build exits 0.
```

- [ ] **Step 9: 提交 scaffold**

Run:

```bash
git add mise.toml extension
git commit -m "build(extension): scaffold WXT Vue project"
```

## Task 2: Add Shared Types And YouTube Video ID Parser

**Files:**
- Create: `extension/src/api/messages.ts`
- Create: `extension/src/youtube/video-id.ts`
- Test: `extension/src/youtube/video-id.test.ts`

- [ ] **Step 1: 写 YouTube videoId 解析测试**

Create `extension/src/youtube/video-id.test.ts`:

```ts
import { describe, expect, it } from 'vitest'
import { parseYouTubeWatchVideoId } from './video-id'

describe('parseYouTubeWatchVideoId', () => {
  it('extracts videoId from a YouTube watch URL', () => {
    const videoId = parseYouTubeWatchVideoId(
      'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
    )

    expect(videoId).toBe('dQw4w9WgXcQ')
  })

  it('extracts videoId from youtube.com without www', () => {
    const videoId = parseYouTubeWatchVideoId(
      'https://youtube.com/watch?v=dQw4w9WgXcQ&t=42',
    )

    expect(videoId).toBe('dQw4w9WgXcQ')
  })

  it('returns null for non-watch YouTube URLs', () => {
    const videoId = parseYouTubeWatchVideoId('https://www.youtube.com/shorts/abc')

    expect(videoId).toBeNull()
  })

  it('returns null for unsupported hosts', () => {
    const videoId = parseYouTubeWatchVideoId('https://example.com/watch?v=abc')

    expect(videoId).toBeNull()
  })

  it('returns null for invalid URLs', () => {
    const videoId = parseYouTubeWatchVideoId('not a url')

    expect(videoId).toBeNull()
  })
})
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
cd extension
mise exec -- npm run test -- src/youtube/video-id.test.ts
```

Expected:

```text
FAIL because src/youtube/video-id.ts does not exist.
```

- [ ] **Step 3: 添加消息类型和语言约束**

Create `extension/src/api/messages.ts`:

```ts
import { browser } from 'wxt/browser'

export const SUPPORTED_LANGUAGES = ['en', 'zh-CN'] as const
export type LanguageCode = (typeof SUPPORTED_LANGUAGES)[number]

export const SUBTITLE_MODES = ['translated', 'bilingual'] as const
export type SubtitleMode = (typeof SUBTITLE_MODES)[number]

export type JobStatus =
  | 'queued'
  | 'downloading'
  | 'transcribing'
  | 'translating'
  | 'packaging'
  | 'completed'
  | 'failed'

export type Job = {
  id: string
  videoId: string
  youtubeUrl: string
  sourceLanguage: LanguageCode
  targetLanguage: LanguageCode
  status: JobStatus
  stage: JobStatus
  progressText: string
  errorMessage: string | null
  createdAt: string
  updatedAt: string
}

export type SubtitleAsset = {
  jobId: string
  videoId: string
  targetLanguage: LanguageCode
  sourceLanguage: LanguageCode
  files: {
    source: string
    translated: string
    bilingual: string
  }
  createdAt: string
}

export type Settings = {
  backendBaseUrl: string
  sourceLanguage: LanguageCode
  targetLanguage: LanguageCode
}

export type CreateJobInput = {
  youtubeUrl: string
  sourceLanguage: LanguageCode
  targetLanguage: LanguageCode
}

export type ExtensionMessage =
  | { type: 'settings:get' }
  | { type: 'settings:update'; payload: Partial<Settings> }
  | { type: 'job:create'; payload: CreateJobInput }
  | { type: 'job:get'; payload: { jobId: string } }
  | { type: 'subtitle:resolve'; payload: { videoId: string } }
  | {
      type: 'subtitle:fetch-file'
      payload: { jobId: string; mode: SubtitleMode }
    }
  | {
      type: 'subtitle:update-mode'
      payload: { videoId: string; targetLanguage: LanguageCode; mode: SubtitleMode }
    }

export type MessageError = {
  code: string
  message: string
}

export type MessageResult<T> =
  | { ok: true; data: T }
  | { ok: false; error: MessageError }

export function isSupportedLanguage(value: string): value is LanguageCode {
  return SUPPORTED_LANGUAGES.includes(value as LanguageCode)
}

export function assertDifferentLanguages(source: LanguageCode, target: LanguageCode) {
  if (source === target) {
    throw new Error('sourceLanguage and targetLanguage must be different')
  }
}

export async function sendExtensionMessage<T>(
  message: ExtensionMessage,
): Promise<MessageResult<T>> {
  return browser.runtime.sendMessage(message) as Promise<MessageResult<T>>
}
```

- [ ] **Step 4: 添加 YouTube watch parser**

Create `extension/src/youtube/video-id.ts`:

```ts
export function parseYouTubeWatchVideoId(input: string): string | null {
  let url: URL
  try {
    url = new URL(input)
  } catch {
    return null
  }

  const host = url.hostname.toLowerCase()
  const isYouTubeHost = host === 'youtube.com' || host === 'www.youtube.com'
  if (!isYouTubeHost || url.pathname !== '/watch') {
    return null
  }

  const videoId = url.searchParams.get('v')
  if (!videoId) {
    return null
  }

  return videoId
}
```

- [ ] **Step 5: 运行测试确认通过**

Run:

```bash
cd extension
mise exec -- npm run test -- src/youtube/video-id.test.ts
```

Expected:

```text
5 tests pass.
```

- [ ] **Step 6: 提交共享类型和 videoId parser**

Run:

```bash
git add extension/src/api/messages.ts extension/src/youtube/video-id.ts extension/src/youtube/video-id.test.ts
git commit -m "feat(extension): add shared message types and video parser"
```

## Task 3: Add Settings And Subtitle Cache Storage

**Files:**
- Create: `extension/src/storage/settings.ts`
- Create: `extension/src/storage/subtitle-cache.ts`
- Test: `extension/src/storage/settings.test.ts`
- Test: `extension/src/storage/subtitle-cache.test.ts`

- [ ] **Step 1: 写 settings storage 测试**

Create `extension/src/storage/settings.test.ts`:

```ts
import { beforeEach, describe, expect, it } from 'vitest'
import { fakeBrowser } from 'wxt/testing/fake-browser'
import { DEFAULT_SETTINGS, getSettings, updateSettings } from './settings'

describe('settings storage', () => {
  beforeEach(() => {
    fakeBrowser.reset()
  })

  it('returns default settings when storage is empty', async () => {
    await expect(getSettings()).resolves.toEqual(DEFAULT_SETTINGS)
  })

  it('updates backend URL and language settings', async () => {
    const settings = await updateSettings({
      backendBaseUrl: 'http://localhost:9090',
      sourceLanguage: 'zh-CN',
      targetLanguage: 'en',
    })

    expect(settings).toEqual({
      backendBaseUrl: 'http://localhost:9090',
      sourceLanguage: 'zh-CN',
      targetLanguage: 'en',
    })
    await expect(getSettings()).resolves.toEqual(settings)
  })

  it('rejects same source and target languages', async () => {
    await expect(
      updateSettings({ sourceLanguage: 'en', targetLanguage: 'en' }),
    ).rejects.toThrow('sourceLanguage and targetLanguage must be different')
  })
})
```

- [ ] **Step 2: 写 subtitle cache storage 测试**

Create `extension/src/storage/subtitle-cache.test.ts`:

```ts
import { beforeEach, describe, expect, it } from 'vitest'
import { fakeBrowser } from 'wxt/testing/fake-browser'
import type { SubtitleAsset } from '@/api/messages'
import {
  getCachedSubtitleAsset,
  getVideoPreference,
  setCachedSubtitleAsset,
  updateCachedSubtitleMode,
} from './subtitle-cache'

const asset: SubtitleAsset = {
  jobId: 'job_123',
  videoId: 'video_123',
  sourceLanguage: 'en',
  targetLanguage: 'zh-CN',
  files: {
    source: '/subtitle-files/job_123/source',
    translated: '/subtitle-files/job_123/translated',
    bilingual: '/subtitle-files/job_123/bilingual',
  },
  createdAt: '2026-04-25T00:00:00Z',
}

describe('subtitle cache storage', () => {
  beforeEach(() => {
    fakeBrowser.reset()
  })

  it('stores and reads an asset by videoId and targetLanguage', async () => {
    await setCachedSubtitleAsset(asset, 'translated', '2026-04-25T00:01:00Z')

    await expect(getCachedSubtitleAsset('video_123', 'zh-CN')).resolves.toEqual({
      ...asset,
      selectedMode: 'translated',
      lastSyncedAt: '2026-04-25T00:01:00Z',
    })
  })

  it('stores the video preference when an asset is cached', async () => {
    await setCachedSubtitleAsset(asset, 'translated', '2026-04-25T00:01:00Z')

    await expect(getVideoPreference('video_123')).resolves.toEqual({
      videoId: 'video_123',
      targetLanguage: 'zh-CN',
      selectedMode: 'translated',
    })
  })

  it('updates the selected subtitle mode', async () => {
    await setCachedSubtitleAsset(asset, 'translated', '2026-04-25T00:01:00Z')

    const updated = await updateCachedSubtitleMode('video_123', 'zh-CN', 'bilingual')

    expect(updated?.selectedMode).toBe('bilingual')
    await expect(getVideoPreference('video_123')).resolves.toEqual({
      videoId: 'video_123',
      targetLanguage: 'zh-CN',
      selectedMode: 'bilingual',
    })
  })

  it('separates cache entries by targetLanguage', async () => {
    await setCachedSubtitleAsset(asset, 'translated', '2026-04-25T00:01:00Z')

    await expect(getCachedSubtitleAsset('video_123', 'en')).resolves.toBeNull()
  })
})
```

- [ ] **Step 3: 运行 storage 测试确认失败**

Run:

```bash
cd extension
mise exec -- npm run test -- src/storage/settings.test.ts src/storage/subtitle-cache.test.ts
```

Expected:

```text
FAIL because storage modules do not exist.
```

- [ ] **Step 4: 实现 settings storage**

Create `extension/src/storage/settings.ts`:

```ts
import { storage } from 'wxt/storage'
import {
  assertDifferentLanguages,
  type LanguageCode,
  type Settings,
} from '@/api/messages'

export const DEFAULT_SETTINGS: Settings = {
  backendBaseUrl: 'http://127.0.0.1:8080',
  sourceLanguage: 'en',
  targetLanguage: 'zh-CN',
}

const settingsItem = storage.defineItem<Settings>('local:settings', {
  fallback: DEFAULT_SETTINGS,
})

export async function getSettings(): Promise<Settings> {
  return settingsItem.getValue()
}

export async function updateSettings(patch: Partial<Settings>): Promise<Settings> {
  const current = await getSettings()
  const next: Settings = {
    ...current,
    ...patch,
  }

  assertDifferentLanguages(next.sourceLanguage, next.targetLanguage)
  await settingsItem.setValue(next)
  return next
}

export function createLanguagePair(
  sourceLanguage: LanguageCode,
  targetLanguage: LanguageCode,
) {
  assertDifferentLanguages(sourceLanguage, targetLanguage)
  return { sourceLanguage, targetLanguage }
}
```

- [ ] **Step 5: 实现 subtitle cache storage**

Create `extension/src/storage/subtitle-cache.ts`:

```ts
import { storage } from 'wxt/storage'
import type { LanguageCode, SubtitleAsset, SubtitleMode } from '@/api/messages'

export type VideoPreference = {
  videoId: string
  targetLanguage: LanguageCode
  selectedMode: SubtitleMode
}

export type SubtitleAssetCacheEntry = SubtitleAsset & {
  selectedMode: SubtitleMode
  lastSyncedAt: string
}

function subtitleAssetKey(videoId: string, targetLanguage: LanguageCode) {
  return `local:subtitleAssets:${videoId}:${targetLanguage}` as const
}

function videoPreferenceKey(videoId: string) {
  return `local:videoPreferences:${videoId}` as const
}

export async function getCachedSubtitleAsset(
  videoId: string,
  targetLanguage: LanguageCode,
): Promise<SubtitleAssetCacheEntry | null> {
  const item = storage.defineItem<SubtitleAssetCacheEntry>(
    subtitleAssetKey(videoId, targetLanguage),
  )
  return item.getValue()
}

export async function setCachedSubtitleAsset(
  asset: SubtitleAsset,
  selectedMode: SubtitleMode,
  lastSyncedAt: string,
): Promise<SubtitleAssetCacheEntry> {
  const entry: SubtitleAssetCacheEntry = {
    ...asset,
    selectedMode,
    lastSyncedAt,
  }
  const item = storage.defineItem<SubtitleAssetCacheEntry>(
    subtitleAssetKey(asset.videoId, asset.targetLanguage),
  )
  await item.setValue(entry)
  await setVideoPreference({
    videoId: asset.videoId,
    targetLanguage: asset.targetLanguage,
    selectedMode,
  })
  return entry
}

export async function getVideoPreference(
  videoId: string,
): Promise<VideoPreference | null> {
  const item = storage.defineItem<VideoPreference>(videoPreferenceKey(videoId))
  return item.getValue()
}

export async function setVideoPreference(
  preference: VideoPreference,
): Promise<VideoPreference> {
  const item = storage.defineItem<VideoPreference>(
    videoPreferenceKey(preference.videoId),
  )
  await item.setValue(preference)
  return preference
}

export async function updateCachedSubtitleMode(
  videoId: string,
  targetLanguage: LanguageCode,
  mode: SubtitleMode,
): Promise<SubtitleAssetCacheEntry | null> {
  const current = await getCachedSubtitleAsset(videoId, targetLanguage)
  if (!current) {
    return null
  }

  const next: SubtitleAssetCacheEntry = {
    ...current,
    selectedMode: mode,
  }
  const item = storage.defineItem<SubtitleAssetCacheEntry>(
    subtitleAssetKey(videoId, targetLanguage),
  )
  await item.setValue(next)
  await setVideoPreference({ videoId, targetLanguage, selectedMode: mode })
  return next
}
```

- [ ] **Step 6: 运行 storage 测试确认通过**

Run:

```bash
cd extension
mise exec -- npm run test -- src/storage/settings.test.ts src/storage/subtitle-cache.test.ts
```

Expected:

```text
7 tests pass.
```

- [ ] **Step 7: 提交 storage 模块**

Run:

```bash
git add extension/src/storage extension/src/api/messages.ts
git commit -m "feat(extension): add settings and subtitle cache storage"
```

## Task 4: Add Backend HTTP Client

**Files:**
- Create: `extension/src/api/backend-client.ts`
- Test: `extension/src/api/backend-client.test.ts`

- [ ] **Step 1: 写 backend client 测试**

Create `extension/src/api/backend-client.test.ts`:

```ts
import { describe, expect, it, vi } from 'vitest'
import { BackendClientError, createBackendClient } from './backend-client'

describe('createBackendClient', () => {
  it('creates a job through POST /jobs', async () => {
    const fetchImpl = vi.fn(async () =>
      new Response(
        JSON.stringify({
          job: {
            id: 'job_123',
            videoId: 'video_123',
            youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
            sourceLanguage: 'en',
            targetLanguage: 'zh-CN',
            status: 'queued',
            stage: 'queued',
            progressText: '等待处理',
            errorMessage: null,
            createdAt: '2026-04-25T00:00:00Z',
            updatedAt: '2026-04-25T00:00:00Z',
          },
          reused: false,
        }),
        { status: 201, headers: { 'Content-Type': 'application/json' } },
      ),
    )
    const client = createBackendClient('http://127.0.0.1:8080/', fetchImpl)

    const response = await client.createJob({
      youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
      sourceLanguage: 'en',
      targetLanguage: 'zh-CN',
    })

    expect(response.reused).toBe(false)
    expect(response.job.id).toBe('job_123')
    expect(fetchImpl).toHaveBeenCalledWith('http://127.0.0.1:8080/jobs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
        sourceLanguage: 'en',
        targetLanguage: 'zh-CN',
      }),
    })
  })

  it('fetches VTT text from subtitle file endpoint', async () => {
    const fetchImpl = vi.fn(async () => new Response('WEBVTT\n', { status: 200 }))
    const client = createBackendClient('http://localhost:8080', fetchImpl)

    await expect(client.fetchSubtitleFile('job_123', 'translated')).resolves.toBe(
      'WEBVTT\n',
    )
  })

  it('converts backend JSON errors into BackendClientError', async () => {
    const fetchImpl = vi.fn(async () =>
      new Response(
        JSON.stringify({
          error: { code: 'invalid_request', message: 'bad input' },
        }),
        { status: 400, headers: { 'Content-Type': 'application/json' } },
      ),
    )
    const client = createBackendClient('http://127.0.0.1:8080', fetchImpl)

    await expect(client.getJob('missing')).rejects.toMatchObject({
      code: 'invalid_request',
      message: 'bad input',
    })
  })

  it('rejects non-local backend URLs', () => {
    expect(() => createBackendClient('https://api.example.com', fetch)).toThrow(
      'backendBaseUrl must use localhost or 127.0.0.1',
    )
  })

  it('uses network_error for failed fetch calls', async () => {
    const fetchImpl = vi.fn(async () => {
      throw new TypeError('failed to fetch')
    })
    const client = createBackendClient('http://127.0.0.1:8080', fetchImpl)

    await expect(client.getJob('job_123')).rejects.toBeInstanceOf(BackendClientError)
    await expect(client.getJob('job_123')).rejects.toMatchObject({
      code: 'network_error',
    })
  })
})
```

- [ ] **Step 2: 运行 backend client 测试确认失败**

Run:

```bash
cd extension
mise exec -- npm run test -- src/api/backend-client.test.ts
```

Expected:

```text
FAIL because src/api/backend-client.ts does not exist.
```

- [ ] **Step 3: 实现 backend client**

Create `extension/src/api/backend-client.ts`:

```ts
import type {
  CreateJobInput,
  Job,
  LanguageCode,
  SubtitleAsset,
  SubtitleMode,
} from './messages'

export type CreateJobResponse = {
  job: Job
  reused: boolean
}

export type GetJobResponse = {
  job: Job
}

export type GetSubtitleAssetResponse = {
  asset: SubtitleAsset | null
}

type FetchLike = typeof fetch

export class BackendClientError extends Error {
  constructor(
    readonly code: string,
    message: string,
  ) {
    super(message)
    this.name = 'BackendClientError'
  }
}

export type BackendClient = {
  createJob(input: CreateJobInput): Promise<CreateJobResponse>
  getJob(jobId: string): Promise<GetJobResponse>
  getSubtitleAsset(
    videoId: string,
    targetLanguage: LanguageCode,
  ): Promise<GetSubtitleAssetResponse>
  fetchSubtitleFile(jobId: string, mode: SubtitleMode): Promise<string>
}

export function createBackendClient(
  backendBaseUrl: string,
  fetchImpl: FetchLike = fetch,
): BackendClient {
  const baseUrl = normalizeBackendBaseUrl(backendBaseUrl)

  return {
    createJob(input) {
      return requestJson<CreateJobResponse>(fetchImpl, `${baseUrl}/jobs`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(input),
      })
    },
    getJob(jobId) {
      return requestJson<GetJobResponse>(
        fetchImpl,
        `${baseUrl}/jobs/${encodeURIComponent(jobId)}`,
      )
    },
    getSubtitleAsset(videoId, targetLanguage) {
      const params = new URLSearchParams({ videoId, targetLanguage })
      return requestJson<GetSubtitleAssetResponse>(
        fetchImpl,
        `${baseUrl}/subtitle-assets?${params.toString()}`,
      )
    },
    async fetchSubtitleFile(jobId, mode) {
      const response = await request(
        fetchImpl,
        `${baseUrl}/subtitle-files/${encodeURIComponent(jobId)}/${mode}`,
      )
      return response.text()
    },
  }
}

function normalizeBackendBaseUrl(value: string): string {
  let url: URL
  try {
    url = new URL(value)
  } catch {
    throw new BackendClientError('invalid_backend_url', 'backendBaseUrl is invalid')
  }

  const isLocalHost = url.hostname === '127.0.0.1' || url.hostname === 'localhost'
  if (!isLocalHost || url.protocol !== 'http:') {
    throw new BackendClientError(
      'invalid_backend_url',
      'backendBaseUrl must use localhost or 127.0.0.1',
    )
  }

  return url.toString().replace(/\/$/, '')
}

async function requestJson<T>(
  fetchImpl: FetchLike,
  url: string,
  init?: RequestInit,
): Promise<T> {
  const response = await request(fetchImpl, url, init)
  return response.json() as Promise<T>
}

async function request(
  fetchImpl: FetchLike,
  url: string,
  init?: RequestInit,
): Promise<Response> {
  let response: Response
  try {
    response = await fetchImpl(url, init)
  } catch {
    throw new BackendClientError('network_error', 'Cannot connect to backend')
  }

  if (!response.ok) {
    throw await errorFromResponse(response)
  }

  return response
}

async function errorFromResponse(response: Response): Promise<BackendClientError> {
  try {
    const payload = (await response.json()) as {
      error?: { code?: string; message?: string }
    }
    if (payload.error?.code && payload.error.message) {
      return new BackendClientError(payload.error.code, payload.error.message)
    }
  } catch {
    return new BackendClientError('backend_error', `Backend returned ${response.status}`)
  }

  return new BackendClientError('backend_error', `Backend returned ${response.status}`)
}
```

- [ ] **Step 4: 修正 failed fetch 测试里的重复调用**

Modify the final test in `extension/src/api/backend-client.test.ts`:

```ts
  it('uses network_error for failed fetch calls', async () => {
    const fetchImpl = vi.fn(async () => {
      throw new TypeError('failed to fetch')
    })
    const client = createBackendClient('http://127.0.0.1:8080', fetchImpl)

    const promise = client.getJob('job_123')

    await expect(promise).rejects.toBeInstanceOf(BackendClientError)
    await expect(promise).rejects.toMatchObject({
      code: 'network_error',
    })
  })
```

- [ ] **Step 5: 运行 backend client 测试确认通过**

Run:

```bash
cd extension
mise exec -- npm run test -- src/api/backend-client.test.ts
```

Expected:

```text
5 tests pass.
```

- [ ] **Step 6: 提交 backend client**

Run:

```bash
git add extension/src/api/backend-client.ts extension/src/api/backend-client.test.ts
git commit -m "feat(extension): add backend API client"
```

## Task 5: Add Background Message Handler

**Files:**
- Create: `extension/src/api/message-handler.ts`
- Modify: `extension/entrypoints/background.ts`
- Test: `extension/src/api/message-handler.test.ts`

- [ ] **Step 1: 写 message handler 测试**

Create `extension/src/api/message-handler.test.ts`:

```ts
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { fakeBrowser } from 'wxt/testing/fake-browser'
import { getCachedSubtitleAsset } from '@/storage/subtitle-cache'
import { handleExtensionMessage } from './message-handler'

describe('handleExtensionMessage', () => {
  beforeEach(() => {
    fakeBrowser.reset()
  })

  it('rejects job creation when source and target languages are equal', async () => {
    const fetchImpl = vi.fn()

    const result = await handleExtensionMessage(
      {
        type: 'job:create',
        payload: {
          youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
          sourceLanguage: 'en',
          targetLanguage: 'en',
        },
      },
      { fetchImpl, now: () => '2026-04-25T00:00:00Z' },
    )

    expect(result).toEqual({
      ok: false,
      error: {
        code: 'invalid_language_pair',
        message: 'sourceLanguage and targetLanguage must be different',
      },
    })
    expect(fetchImpl).not.toHaveBeenCalled()
  })

  it('creates a job through the backend client', async () => {
    const fetchImpl = vi.fn(async () =>
      new Response(
        JSON.stringify({
          job: {
            id: 'job_123',
            videoId: 'video_123',
            youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
            sourceLanguage: 'en',
            targetLanguage: 'zh-CN',
            status: 'queued',
            stage: 'queued',
            progressText: '等待处理',
            errorMessage: null,
            createdAt: '2026-04-25T00:00:00Z',
            updatedAt: '2026-04-25T00:00:00Z',
          },
          reused: false,
        }),
        { status: 201, headers: { 'Content-Type': 'application/json' } },
      ),
    )

    const result = await handleExtensionMessage(
      {
        type: 'job:create',
        payload: {
          youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
          sourceLanguage: 'en',
          targetLanguage: 'zh-CN',
        },
      },
      { fetchImpl, now: () => '2026-04-25T00:00:00Z' },
    )

    expect(result.ok).toBe(true)
    expect(fetchImpl).toHaveBeenCalledOnce()
  })

  it('resolves and caches a subtitle asset from backend when local cache is empty', async () => {
    const fetchImpl = vi.fn(async () =>
      new Response(
        JSON.stringify({
          asset: {
            jobId: 'job_123',
            videoId: 'video_123',
            sourceLanguage: 'en',
            targetLanguage: 'zh-CN',
            files: {
              source: '/subtitle-files/job_123/source',
              translated: '/subtitle-files/job_123/translated',
              bilingual: '/subtitle-files/job_123/bilingual',
            },
            createdAt: '2026-04-25T00:00:00Z',
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    )

    const result = await handleExtensionMessage(
      { type: 'subtitle:resolve', payload: { videoId: 'video_123' } },
      { fetchImpl, now: () => '2026-04-25T00:01:00Z' },
    )

    expect(result.ok).toBe(true)
    await expect(getCachedSubtitleAsset('video_123', 'zh-CN')).resolves.toMatchObject({
      jobId: 'job_123',
      selectedMode: 'translated',
    })
  })
})
```

- [ ] **Step 2: 运行 message handler 测试确认失败**

Run:

```bash
cd extension
mise exec -- npm run test -- src/api/message-handler.test.ts
```

Expected:

```text
FAIL because src/api/message-handler.ts does not exist.
```

- [ ] **Step 3: 实现 message handler**

Create `extension/src/api/message-handler.ts`:

```ts
import { BackendClientError, createBackendClient } from './backend-client'
import {
  assertDifferentLanguages,
  type ExtensionMessage,
  type MessageError,
  type MessageResult,
} from './messages'
import { getSettings, updateSettings } from '@/storage/settings'
import {
  getCachedSubtitleAsset,
  getVideoPreference,
  setCachedSubtitleAsset,
  updateCachedSubtitleMode,
} from '@/storage/subtitle-cache'

export type MessageHandlerDeps = {
  fetchImpl?: typeof fetch
  now?: () => string
}

export async function handleExtensionMessage(
  message: ExtensionMessage,
  deps: MessageHandlerDeps = {},
): Promise<MessageResult<unknown>> {
  try {
    const fetchImpl = deps.fetchImpl ?? fetch
    const now = deps.now ?? (() => new Date().toISOString())

    switch (message.type) {
      case 'settings:get':
        return ok(await getSettings())
      case 'settings:update':
        return ok(await updateSettings(message.payload))
      case 'job:create': {
        assertDifferentLanguages(
          message.payload.sourceLanguage,
          message.payload.targetLanguage,
        )
        await updateSettings({
          sourceLanguage: message.payload.sourceLanguage,
          targetLanguage: message.payload.targetLanguage,
        })
        const client = await clientFromSettings(fetchImpl)
        return ok(await client.createJob(message.payload))
      }
      case 'job:get': {
        const client = await clientFromSettings(fetchImpl)
        return ok(await client.getJob(message.payload.jobId))
      }
      case 'subtitle:resolve': {
        const settings = await getSettings()
        const preference = await getVideoPreference(message.payload.videoId)
        const targetLanguage = preference?.targetLanguage ?? settings.targetLanguage
        const cached = await getCachedSubtitleAsset(
          message.payload.videoId,
          targetLanguage,
        )
        if (cached) {
          return ok(cached)
        }

        const client = createBackendClient(settings.backendBaseUrl, fetchImpl)
        const response = await client.getSubtitleAsset(
          message.payload.videoId,
          targetLanguage,
        )
        if (!response.asset) {
          return ok(null)
        }

        const entry = await setCachedSubtitleAsset(
          response.asset,
          preference?.selectedMode ?? 'translated',
          now(),
        )
        return ok(entry)
      }
      case 'subtitle:fetch-file': {
        const client = await clientFromSettings(fetchImpl)
        return ok(
          await client.fetchSubtitleFile(message.payload.jobId, message.payload.mode),
        )
      }
      case 'subtitle:update-mode':
        return ok(
          await updateCachedSubtitleMode(
            message.payload.videoId,
            message.payload.targetLanguage,
            message.payload.mode,
          ),
        )
    }
  } catch (error) {
    return { ok: false, error: errorToMessage(error) }
  }
}

async function clientFromSettings(fetchImpl: typeof fetch) {
  const settings = await getSettings()
  return createBackendClient(settings.backendBaseUrl, fetchImpl)
}

function ok<T>(data: T): MessageResult<T> {
  return { ok: true, data }
}

function errorToMessage(error: unknown): MessageError {
  if (error instanceof BackendClientError) {
    return { code: error.code, message: error.message }
  }
  if (
    error instanceof Error &&
    error.message === 'sourceLanguage and targetLanguage must be different'
  ) {
    return {
      code: 'invalid_language_pair',
      message: 'sourceLanguage and targetLanguage must be different',
    }
  }
  if (error instanceof Error) {
    return { code: 'internal_error', message: error.message }
  }
  return { code: 'internal_error', message: 'Unknown extension error' }
}
```

- [ ] **Step 4: 接入 background runtime listener**

Replace `extension/entrypoints/background.ts`:

```ts
import { browser } from 'wxt/browser'
import { handleExtensionMessage } from '@/api/message-handler'
import type { ExtensionMessage } from '@/api/messages'

export default defineBackground(() => {
  browser.runtime.onMessage.addListener((message, _sender, sendResponse) => {
    handleExtensionMessage(message as ExtensionMessage).then(sendResponse)
    return true
  })
})
```

- [ ] **Step 5: 运行 message handler 测试确认通过**

Run:

```bash
cd extension
mise exec -- npm run test -- src/api/message-handler.test.ts
```

Expected:

```text
3 tests pass.
```

- [ ] **Step 6: 提交 background message handler**

Run:

```bash
git add extension/entrypoints/background.ts extension/src/api/message-handler.ts extension/src/api/message-handler.test.ts
git commit -m "feat(extension): route API calls through background"
```

## Task 6: Add WebVTT Parser And Active Cue Selection

**Files:**
- Create: `extension/src/subtitles/vtt.ts`
- Create: `extension/src/subtitles/active-cue.ts`
- Test: `extension/src/subtitles/vtt.test.ts`
- Test: `extension/src/subtitles/active-cue.test.ts`

- [ ] **Step 1: 写 WebVTT parser 测试**

Create `extension/src/subtitles/vtt.test.ts`:

```ts
import { describe, expect, it } from 'vitest'
import { parseVtt } from './vtt'

describe('parseVtt', () => {
  it('parses WebVTT cues', () => {
    const cues = parseVtt(`WEBVTT

00:00:00.000 --> 00:00:02.000
Hello

00:00:02.000 --> 00:00:04.500
World
`)

    expect(cues).toEqual([
      { start: 0, end: 2, text: 'Hello' },
      { start: 2, end: 4.5, text: 'World' },
    ])
  })

  it('keeps multiline cue text', () => {
    const cues = parseVtt(`WEBVTT

00:00:00.000 --> 00:00:02.000
Hello
你好
`)

    expect(cues[0].text).toBe('Hello\n你好')
  })

  it('throws when no cues are present', () => {
    expect(() => parseVtt('WEBVTT\n')).toThrow('VTT contains no cues')
  })
})
```

- [ ] **Step 2: 写 active cue 测试**

Create `extension/src/subtitles/active-cue.test.ts`:

```ts
import { describe, expect, it } from 'vitest'
import { findActiveCue } from './active-cue'
import type { VttCue } from './vtt'

const cues: VttCue[] = [
  { start: 0, end: 2, text: 'first' },
  { start: 2, end: 4, text: 'second' },
]

describe('findActiveCue', () => {
  it('returns the cue that contains the current time', () => {
    expect(findActiveCue(cues, 1)?.text).toBe('first')
  })

  it('treats cue end time as exclusive', () => {
    expect(findActiveCue(cues, 2)?.text).toBe('second')
  })

  it('returns null when no cue matches', () => {
    expect(findActiveCue(cues, 5)).toBeNull()
  })
})
```

- [ ] **Step 3: 运行字幕测试确认失败**

Run:

```bash
cd extension
mise exec -- npm run test -- src/subtitles/vtt.test.ts src/subtitles/active-cue.test.ts
```

Expected:

```text
FAIL because subtitle modules do not exist.
```

- [ ] **Step 4: 实现 WebVTT parser**

Create `extension/src/subtitles/vtt.ts`:

```ts
export type VttCue = {
  start: number
  end: number
  text: string
}

const TIMING_RE =
  /^(?<start>\d{2}:\d{2}:\d{2}\.\d{3})\s+-->\s+(?<end>\d{2}:\d{2}:\d{2}\.\d{3})/

export function parseVtt(input: string): VttCue[] {
  const lines = input.replace(/\r\n/g, '\n').split('\n')
  const cues: VttCue[] = []
  let index = 0

  while (index < lines.length) {
    const line = lines[index].trim()
    const match = TIMING_RE.exec(line)

    if (!match?.groups) {
      index += 1
      continue
    }

    const textLines: string[] = []
    index += 1
    while (index < lines.length && lines[index].trim() !== '') {
      textLines.push(lines[index])
      index += 1
    }

    const text = textLines.join('\n').trim()
    if (text) {
      cues.push({
        start: parseTimestamp(match.groups.start),
        end: parseTimestamp(match.groups.end),
        text,
      })
    }
  }

  if (cues.length === 0) {
    throw new Error('VTT contains no cues')
  }

  return cues
}

function parseTimestamp(value: string): number {
  const [hours, minutes, secondsWithMillis] = value.split(':')
  const [seconds, millis] = secondsWithMillis.split('.')
  return (
    Number(hours) * 3600 +
    Number(minutes) * 60 +
    Number(seconds) +
    Number(millis) / 1000
  )
}
```

- [ ] **Step 5: 实现 active cue selection**

Create `extension/src/subtitles/active-cue.ts`:

```ts
import type { VttCue } from './vtt'

export function findActiveCue(cues: VttCue[], currentTime: number): VttCue | null {
  return cues.find((cue) => cue.start <= currentTime && currentTime < cue.end) ?? null
}
```

- [ ] **Step 6: 运行字幕测试确认通过**

Run:

```bash
cd extension
mise exec -- npm run test -- src/subtitles/vtt.test.ts src/subtitles/active-cue.test.ts
```

Expected:

```text
6 tests pass.
```

- [ ] **Step 7: 提交字幕模块**

Run:

```bash
git add extension/src/subtitles
git commit -m "feat(extension): parse VTT subtitles"
```

## Task 7: Build Popup UI And Job Polling

**Files:**
- Create: `extension/src/popup/form-validation.ts`
- Test: `extension/src/popup/form-validation.test.ts`
- Modify: `extension/entrypoints/popup/App.vue`

- [ ] **Step 1: 写 popup 表单校验测试**

Create `extension/src/popup/form-validation.test.ts`:

```ts
import { describe, expect, it } from 'vitest'
import { validateCreateJobForm } from './form-validation'

describe('validateCreateJobForm', () => {
  it('accepts a valid form', () => {
    expect(
      validateCreateJobForm({
        backendBaseUrl: 'http://127.0.0.1:8080',
        youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
        sourceLanguage: 'en',
        targetLanguage: 'zh-CN',
      }),
    ).toBeNull()
  })

  it('rejects empty YouTube URL', () => {
    expect(
      validateCreateJobForm({
        backendBaseUrl: 'http://127.0.0.1:8080',
        youtubeUrl: '',
        sourceLanguage: 'en',
        targetLanguage: 'zh-CN',
      }),
    ).toBe('请输入 YouTube URL')
  })

  it('rejects equal source and target languages', () => {
    expect(
      validateCreateJobForm({
        backendBaseUrl: 'http://127.0.0.1:8080',
        youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
        sourceLanguage: 'en',
        targetLanguage: 'en',
      }),
    ).toBe('源语言和目标语言不能相同')
  })
})
```

- [ ] **Step 2: 运行 popup 校验测试确认失败**

Run:

```bash
cd extension
mise exec -- npm run test -- src/popup/form-validation.test.ts
```

Expected:

```text
FAIL because src/popup/form-validation.ts does not exist.
```

- [ ] **Step 3: 实现 popup 表单校验**

Create `extension/src/popup/form-validation.ts`:

```ts
import type { LanguageCode } from '@/api/messages'

export type CreateJobForm = {
  backendBaseUrl: string
  youtubeUrl: string
  sourceLanguage: LanguageCode
  targetLanguage: LanguageCode
}

export function validateCreateJobForm(form: CreateJobForm): string | null {
  if (!form.backendBaseUrl.trim()) {
    return '请输入 backend URL'
  }
  if (!form.youtubeUrl.trim()) {
    return '请输入 YouTube URL'
  }
  if (form.sourceLanguage === form.targetLanguage) {
    return '源语言和目标语言不能相同'
  }
  return null
}
```

- [ ] **Step 4: 替换 popup App**

Replace `extension/entrypoints/popup/App.vue`:

```vue
<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { browser } from 'wxt/browser'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import type { CreateJobInput, Job, LanguageCode, Settings } from '@/api/messages'
import { sendExtensionMessage } from '@/api/messages'
import { validateCreateJobForm } from '@/popup/form-validation'

const languages: LanguageCode[] = ['en', 'zh-CN']

const backendBaseUrl = ref('http://127.0.0.1:8080')
const youtubeUrl = ref('')
const sourceLanguage = ref<LanguageCode>('en')
const targetLanguage = ref<LanguageCode>('zh-CN')
const currentJob = ref<Job | null>(null)
const errorMessage = ref<string | null>(null)
const isSubmitting = ref(false)
const elapsedSeconds = ref(0)
let pollTimer: ReturnType<typeof window.setInterval> | null = null
let elapsedTimer: ReturnType<typeof window.setInterval> | null = null

const validationError = computed(() =>
  validateCreateJobForm({
    backendBaseUrl: backendBaseUrl.value,
    youtubeUrl: youtubeUrl.value,
    sourceLanguage: sourceLanguage.value,
    targetLanguage: targetLanguage.value,
  }),
)

const canSubmit = computed(() => !validationError.value && !isSubmitting.value)

onMounted(async () => {
  const settings = await sendExtensionMessage<Settings>({ type: 'settings:get' })
  if (settings.ok) {
    backendBaseUrl.value = settings.data.backendBaseUrl
    sourceLanguage.value = settings.data.sourceLanguage
    targetLanguage.value = settings.data.targetLanguage
  }

  const [tab] = await browser.tabs.query({ active: true, currentWindow: true })
  if (tab?.url?.startsWith('https://www.youtube.com/watch')) {
    youtubeUrl.value = tab.url
  }
})

onUnmounted(() => {
  stopPolling()
})

async function submitJob() {
  const formError = validationError.value
  if (formError) {
    errorMessage.value = formError
    return
  }

  isSubmitting.value = true
  errorMessage.value = null
  currentJob.value = null
  elapsedSeconds.value = 0

  const settingsResult = await sendExtensionMessage<Settings>({
    type: 'settings:update',
    payload: {
      backendBaseUrl: backendBaseUrl.value,
      sourceLanguage: sourceLanguage.value,
      targetLanguage: targetLanguage.value,
    },
  })
  if (!settingsResult.ok) {
    finishWithError(settingsResult.error.message)
    return
  }

  const createResult = await sendExtensionMessage<{ job: Job; reused: boolean }>({
    type: 'job:create',
    payload: {
      youtubeUrl: youtubeUrl.value,
      sourceLanguage: sourceLanguage.value,
      targetLanguage: targetLanguage.value,
    } satisfies CreateJobInput,
  })
  if (!createResult.ok) {
    finishWithError(createResult.error.message)
    return
  }

  currentJob.value = createResult.data.job
  startPolling(createResult.data.job.id)
}

function startPolling(jobId: string) {
  stopPolling()
  elapsedTimer = window.setInterval(() => {
    elapsedSeconds.value += 1
  }, 1000)
  pollTimer = window.setInterval(() => {
    pollJob(jobId)
  }, 1000)
  pollJob(jobId)
}

async function pollJob(jobId: string) {
  const result = await sendExtensionMessage<{ job: Job }>({
    type: 'job:get',
    payload: { jobId },
  })
  if (!result.ok) {
    finishWithError(result.error.message)
    return
  }

  currentJob.value = result.data.job
  if (result.data.job.status === 'completed') {
    await sendExtensionMessage({
      type: 'subtitle:resolve',
      payload: { videoId: result.data.job.videoId },
    })
    isSubmitting.value = false
    stopPolling()
  }
  if (result.data.job.status === 'failed') {
    finishWithError(result.data.job.errorMessage ?? '任务失败')
  }
}

function stopPolling() {
  if (pollTimer) {
    window.clearInterval(pollTimer)
    pollTimer = null
  }
  if (elapsedTimer) {
    window.clearInterval(elapsedTimer)
    elapsedTimer = null
  }
}

function finishWithError(message: string) {
  errorMessage.value = message
  isSubmitting.value = false
  stopPolling()
}
</script>

<template>
  <main class="w-[380px] bg-background p-4 text-foreground">
    <Card class="rounded-lg">
      <CardHeader class="space-y-1">
        <CardTitle class="text-base">Lets Sub It</CardTitle>
        <p class="text-xs text-muted-foreground">提交 YouTube 视频并加载本地字幕。</p>
      </CardHeader>
      <CardContent class="space-y-3">
        <label class="space-y-1 text-xs font-medium">
          <span>Backend URL</span>
          <Input v-model="backendBaseUrl" />
        </label>

        <label class="space-y-1 text-xs font-medium">
          <span>YouTube URL</span>
          <Input v-model="youtubeUrl" />
        </label>

        <div class="grid grid-cols-2 gap-2">
          <label class="space-y-1 text-xs font-medium">
            <span>源语言</span>
            <Select v-model="sourceLanguage">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="language in languages" :key="language" :value="language">
                  {{ language }}
                </SelectItem>
              </SelectContent>
            </Select>
          </label>

          <label class="space-y-1 text-xs font-medium">
            <span>目标语言</span>
            <Select v-model="targetLanguage">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="language in languages" :key="language" :value="language">
                  {{ language }}
                </SelectItem>
              </SelectContent>
            </Select>
          </label>
        </div>

        <Alert v-if="errorMessage || validationError" variant="destructive">
          <AlertDescription>{{ errorMessage || validationError }}</AlertDescription>
        </Alert>

        <Button class="w-full" :disabled="!canSubmit" @click="submitJob">
          {{ isSubmitting ? '处理中' : '创建字幕任务' }}
        </Button>

        <template v-if="currentJob">
          <Separator />
          <section class="space-y-2 text-sm">
            <div class="flex items-center justify-between">
              <span class="text-muted-foreground">状态</span>
              <Badge>{{ currentJob.status }}</Badge>
            </div>
            <p class="text-xs text-muted-foreground">{{ currentJob.progressText }}</p>
            <p v-if="currentJob.status === 'transcribing'" class="text-xs text-muted-foreground">
              已耗时 {{ elapsedSeconds }} 秒
            </p>
            <p v-if="currentJob.status === 'completed'" class="text-xs text-muted-foreground">
              字幕结果已缓存，刷新播放页即可恢复。
            </p>
          </section>
        </template>
      </CardContent>
    </Card>
  </main>
</template>
```

- [ ] **Step 5: 运行 popup 校验测试和构建**

Run:

```bash
cd extension
mise exec -- npm run test -- src/popup/form-validation.test.ts
mise exec -- npm run build
```

Expected:

```text
3 form-validation tests pass.
WXT build exits 0.
```

- [ ] **Step 6: 提交 popup UI**

Run:

```bash
git add extension/entrypoints/popup/App.vue extension/src/popup
git commit -m "feat(extension): add popup job form"
```

## Task 8: Build YouTube Content Script Subtitle UI

**Files:**
- Create: `extension/src/youtube/page-watch.ts`
- Test: `extension/src/youtube/page-watch.test.ts`
- Create: `extension/src/content/YoutubeOverlay.vue`
- Modify: `extension/entrypoints/youtube.content.ts`

- [ ] **Step 1: 写 page watch 测试**

Create `extension/src/youtube/page-watch.test.ts`:

```ts
import { describe, expect, it } from 'vitest'
import { getVideoIdFromLocationHref } from './page-watch'

describe('getVideoIdFromLocationHref', () => {
  it('returns videoId for YouTube watch URLs', () => {
    expect(
      getVideoIdFromLocationHref('https://www.youtube.com/watch?v=video_123'),
    ).toBe('video_123')
  })

  it('returns null outside YouTube watch pages', () => {
    expect(getVideoIdFromLocationHref('https://www.youtube.com/')).toBeNull()
  })
})
```

- [ ] **Step 2: 运行 page watch 测试确认失败**

Run:

```bash
cd extension
mise exec -- npm run test -- src/youtube/page-watch.test.ts
```

Expected:

```text
FAIL because src/youtube/page-watch.ts does not exist.
```

- [ ] **Step 3: 实现 page watch helper**

Create `extension/src/youtube/page-watch.ts`:

```ts
import { parseYouTubeWatchVideoId } from './video-id'

export function getVideoIdFromLocationHref(href: string): string | null {
  return parseYouTubeWatchVideoId(href)
}

export function getCurrentVideoId(): string | null {
  return getVideoIdFromLocationHref(window.location.href)
}

export function watchVideoIdChanges(onChange: (videoId: string | null) => void) {
  let current = getCurrentVideoId()

  const check = () => {
    const next = getCurrentVideoId()
    if (next !== current) {
      current = next
      onChange(next)
    }
  }

  window.addEventListener('yt-navigate-finish', check)
  window.addEventListener('popstate', check)

  return () => {
    window.removeEventListener('yt-navigate-finish', check)
    window.removeEventListener('popstate', check)
  }
}
```

- [ ] **Step 4: 创建播放页 Vue overlay**

Create `extension/src/content/YoutubeOverlay.vue`:

```vue
<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { SubtitleMode } from '@/api/messages'
import { sendExtensionMessage } from '@/api/messages'
import type { SubtitleAssetCacheEntry } from '@/storage/subtitle-cache'
import { findActiveCue } from '@/subtitles/active-cue'
import { parseVtt, type VttCue } from '@/subtitles/vtt'
import { getCurrentVideoId, watchVideoIdChanges } from '@/youtube/page-watch'

const enabled = ref(true)
const status = ref('查找字幕')
const currentVideoId = ref<string | null>(null)
const currentAsset = ref<SubtitleAssetCacheEntry | null>(null)
const selectedMode = ref<SubtitleMode>('translated')
const cues = ref<VttCue[]>([])
const activeText = ref('')
let removeVideoListener: (() => void) | null = null
let removeVideoWatch: (() => void) | null = null

const hasSubtitle = computed(() => currentAsset.value !== null && cues.value.length > 0)

onMounted(() => {
  loadForVideo(getCurrentVideoId())
  removeVideoWatch = watchVideoIdChanges((videoId) => {
    loadForVideo(videoId)
  })
})

onUnmounted(() => {
  removeVideoListener?.()
  removeVideoWatch?.()
})

async function loadForVideo(videoId: string | null) {
  currentVideoId.value = videoId
  currentAsset.value = null
  cues.value = []
  activeText.value = ''

  if (!videoId) {
    status.value = '非 watch 页面'
    return
  }

  status.value = '查找字幕'
  const resolved = await sendExtensionMessage<SubtitleAssetCacheEntry | null>({
    type: 'subtitle:resolve',
    payload: { videoId },
  })
  if (!resolved.ok) {
    status.value = resolved.error.message
    return
  }
  if (!resolved.data) {
    status.value = '未找到字幕'
    return
  }

  currentAsset.value = resolved.data
  selectedMode.value = resolved.data.selectedMode
  await loadVtt()
}

async function loadVtt() {
  if (!currentAsset.value) {
    return
  }

  status.value = '加载字幕'
  const result = await sendExtensionMessage<string>({
    type: 'subtitle:fetch-file',
    payload: { jobId: currentAsset.value.jobId, mode: selectedMode.value },
  })
  if (!result.ok) {
    status.value = result.error.message
    return
  }

  try {
    cues.value = parseVtt(result.data)
    status.value = '字幕已加载'
    bindVideo()
  } catch {
    status.value = '字幕解析失败'
  }
}

async function changeMode(mode: SubtitleMode) {
  selectedMode.value = mode
  if (currentAsset.value) {
    await sendExtensionMessage({
      type: 'subtitle:update-mode',
      payload: {
        videoId: currentAsset.value.videoId,
        targetLanguage: currentAsset.value.targetLanguage,
        mode,
      },
    })
  }
  await loadVtt()
}

function bindVideo() {
  removeVideoListener?.()
  const video = document.querySelector('video')
  if (!video) {
    status.value = '未找到视频'
    return
  }

  const update = () => {
    activeText.value = findActiveCue(cues.value, video.currentTime)?.text ?? ''
  }

  video.addEventListener('timeupdate', update)
  video.addEventListener('seeked', update)
  update()

  removeVideoListener = () => {
    video.removeEventListener('timeupdate', update)
    video.removeEventListener('seeked', update)
  }
}
</script>

<template>
  <section class="fixed bottom-20 left-1/2 z-[2147483647] w-[min(720px,calc(100vw-32px))] -translate-x-1/2 space-y-2 text-center">
    <div class="mx-auto flex w-fit items-center gap-2 rounded-md border bg-background/95 px-2 py-1 text-xs text-foreground shadow">
      <Button size="sm" variant="ghost" @click="enabled = !enabled">
        {{ enabled ? '字幕开' : '字幕关' }}
      </Button>
      <Button size="sm" :variant="selectedMode === 'translated' ? 'default' : 'ghost'" @click="changeMode('translated')">
        translated
      </Button>
      <Button size="sm" :variant="selectedMode === 'bilingual' ? 'default' : 'ghost'" @click="changeMode('bilingual')">
        bilingual
      </Button>
      <Badge variant="secondary">{{ status }}</Badge>
    </div>

    <div v-if="enabled && hasSubtitle && activeText" class="mx-auto whitespace-pre-line rounded-md bg-black/80 px-4 py-2 text-xl leading-relaxed text-white shadow">
      {{ activeText }}
    </div>
  </section>
</template>
```

- [ ] **Step 5: 接入 content script Shadow DOM**

Replace `extension/entrypoints/youtube.content.ts`:

```ts
import '@/style.css'
import { createApp } from 'vue'
import { createShadowRootUi } from 'wxt/utils/content-script-ui/shadow-root'
import YoutubeOverlay from '@/content/YoutubeOverlay.vue'

export default defineContentScript({
  matches: ['https://www.youtube.com/watch*'],
  cssInjectionMode: 'ui',
  async main(ctx) {
    const ui = await createShadowRootUi(ctx, {
      name: 'lets-sub-it-youtube-ui',
      position: 'inline',
      anchor: 'body',
      onMount: (container) => {
        const app = createApp(YoutubeOverlay)
        app.mount(container)
        return app
      },
      onRemove: (app) => {
        app?.unmount()
      },
    })

    ui.mount()
  },
})
```

- [ ] **Step 6: 运行 page watch 测试和构建**

Run:

```bash
cd extension
mise exec -- npm run test -- src/youtube/page-watch.test.ts
mise exec -- npm run build
```

Expected:

```text
2 page-watch tests pass.
WXT build exits 0.
```

- [ ] **Step 7: 提交播放页字幕 UI**

Run:

```bash
git add extension/entrypoints/youtube.content.ts extension/src/content extension/src/youtube/page-watch.ts extension/src/youtube/page-watch.test.ts
git commit -m "feat(extension): render subtitles on YouTube watch pages"
```

## Task 9: Update Documentation And Run Full Verification

**Files:**
- Create: `extension/README.md`
- Modify: `README.md`
- Modify: `docs/superpowers/specs/2026-04-25-extension-mvp-design.md`

- [ ] **Step 1: 创建 extension README**

Create `extension/README.md`:

```markdown
# Lets Sub It Extension

Chrome Manifest V3 extension for submitting YouTube subtitle jobs and rendering completed subtitles on YouTube watch pages.

## Setup

```bash
mise install
cd extension
mise exec -- npm install
```

## Development

```bash
cd extension
mise exec -- npm run dev
```

Load the generated Chrome extension output from WXT in Chrome's extension developer mode.

## Test

```bash
cd extension
mise exec -- npm run test
```

## Build

```bash
cd extension
mise exec -- npm run build
```

## Backend

The popup defaults to `http://127.0.0.1:8080`. The first MVP supports `localhost` and `127.0.0.1` backend URLs only.
```

- [ ] **Step 2: 更新根 README 中 extension 状态**

Modify the note in `README.md` to state:

```markdown
> [!NOTE]
> 项目仍处于 MVP 阶段。当前可运行部分包括 `backend/` 的 mock API server、`whisper/` 的本地 `faster-whisper` 转写 CLI，以及 `extension/` 的 Chrome MV3 前端工程。真实 `yt-dlp`、`ffmpeg`、`whisper-cli` runner 和 LLM 翻译仍在路线图中。
```

Modify the current implementation list to include:

```markdown
- `extension/`：Chrome MV3 extension 工程，支持 popup 提交/轮询、background API 网关、storage 缓存和 YouTube watch 页面字幕层。
```

Modify the route map entry:

```text
├── extension/               # Chrome MV3 extension
```

Modify roadmap checkboxes:

```markdown
- [x] Chrome extension 任务提交、状态轮询和播放页字幕层
```

- [ ] **Step 3: 确认 spec 权限修正已保留**

Verify `docs/superpowers/specs/2026-04-25-extension-mvp-design.md` contains:

```text
permissions:
  - storage
  - activeTab
```

- [ ] **Step 4: 运行全量 extension 验证**

Run:

```bash
cd extension
mise exec -- npm run test
mise exec -- npm run typecheck
mise exec -- npm run build
```

Expected:

```text
All Vitest tests pass.
vue-tsc exits 0.
WXT build exits 0.
```

- [ ] **Step 5: 检查仓库状态和忽略文件**

Run:

```bash
git status --short
```

Expected:

```text
No generated output directories such as extension/.output or extension/node_modules are staged.
Only source, lockfile, README, mise, spec, and plan-related files are changed.
```

- [ ] **Step 6: 提交文档和最终验证变更**

Run:

```bash
git add README.md extension/README.md docs/superpowers/specs/2026-04-25-extension-mvp-design.md
git commit -m "docs(extension): document extension MVP workflow"
```

## 最终验收

- [ ] `cd extension && mise exec -- npm run test` 通过。
- [ ] `cd extension && mise exec -- npm run typecheck` 通过。
- [ ] `cd extension && mise exec -- npm run build` 通过。
- [ ] `cd backend && mise exec -- go test ./...` 仍通过。
- [ ] popup 可以保存 backend URL。
- [ ] popup 创建 job 时拒绝相同的源语言和目标语言。
- [ ] popup 可以创建或复用 mock backend job，并展示阶段状态。
- [ ] job completed 后，subtitle asset 写入 extension storage。
- [ ] YouTube watch 页面可以从缓存或 backend asset 查询恢复字幕。
- [ ] 播放页可以切换 `translated` 和 `bilingual`。
- [ ] `git status --short` 不包含未解释的生成产物。
