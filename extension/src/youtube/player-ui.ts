export const PLAYER_OVERLAY_HOST_ID = 'lets-sub-it-player-overlay-host'
export const SUBTITLE_TOGGLE_BUTTON_ID = 'lets-sub-it-subtitle-toggle'

export type PlayerOverlayHost = HTMLElement & {
  __letsSubItCleanedUp?: boolean
  __letsSubItCleanup?: () => void
}

const PLAYER_SELECTOR = '#movie_player.html5-video-player'
const RIGHT_CONTROLS_SELECTOR = '#movie_player .ytp-right-controls'
const STOPPED_EVENTS = ['click', 'mousedown', 'pointerdown', 'dblclick'] as const

export function findYouTubePlayer(): HTMLElement | null {
  return document.querySelector<HTMLElement>(PLAYER_SELECTOR)
}

export function cleanupPlayerOverlayHost(host: PlayerOverlayHost | null): void {
  if (!host || host.__letsSubItCleanedUp) {
    return
  }

  host.__letsSubItCleanedUp = true
  host.__letsSubItCleanup?.()
  host.remove()
}

export function ensurePlayerOverlayHost(player: HTMLElement | null): PlayerOverlayHost | null {
  if (!player) {
    return null
  }

  const existing = document.getElementById(PLAYER_OVERLAY_HOST_ID)
  if (existing?.parentElement === player) {
    return existing as PlayerOverlayHost
  }
  if (existing) {
    cleanupPlayerOverlayHost(existing as PlayerOverlayHost)
  }

  if (window.getComputedStyle(player).position === 'static') {
    player.style.position = 'relative'
  }

  const host = document.createElement('div') as PlayerOverlayHost
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

export function removeSubtitleToggleButton(): void {
  document.getElementById(SUBTITLE_TOGGLE_BUTTON_ID)?.remove()
}
