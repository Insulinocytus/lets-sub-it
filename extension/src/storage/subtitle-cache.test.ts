import { beforeEach, describe, expect, it } from 'vitest'
import { fakeBrowser } from 'wxt/testing/fake-browser'
import type { SubtitleAsset } from '@/api/messages'
import {
  getCachedSubtitleAsset,
  getVideoPreference,
  setCachedSubtitleAsset,
  updateCachedSubtitleMode,
} from './subtitle-cache'

const asset: SubtitleAsset = {
  jobId: 'job_123',
  videoId: 'video_123',
  sourceLanguage: 'en',
  targetLanguage: 'zh-CN',
  files: {
    source: '/subtitle-files/job_123/source',
    translated: '/subtitle-files/job_123/translated',
    bilingual: '/subtitle-files/job_123/bilingual',
  },
  createdAt: '2026-04-25T00:00:00Z',
}

describe('subtitle cache storage', () => {
  beforeEach(() => {
    fakeBrowser.reset()
  })

  it('stores and reads an asset by videoId and targetLanguage', async () => {
    await setCachedSubtitleAsset(asset, 'translated', '2026-04-25T00:01:00Z')

    await expect(getCachedSubtitleAsset('video_123', 'zh-CN')).resolves.toEqual({
      ...asset,
      selectedMode: 'translated',
      lastSyncedAt: '2026-04-25T00:01:00Z',
    })
  })

  it('stores the video preference when an asset is cached', async () => {
    await setCachedSubtitleAsset(asset, 'translated', '2026-04-25T00:01:00Z')

    await expect(getVideoPreference('video_123')).resolves.toEqual({
      videoId: 'video_123',
      targetLanguage: 'zh-CN',
      selectedMode: 'translated',
    })
  })

  it('updates the selected subtitle mode', async () => {
    await setCachedSubtitleAsset(asset, 'translated', '2026-04-25T00:01:00Z')

    const updated = await updateCachedSubtitleMode('video_123', 'zh-CN', 'bilingual')

    expect(updated?.selectedMode).toBe('bilingual')
    await expect(getVideoPreference('video_123')).resolves.toEqual({
      videoId: 'video_123',
      targetLanguage: 'zh-CN',
      selectedMode: 'bilingual',
    })
  })

  it('separates cache entries by targetLanguage', async () => {
    await setCachedSubtitleAsset(asset, 'translated', '2026-04-25T00:01:00Z')

    await expect(getCachedSubtitleAsset('video_123', 'en')).resolves.toBeNull()
  })

  it('keeps special videoId keys isolated from ordinary keys', async () => {
    const specialAsset: SubtitleAsset = {
      ...asset,
      videoId: 'video/123?ref=abc#frag',
      targetLanguage: 'zh-CN',
    }
    const ordinaryAsset: SubtitleAsset = {
      ...asset,
      videoId: 'video_ordinary',
      targetLanguage: 'zh-CN',
    }

    await setCachedSubtitleAsset(ordinaryAsset, 'translated', '2026-04-25T00:01:00Z')
    await setCachedSubtitleAsset(specialAsset, 'bilingual', '2026-04-25T00:02:00Z')

    await expect(getCachedSubtitleAsset('video_ordinary', 'zh-CN')).resolves.toEqual({
      ...ordinaryAsset,
      selectedMode: 'translated',
      lastSyncedAt: '2026-04-25T00:01:00Z',
    })
    await expect(
      getCachedSubtitleAsset('video/123?ref=abc#frag', 'zh-CN'),
    ).resolves.toEqual({
      ...specialAsset,
      selectedMode: 'bilingual',
      lastSyncedAt: '2026-04-25T00:02:00Z',
    })
  })
})
