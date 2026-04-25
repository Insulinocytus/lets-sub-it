# Chrome Extension 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Chrome extension (WXT + Vue 3 + Tailwind + shadcn-vue) that allows users to submit YouTube URLs for subtitle generation and display subtitles on YouTube watch pages.

**Architecture:** Three WXT entrypoints — popup (Vue SPA for job creation/status/results), background (API proxy + cache + message router), content (shadow DOM subtitle overlay on YouTube). Background acts as the central data layer; popup communicates via `runtime.sendMessage`, content script listens via `storage.onChanged`.

**Tech Stack:** WXT (Chrome MV3), Vue 3 Composition API, TypeScript, Tailwind CSS, shadcn-vue, Vitest + @vue/test-utils, npm

---

### Task 1: Scaffold WXT + Vue + Tailwind + shadcn-vue project

**Files:**
- Create: `extension/package.json`
- Create: `extension/wxt.config.ts`
- Create: `extension/tsconfig.json`
- Create: `extension/tailwind.config.ts`
- Create: `extension/postcss.config.js`
- Create: `extension/components.json`
- Create: `extension/src/styles/globals.css`
- Create: `extension/src/lib/utils.ts`
- Create: `extension/src/entrypoints/popup/index.html`
- Create: `extension/.gitignore` (merge into existing)
- Create: `extension/src/types/index.ts`
- Modify: `extension/.gitignore` (add `src/.wxt/` if needed)

- [ ] **Step 1: Create package.json with all dependencies**

```json
{
  "name": "lets-sub-it-extension",
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "wxt",
    "build": "wxt build",
    "zip": "wxt zip",
    "compile": "wxt prepare",
    "test": "vitest run",
    "test:watch": "vitest"
  },
  "dependencies": {
    "vue": "^3.5.13",
    "@wxt-dev/storage": "^1.0.3",
    "clsx": "^2.1.1",
    "tailwind-merge": "^2.6.0",
    "class-variance-authority": "^0.7.1",
    "lucide-vue-next": "^0.468.0"
  },
  "devDependencies": {
    "wxt": "^0.19.26",
    "@wxt-dev/auto-icons": "^0.3.4",
    "typescript": "^5.7.3",
    "tailwindcss": "^3.4.17",
    "postcss": "^8.5.1",
    "autoprefixer": "^10.4.20",
    "vitest": "^3.0.4",
    "@vue/test-utils": "^2.4.6",
    "jsdom": "^26.0.0",
    "@types/webextension-polyfill": "^0.12.1"
  }
}
```

- [ ] **Step 2: Create wxt.config.ts**

```typescript
import { defineConfig } from 'wxt'

export default defineConfig({
  srcDir: 'src',
  extensionApi: 'chrome',
  modules: ['@wxt-dev/storage'],
  manifest: {
    name: "Let's Sub It",
    version: '0.1.0',
    permissions: ['storage', 'tabs'],
    host_permissions: [
      'http://localhost:8080/*',
      'https://www.youtube.com/*',
    ],
    action: {
      default_title: "Let's Sub It",
      default_popup: 'entrypoints/popup/index.html',
    },
  },
})
```

- [ ] **Step 3: Create tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2021",
    "lib": ["ES2021", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "jsx": "preserve",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "types": ["wxt/client"],
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src/**/*.ts", "src/**/*.vue", "src/**/*.d.ts"]
}
```

- [ ] **Step 4: Create tailwind.config.ts**

```typescript
import type { Config } from 'tailwindcss'

export default {
  content: ['./src/**/*.{vue,ts,html}'],
  theme: {
    extend: {},
  },
  plugins: [],
} satisfies Config
```

- [ ] **Step 5: Create postcss.config.js**

```javascript
export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
```

- [ ] **Step 6: Create components.json for shadcn-vue**

```json
{
  "$schema": "https://shadcn-vue.com/schema.json",
  "style": "default",
  "typescript": true,
  "tailwind": {
    "config": "tailwind.config.ts",
    "css": "src/styles/globals.css",
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

- [ ] **Step 7: Create src/styles/globals.css**

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 0 0% 3.9%;
    --card: 0 0% 100%;
    --card-foreground: 0 0% 3.9%;
    --popover: 0 0% 100%;
    --popover-foreground: 0 0% 3.9%;
    --primary: 0 0% 9%;
    --primary-foreground: 0 0% 98%;
    --secondary: 0 0% 96.1%;
    --secondary-foreground: 0 0% 9%;
    --muted: 0 0% 96.1%;
    --muted-foreground: 0 0% 45.1%;
    --accent: 0 0% 96.1%;
    --accent-foreground: 0 0% 9%;
    --destructive: 0 84.2% 60.2%;
    --destructive-foreground: 0 0% 98%;
    --border: 0 0% 89.8%;
    --input: 0 0% 89.8%;
    --ring: 0 0% 3.9%;
    --radius: 0.5rem;
  }
}

@layer base {
  * {
    @apply border-border;
  }
  body {
    @apply bg-background text-foreground;
  }
}
```

- [ ] **Step 8: Create src/lib/utils.ts**

```typescript
import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

- [ ] **Step 9: Create src/entrypoints/popup/index.html**

```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=400" />
    <title>Let's Sub It</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="./main.ts"></script>
  </body>
</html>
```

- [ ] **Step 10: Install dependencies**

Run:
```bash
cd extension
npm install
```

Expected: `node_modules/` created, `package-lock.json` written.

- [ ] **Step 11: Run WXT prepare to validate config**

Run:
```bash
cd extension
npx wxt prepare
```

Expected: `.wxt/` directory created with generated types. No errors.

- [ ] **Step 12: Add shadcn-vue button component**

Run:
```bash
cd extension
npx shadcn-vue@latest add button
```

Expected: `src/components/ui/button/` created with Button.vue.

- [ ] **Step 13: Commit**

```bash
git add extension/package.json extension/package-lock.json extension/wxt.config.ts extension/tsconfig.json extension/tailwind.config.ts extension/postcss.config.js extension/components.json extension/src/
git commit -m "feat(extension): scaffold WXT + Vue + Tailwind + shadcn-vue project"
```

---

### Task 2: Define TypeScript types and backend API client interface

**Files:**
- Create: `extension/src/types/index.ts`
- Create: `extension/src/types/vue-shim.d.ts`

- [ ] **Step 1: Write TypeScript type definitions**

Write to `src/types/index.ts`:

```typescript
// --- Backend API response types (matching backend mock API) ---

export interface JobResponse {
  id: string
  videoId: string
  youtubeUrl: string
  sourceLanguage: string
  targetLanguage: string
  status: JobStatus
  stage: string
  progressText: string
  errorMessage: string | null
  createdAt: string
  updatedAt: string
}

export type JobStatus =
  | 'queued'
  | 'downloading'
  | 'transcribing'
  | 'translating'
  | 'packaging'
  | 'completed'
  | 'failed'

export type SubtitleMode = 'translated' | 'bilingual'

export interface CreateJobParams {
  youtubeUrl: string
  sourceLanguage: string
  targetLanguage: string
}

export interface CreateJobResult {
  job: JobResponse
  reused: boolean
}

export interface SubtitleAssetResponse {
  asset: {
    jobId: string
    videoId: string
    targetLanguage: string
    sourceLanguage: string
    files: {
      source: string
      translated: string
      bilingual: string
    }
    createdAt: string
  } | null
}

// --- Extension storage types ---

export interface LocalCacheEntry {
  videoId: string
  targetLanguage: string
  jobId: string
  selectedMode: SubtitleMode
  lastSyncedAt: string
}

export interface UserPreferences {
  videoId: string
  targetLanguage: string | null
  selectedMode: SubtitleMode
}

// --- Message types for popup <-> background communication ---

export type MessageType =
  | 'CREATE_JOB'
  | 'GET_JOB'
  | 'GET_SUBTITLE_ASSETS'

export interface ExtensionMessage {
  type: MessageType
  payload: Record<string, unknown>
}

export interface ExtensionResponse {
  success: boolean
  data?: unknown
  error?: string
}

// --- Content script types ---

export interface SubtitleCue {
  start: number  // seconds
  end: number    // seconds
  text: string
}

export interface SubtitleData {
  videoId: string
  cues: SubtitleCue[]
  mode: SubtitleMode
  vttUrl: string
}

// --- VTT parser ---

export interface VttParseResult {
  cues: SubtitleCue[]
}
```

- [ ] **Step 2: Create Vue SFC type shim**

Write to `src/types/vue-shim.d.ts`:

```typescript
declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<object, object, unknown>
  export default component
}
```

- [ ] **Step 3: Commit**

```bash
git add extension/src/types/
git commit -m "feat(extension): add TypeScript types"
```

---

### Task 3: Implement VTT parser

**Files:**
- Create: `extension/src/lib/vtt-parser.ts`
- Create: `extension/src/lib/__tests__/vtt-parser.spec.ts`

- [ ] **Step 1: Write failing tests for VTT parser**

Write to `src/lib/__tests__/vtt-parser.spec.ts`:

```typescript
import { describe, it, expect } from 'vitest'
import { parseVtt } from '../vtt-parser'

describe('parseVtt', () => {
  it('parses WEBVTT header and cues', () => {
    const vtt = `WEBVTT

00:00:01.000 --> 00:00:03.500
Hello world

00:00:04.000 --> 00:00:06.000
Second cue
`
    const result = parseVtt(vtt)
    expect(result.cues).toHaveLength(2)
    expect(result.cues[0]).toEqual({ start: 1, end: 3.5, text: 'Hello world' })
    expect(result.cues[1]).toEqual({ start: 4, end: 6, text: 'Second cue' })
  })

  it('handles cue text with multiple lines', () => {
    const vtt = `WEBVTT

00:00:00.000 --> 00:00:02.000
Line one
Line two
`
    const result = parseVtt(vtt)
    expect(result.cues).toHaveLength(1)
    expect(result.cues[0].text).toBe('Line one\nLine two')
  })

  it('returns empty cues for empty VTT', () => {
    const result = parseVtt('WEBVTT\n\n')
    expect(result.cues).toHaveLength(0)
  })

  it('ignores cue identifiers', () => {
    const vtt = `WEBVTT

1
00:00:00.000 --> 00:00:01.000
Text
`
    const result = parseVtt(vtt)
    expect(result.cues).toHaveLength(1)
    expect(result.cues[0].text).toBe('Text')
  })

  it('parses timestamps with hours', () => {
    const vtt = `WEBVTT

01:02:03.000 --> 01:02:05.500
Long video
`
    const result = parseVtt(vtt)
    expect(result.cues[0].start).toBe(3723)
    expect(result.cues[0].end).toBe(3725.5)
  })

  it('parses bilingual VTT with two text lines per cue', () => {
    const vtt = `WEBVTT

00:00:00.000 --> 00:00:02.000
source text
翻译文本
`
    const result = parseVtt(vtt)
    expect(result.cues).toHaveLength(1)
    expect(result.cues[0].text).toBe('source text\n翻译文本')
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
cd extension
npx vitest run
```

Expected: FAIL with "Cannot find module '../vtt-parser'" or similar.

- [ ] **Step 3: Write minimal VTT parser implementation**

Write to `src/lib/vtt-parser.ts`:

```typescript
import type { SubtitleCue, VttParseResult } from '@/types'

function parseTimestamp(ts: string): number {
  const parts = ts.split(':')
  if (parts.length === 3) {
    const [h, m, s] = parts
    return parseInt(h) * 3600 + parseInt(m) * 60 + parseFloat(s)
  }
  // mm:ss.xxx
  return parseInt(parts[0]) * 60 + parseFloat(parts[1])
}

export function parseVtt(vttContent: string): VttParseResult {
  const lines = vttContent.split('\n')
  const cues: SubtitleCue[] = []

  let i = 0
  // Skip WEBVTI header and blank lines
  while (i < lines.length && !lines[i].includes('-->')) {
    i++
  }

  while (i < lines.length) {
    // Find a timing line
    const timingLine = lines[i]
    if (!timingLine.includes('-->')) {
      i++
      continue
    }

    const [startStr, , endStr] = timingLine.split(/\s+-->\s+|(?:(?!\s)|\s)-->|\s+-->/)
    const start = parseTimestamp(startStr.trim())

    // endStr might have whitespace after -->
    const end = parseTimestamp(endStr.trim().split(/\s+/)[0])

    i++
    // Collect cue text lines (until blank line or end)
    const textLines: string[] = []
    while (i < lines.length && lines[i].trim() !== '') {
      textLines.push(lines[i].trim())
      i++
    }

    if (textLines.length > 0) {
      cues.push({
        start,
        end,
        text: textLines.join('\n'),
      })
    }

    i++
  }

  return { cues }
}

export function findCueAtTime(cues: SubtitleCue[], time: number): SubtitleCue | null {
  return cues.find((cue) => time >= cue.start && time < cue.end) ?? null
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
cd extension
npx vitest run
```

Expected: PASS — all 6 tests pass.

- [ ] **Step 5: Commit**

```bash
git add extension/src/lib/vtt-parser.ts extension/src/lib/__tests__/vtt-parser.spec.ts
git commit -m "feat(extension): add VTT parser"
```

---

### Task 4: Implement composables — useApi, useCache, useJobPolling

**Files:**
- Create: `extension/src/composables/useApi.ts`
- Create: `extension/src/composables/useCache.ts`
- Create: `extension/src/composables/useJobPolling.ts`
- Create: `extension/src/composables/__tests__/useCache.spec.ts`

- [ ] **Step 1: Implement useApi composable**

Write to `src/composables/useApi.ts`:

```typescript
import type {
  CreateJobParams,
  CreateJobResult,
  ExtensionMessage,
  ExtensionResponse,
  JobResponse,
  SubtitleAssetResponse,
} from '@/types'

let inflightMap = new Map<string, Promise<unknown>>()

async function sendMessage(type: string, payload: Record<string, unknown>): Promise<unknown> {
  const key = `${type}:${JSON.stringify(payload)}`
  const inflight = inflightMap.get(key)
  if (inflight) return inflight

  const promise = (async () => {
    try {
      const response: ExtensionResponse = await chrome.runtime.sendMessage({
        type,
        payload,
      } as ExtensionMessage)
      if (!response.success) {
        throw new Error(response.error ?? 'unknown error')
      }
      return response.data
    } finally {
      inflightMap.delete(key)
    }
  })()

  inflightMap.set(key, promise)
  return promise
}

export function useApi() {
  async function createJob(params: CreateJobParams): Promise<CreateJobResult> {
    return sendMessage('CREATE_JOB', params as unknown as Record<string, unknown>) as Promise<CreateJobResult>
  }

  async function getJob(jobId: string): Promise<JobResponse> {
    return sendMessage('GET_JOB', { jobId }) as Promise<JobResponse>
  }

  async function getSubtitleAssets(videoId: string, targetLanguage: string): Promise<SubtitleAssetResponse> {
    return sendMessage('GET_SUBTITLE_ASSETS', { videoId, targetLanguage }) as Promise<SubtitleAssetResponse>
  }

  return { createJob, getJob, getSubtitleAssets }
}
```

- [ ] **Step 2: Implement useCache composable**

Write to `src/composables/useCache.ts`:

```typescript
import type { LocalCacheEntry, SubtitleMode, UserPreferences } from '@/types'

function cacheKey(videoId: string, targetLanguage: string): string {
  return `cache:${videoId}:${targetLanguage}`
}

function prefsKey(videoId: string): string {
  return `prefs:${videoId}`
}

export function useCache() {
  async function getCacheEntry(videoId: string, targetLanguage: string): Promise<LocalCacheEntry | null> {
    const result = await chrome.storage.local.get(cacheKey(videoId, targetLanguage))
    return (result[cacheKey(videoId, targetLanguage)] as LocalCacheEntry) ?? null
  }

  async function setCacheEntry(entry: LocalCacheEntry): Promise<void> {
    await chrome.storage.local.set({ [cacheKey(entry.videoId, entry.targetLanguage)]: entry })
  }

  async function getPreferences(videoId: string): Promise<UserPreferences | null> {
    const result = await chrome.storage.local.get(prefsKey(videoId))
    return (result[prefsKey(videoId)] as UserPreferences) ?? null
  }

  async function setPreferences(prefs: UserPreferences): Promise<void> {
    await chrome.storage.local.set({ [prefsKey(prefs.videoId)]: prefs })
  }

  return { getCacheEntry, setCacheEntry, getPreferences, setPreferences }
}
```

- [ ] **Step 3: Write tests for useCache**

Write to `src/composables/__tests__/useCache.spec.ts`:

```typescript
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useCache } from '../useCache'
import type { LocalCacheEntry, UserPreferences } from '@/types'

const mockStorage: Record<string, unknown> = {}

beforeEach(() => {
  Object.keys(mockStorage).forEach((k) => delete mockStorage[k])

  vi.stubGlobal('chrome', {
    storage: {
      local: {
        get: vi.fn((keys: string | string[] | Record<string, unknown>) => {
          if (typeof keys === 'string') {
            return { [keys]: mockStorage[keys] ?? null }
          }
          if (Array.isArray(keys)) {
            const result: Record<string, unknown> = {}
            for (const k of keys) result[k] = mockStorage[k] ?? null
            return result
          }
          return {}
        }),
        set: vi.fn((items: Record<string, unknown>) => {
          Object.assign(mockStorage, items)
        }),
      },
    },
  })
})

describe('useCache', () => {
  it('returns null for missing cache entry', async () => {
    const { getCacheEntry } = useCache()
    const result = await getCacheEntry('video1', 'zh-CN')
    expect(result).toBeNull()
  })

  it('stores and retrieves a cache entry', async () => {
    const { getCacheEntry, setCacheEntry } = useCache()
    const entry: LocalCacheEntry = {
      videoId: 'abc123',
      targetLanguage: 'zh-CN',
      jobId: 'job_xyz',
      selectedMode: 'bilingual',
      lastSyncedAt: '2026-04-25T00:00:00Z',
    }
    await setCacheEntry(entry)
    const result = await getCacheEntry('abc123', 'zh-CN')
    expect(result).toEqual(entry)
  })

  it('stores and retrieves user preferences', async () => {
    const { getPreferences, setPreferences } = useCache()
    const prefs: UserPreferences = {
      videoId: 'abc123',
      targetLanguage: 'zh-CN',
      selectedMode: 'translated',
    }
    await setPreferences(prefs)
    const result = await getPreferences('abc123')
    expect(result).toEqual(prefs)
  })

  it('returns null for missing preferences', async () => {
    const { getPreferences } = useCache()
    const result = await getPreferences('nonexistent')
    expect(result).toBeNull()
  })
})
```

- [ ] **Step 4: Run tests to verify useCache tests pass**

Run:
```bash
cd extension
npx vitest run
```

Expected: 4 new tests (from useCache) PASS. Previous VTT parser tests still PASS.

- [ ] **Step 5: Implement useJobPolling composable**

Write to `src/composables/useJobPolling.ts`:

```typescript
import { ref, watch, onUnmounted } from 'vue'
import { useApi } from './useApi'
import type { JobResponse, JobStatus } from '@/types'

export function useJobPolling() {
  const jobId = ref<string | null>(null)
  const status = ref<JobStatus | null>(null)
  const job = ref<JobResponse | null>(null)
  const error = ref<string | null>(null)
  const isPolling = ref(false)

  const { getJob } = useApi()

  let intervalId: ReturnType<typeof setInterval> | null = null

  const terminalStatuses: JobStatus[] = ['completed', 'failed']

  function start(id: string, intervalMs = 1000) {
    jobId.value = id
    isPolling.value = true
    error.value = null
    poll()
    intervalId = setInterval(poll, intervalMs)
  }

  function stop() {
    if (intervalId !== null) {
      clearInterval(intervalId)
      intervalId = null
    }
    isPolling.value = false
  }

  async function poll() {
    if (!jobId.value) return
    try {
      const result = await getJob(jobId.value)
      job.value = result
      status.value = result.status as JobStatus
      if (terminalStatuses.includes(result.status as JobStatus)) {
        stop()
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'polling failed'
      stop()
    }
  }

  function reset() {
    stop()
    jobId.value = null
    status.value = null
    job.value = null
    error.value = null
  }

  onUnmounted(() => {
    stop()
  })

  return { jobId, status, job, error, isPolling, start, stop, reset }
}
```

- [ ] **Step 6: Commit all composables**

```bash
git add extension/src/composables/
git commit -m "feat(extension): add useApi, useCache, useJobPolling composables"
```

---

### Task 5: Implement background service worker

**Files:**
- Create: `extension/src/entrypoints/background.ts`

- [ ] **Step 1: Write background script with message handler and API proxy**

Write to `src/entrypoints/background.ts`:

```typescript
import type { ExtensionMessage, ExtensionResponse } from '@/types'

const BACKEND_BASE = 'http://127.0.0.1:8080'

export default defineBackground(() => {
  // In-memory state cache for popup recovery
  const stateCache = new Map<string, unknown>()

  chrome.runtime.onMessage.addListener(
    (message: ExtensionMessage, _sender, sendResponse: (response: ExtensionResponse) => void) => {
      switch (message.type) {
        case 'CREATE_JOB':
          handleCreateJob(message.payload, sendResponse)
          return true // keep channel open for async
        case 'GET_JOB':
          handleGetJob(message.payload, sendResponse)
          return true
        case 'GET_SUBTITLE_ASSETS':
          handleGetSubtitleAssets(message.payload, sendResponse)
          return true
        default:
          sendResponse({ success: false, error: `unknown message type: ${message.type}` })
          return false
      }
    },
  )

  async function handleCreateJob(
    payload: Record<string, unknown>,
    sendResponse: (response: ExtensionResponse) => void,
  ) {
    try {
      const res = await fetch(`${BACKEND_BASE}/jobs`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          youtubeUrl: payload.youtubeUrl,
          sourceLanguage: payload.sourceLanguage,
          targetLanguage: payload.targetLanguage,
        }),
      })
      const data = await res.json()
      if (!res.ok) {
        sendResponse({ success: false, error: data.error?.message ?? 'create job failed' })
        return
      }
      const jobId = (data.job as { id: string }).id
      stateCache.set(jobId, data)
      sendResponse({ success: true, data })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'create job failed'
      sendResponse({ success: false, error: message })
    }
  }

  async function handleGetJob(
    payload: Record<string, unknown>,
    sendResponse: (response: ExtensionResponse) => void,
  ) {
    const jobId = payload.jobId as string
    if (!jobId) {
      sendResponse({ success: false, error: 'jobId is required' })
      return
    }
    try {
      const res = await fetch(`${BACKEND_BASE}/jobs/${jobId}`)
      const data = await res.json()
      if (!res.ok) {
        sendResponse({ success: false, error: data.error?.message ?? 'get job failed' })
        return
      }
      stateCache.set(jobId, data)
      sendResponse({ success: true, data })
    } catch (err) {
      // Retry once on network error
      try {
        const res = await fetch(`${BACKEND_BASE}/jobs/${jobId}`)
        const data = await res.json()
        if (!res.ok) {
          sendResponse({ success: false, error: data.error?.message ?? 'get job failed' })
          return
        }
        stateCache.set(jobId, data)
        sendResponse({ success: true, data })
      } catch (err2) {
        const message = err2 instanceof Error ? err2.message : 'get job failed'
        sendResponse({ success: false, error: message })
      }
    }
  }

  async function handleGetSubtitleAssets(
    payload: Record<string, unknown>,
    sendResponse: (response: ExtensionResponse) => void,
  ) {
    const videoId = payload.videoId as string
    const targetLanguage = payload.targetLanguage as string
    if (!videoId || !targetLanguage) {
      sendResponse({ success: false, error: 'videoId and targetLanguage are required' })
      return
    }
    try {
      const url = `${BACKEND_BASE}/subtitle-assets?videoId=${encodeURIComponent(videoId)}&targetLanguage=${encodeURIComponent(targetLanguage)}`
      const res = await fetch(url)
      const data = await res.json()
      if (!res.ok) {
        sendResponse({ success: false, error: data.error?.message ?? 'get subtitle assets failed' })
        return
      }
      sendResponse({ success: true, data })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'get subtitle assets failed'
      sendResponse({ success: false, error: message })
    }
  }
})
```

- [ ] **Step 2: Commit**

```bash
git add extension/src/entrypoints/background.ts
git commit -m "feat(extension): add background service worker with API proxy"
```

---

### Task 6: Implement popup entry point and App.vue

**Files:**
- Create: `extension/src/entrypoints/popup/main.ts`
- Create: `extension/src/entrypoints/popup/App.vue`
- Create: `extension/src/entrypoints/popup/components/JobForm.vue`
- Create: `extension/src/entrypoints/popup/components/JobStatus.vue`
- Create: `extension/src/entrypoints/popup/components/JobResult.vue`

- [ ] **Step 1: Create popup main.ts**

Write to `src/entrypoints/popup/main.ts`:

```typescript
import { createApp } from 'vue'
import App from './App.vue'
import '@/styles/globals.css'

const app = createApp(App)
app.mount('#app')
```

- [ ] **Step 2: Create App.vue (root component with three-view switching)**

Write to `src/entrypoints/popup/App.vue`:

```vue
<script setup lang="ts">
import { ref } from 'vue'
import type { JobStatus, JobResponse, CreateJobParams, SubtitleMode } from '@/types'
import { useApi } from '@/composables/useApi'
import { useJobPolling } from '@/composables/useJobPolling'
import { useCache } from '@/composables/useCache'
import JobForm from './components/JobForm.vue'
import JobStatusView from './components/JobStatus.vue'
import JobResultView from './components/JobResult.vue'

type ViewState = 'form' | 'polling' | 'result'

const viewState = ref<ViewState>('form')
const failedJob = ref<JobResponse | null>(null)

const api = useApi()
const polling = useJobPolling()
const cache = useCache()

async function handleSubmit(params: CreateJobParams) {
  try {
    const result = await api.createJob(params)
    if (result.reused && result.job.status === 'completed') {
      // Already completed — skip polling, show result
      pollAndFinish(result.job)
    } else {
      viewState.value = 'polling'
      polling.start(result.job.id)
    }
  } catch (err) {
    // Error is handled by the component
    throw err
  }
}

function pollAndFinish(job: JobResponse) {
  viewState.value = 'result'
  if (job.status === 'failed') {
    failedJob.value = job
  }
}

function handleJobComplete(job: JobResponse, selectedMode: SubtitleMode) {
  cache.setCacheEntry({
    videoId: job.videoId,
    targetLanguage: job.targetLanguage,
    jobId: job.id,
    selectedMode,
    lastSyncedAt: new Date().toISOString(),
  })
  cache.setPreferences({
    videoId: job.videoId,
    targetLanguage: job.targetLanguage,
    selectedMode,
  })
  viewState.value = 'result'
}

function handleJobFailed(job: JobResponse) {
  failedJob.value = job
  viewState.value = 'result'
}

function handleRetry() {
  failedJob.value = null
  polling.reset()
  viewState.value = 'form'
}
</script>

<template>
  <div class="w-[400px] min-h-[300px] p-4">
    <JobForm
      v-if="viewState === 'form'"
      @submit="handleSubmit"
    />
    <JobStatusView
      v-else-if="viewState === 'polling'"
      :job="polling.job.value"
      :status="polling.status.value"
      :error="polling.error.value"
      @complete="handleJobComplete"
      @failed="handleJobFailed"
    />
    <JobResultView
      v-else
      :job="polling.job.value ?? failedJob"
      :failed-job="failedJob"
      @retry="handleRetry"
    />
  </div>
</template>
```

- [ ] **Step 3: Create JobForm.vue**

Write to `src/entrypoints/popup/components/JobForm.vue`:

```vue
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useApi } from '@/composables/useApi'
import { Button } from '@/components/ui/button'

const emit = defineEmits<{
  submit: [params: { youtubeUrl: string; sourceLanguage: string; targetLanguage: string }]
}>()

const youtubeUrl = ref('')
const sourceLanguage = ref('ja')
const targetLanguage = ref('zh-CN')
const isLoading = ref(false)
const error = ref<string | null>(null)

const isValidUrl = computed(() => {
  return /^(https?:\/\/)?(www\.)?(youtube\.com\/watch\?v=|youtu\.be\/)[\w-]{11}/.test(youtubeUrl.value)
})

const languages = [
  { code: 'ja', label: '日本語' },
  { code: 'en', label: 'English' },
  { code: 'ko', label: '한국어' },
  { code: 'zh-CN', label: '简体中文' },
  { code: 'zh-TW', label: '繁体中文' },
]

async function handleSubmit() {
  if (!isValidUrl.value || isLoading.value) return
  isLoading.value = true
  error.value = null

  try {
    emit('submit', {
      youtubeUrl: youtubeUrl.value,
      sourceLanguage: sourceLanguage.value,
      targetLanguage: targetLanguage.value,
    })
  } catch (err) {
    error.value = err instanceof Error ? err.message : '无法连接服务器，请检查服务是否启动'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <h1 class="text-lg font-semibold">Let's Sub It</h1>

    <div class="flex flex-col gap-2">
      <label class="text-sm font-medium">YouTube URL</label>
      <input
        v-model="youtubeUrl"
        type="url"
        placeholder="https://www.youtube.com/watch?v=..."
        class="px-3 py-2 border rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-ring"
        :class="{ 'border-destructive': youtubeUrl && !isValidUrl }"
      />
      <p v-if="youtubeUrl && !isValidUrl" class="text-xs text-destructive">
        请输入有效的 YouTube URL
      </p>
    </div>

    <div class="flex gap-3">
      <div class="flex flex-col gap-2 flex-1">
        <label class="text-sm font-medium">源语言</label>
        <select
          v-model="sourceLanguage"
          class="px-3 py-2 border rounded-md text-sm bg-background"
        >
          <option v-for="lang in languages" :key="lang.code" :value="lang.code">
            {{ lang.label }}
          </option>
        </select>
      </div>

      <div class="flex flex-col gap-2 flex-1">
        <label class="text-sm font-medium">目标语言</label>
        <select
          v-model="targetLanguage"
          class="px-3 py-2 border rounded-md text-sm bg-background"
        >
          <option v-for="lang in languages" :key="lang.code" :value="lang.code">
            {{ lang.label }}
          </option>
        </select>
      </div>
    </div>

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <Button
      :disabled="!isValidUrl || isLoading"
      @click="handleSubmit"
    >
      {{ isLoading ? '提交中...' : '生成字幕' }}
    </Button>
  </div>
</template>
```

- [ ] **Step 4: Create JobStatus.vue**

Write to `src/entrypoints/popup/components/JobStatus.vue`:

```vue
<script setup lang="ts">
import { watch } from 'vue'
import type { JobResponse, JobStatus as JobStatusType, SubtitleMode } from '@/types'

const props = defineProps<{
  job: JobResponse | null
  status: JobStatusType | null
  error: string | null
}>()

const emit = defineEmits<{
  complete: [job: JobResponse, mode: SubtitleMode]
  failed: [job: JobResponse]
}>()

const stageLabels: Record<string, string> = {
  queued: '等待处理',
  downloading: '准备媒体',
  transcribing: '转写中',
  translating: '翻译中',
  packaging: '生成字幕',
}

watch(
  () => props.status,
  (newStatus) => {
    if (newStatus === 'completed' && props.job) {
      emit('complete', props.job, 'bilingual')
    } else if (newStatus === 'failed' && props.job) {
      emit('failed', props.job)
    }
  },
)
</script>

<template>
  <div class="flex flex-col gap-4 items-center justify-center min-h-[250px]">
    <div v-if="error" class="text-center">
      <p class="text-destructive text-sm">{{ error }}</p>
    </div>

    <template v-else-if="job">
      <div class="animate-spin h-8 w-8 border-2 border-primary border-t-transparent rounded-full" />

      <div class="text-center">
        <p class="text-sm font-medium">
          {{ stageLabels[job.stage] || job.stage }}
        </p>
        <p class="text-xs text-muted-foreground mt-1">
          {{ job.progressText }}
        </p>
      </div>

      <!-- Progress bar -->
      <div class="w-full bg-secondary rounded-full h-2">
        <div
          class="bg-primary h-2 rounded-full transition-all duration-500"
          :style="{ width: job.stage === 'packaging' ? '90%' : job.stage === 'translating' ? '60%' : '30%' }"
        />
      </div>
    </template>

    <div v-else class="text-sm text-muted-foreground">
      准备中...
    </div>
  </div>
</template>
```

- [ ] **Step 5: Create JobResult.vue**

Write to `src/entrypoints/popup/components/JobResult.vue`:

```vue
<script setup lang="ts">
import { ref } from 'vue'
import type { JobResponse, SubtitleMode } from '@/types'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  job: JobResponse | null
  failedJob: JobResponse | null
}>()

const emit = defineEmits<{
  retry: []
}>()

const selectedMode = ref<SubtitleMode>('bilingual')

const isFailed = props.failedJob !== null
const displayJob = props.failedJob ?? props.job
</script>

<template>
  <div class="flex flex-col gap-4 min-h-[250px]">
    <div v-if="isFailed" class="text-center">
      <p class="text-destructive font-medium">处理失败</p>
      <p class="text-sm text-muted-foreground mt-1">
        阶段：{{ displayJob?.stage }}
      </p>
      <p class="text-xs text-muted-foreground mt-1" v-if="displayJob?.errorMessage">
        {{ displayJob.errorMessage }}
      </p>
      <Button class="mt-4" variant="outline" @click="emit('retry')">
        重新提交
      </Button>
    </div>

    <div v-else class="flex flex-col gap-4">
      <div class="text-center">
        <p class="text-green-600 font-medium">字幕已生成</p>
      </div>

      <div class="flex flex-col gap-2">
        <label class="text-sm font-medium">字幕模式</label>
        <div class="flex gap-2">
          <Button
            :variant="selectedMode === 'translated' ? 'default' : 'outline'"
            size="sm"
            @click="selectedMode = 'translated'"
          >
            仅翻译
          </Button>
          <Button
            :variant="selectedMode === 'bilingual' ? 'default' : 'outline'"
            size="sm"
            @click="selectedMode = 'bilingual'"
          >
            双语
          </Button>
        </div>
      </div>

      <div class="text-xs text-muted-foreground">
        <p>视频 ID: {{ displayJob?.videoId }}</p>
        <p>{{ displayJob?.sourceLanguage }} → {{ displayJob?.targetLanguage }}</p>
      </div>

      <p class="text-xs text-muted-foreground">
        打开 YouTube 视频页面，字幕将自动显示。
      </p>
    </div>
  </div>
</template>
```

- [ ] **Step 6: Verify build compiles**

Run:
```bash
cd extension
npx wxt build
```

Expected: Build succeeds. Output in `.output/chrome-mv3/`.

- [ ] **Step 7: Commit**

```bash
git add extension/src/entrypoints/popup/
git commit -m "feat(extension): add popup entry point and components"
```

---

### Task 7: Implement content script

**Files:**
- Create: `extension/src/entrypoints/content.ts`
- Create: `extension/src/lib/__tests__/find-cue.spec.ts`

- [ ] **Step 1: Write test for findCueAtTime**

Write to `src/lib/__tests__/find-cue.spec.ts`:

```typescript
import { describe, it, expect } from 'vitest'
import { findCueAtTime } from '../vtt-parser'
import type { SubtitleCue } from '@/types'

describe('findCueAtTime', () => {
  const cues: SubtitleCue[] = [
    { start: 0, end: 2, text: 'First' },
    { start: 2, end: 5, text: 'Second' },
    { start: 5, end: 10, text: 'Third' },
  ]

  it('finds cue by current time', () => {
    expect(findCueAtTime(cues, 0)).toEqual(cues[0])
    expect(findCueAtTime(cues, 1.5)).toEqual(cues[0])
    expect(findCueAtTime(cues, 2)).toEqual(cues[1])
    expect(findCueAtTime(cues, 7)).toEqual(cues[2])
  })

  it('returns null for time before first cue', () => {
    expect(findCueAtTime(cues, -1)).toBeNull()
  })

  it('returns null for time after last cue', () => {
    expect(findCueAtTime(cues, 10)).toBeNull()
  })

  it('returns null for empty cues array', () => {
    expect(findCueAtTime([], 5)).toBeNull()
  })
})
```

- [ ] **Step 2: Run tests to verify**

Run:
```bash
cd extension
npx vitest run src/lib/__tests__/find-cue.spec.ts
```

Expected: PASS.

- [ ] **Step 3: Write content script**

Write to `src/entrypoints/content.ts`:

```typescript
import type { SubtitleCue, SubtitleMode, UserPreferences } from '@/types'
import { parseVtt, findCueAtTime } from '@/lib/vtt-parser'

const BACKEND_BASE = 'http://127.0.0.1:8080'

let currentCues: SubtitleCue[] = []
let currentMode: SubtitleMode = 'bilingual'
let currentVideoId: string | null = null
let videoElement: HTMLVideoElement | null = null
let shadowRoot: ShadowRoot | null = null
let containerEl: HTMLElement | null = null
let animationFrameId: number | null = null
let retryCount = 0
const MAX_RETRIES = 3

function getVideoId(): string | null {
  const url = new URL(window.location.href)
  return url.searchParams.get('v')
}

function waitForPlayer(): Promise<HTMLVideoElement | null> {
  return new Promise((resolve) => {
    const video = document.querySelector('video')
    if (video) {
      resolve(video as HTMLVideoElement)
      return
    }

    const observer = new MutationObserver(() => {
      const el = document.querySelector('video')
      if (el) {
        observer.disconnect()
        resolve(el as HTMLVideoElement)
      }
    })

    observer.observe(document.body, { childList: true, subtree: true })

    // Timeout after 10s
    setTimeout(() => {
      observer.disconnect()
      resolve(null)
    }, 10000)
  })
}

function injectShadowContainer(): ShadowRoot | null {
  const player = document.querySelector('#movie_player') ?? document.querySelector('.video-stream')
  if (!player) return null

  // Remove old container if exists
  const old = document.getElementById('lsi-subtitle-host')
  if (old) old.remove()

  const host = document.createElement('div')
  host.id = 'lsi-subtitle-host'
  host.style.cssText = 'position: absolute; bottom: 60px; left: 0; right: 0; z-index: 9999; pointer-events: none; text-align: center;'
  player.appendChild(host)

  const shadow = host.attachShadow({ mode: 'open' })

  // Inject styles
  const style = document.createElement('style')
  style.textContent = `
    .subtitle-container {
      display: inline-block;
      background: rgba(0, 0, 0, 0.75);
      color: #fff;
      padding: 4px 12px;
      border-radius: 4px;
      font-size: 18px;
      font-family: 'YouTube Noto', Roboto, Arial, sans-serif;
      line-height: 1.5;
      text-align: center;
      max-width: 80%;
      margin: 0 auto;
      user-select: none;
    }
  `
  shadow.appendChild(style)

  const container = document.createElement('div')
  container.className = 'subtitle-container'
  shadow.appendChild(container)

  containerEl = container
  shadowRoot = shadow
  return shadow
}

function renderCue(cue: SubtitleCue | null) {
  if (!containerEl) return
  containerEl.textContent = cue ? cue.text : ''
}

function onTimeUpdate() {
  if (!videoElement || currentCues.length === 0) return

  if (animationFrameId !== null) {
    cancelAnimationFrame(animationFrameId)
  }

  animationFrameId = requestAnimationFrame(() => {
    const cue = findCueAtTime(currentCues, videoElement!.currentTime)
    renderCue(cue)
    animationFrameId = null
  })
}

function setupVideoListener() {
  if (videoElement) {
    videoElement.removeEventListener('timeupdate', onTimeUpdate)
  }
  videoElement?.addEventListener('timeupdate', onTimeUpdate)
}

async function fetchVtt(videoId: string, mode: SubtitleMode): Promise<SubtitleCue[]> {
  const cacheKey = `cache:${videoId}`
  const result = await chrome.storage.local.get(cacheKey)
  const entry = result[cacheKey] as { jobId: string } | undefined
  if (!entry?.jobId) return []

  const url = `${BACKEND_BASE}/subtitle-files/${entry.jobId}/${mode}`
  const res = await fetch(url)
  if (!res.ok) return []

  const vtt = await res.text()
  const parsed = parseVtt(vtt)
  return parsed.cues
}

async function loadSubtitles() {
  if (!currentVideoId) return
  currentCues = await fetchVtt(currentVideoId, currentMode)
}

async function init() {
  const videoId = getVideoId()
  if (!videoId) return
  currentVideoId = videoId

  // Check cache for existing preferences
  const prefsKey = `prefs:${videoId}`
  const prefsResult = await chrome.storage.local.get(prefsKey)
  const prefs = prefsResult[prefsKey] as UserPreferences | undefined
  if (prefs) {
    currentMode = prefs.selectedMode
  }

  // Wait for player and inject shadow DOM
  videoElement = await waitForPlayer()
  if (!videoElement) {
    if (retryCount < MAX_RETRIES) {
      retryCount++
      setTimeout(init, 1000)
    }
    return
  }

  const shadow = injectShadowContainer()
  if (!shadow) {
    if (retryCount < MAX_RETRIES) {
      retryCount++
      setTimeout(init, 1000)
    }
    return
  }

  setupVideoListener()

  // Load subtitles if we have cached data
  await loadSubtitles()
}

// Listen for storage changes (when popup completes a job)
chrome.storage.onChanged.addListener(async (changes) => {
  for (const key of Object.keys(changes)) {
    if (key.startsWith('cache:') && currentVideoId && key.includes(currentVideoId)) {
      await loadSubtitles()
    }
    if (key.startsWith('prefs:') && currentVideoId && key.includes(currentVideoId)) {
      const prefs = changes[key].newValue as UserPreferences | undefined
      if (prefs) {
        currentMode = prefs.selectedMode
        await loadSubtitles()
      }
    }
  }
})

// SPA navigation detection
let lastUrl = window.location.href
new MutationObserver(() => {
  if (window.location.href !== lastUrl) {
    lastUrl = window.location.href
    // Re-init for new video
    retryCount = 0
    const shadow = document.getElementById('lsi-subtitle-host')
    if (shadow) shadow.remove()
    init()
  }
}).observe(document.body, { childList: true, subtree: true })

// Start
init()
```

- [ ] **Step 4: Verify build compiles**

Run:
```bash
cd extension
npx wxt build
```

Expected: Build succeeds. `.output/chrome-mv3/content-scripts/content.js` exists.

- [ ] **Step 5: Run all tests**

Run:
```bash
cd extension
npx vitest run
```

Expected: ALL tests PASS.

- [ ] **Step 6: Commit**

```bash
git add extension/src/entrypoints/content.ts extension/src/lib/__tests__/find-cue.spec.ts
git commit -m "feat(extension): add content script for YouTube subtitle overlay"
```

---

### Self-Review

After writing the complete plan, verify against the spec:

1. **Spec coverage:**
   - Popup three views (form/polling/result): Task 6
   - Background API proxy + cache + message router: Task 5
   - Content script shadow DOM overlay: Task 7
   - Data flow (popup→message→background→fetch): Tasks 4, 5
   - Storage cache + prefs: Tasks 4, 7
   - Error handling (retries, timeouts, failed states): Tasks 5, 6, 7
   - SPA navigation detection: Task 7

2. **Placeholder check:** No TBD, TODO, or incomplete sections.

3. **Type consistency:** `SubtitleCue`, `SubtitleMode`, `LocalCacheEntry`, `UserPreferences`, `JobStatus` all match the spec definitions.

4. **Scope check:** Focused on extension MVP only. No out-of-scope features.
