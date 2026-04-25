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
