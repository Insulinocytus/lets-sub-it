# Extension YouTube Subtitle UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move Lets Sub It subtitles into the YouTube player, add a player-control subtitle toggle, and add popup subtitle usage settings for font size and display mode.

**Architecture:** Keep the existing WXT + Vue extension architecture. Add persistent subtitle settings to the existing settings storage, split popup UI into generation/settings tabs, mount the content UI inside `#movie_player.html5-video-player`, and insert one small button into YouTube's right controls.

**Tech Stack:** WXT, Vue 3 SFC, TypeScript, Vitest, jsdom, `wxt/utils/storage`, existing extension runtime messaging.

---

## File Structure

- Modify `extension/src/api/messages.ts`: extend `Settings` with `subtitleFontSizePx` and `subtitleMode`; define the settings update runtime message shape used by content script tests.
- Modify `extension/src/storage/settings.ts`: add defaults, merge old stored settings with defaults, validate subtitle font size.
- Modify `extension/src/storage/settings.test.ts`: cover default subtitle settings, old storage merge, positive numeric validation, and unlimited large px values.
- Modify `extension/src/api/message-handler.ts`: use global `settings.subtitleMode` when resolving new subtitle assets instead of hard-coded `translated`.
- Modify `extension/src/api/message-handler.test.ts`: cover that `subtitle:resolve` uses the global subtitle mode default.
- Create `extension/src/youtube/player-ui.ts`: focused DOM helpers for finding the YouTube player, mounting/removing the player overlay host, and inserting/removing the control-bar toggle button.
- Create `extension/src/youtube/player-ui.test.ts`: unit tests for helper behavior without mounting Vue.
- Modify `extension/entrypoints/youtube.content.ts`: mount the Vue overlay into the player overlay host and insert the control-bar toggle button; retry on YouTube SPA navigation and DOM changes.
- Modify `extension/src/content/YoutubeOverlay.vue`: remove the old floating control panel, read/apply subtitle settings, react to runtime settings/toggle messages, and render only subtitles inside the player layer.
- Modify `extension/src/content/YoutubeOverlay.test.ts`: cover font-size application, settings update reload, toggle message behavior, and absence of old floating buttons.
- Modify `extension/entrypoints/popup/App.vue`: add two tabs, keep generation UI in the first tab, add subtitle settings form in the second tab, persist settings, and notify the active YouTube tab.
- Modify `extension/entrypoints/popup/App.test.ts`: cover tab switching, settings save, validation, and tab notification.

---

### Task 1: Persist Subtitle Usage Settings

**Files:**
- Modify: `extension/src/api/messages.ts`
- Modify: `extension/src/storage/settings.ts`
- Modify: `extension/src/storage/settings.test.ts`

- [ ] **Step 1: Write failing tests for default subtitle settings**

Add these tests to `extension/src/storage/settings.test.ts` after the existing default settings test:

```ts
  it('includes default subtitle display settings', async () => {
    await expect(getSettings()).resolves.toEqual({
      backendBaseUrl: 'http://127.0.0.1:8080',
      sourceLanguage: 'en',
      targetLanguage: 'zh',
      subtitleFontSizePx: 20,
      subtitleMode: 'translated',
    })
  })

  it('merges subtitle defaults into previously stored settings', async () => {
    await fakeBrowser.storage.local.set({
      settings: {
        backendBaseUrl: 'http://localhost:9090',
        sourceLanguage: 'zh',
        targetLanguage: 'en',
      },
    })

    await expect(getSettings()).resolves.toEqual({
      backendBaseUrl: 'http://localhost:9090',
      sourceLanguage: 'zh',
      targetLanguage: 'en',
      subtitleFontSizePx: 20,
      subtitleMode: 'translated',
    })
  })
```

- [ ] **Step 2: Run focused settings tests and verify failure**

Run:

```bash
cd extension && mise exec -- npx vitest run src/storage/settings.test.ts
```

Expected: FAIL with assertions showing `subtitleFontSizePx` and `subtitleMode` are missing from returned settings.

- [ ] **Step 3: Extend the `Settings` type**

In `extension/src/api/messages.ts`, replace the `Settings` type with:

```ts
export type Settings = {
  backendBaseUrl: string
  sourceLanguage: LanguageCode
  targetLanguage: LanguageCode
  subtitleFontSizePx: number
  subtitleMode: SubtitleMode
}
```

- [ ] **Step 4: Add defaults and merge stored settings**

In `extension/src/storage/settings.ts`, replace `DEFAULT_SETTINGS` with:

```ts
export const DEFAULT_SETTINGS: Settings = Object.freeze({
  backendBaseUrl: 'http://127.0.0.1:8080',
  sourceLanguage: 'en',
  targetLanguage: 'zh',
  subtitleFontSizePx: 20,
  subtitleMode: 'translated',
})
```

Replace `getSettings()` with:

```ts
export async function getSettings(): Promise<Settings> {
  const settings = await settingsItem.getValue()
  return { ...DEFAULT_SETTINGS, ...settings }
}
```

- [ ] **Step 5: Add validation tests for subtitle font size**

Add these tests to `extension/src/storage/settings.test.ts` after the backend URL normalization test:

```ts
  it('updates subtitle display settings without imposing an upper bound', async () => {
    const settings = await updateSettings({
      subtitleFontSizePx: 240,
      subtitleMode: 'bilingual',
    })

    expect(settings).toEqual({
      ...DEFAULT_SETTINGS,
      subtitleFontSizePx: 240,
      subtitleMode: 'bilingual',
    })
    await expect(getSettings()).resolves.toEqual(settings)
  })

  it('rejects non-positive subtitle font sizes without persisting settings', async () => {
    await expect(updateSettings({ subtitleFontSizePx: 0 })).rejects.toThrow(
      'subtitleFontSizePx must be a positive number',
    )
    await expect(updateSettings({ subtitleFontSizePx: -4 })).rejects.toThrow(
      'subtitleFontSizePx must be a positive number',
    )
    await expect(getSettings()).resolves.toEqual(DEFAULT_SETTINGS)
  })
```

- [ ] **Step 6: Implement font-size validation**

In `extension/src/storage/settings.ts`, add this helper near `createLanguagePair`:

```ts
function assertPositiveSubtitleFontSize(value: number) {
  if (!Number.isFinite(value) || value <= 0) {
    throw new Error('subtitleFontSizePx must be a positive number')
  }
}
```

In `updateSettings()`, after `assertDifferentLanguages(next.sourceLanguage, next.targetLanguage)`, add:

```ts
  assertPositiveSubtitleFontSize(next.subtitleFontSizePx)
```

- [ ] **Step 7: Run focused settings tests and verify pass**

Run:

```bash
cd extension && mise exec -- npx vitest run src/storage/settings.test.ts
```

Expected: PASS.

- [ ] **Step 8: Commit settings changes**

Run:

```bash
git add extension/src/api/messages.ts extension/src/storage/settings.ts extension/src/storage/settings.test.ts
git commit -m "feat(extension): persist subtitle display settings"
```

---

### Task 2: Use Global Subtitle Mode When Resolving Assets

**Files:**
- Modify: `extension/src/api/message-handler.ts`
- Modify: `extension/src/api/message-handler.test.ts`

- [ ] **Step 1: Write failing test for global default subtitle mode**

Add this test to `extension/src/api/message-handler.test.ts` after `resolves and caches a subtitle asset from backend when local cache is empty`:

```ts
  it('uses the global subtitle mode when caching a newly resolved subtitle asset', async () => {
    await handleExtensionMessage({
      type: 'settings:update',
      payload: { subtitleMode: 'bilingual' },
    })

    const fetchImpl = vi.fn(async () =>
      new Response(
        JSON.stringify({
          asset: {
            jobId: 'job_123',
            videoId: 'video_123',
            sourceLanguage: 'en',
            targetLanguage: 'zh',
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

    await handleExtensionMessage(
      { type: 'subtitle:resolve', payload: { videoId: 'video_123' } },
      { fetchImpl, now: () => '2026-04-25T00:01:00Z' },
    )

    await expect(
      getCachedSubtitleAsset(DEFAULT_SETTINGS.backendBaseUrl, 'video_123', 'zh'),
    ).resolves.toMatchObject({
      jobId: 'job_123',
      selectedMode: 'bilingual',
    })
  })
```

- [ ] **Step 2: Run focused handler test and verify failure**

Run:

```bash
cd extension && mise exec -- npx vitest run src/api/message-handler.test.ts -t "global subtitle mode"
```

Expected: FAIL because `selectedMode` is still `translated`.

- [ ] **Step 3: Use `settings.subtitleMode` in `subtitle:resolve`**

In `extension/src/api/message-handler.ts`, replace:

```ts
          preference?.selectedMode ?? 'translated',
```

with:

```ts
          preference?.selectedMode ?? settings.subtitleMode,
```

- [ ] **Step 4: Run focused handler test and verify pass**

Run:

```bash
cd extension && mise exec -- npx vitest run src/api/message-handler.test.ts -t "global subtitle mode"
```

Expected: PASS.

- [ ] **Step 5: Commit handler change**

Run:

```bash
git add extension/src/api/message-handler.ts extension/src/api/message-handler.test.ts
git commit -m "feat(extension): use default subtitle mode for new assets"
```

---

### Task 3: Add YouTube Player UI DOM Helpers

**Files:**
- Create: `extension/src/youtube/player-ui.ts`
- Create: `extension/src/youtube/player-ui.test.ts`

- [ ] **Step 1: Write helper tests**

Create `extension/src/youtube/player-ui.test.ts` with:

```ts
import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  createSubtitleToggleButton,
  ensurePlayerOverlayHost,
  findYouTubePlayer,
  mountSubtitleToggleButton,
  PLAYER_OVERLAY_HOST_ID,
  SUBTITLE_TOGGLE_BUTTON_ID,
} from './player-ui'

describe('YouTube player UI helpers', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
  })

  it('finds the YouTube movie player', () => {
    document.body.innerHTML = '<div id="movie_player" class="html5-video-player"></div>'

    expect(findYouTubePlayer()).toBe(document.querySelector('#movie_player'))
  })

  it('creates one absolute overlay host inside the player', () => {
    document.body.innerHTML = '<div id="movie_player" class="html5-video-player"></div>'
    const player = findYouTubePlayer()

    const first = ensurePlayerOverlayHost(player)
    const second = ensurePlayerOverlayHost(player)

    expect(first).toBe(second)
    expect(first.id).toBe(PLAYER_OVERLAY_HOST_ID)
    expect(first.parentElement).toBe(player)
    expect(first.style.position).toBe('absolute')
    expect(first.style.pointerEvents).toBe('none')
    expect(window.getComputedStyle(player!).position).toBe('relative')
  })

  it('moves the overlay host when the player changes', () => {
    document.body.innerHTML = `
      <div id="old-player" class="html5-video-player"></div>
      <div id="movie_player" class="html5-video-player"></div>
    `
    const oldPlayer = document.querySelector<HTMLElement>('#old-player')!
    const newPlayer = findYouTubePlayer()
    const oldHost = document.createElement('div')
    oldHost.id = PLAYER_OVERLAY_HOST_ID
    oldPlayer.append(oldHost)

    const host = ensurePlayerOverlayHost(newPlayer)

    expect(host).not.toBe(oldHost)
    expect(document.querySelectorAll(`#${PLAYER_OVERLAY_HOST_ID}`)).toHaveLength(1)
    expect(host.parentElement).toBe(newPlayer)
  })

  it('creates a toggle button that stops player event propagation', () => {
    const onToggle = vi.fn()
    const button = createSubtitleToggleButton(true, onToggle)
    const parent = document.createElement('div')
    const parentClick = vi.fn()
    parent.addEventListener('click', parentClick)
    parent.append(button)

    button.click()

    expect(button.id).toBe(SUBTITLE_TOGGLE_BUTTON_ID)
    expect(button.getAttribute('aria-pressed')).toBe('true')
    expect(onToggle).toHaveBeenCalledOnce()
    expect(parentClick).not.toHaveBeenCalled()
  })

  it('inserts one toggle button into YouTube right controls', () => {
    document.body.innerHTML = `
      <div id="movie_player" class="html5-video-player">
        <div class="ytp-right-controls"></div>
      </div>
    `
    const onToggle = vi.fn()

    const first = mountSubtitleToggleButton(true, onToggle)
    const second = mountSubtitleToggleButton(false, onToggle)

    expect(first).toBe(second)
    expect(document.querySelectorAll(`#${SUBTITLE_TOGGLE_BUTTON_ID}`)).toHaveLength(1)
    expect(first?.getAttribute('aria-pressed')).toBe('false')
  })
})
```

- [ ] **Step 2: Run helper tests and verify failure**

Run:

```bash
cd extension && mise exec -- npx vitest run src/youtube/player-ui.test.ts
```

Expected: FAIL because `player-ui.ts` does not exist.

- [ ] **Step 3: Implement DOM helpers**

Create `extension/src/youtube/player-ui.ts` with:

```ts
export const PLAYER_OVERLAY_HOST_ID = 'lets-sub-it-player-overlay-host'
export const SUBTITLE_TOGGLE_BUTTON_ID = 'lets-sub-it-subtitle-toggle'

const PLAYER_SELECTOR = '#movie_player.html5-video-player'
const RIGHT_CONTROLS_SELECTOR = '#movie_player .ytp-right-controls'
const STOPPED_EVENTS = ['click', 'mousedown', 'pointerdown', 'dblclick'] as const

export function findYouTubePlayer(): HTMLElement | null {
  return document.querySelector<HTMLElement>(PLAYER_SELECTOR)
}

export function ensurePlayerOverlayHost(player: HTMLElement | null): HTMLElement | null {
  if (!player) {
    return null
  }

  const existing = document.getElementById(PLAYER_OVERLAY_HOST_ID)
  if (existing?.parentElement === player) {
    return existing
  }
  existing?.remove()

  if (window.getComputedStyle(player).position === 'static') {
    player.style.position = 'relative'
  }

  const host = document.createElement('div')
  host.id = PLAYER_OVERLAY_HOST_ID
  Object.assign(host.style, {
    position: 'absolute',
    inset: '0',
    pointerEvents: 'none',
    zIndex: '9999',
    overflow: 'visible',
  })
  player.append(host)
  return host
}

export function createSubtitleToggleButton(
  enabled: boolean,
  onToggle: () => void,
): HTMLButtonElement {
  const button = document.createElement('button')
  button.id = SUBTITLE_TOGGLE_BUTTON_ID
  button.type = 'button'
  button.className = 'ytp-button lets-sub-it-subtitle-toggle'
  button.setAttribute('aria-label', 'Lets Sub It subtitles')
  button.setAttribute('aria-pressed', String(enabled))
  button.title = enabled ? '关闭 Lets Sub It 字幕' : '开启 Lets Sub It 字幕'
  button.textContent = 'LSI'
  button.style.cssText = `
    color: white;
    font-size: 11px;
    font-weight: 700;
    line-height: 36px;
    text-align: center;
  `

  for (const eventType of STOPPED_EVENTS) {
    button.addEventListener(eventType, event => event.stopPropagation())
  }
  button.addEventListener('click', onToggle)

  return button
}

export function mountSubtitleToggleButton(
  enabled: boolean,
  onToggle: () => void,
): HTMLButtonElement | null {
  const controls = document.querySelector<HTMLElement>(RIGHT_CONTROLS_SELECTOR)
  if (!controls) {
    return null
  }

  const existing = document.getElementById(SUBTITLE_TOGGLE_BUTTON_ID) as HTMLButtonElement | null
  if (existing?.parentElement === controls) {
    existing.setAttribute('aria-pressed', String(enabled))
    existing.title = enabled ? '关闭 Lets Sub It 字幕' : '开启 Lets Sub It 字幕'
    return existing
  }
  existing?.remove()

  const button = createSubtitleToggleButton(enabled, onToggle)
  controls.prepend(button)
  return button
}
```

- [ ] **Step 4: Run helper tests and verify pass**

Run:

```bash
cd extension && mise exec -- npx vitest run src/youtube/player-ui.test.ts
```

Expected: PASS.

- [ ] **Step 5: Commit helper files**

Run:

```bash
git add extension/src/youtube/player-ui.ts extension/src/youtube/player-ui.test.ts
git commit -m "feat(extension): add YouTube player UI helpers"
```

---

### Task 4: Mount Content UI Inside the YouTube Player

**Files:**
- Modify: `extension/entrypoints/youtube.content.ts`
- Modify: `extension/src/youtube/player-ui.test.ts`
- Modify: `extension/src/youtube/player-ui.ts`

- [ ] **Step 1: Add cleanup support test for overlay hosts**

Add this test to `extension/src/youtube/player-ui.test.ts` after `moves the overlay host when the player changes`:

```ts
  it('runs overlay cleanup before replacing an old host', () => {
    document.body.innerHTML = `
      <div id="old-player" class="html5-video-player"></div>
      <div id="movie_player" class="html5-video-player"></div>
    `
    const cleanup = vi.fn()
    const oldHost = document.createElement('div') as HTMLElement & {
      __letsSubItCleanup?: () => void
    }
    oldHost.id = PLAYER_OVERLAY_HOST_ID
    oldHost.__letsSubItCleanup = cleanup
    document.querySelector('#old-player')!.append(oldHost)

    ensurePlayerOverlayHost(findYouTubePlayer())

    expect(cleanup).toHaveBeenCalledOnce()
  })
```

- [ ] **Step 2: Run helper test and verify failure**

Run:

```bash
cd extension && mise exec -- npx vitest run src/youtube/player-ui.test.ts -t "runs overlay cleanup"
```

Expected: FAIL because `ensurePlayerOverlayHost` removes old host without cleanup.

- [ ] **Step 3: Add cleanup type and call cleanup**

In `extension/src/youtube/player-ui.ts`, add this type near constants:

```ts
export type PlayerOverlayHost = HTMLElement & {
  __letsSubItCleanup?: () => void
}
```

In `ensurePlayerOverlayHost`, replace:

```ts
  existing?.remove()
```

with:

```ts
  if (existing) {
    ;(existing as PlayerOverlayHost).__letsSubItCleanup?.()
    existing.remove()
  }
```

Also change the return type to:

```ts
export function ensurePlayerOverlayHost(player: HTMLElement | null): PlayerOverlayHost | null {
```

and create host as:

```ts
  const host = document.createElement('div') as PlayerOverlayHost
```

- [ ] **Step 4: Run helper tests and verify pass**

Run:

```bash
cd extension && mise exec -- npx vitest run src/youtube/player-ui.test.ts
```

Expected: PASS.

- [ ] **Step 5: Update content script mounting**

Replace `extension/entrypoints/youtube.content.ts` with:

```ts
import '@/style.css'
import { createApp } from 'vue'
import YoutubeOverlay from '@/content/YoutubeOverlay.vue'
import {
  ensurePlayerOverlayHost,
  findYouTubePlayer,
  mountSubtitleToggleButton,
  type PlayerOverlayHost,
} from '@/youtube/player-ui'

export default defineContentScript({
  matches: ['https://www.youtube.com/watch*'],
  cssInjectionMode: 'ui',
  main(ctx) {
    let mountedHost: PlayerOverlayHost | null = null
    let observer: MutationObserver | null = null

    const dispatchToggle = () => {
      window.dispatchEvent(new CustomEvent('lets-sub-it:toggle-subtitles'))
    }

    const mount = () => {
      const host = ensurePlayerOverlayHost(findYouTubePlayer())
      if (host && host !== mountedHost) {
        mountedHost = host
        const app = createApp(YoutubeOverlay)
        app.mount(host)
        host.__letsSubItCleanup = () => app.unmount()
      }
      mountSubtitleToggleButton(true, dispatchToggle)
    }

    mount()

    observer = new MutationObserver(() => mount())
    observer.observe(document.documentElement, { childList: true, subtree: true })

    ctx.onInvalidated(() => {
      observer?.disconnect()
      observer = null
      mountedHost?.__letsSubItCleanup?.()
      mountedHost?.remove()
      mountedHost = null
    })
  },
})
```

Note: the button `aria-pressed` will be synchronized with actual overlay state in Task 7.

- [ ] **Step 6: Run helper tests and content typecheck**

Run:

```bash
cd extension && mise exec -- npx vitest run src/youtube/player-ui.test.ts
cd extension && mise exec -- npm run typecheck
```

Expected: helper tests PASS and typecheck PASS.

- [ ] **Step 7: Commit content mounting changes**

Run:

```bash
git add extension/entrypoints/youtube.content.ts extension/src/youtube/player-ui.ts extension/src/youtube/player-ui.test.ts
git commit -m "feat(extension): mount subtitles inside YouTube player"
```

---

### Task 5: Apply Settings in the Player Overlay

**Files:**
- Modify: `extension/src/content/YoutubeOverlay.vue`
- Modify: `extension/src/content/YoutubeOverlay.test.ts`

- [ ] **Step 1: Update overlay test fixtures for settings**

In `extension/src/content/YoutubeOverlay.test.ts`, add `Settings`-like fields to mocked settings responses by extending the existing asset area with:

```ts
const settings = {
  backendBaseUrl: 'http://127.0.0.1:8080',
  sourceLanguage: 'en',
  targetLanguage: 'zh',
  subtitleFontSizePx: 20,
  subtitleMode: 'translated' as const,
}
```

In `beforeEach`, change `sendExtensionMessage.mockReset()` setup so tests can respond to `settings:get` by default:

```ts
    sendExtensionMessage.mockImplementation(async (message) => {
      if (message.type === 'settings:get') {
        return { ok: true, data: settings }
      }
      return { ok: true, data: null }
    })
```

For existing tests that use chained `mockResolvedValueOnce`, prepend one settings response before the existing `asset` response:

```ts
      .mockResolvedValueOnce({ ok: true, data: settings })
```

- [ ] **Step 2: Write failing test for no old floating controls**

Add this test near the top of the `describe('YoutubeOverlay')` block:

```ts
  it('renders subtitles without the old floating control buttons', async () => {
    sendExtensionMessage
      .mockResolvedValueOnce({ ok: true, data: settings })
      .mockResolvedValueOnce({ ok: true, data: asset })
      .mockResolvedValueOnce({ ok: true, data: validVtt })

    const wrapper = mount(YoutubeOverlay)
    await flushPromises()

    expect(wrapper.text()).toContain('hello')
    expect(wrapper.text()).not.toContain('字幕开')
    expect(wrapper.text()).not.toContain('翻译')
    expect(wrapper.text()).not.toContain('双语')
    expect(wrapper.text()).not.toContain('字幕已加载')
  })
```

- [ ] **Step 3: Run overlay test and verify failure**

Run:

```bash
cd extension && mise exec -- npx vitest run src/content/YoutubeOverlay.test.ts -t "without the old floating"
```

Expected: FAIL because old buttons/status are still rendered.

- [ ] **Step 4: Load settings and remove old floating controls**

In `extension/src/content/YoutubeOverlay.vue`, import `type Settings`:

```ts
  type Settings,
```

Add refs after `activeText`:

```ts
const subtitleFontSizePx = ref(20)
const settingsSubtitleMode = ref<SubtitleMode>('translated')
```

Change the `onMounted` callback signature from `onMounted(() => {` to:

```ts
onMounted(async () => {
```

In `onMounted`, before `const videoId = getCurrentVideoId()`, load settings:

```ts
  await loadSettings()
```

Add this function:

```ts
async function loadSettings() {
  try {
    const result = await sendExtensionMessage<Settings>({ type: 'settings:get' })
    if (!result.ok) {
      return
    }
    applySettings(result.data)
  } catch {
    // Keep defaults when settings are unavailable in content tests or startup.
  }
}

function applySettings(settings: Settings) {
  subtitleFontSizePx.value = settings.subtitleFontSizePx
  settingsSubtitleMode.value = settings.subtitleMode
}
```

In `loadForVideo`, replace:

```ts
  selectedMode.value = result.data.selectedMode
```

with:

```ts
  selectedMode.value = settingsSubtitleMode.value
```

Replace the full `<template>` with:

```vue
<template>
  <div
    v-if="isWatchPage"
    class="pointer-events-none absolute inset-0 z-[9999] flex items-end justify-center overflow-hidden px-4 pb-16"
  >
    <div
      v-if="enabled && hasSubtitle && activeText"
      class="pointer-events-auto max-w-[90%] whitespace-pre-line rounded bg-black/78 px-4 py-2 text-center font-semibold leading-snug text-white shadow-lg [text-shadow:0_1px_2px_rgb(0_0_0/0.85)]"
      :style="{ fontSize: `${subtitleFontSizePx}px` }"
    >
      {{ activeText }}
    </div>
  </div>
</template>
```

- [ ] **Step 5: Run no-floating-controls test and verify pass**

Run:

```bash
cd extension && mise exec -- npx vitest run src/content/YoutubeOverlay.test.ts -t "without the old floating"
```

Expected: PASS.

- [ ] **Step 6: Write failing test for settings update message**

Add this test to `YoutubeOverlay.test.ts`:

```ts
  it('applies subtitle settings updates and reloads the requested mode', async () => {
    sendExtensionMessage
      .mockResolvedValueOnce({ ok: true, data: settings })
      .mockResolvedValueOnce({ ok: true, data: asset })
      .mockResolvedValueOnce({ ok: true, data: validVtt })
      .mockResolvedValueOnce({
        ok: true,
        data: { ...asset, selectedMode: 'bilingual' as const },
      })
      .mockResolvedValueOnce({ ok: true, data: validVtt.replace('hello', 'hello\n你好') })

    const wrapper = mount(YoutubeOverlay)
    await flushPromises()

    const listener = addRuntimeListener.mock.calls[0]?.[0]
    listener?.({
      type: 'lets-sub-it:settings-updated',
      settings: {
        ...settings,
        subtitleFontSizePx: 32,
        subtitleMode: 'bilingual',
      },
    })
    await flushPromises()

    const subtitle = wrapper.find('[style*="font-size"]')
    expect(subtitle.attributes('style')).toContain('font-size: 32px')
    expect(getMessagesByType('subtitle:update-mode')).toContainEqual({
      type: 'subtitle:update-mode',
      payload: {
        videoId: 'video_123',
        targetLanguage: 'zh',
        mode: 'bilingual',
      },
    })
    expect(
      getMessagesByType('subtitle:fetch-file').some(
        message => message?.payload?.mode === 'bilingual',
      ),
    ).toBe(true)
  })
```

- [ ] **Step 7: Write failing test for initial global subtitle mode**

Add this test to `YoutubeOverlay.test.ts`:

```ts
  it('uses settings subtitle mode for the initial VTT request', async () => {
    sendExtensionMessage
      .mockResolvedValueOnce({
        ok: true,
        data: { ...settings, subtitleMode: 'bilingual' as const },
      })
      .mockResolvedValueOnce({ ok: true, data: asset })
      .mockResolvedValueOnce({ ok: true, data: validVtt })

    mount(YoutubeOverlay)
    await flushPromises()

    expect(getMessagesByType('subtitle:fetch-file')).toContainEqual({
      type: 'subtitle:fetch-file',
      payload: {
        jobId: 'job_123',
        mode: 'bilingual',
      },
    })
  })
```

- [ ] **Step 8: Write failing test for external toggle event**

Add this test to `YoutubeOverlay.test.ts`:

```ts
  it('toggles subtitle visibility from the player control button event', async () => {
    sendExtensionMessage
      .mockResolvedValueOnce({ ok: true, data: settings })
      .mockResolvedValueOnce({ ok: true, data: asset })
      .mockResolvedValueOnce({ ok: true, data: validVtt })

    const wrapper = mount(YoutubeOverlay)
    await flushPromises()

    expect(wrapper.text()).toContain('hello')

    window.dispatchEvent(new CustomEvent('lets-sub-it:toggle-subtitles'))
    await flushPromises()

    expect(wrapper.text()).not.toContain('hello')

    window.dispatchEvent(new CustomEvent('lets-sub-it:toggle-subtitles'))
    await flushPromises()

    expect(wrapper.text()).toContain('hello')
  })
```

- [ ] **Step 9: Run new overlay tests and verify failure**

Run:

```bash
cd extension && mise exec -- npx vitest run src/content/YoutubeOverlay.test.ts -t "settings updates|initial VTT|player control button"
```

Expected: FAIL because settings update, initial settings ordering, and external toggle handling are not implemented.

- [ ] **Step 10: Implement runtime settings and toggle handling**

In `YoutubeOverlay.vue`, add a new variable near other cleanup variables:

```ts
let removeToggleListener: (() => void) | null = null
```

In `onMounted`, replace the existing `handleRuntimeMessage` body with support for both message types:

```ts
  const handleRuntimeMessage = (message: unknown) => {
    if (isSettingsUpdatedMessage(message)) {
      void applySettingsUpdate(message.settings)
      return
    }

    if (!isSubtitleUpdatedMessage(message)) {
      return
    }
    if (message.videoId !== currentVideoId.value) {
      return
    }

    void loadForVideo(currentVideoId.value)
  }
```

After runtime listener setup in `onMounted`, add:

```ts
  const handleToggle = () => {
    enabled.value = !enabled.value
  }
  window.addEventListener('lets-sub-it:toggle-subtitles', handleToggle)
  removeToggleListener = () => {
    window.removeEventListener('lets-sub-it:toggle-subtitles', handleToggle)
  }
```

In `onUnmounted`, add:

```ts
  removeToggleListener?.()
  removeToggleListener = null
```

Add these functions near `handleModeClick`:

```ts
async function applySettingsUpdate(settings: Settings) {
  const previousMode = settingsSubtitleMode.value
  applySettings(settings)
  if (settings.subtitleMode === previousMode) {
    return
  }
  await changeMode(settings.subtitleMode)
}
```

Add the settings-updated type guard near `SubtitleUpdatedMessage`:

```ts
type SettingsUpdatedMessage = {
  type: 'lets-sub-it:settings-updated'
  settings: Settings
}

function isSettingsUpdatedMessage(message: unknown): message is SettingsUpdatedMessage {
  if (typeof message !== 'object' || message === null) {
    return false
  }

  const candidate = message as Partial<SettingsUpdatedMessage>
  return candidate.type === 'lets-sub-it:settings-updated' && typeof candidate.settings === 'object' && candidate.settings !== null
}
```

- [ ] **Step 11: Normalize chained overlay mocks and run overlay test file**

Before running the full file, inspect every existing test in `YoutubeOverlay.test.ts` that starts with `sendExtensionMessage.mockResolvedValueOnce(...)`. The first mocked result in each chain must be:

```ts
      .mockResolvedValueOnce({ ok: true, data: settings })
```

Then keep the previous mocked `asset` and VTT results in their original order.

Run:

```bash
cd extension && mise exec -- npx vitest run src/content/YoutubeOverlay.test.ts
```

Expected: PASS.

- [ ] **Step 12: Commit overlay changes**

Run:

```bash
git add extension/src/content/YoutubeOverlay.vue extension/src/content/YoutubeOverlay.test.ts
git commit -m "feat(extension): apply subtitle settings in player overlay"
```

---

### Task 6: Add Popup Tabs and Subtitle Settings Form

**Files:**
- Modify: `extension/entrypoints/popup/App.vue`
- Modify: `extension/entrypoints/popup/App.test.ts`

- [ ] **Step 1: Update popup test mocks for settings and tab notification**

In `extension/entrypoints/popup/App.test.ts`, change the `vi.hoisted` block to include `sendTabMessage`:

```ts
const { sendExtensionMessage, queryTabs, sendTabMessage } = vi.hoisted(() => ({
  sendExtensionMessage: vi.fn(),
  queryTabs: vi.fn(),
  sendTabMessage: vi.fn(),
}))
```

In the `wxt/browser` mock, replace `sendMessage: vi.fn()` with:

```ts
      sendMessage: sendTabMessage,
```

Replace the `settings` constant with:

```ts
const settings = {
  backendBaseUrl: 'http://127.0.0.1:8080',
  sourceLanguage: 'en',
  targetLanguage: 'zh',
  subtitleFontSizePx: 20,
  subtitleMode: 'translated' as const,
}
```

In `beforeEach`, reset `sendTabMessage` and include an id on the queried tab:

```ts
    sendTabMessage.mockReset()
    queryTabs.mockResolvedValue([
      { id: 7, url: 'https://www.youtube.com/watch?v=video_123' },
    ])
```

- [ ] **Step 2: Write failing test for tabs**

Add this test after the existing restore test:

```ts
  it('shows generation and subtitle settings tabs', async () => {
    const wrapper = mount(App)
    await flushPromises()

    expect(wrapper.text()).toContain('生成字幕')
    expect(wrapper.text()).toContain('字幕设置')
    expect(wrapper.text()).toContain('backend URL')

    await wrapper.get('button[data-testid="subtitle-settings-tab"]').trigger('click')

    expect(wrapper.text()).toContain('字体大小')
    expect(wrapper.text()).toContain('显示模式')
    expect(wrapper.text()).toContain('翻译 only')
    expect(wrapper.text()).toContain('双语')
  })
```

- [ ] **Step 3: Write failing test for saving settings and notifying tab**

Add this test:

```ts
  it('saves subtitle settings and notifies the current YouTube tab', async () => {
    sendExtensionMessage.mockImplementation(async (message) => {
      if (message.type === 'settings:get') {
        return { ok: true, data: settings }
      }
      if (message.type === 'settings:update') {
        return {
          ok: true,
          data: {
            ...settings,
            ...message.payload,
          },
        }
      }
      if (message.type === 'job:active') {
        return { ok: true, data: { job: activeJob } }
      }
      return { ok: true, data: null }
    })

    const wrapper = mount(App)
    await flushPromises()
    await wrapper.get('button[data-testid="subtitle-settings-tab"]').trigger('click')

    await wrapper.get('input[data-testid="subtitle-font-size-input"]').setValue('32')
    await wrapper.get('button[data-testid="subtitle-mode-bilingual"]').trigger('click')
    await wrapper.get('button[data-testid="save-subtitle-settings"]').trigger('click')
    await flushPromises()

    expect(sendExtensionMessage).toHaveBeenCalledWith({
      type: 'settings:update',
      payload: {
        subtitleFontSizePx: 32,
        subtitleMode: 'bilingual',
      },
    })
    expect(sendTabMessage).toHaveBeenCalledWith(7, {
      type: 'lets-sub-it:settings-updated',
      settings: {
        ...settings,
        subtitleFontSizePx: 32,
        subtitleMode: 'bilingual',
      },
    })
  })
```

- [ ] **Step 4: Write failing test for invalid font size**

Add this test:

```ts
  it('rejects invalid subtitle font size input before saving', async () => {
    const wrapper = mount(App)
    await flushPromises()
    await wrapper.get('button[data-testid="subtitle-settings-tab"]').trigger('click')

    await wrapper.get('input[data-testid="subtitle-font-size-input"]').setValue('0')
    await wrapper.get('button[data-testid="save-subtitle-settings"]').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('字幕字体大小必须是正数')
    expect(sendExtensionMessage).not.toHaveBeenCalledWith({
      type: 'settings:update',
      payload: expect.objectContaining({ subtitleFontSizePx: 0 }),
    })
  })
```

- [ ] **Step 5: Run popup tests and verify failure**

Run:

```bash
cd extension && mise exec -- npx vitest run entrypoints/popup/App.test.ts
```

Expected: FAIL because tabs and settings form are not implemented.

- [ ] **Step 6: Add popup state and save function**

In `extension/entrypoints/popup/App.vue`, add `type SubtitleMode` to the import from `@/api/messages`:

```ts
  type SubtitleMode,
```

Add these refs after `pollTimer`:

```ts
const activeTab = ref<'generate' | 'settings'>('generate')
const subtitleFontSizeInput = ref('20')
const subtitleMode = ref<SubtitleMode>('translated')
const settingsMessage = ref('')
```

In `onMounted`, after loading settings, set subtitle settings values:

```ts
      subtitleFontSizeInput.value = String(settingsResult.data.subtitleFontSizePx)
      subtitleMode.value = settingsResult.data.subtitleMode
```

Add these functions before `notifySubtitleUpdated`:

```ts
async function saveSubtitleSettings() {
  settingsMessage.value = ''
  const fontSize = Number(subtitleFontSizeInput.value)
  if (!Number.isFinite(fontSize) || fontSize <= 0) {
    settingsMessage.value = '字幕字体大小必须是正数'
    return
  }

  try {
    const result = await sendExtensionMessage<Settings>({
      type: 'settings:update',
      payload: {
        subtitleFontSizePx: fontSize,
        subtitleMode: subtitleMode.value,
      },
    })
    if (!result.ok) {
      settingsMessage.value = result.error.message
      return
    }

    subtitleFontSizeInput.value = String(result.data.subtitleFontSizePx)
    subtitleMode.value = result.data.subtitleMode
    settingsMessage.value = '字幕设置已保存'
    await notifySettingsUpdated(result.data)
  } catch (error) {
    settingsMessage.value = readableError(error)
  }
}

async function notifySettingsUpdated(settings: Settings) {
  try {
    const [tab] = await browser.tabs.query({ active: true, currentWindow: true })
    if (!tab?.id || !tab.url?.startsWith('https://www.youtube.com/watch')) {
      return
    }

    await browser.tabs.sendMessage(tab.id, {
      type: 'lets-sub-it:settings-updated',
      settings,
    })
  } catch {
    // Settings are already persisted; current-tab notification is best effort.
  }
}
```

- [ ] **Step 7: Replace popup template with tabbed layout**

In `App.vue`, replace the full `<template>` block with:

```vue
<template>
  <main class="w-[380px] bg-background p-4 text-foreground">
    <Card class="gap-4 rounded-lg py-4 shadow-none">
      <CardHeader class="gap-1 px-4">
        <CardTitle class="text-base">Lets Sub It</CardTitle>
        <CardDescription class="text-xs">
          提交当前 YouTube 视频并生成双语字幕。
        </CardDescription>
      </CardHeader>

      <CardContent class="space-y-4 px-4">
        <div class="grid grid-cols-2 gap-1 rounded-md bg-muted p-1 text-xs">
          <Button
            type="button"
            size="sm"
            data-testid="generate-tab"
            :variant="activeTab === 'generate' ? 'secondary' : 'ghost'"
            class="h-7 text-xs"
            @click="activeTab = 'generate'"
          >
            生成字幕
          </Button>
          <Button
            type="button"
            size="sm"
            data-testid="subtitle-settings-tab"
            :variant="activeTab === 'settings' ? 'secondary' : 'ghost'"
            class="h-7 text-xs"
            @click="activeTab = 'settings'"
          >
            字幕设置
          </Button>
        </div>

        <form v-if="activeTab === 'generate'" class="space-y-3" @submit.prevent="submitJob">
          <label class="grid gap-1.5 text-xs font-medium">
            backend URL
            <Input
              v-model="backendBaseUrl"
              class="h-8 text-xs"
              placeholder="http://127.0.0.1:8080"
            />
          </label>

          <label class="grid gap-1.5 text-xs font-medium">
            YouTube URL
            <Input
              v-model="youtubeUrl"
              class="h-8 text-xs"
              placeholder="https://www.youtube.com/watch?v=..."
            />
          </label>

          <div class="grid grid-cols-2 gap-3">
            <label class="grid min-w-0 gap-1.5 text-xs font-medium">
              源语言
              <Select v-model="sourceLanguage">
                <SelectTrigger class="h-8 w-full text-xs">
                  <SelectValue placeholder="源语言" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="language in languages"
                    :key="language"
                    :value="language"
                  >
                    {{ languageLabels[language] }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </label>

            <label class="grid min-w-0 gap-1.5 text-xs font-medium">
              目标语言
              <Select v-model="targetLanguage">
                <SelectTrigger class="h-8 w-full text-xs">
                  <SelectValue placeholder="目标语言" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="language in languages"
                    :key="language"
                    :value="language"
                  >
                    {{ languageLabels[language] }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </label>
          </div>

          <Alert v-if="alertMessage" variant="destructive" class="py-2">
            <AlertDescription class="text-xs">
              {{ alertMessage }}
            </AlertDescription>
          </Alert>

          <Button
            class="h-8 w-full text-xs"
            type="submit"
            :disabled="isSubmitDisabled"
          >
            {{ isSubmitting ? '处理中...' : '生成字幕' }}
          </Button>
        </form>

        <section v-if="activeTab === 'generate' && currentJob" class="space-y-2 rounded-md border bg-muted/40 p-3 text-xs">
          <div class="flex items-center justify-between gap-2">
            <span class="font-medium">任务状态</span>
            <Badge :variant="statusBadgeVariant">
              {{ statusLabel }}
            </Badge>
          </div>
          <p class="break-words text-muted-foreground">
            {{ currentJob.progressText }}
          </p>
          <p
            v-if="currentJob.status === 'completed' && subtitleReady"
            class="text-muted-foreground"
          >
            字幕已生成并写入本地缓存。
          </p>
        </section>

        <section v-if="activeTab === 'settings'" class="space-y-3 text-xs">
          <label class="grid gap-1.5 font-medium">
            字体大小
            <div class="flex items-center gap-2">
              <Input
                v-model="subtitleFontSizeInput"
                data-testid="subtitle-font-size-input"
                class="h-8 text-xs"
                inputmode="decimal"
                placeholder="20"
              />
              <span class="text-muted-foreground">px</span>
            </div>
          </label>

          <div class="grid gap-1.5 font-medium">
            显示模式
            <div class="grid grid-cols-2 gap-2">
              <Button
                type="button"
                size="sm"
                data-testid="subtitle-mode-translated"
                :variant="subtitleMode === 'translated' ? 'secondary' : 'outline'"
                class="h-8 text-xs"
                @click="subtitleMode = 'translated'"
              >
                翻译 only
              </Button>
              <Button
                type="button"
                size="sm"
                data-testid="subtitle-mode-bilingual"
                :variant="subtitleMode === 'bilingual' ? 'secondary' : 'outline'"
                class="h-8 text-xs"
                @click="subtitleMode = 'bilingual'"
              >
                双语
              </Button>
            </div>
          </div>

          <Alert v-if="settingsMessage" class="py-2">
            <AlertDescription class="text-xs">
              {{ settingsMessage }}
            </AlertDescription>
          </Alert>

          <Button
            type="button"
            data-testid="save-subtitle-settings"
            class="h-8 w-full text-xs"
            @click="saveSubtitleSettings"
          >
            保存字幕设置
          </Button>
        </section>
      </CardContent>
    </Card>
  </main>
</template>
```

- [ ] **Step 8: Run popup tests and verify pass**

Run:

```bash
cd extension && mise exec -- npx vitest run entrypoints/popup/App.test.ts
```

Expected: PASS.

- [ ] **Step 9: Commit popup changes**

Run:

```bash
git add extension/entrypoints/popup/App.vue extension/entrypoints/popup/App.test.ts
git commit -m "feat(extension): add subtitle settings tab"
```

---

### Task 7: Synchronize Player Toggle Button State

**Files:**
- Modify: `extension/entrypoints/youtube.content.ts`
- Modify: `extension/src/content/YoutubeOverlay.vue`
- Modify: `extension/src/content/YoutubeOverlay.test.ts`

- [ ] **Step 1: Write failing overlay test for enabled-state event**

Add this test to `YoutubeOverlay.test.ts` after the external toggle test:

```ts
  it('publishes subtitle enabled state changes for the player button', async () => {
    const events: boolean[] = []
    window.addEventListener('lets-sub-it:subtitle-enabled-changed', (event) => {
      events.push((event as CustomEvent<{ enabled: boolean }>).detail.enabled)
    })
    sendExtensionMessage
      .mockResolvedValueOnce({ ok: true, data: settings })
      .mockResolvedValueOnce({ ok: true, data: asset })
      .mockResolvedValueOnce({ ok: true, data: validVtt })

    mount(YoutubeOverlay)
    await flushPromises()

    window.dispatchEvent(new CustomEvent('lets-sub-it:toggle-subtitles'))
    await flushPromises()
    window.dispatchEvent(new CustomEvent('lets-sub-it:toggle-subtitles'))
    await flushPromises()

    expect(events).toEqual([true, false, true])
  })
```

- [ ] **Step 2: Run overlay test and verify failure**

Run:

```bash
cd extension && mise exec -- npx vitest run src/content/YoutubeOverlay.test.ts -t "publishes subtitle enabled"
```

Expected: FAIL because no enabled-state event is emitted.

- [ ] **Step 3: Emit enabled-state changes from overlay**

In `YoutubeOverlay.vue`, import `watch` from Vue:

```ts
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
```

After variable declarations, add:

```ts
watch(enabled, (next) => {
  window.dispatchEvent(new CustomEvent('lets-sub-it:subtitle-enabled-changed', {
    detail: { enabled: next },
  }))
}, { immediate: true })
```

- [ ] **Step 4: Run overlay test and verify pass**

Run:

```bash
cd extension && mise exec -- npx vitest run src/content/YoutubeOverlay.test.ts -t "publishes subtitle enabled"
```

Expected: PASS.

- [ ] **Step 5: Update content script to refresh button state**

In `extension/entrypoints/youtube.content.ts`, add `let subtitlesEnabled = true` near `mountedHost`:

```ts
    let subtitlesEnabled = true
```

Change `mountSubtitleToggleButton(true, dispatchToggle)` to:

```ts
      mountSubtitleToggleButton(subtitlesEnabled, dispatchToggle)
```

After MutationObserver setup, add:

```ts
    const handleEnabledChanged = (event: Event) => {
      subtitlesEnabled = (event as CustomEvent<{ enabled: boolean }>).detail.enabled
      mountSubtitleToggleButton(subtitlesEnabled, dispatchToggle)
    }
    window.addEventListener('lets-sub-it:subtitle-enabled-changed', handleEnabledChanged)
```

In `ctx.onInvalidated`, add:

```ts
      window.removeEventListener('lets-sub-it:subtitle-enabled-changed', handleEnabledChanged)
```

- [ ] **Step 6: Run overlay tests and typecheck**

Run:

```bash
cd extension && mise exec -- npx vitest run src/content/YoutubeOverlay.test.ts
cd extension && mise exec -- npm run typecheck
```

Expected: PASS.

- [ ] **Step 7: Commit toggle synchronization**

Run:

```bash
git add extension/entrypoints/youtube.content.ts extension/src/content/YoutubeOverlay.vue extension/src/content/YoutubeOverlay.test.ts
git commit -m "feat(extension): sync YouTube subtitle toggle state"
```

---

### Task 8: Full Extension Verification

**Files:**
- No planned source edits unless verification reveals failures.

- [ ] **Step 1: Run full extension test suite**

Run:

```bash
cd extension && mise exec -- npm run test
```

Expected: PASS.

- [ ] **Step 2: Run extension typecheck**

Run:

```bash
cd extension && mise exec -- npm run typecheck
```

Expected: PASS.

- [ ] **Step 3: Run repository diff check**

Run:

```bash
git diff --check
```

Expected: no output.

- [ ] **Step 4: Inspect final status**

Run:

```bash
git status --short
```

Expected: no uncommitted files, unless verification forced a fix commit.

- [ ] **Step 5: Commit verification fixes when verification produced edits**

When Step 1, 2, or 3 produces edits, run the focused command that originally failed, then commit only those fixes:

```bash
git add extension
git commit -m "fix(extension): stabilize YouTube subtitle UI settings"
```

Expected: commit succeeds and `git status --short` is clean.

---

## Self-Review Notes

- Spec coverage: Tasks 3 and 4 cover player-container mounting and control-bar insertion; Task 5 covers removing old floating controls, font-size application, settings update, and mode reload; Task 6 covers popup tabs and settings persistence/notification; Task 8 covers required verification.
- Placeholder scan: Plan contains no unfinished placeholder markers, no open-ended implementation instructions without concrete code, and no missing file paths.
- Type consistency: `subtitleFontSizePx`, `subtitleMode`, `Settings`, `SubtitleMode`, `lets-sub-it:settings-updated`, `lets-sub-it:toggle-subtitles`, and `lets-sub-it:subtitle-enabled-changed` are used consistently across tasks.
