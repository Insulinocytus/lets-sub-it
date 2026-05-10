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
    expect(first?.id).toBe(PLAYER_OVERLAY_HOST_ID)
    expect(first?.parentElement).toBe(player)
    expect(first?.style.position).toBe('absolute')
    expect(first?.style.pointerEvents).toBe('none')
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
    expect(host?.parentElement).toBe(newPlayer)
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
