import { describe, expect, it } from 'vitest'
import { getVideoIdFromLocationHref } from './page-watch'

describe('getVideoIdFromLocationHref', () => {
  it('returns videoId for YouTube watch URLs', () => {
    expect(
      getVideoIdFromLocationHref('https://www.youtube.com/watch?v=video_123'),
    ).toBe('video_123')
  })

  it('returns null outside YouTube watch pages', () => {
    expect(getVideoIdFromLocationHref('https://www.youtube.com/')).toBeNull()
  })
})
