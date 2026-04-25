import { defineContentScript } from 'wxt/sandbox'
import { parseVtt, findCueAtTime } from '@/lib/vtt-parser'
import type { SubtitleCue, SubtitleMode, LocalCacheEntry, UserPreferences } from '@/types'

const BACKEND_BASE = 'http://127.0.0.1:8080'
const CONTAINER_ID = 'lsi-subtitle-host'
const MAX_RETRIES = 3
const RETRY_INTERVAL_MS = 1000

export default defineContentScript({
  matches: ['*://www.youtube.com/watch*'],
  main() {
    // ── State ──
    let videoElement: HTMLVideoElement | null = null
    let shadowHost: HTMLDivElement | null = null
    let cueDisplayEl: HTMLSpanElement | null = null
    let subtitleCues: SubtitleCue[] = []
    let subtitleMode: SubtitleMode = 'translated'
    let currentCueText: string | null = null
    let retryCount = 0
    let rAFHandle: number | null = null
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

    async function fetchSubtitles(entry: LocalCacheEntry): Promise<void> {
      const url = `${BACKEND_BASE}/subtitle-files/${entry.jobId}/${subtitleMode}`
      try {
        const res = await fetch(url)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const vtt = await res.text()
        const result = parseVtt(vtt)
        subtitleCues = result.cues
        currentCueText = null // force re-render on next tick
        renderCue()
      } catch (err) {
        console.error('[LSI] Failed to fetch subtitles:', err)
      }
    }

    async function loadFromCache(): Promise<void> {
      const videoId = getVideoId()
      if (!videoId) return

      const prefsKey = `prefs:${videoId}`
      const { [prefsKey]: prefs } = await chrome.storage.local.get(prefsKey)
      const userPrefs = prefs as UserPreferences | undefined

      if (userPrefs?.selectedMode) {
        subtitleMode = userPrefs.selectedMode
      }

      const targetLang = userPrefs?.targetLanguage
      if (targetLang) {
        const cacheKey = `cache:${videoId}:${targetLang}`
        const { [cacheKey]: cacheEntry } = await chrome.storage.local.get(cacheKey)
        const entry = cacheEntry as LocalCacheEntry | undefined
        if (entry?.jobId) {
          await fetchSubtitles(entry)
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

        for (const [key, change] of Object.entries(changes)) {
          // Cache key: cache:{videoId}:{targetLanguage}
          if (key.startsWith('cache:')) {
            const entry = change.newValue as LocalCacheEntry | undefined
            if (entry?.videoId === videoId) {
              fetchSubtitles(entry)
            }
          }
          // Prefs key: prefs:{videoId}
          if (key === `prefs:${videoId}`) {
            const prefs = change.newValue as UserPreferences | undefined
            if (prefs?.selectedMode) {
              subtitleMode = prefs.selectedMode
            }
            const targetLang = prefs?.targetLanguage
            if (targetLang) {
              const cacheKey = `cache:${videoId}:${targetLang}`
              chrome.storage.local.get(cacheKey).then((result) => {
                const entry = result[cacheKey] as LocalCacheEntry | undefined
                if (entry?.jobId) {
                  fetchSubtitles(entry)
                }
              })
            }
          }
        }
      }

      chrome.storage.onChanged.addListener(handler)
      cleanups.push(() => {
        chrome.storage.onChanged.removeListener(handler)
      })
    }

    // ── Video element detection ──

    function waitForVideo(resolve: () => void): void {
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

      setTimeout(() => waitForVideo(resolve), RETRY_INTERVAL_MS)
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
      createShadowHost()
      setupStorageListener()
      setupSpaNavigation()

      waitForVideo(() => {
        injectShadowHost()
        loadFromCache()
      })
    }

    function teardown(): void {
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
