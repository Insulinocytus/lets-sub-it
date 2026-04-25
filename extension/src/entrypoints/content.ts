import { defineContentScript } from 'wxt/sandbox'
import { parseVtt, findCueAtTime } from '@/lib/vtt-parser'
import type {
  ExtensionMessage,
  ExtensionResponse,
  SubtitleCue,
  SubtitleMode,
  LocalCacheEntry,
  UserPreferences,
} from '@/types'

const CONTAINER_ID = 'lsi-subtitle-host'
const MAX_RETRIES = 3
const RETRY_INTERVAL_MS = 1000

type SendMessage = (message: ExtensionMessage) => Promise<ExtensionResponse>
type GetCurrentGeneration = () => number
type GetCacheEntry = (cacheKey: string) => Promise<LocalCacheEntry | undefined>
type LoadSubtitlesForGeneration = (
  entry: LocalCacheEntry,
  mode: SubtitleMode,
  generation: number,
) => Promise<void>
type LoadSubtitleCues = () => Promise<SubtitleCue[]>
type ApplySubtitleCues = (cues: SubtitleCue[]) => void

export interface SubtitleStorageLoadDecision {
  nextMode: SubtitleMode
  nextTargetLanguage: string | null
  entryToLoad: LocalCacheEntry | null
  cacheKeyToRead: string | null
  modeToLoad: SubtitleMode
}

export function decideSubtitleStorageLoad(
  changes: Record<string, chrome.storage.StorageChange>,
  videoId: string,
  currentTargetLanguage: string | null,
  currentMode: SubtitleMode,
): SubtitleStorageLoadDecision {
  const prefsKey = `prefs:${videoId}`
  const prefs = changes[prefsKey]?.newValue as UserPreferences | undefined
  const nextMode = prefs?.selectedMode ?? currentMode
  const nextTargetLanguage = prefs?.targetLanguage ?? currentTargetLanguage
  let entryToLoad: LocalCacheEntry | null = null
  let cacheKeyToRead =
    prefs?.targetLanguage ? `cache:${videoId}:${prefs.targetLanguage}` : null

  for (const [key, change] of Object.entries(changes)) {
    if (!key.startsWith('cache:')) continue

    const entry = change.newValue as LocalCacheEntry | undefined
    if (
      entry?.jobId &&
      entry.videoId === videoId &&
      entry.targetLanguage === nextTargetLanguage
    ) {
      entryToLoad = entry
      cacheKeyToRead = null
      break
    }
  }

  return {
    nextMode,
    nextTargetLanguage,
    entryToLoad,
    cacheKeyToRead,
    modeToLoad: nextMode,
  }
}

export async function loadCacheEntryForGeneration(
  cacheKey: string,
  mode: SubtitleMode,
  generation: number,
  getCurrentGeneration: GetCurrentGeneration,
  getCacheEntry: GetCacheEntry,
  loadSubtitles: LoadSubtitlesForGeneration,
): Promise<void> {
  const entry = await getCacheEntry(cacheKey)
  if (generation !== getCurrentGeneration()) return
  if (entry?.jobId) {
    await loadSubtitles(entry, mode, generation)
  }
}

export async function applySubtitleLoadForGeneration(
  generation: number,
  getCurrentGeneration: GetCurrentGeneration,
  loadCues: LoadSubtitleCues,
  applyCues: ApplySubtitleCues,
): Promise<void> {
  const cues = await loadCues()
  if (generation !== getCurrentGeneration()) return
  applyCues(cues)
}

export async function fetchSubtitleText(
  jobId: string,
  mode: SubtitleMode,
  sendMessage: SendMessage = (message) => chrome.runtime.sendMessage(message),
): Promise<string> {
  const response = await sendMessage({
    type: 'GET_SUBTITLE_FILE',
    payload: { jobId, mode },
  })
  if (!response.success) {
    throw new Error(response.error ?? 'get subtitle file failed')
  }
  if (typeof response.data !== 'string') {
    throw new Error('invalid subtitle file response')
  }
  return response.data
}

export default defineContentScript({
  matches: ['*://www.youtube.com/watch*'],
  main() {
    // ── State ──
    let videoElement: HTMLVideoElement | null = null
    let shadowHost: HTMLDivElement | null = null
    let cueDisplayEl: HTMLSpanElement | null = null
    let subtitleCues: SubtitleCue[] = []
    let subtitleMode: SubtitleMode = 'bilingual'
    let currentTargetLang: string | null = null
    let currentCueText: string | null = null
    let retryCount = 0
    let rAFHandle: number | null = null
    let subtitleLoadGeneration = 0
    const cleanups: (() => void)[] = []

    // ── Shadow DOM creation and injection ──

    function createShadowHost(): void {
      if (document.getElementById(CONTAINER_ID)) return

      const host = document.createElement('div')
      host.id = CONTAINER_ID
      host.style.cssText = `
        position: absolute;
        bottom: 60px;
        left: 0;
        right: 0;
        pointer-events: none;
        z-index: 100;
        display: flex;
        justify-content: center;
      `
      shadowHost = host

      const shadow = host.attachShadow({ mode: 'open' })

      const style = document.createElement('style')
      style.textContent = `
        .cue-container {
          text-align: center;
          padding: 0 16px;
          max-width: 80%;
        }
        .cue-text {
          background: rgba(0, 0, 0, 0.7);
          color: #fff;
          font-family: "YouTube Noto", Roboto, Arial, sans-serif;
          font-size: 20px;
          line-height: 1.4;
          padding: 4px 12px;
          border-radius: 4px;
          display: inline-block;
          text-align: center;
          white-space: pre-wrap;
          word-break: break-word;
        }
      `
      shadow.appendChild(style)

      const container = document.createElement('div')
      container.className = 'cue-container'
      const textEl = document.createElement('span')
      textEl.className = 'cue-text'
      textEl.style.display = 'none'
      container.appendChild(textEl)
      shadow.appendChild(container)
      cueDisplayEl = textEl
    }

    function injectShadowHost(): void {
      const player =
        document.querySelector('#movie_player') ||
        document.querySelector('.html5-video-player')
      if (player && shadowHost && !shadowHost.parentElement) {
        player.appendChild(shadowHost)
      }
    }

    // ── Rendering ──

    function renderCue(): void {
      const time = videoElement?.currentTime ?? 0
      const cue = findCueAtTime(subtitleCues, time)
      const text = cue?.text ?? null
      if (text !== currentCueText) {
        currentCueText = text
        if (cueDisplayEl) {
          cueDisplayEl.textContent = text
          cueDisplayEl.style.display = text ? 'inline-block' : 'none'
        }
      }
    }

    function onTimeUpdate(): void {
      if (rAFHandle !== null) return
      rAFHandle = requestAnimationFrame(() => {
        rAFHandle = null
        renderCue()
      })
    }

    // ── Subtitle data fetching ──

    function getVideoId(): string | null {
      return new URLSearchParams(location.search).get('v')
    }

    async function fetchSubtitles(
      entry: LocalCacheEntry,
      mode: SubtitleMode,
      generation: number,
    ): Promise<void> {
      try {
        await applySubtitleLoadForGeneration(
          generation,
          () => subtitleLoadGeneration,
          async () => {
            const vtt = await fetchSubtitleText(entry.jobId, mode)
            return parseVtt(vtt).cues
          },
          (cues) => {
            subtitleCues = cues
            currentCueText = null // force re-render on next tick
            renderCue()
          },
        )
      } catch (err) {
        console.error('[LSI] Failed to fetch subtitles:', err)
      }
    }

    async function getCacheEntry(cacheKey: string): Promise<LocalCacheEntry | undefined> {
      const result = await chrome.storage.local.get(cacheKey)
      return result[cacheKey] as LocalCacheEntry | undefined
    }

    async function loadFromCache(generation: number): Promise<void> {
      const videoId = getVideoId()
      if (!videoId || generation !== subtitleLoadGeneration) return

      const prefsKey = `prefs:${videoId}`
      const { [prefsKey]: prefs } = await chrome.storage.local.get(prefsKey)
      if (generation !== subtitleLoadGeneration) return

      const userPrefs = prefs as UserPreferences | undefined

      if (userPrefs?.selectedMode) {
        subtitleMode = userPrefs.selectedMode
      }

      const targetLang = userPrefs?.targetLanguage
      if (targetLang) {
        currentTargetLang = targetLang
        const cacheKey = `cache:${videoId}:${targetLang}`
        const { [cacheKey]: cacheEntry } = await chrome.storage.local.get(cacheKey)
        if (generation !== subtitleLoadGeneration) return

        const entry = cacheEntry as LocalCacheEntry | undefined
        if (entry?.jobId) {
          await fetchSubtitles(entry, subtitleMode, generation)
        }
      }
    }

    // ── Storage change listener ──

    function setupStorageListener(): void {
      const handler = (
        changes: Record<string, chrome.storage.StorageChange>,
      ): void => {
        const videoId = getVideoId()
        if (!videoId) return

        const decision = decideSubtitleStorageLoad(
          changes,
          videoId,
          currentTargetLang,
          subtitleMode,
        )
        subtitleMode = decision.nextMode
        currentTargetLang = decision.nextTargetLanguage

        const hasPrefsChange = Object.hasOwn(changes, `prefs:${videoId}`)
        if (!hasPrefsChange && !decision.entryToLoad && !decision.cacheKeyToRead) {
          return
        }

        const generation = ++subtitleLoadGeneration
        if (decision.entryToLoad) {
          fetchSubtitles(decision.entryToLoad, decision.modeToLoad, generation)
        }
        if (decision.cacheKeyToRead) {
          const { cacheKeyToRead, modeToLoad } = decision
          loadCacheEntryForGeneration(
            cacheKeyToRead,
            modeToLoad,
            generation,
            () => subtitleLoadGeneration,
            getCacheEntry,
            fetchSubtitles,
          )
        }
      }

      chrome.storage.onChanged.addListener(handler)
      cleanups.push(() => {
        chrome.storage.onChanged.removeListener(handler)
      })
    }

    // ── Video element detection ──

    function waitForVideo(generation: number, resolve: () => void): void {
      if (generation !== subtitleLoadGeneration) return

      const el = document.querySelector('video')
      if (el) {
        videoElement = el
        videoElement.addEventListener('timeupdate', onTimeUpdate)
        resolve()
        return
      }

      retryCount++
      if (retryCount >= MAX_RETRIES) {
        console.warn('[LSI] Video element not found after max retries')
        return
      }

      setTimeout(() => waitForVideo(generation, resolve), RETRY_INTERVAL_MS)
    }

    // ── SPA navigation detection ──

    function setupSpaNavigation(): void {
      const origPushState = history.pushState.bind(history)
      const origReplaceState = history.replaceState.bind(history)

      function checkUrlChange(): void {
        const isWatchPage = location.pathname === '/watch'
        if (isWatchPage) {
          reinit()
        } else {
          teardown()
        }
      }

      history.pushState = function (...args: Parameters<typeof origPushState>) {
        origPushState(...args)
        checkUrlChange()
      }

      history.replaceState = function (
        ...args: Parameters<typeof origReplaceState>
      ) {
        origReplaceState(...args)
        checkUrlChange()
      }

      const popstateHandler = (): void => checkUrlChange()
      window.addEventListener('popstate', popstateHandler)

      cleanups.push(() => {
        history.pushState = origPushState
        history.replaceState = origReplaceState
        window.removeEventListener('popstate', popstateHandler)
      })
    }

    // ── Init / Teardown ──

    function init(): void {
      const generation = ++subtitleLoadGeneration
      createShadowHost()
      setupStorageListener()
      setupSpaNavigation()

      waitForVideo(generation, () => {
        if (generation !== subtitleLoadGeneration) return

        injectShadowHost()
        loadFromCache(generation)
      })
    }

    function teardown(): void {
      subtitleLoadGeneration++
      if (rAFHandle !== null) {
        cancelAnimationFrame(rAFHandle)
        rAFHandle = null
      }
      videoElement?.removeEventListener('timeupdate', onTimeUpdate)
      cleanups.forEach((fn) => fn())
      cleanups.length = 0
      shadowHost?.remove()
      shadowHost = null
      cueDisplayEl = null
      videoElement = null
      subtitleCues = []
      currentCueText = null
      currentTargetLang = null
      retryCount = 0
    }

    function reinit(): void {
      teardown()
      init()
    }

    // ── Start ──
    init()
  },
})
