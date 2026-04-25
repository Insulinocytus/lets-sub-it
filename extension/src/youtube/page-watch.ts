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
