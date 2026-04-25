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
