import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import contentScript from '../../entrypoints/youtube.content'
import { PLAYER_OVERLAY_HOST_ID, SUBTITLE_TOGGLE_BUTTON_ID } from './player-ui'

const unmount = vi.fn()

vi.mock('vue', () => ({
  createApp: vi.fn(() => ({
    mount: vi.fn(),
    unmount,
  })),
}))

vi.mock('@/content/YoutubeOverlay.vue', () => ({ default: {} }))
vi.mock('@/youtube/youtube-content.css', () => ({}))

describe('YouTube content script', () => {
  let mutationCallback: MutationCallback | null = null
  const disconnect = vi.fn()
  const onInvalidated = vi.fn()
  const originalMutationObserver = globalThis.MutationObserver
  const originalLocation = window.location

  beforeEach(() => {
    document.body.innerHTML = ''
    mutationCallback = null
    disconnect.mockClear()
    onInvalidated.mockClear()
    unmount.mockClear()
    class FakeMutationObserver {
      constructor(callback: MutationCallback) {
        mutationCallback = callback
      }

      observe = vi.fn()
      disconnect = disconnect
      takeRecords = vi.fn(() => [])
    }
    globalThis.MutationObserver = FakeMutationObserver as unknown as typeof MutationObserver
  })

  afterEach(() => {
    globalThis.MutationObserver = originalMutationObserver
    setLocationHref(originalLocation.href)
  })

  function setLocationHref(href: string) {
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: new URL(href),
    })
  }

  function runContentScript() {
    const run = contentScript.main as (ctx: {
      onInvalidated: typeof onInvalidated
    }) => void
    run({
      onInvalidated,
    })
  }

  it('cleans up the mounted overlay when the YouTube player disappears', () => {
    setLocationHref('https://www.youtube.com/watch?v=video_123')
    document.body.innerHTML = `
      <div id="movie_player" class="html5-video-player">
        <div class="ytp-right-controls"></div>
      </div>
    `

    runContentScript()
    expect(document.getElementById(PLAYER_OVERLAY_HOST_ID)).not.toBeNull()
    expect(document.getElementById(SUBTITLE_TOGGLE_BUTTON_ID)).not.toBeNull()

    document.getElementById('movie_player')?.remove()
    mutationCallback?.([], {} as MutationObserver)

    expect(unmount).toHaveBeenCalledOnce()
    expect(document.getElementById(PLAYER_OVERLAY_HOST_ID)).toBeNull()
    expect(document.getElementById(SUBTITLE_TOGGLE_BUTTON_ID)).toBeNull()
  })

  it('waits to mount the player UI until a SPA navigation reaches a watch page', () => {
    setLocationHref('https://www.youtube.com/')
    document.body.innerHTML = `
      <div id="movie_player" class="html5-video-player">
        <div class="ytp-right-controls"></div>
      </div>
    `

    runContentScript()

    expect(document.getElementById(PLAYER_OVERLAY_HOST_ID)).toBeNull()
    expect(document.getElementById(SUBTITLE_TOGGLE_BUTTON_ID)).toBeNull()

    setLocationHref('https://www.youtube.com/watch?v=video_123')
    mutationCallback?.([], {} as MutationObserver)

    expect(document.getElementById(PLAYER_OVERLAY_HOST_ID)).not.toBeNull()
    expect(document.getElementById(SUBTITLE_TOGGLE_BUTTON_ID)).not.toBeNull()
  })
})
