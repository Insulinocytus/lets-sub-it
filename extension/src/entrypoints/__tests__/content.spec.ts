import { describe, expect, it, vi } from 'vitest'
import {
  applySubtitleLoadForGeneration,
  decideSubtitleStorageLoad,
  fetchSubtitleText,
  loadCacheEntryForGeneration,
} from '../content'
import type { LocalCacheEntry, UserPreferences } from '@/types'

vi.mock('@/lib/vtt-parser', () => ({
  findCueAtTime: vi.fn(),
  parseVtt: vi.fn(),
}))

describe('content subtitle loading', () => {
  it('loads subtitle text through the background runtime message', async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    const sendMessage = vi.fn().mockResolvedValue({
      success: true,
      data: 'WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nHello',
    })

    const text = await fetchSubtitleText('job_123', 'bilingual', sendMessage)

    expect(sendMessage).toHaveBeenCalledWith({
      type: 'GET_SUBTITLE_FILE',
      payload: {
        jobId: 'job_123',
        mode: 'bilingual',
      },
    })
    expect(fetchMock).not.toHaveBeenCalled()
    expect(text).toBe('WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nHello')
  })

  it('uses prefs mode when prefs and cache changes arrive together', () => {
    const cacheEntry: LocalCacheEntry = {
      videoId: 'video-1',
      targetLanguage: 'zh-Hans',
      jobId: 'job_123',
      selectedMode: 'translated',
      lastSyncedAt: '2026-04-25T00:00:00.000Z',
    }
    const prefs: UserPreferences = {
      videoId: 'video-1',
      targetLanguage: 'zh-Hans',
      selectedMode: 'translated',
    }

    const decision = decideSubtitleStorageLoad(
      {
        'cache:video-1:zh-Hans': { newValue: cacheEntry },
        'prefs:video-1': { newValue: prefs },
      } as Record<string, chrome.storage.StorageChange>,
      'video-1',
      'zh-Hans',
      'bilingual',
    )

    expect(decision.entryToLoad).toBe(cacheEntry)
    expect(decision.modeToLoad).toBe('translated')
    expect(decision.nextMode).toBe('translated')
    expect(decision.nextTargetLanguage).toBe('zh-Hans')
    expect(decision.cacheKeyToRead).toBeNull()
  })

  it('keeps loading cache-only changes with the current mode', () => {
    const cacheEntry: LocalCacheEntry = {
      videoId: 'video-1',
      targetLanguage: 'zh-Hans',
      jobId: 'job_123',
      selectedMode: 'translated',
      lastSyncedAt: '2026-04-25T00:00:00.000Z',
    }

    const decision = decideSubtitleStorageLoad(
      {
        'cache:video-1:zh-Hans': { newValue: cacheEntry },
      } as Record<string, chrome.storage.StorageChange>,
      'video-1',
      'zh-Hans',
      'bilingual',
    )

    expect(decision.entryToLoad).toBe(cacheEntry)
    expect(decision.modeToLoad).toBe('bilingual')
    expect(decision.nextMode).toBe('bilingual')
    expect(decision.nextTargetLanguage).toBe('zh-Hans')
    expect(decision.cacheKeyToRead).toBeNull()
  })

  it('requests the matching cache when only prefs changes', () => {
    const prefs: UserPreferences = {
      videoId: 'video-1',
      targetLanguage: 'zh-Hans',
      selectedMode: 'translated',
    }

    const decision = decideSubtitleStorageLoad(
      {
        'prefs:video-1': { newValue: prefs },
      } as Record<string, chrome.storage.StorageChange>,
      'video-1',
      null,
      'bilingual',
    )

    expect(decision.entryToLoad).toBeNull()
    expect(decision.modeToLoad).toBe('translated')
    expect(decision.nextMode).toBe('translated')
    expect(decision.nextTargetLanguage).toBe('zh-Hans')
    expect(decision.cacheKeyToRead).toBe('cache:video-1:zh-Hans')
  })

  it('does not load an older prefs-only cache read after a newer generation starts', async () => {
    const cacheEntry: LocalCacheEntry = {
      videoId: 'video-1',
      targetLanguage: 'zh-Hans',
      jobId: 'job_old',
      selectedMode: 'bilingual',
      lastSyncedAt: '2026-04-25T00:00:00.000Z',
    }
    let currentGeneration = 1
    const getCacheEntry = vi.fn().mockImplementation(async () => {
      currentGeneration = 2
      return cacheEntry
    })
    const loadSubtitles = vi.fn()

    await loadCacheEntryForGeneration(
      'cache:video-1:zh-Hans',
      'bilingual',
      1,
      () => currentGeneration,
      getCacheEntry,
      loadSubtitles,
    )

    expect(getCacheEntry).toHaveBeenCalledWith('cache:video-1:zh-Hans')
    expect(loadSubtitles).not.toHaveBeenCalled()
  })

  it('does not apply a subtitle response after generation invalidation', async () => {
    let currentGeneration = 1
    const loadCues = vi.fn().mockImplementation(async () => {
      currentGeneration = 2
      return [{ start: 0, end: 1, text: 'old video subtitle' }]
    })
    const applyCues = vi.fn()

    await applySubtitleLoadForGeneration(
      1,
      () => currentGeneration,
      loadCues,
      applyCues,
    )

    expect(loadCues).toHaveBeenCalled()
    expect(applyCues).not.toHaveBeenCalled()
  })
})
