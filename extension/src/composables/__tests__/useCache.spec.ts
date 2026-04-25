import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useCache } from '../useCache'
import type { LocalCacheEntry, UserPreferences } from '@/types'

const mockStorage: Record<string, unknown> = {}

beforeEach(() => {
  Object.keys(mockStorage).forEach((k) => delete mockStorage[k])

  vi.stubGlobal('chrome', {
    storage: {
      local: {
        get: vi.fn((keys: string | string[] | Record<string, unknown>) => {
          if (typeof keys === 'string') {
            return Promise.resolve({ [keys]: mockStorage[keys] ?? null })
          }
          if (Array.isArray(keys)) {
            const result: Record<string, unknown> = {}
            for (const k of keys) result[k] = mockStorage[k] ?? null
            return Promise.resolve(result)
          }
          return Promise.resolve({})
        }),
        set: vi.fn((items: Record<string, unknown>) => {
          Object.assign(mockStorage, items)
          return Promise.resolve()
        }),
      },
    },
  })
})

describe('useCache', () => {
  it('returns null for missing cache entry', async () => {
    const { getCacheEntry } = useCache()
    const result = await getCacheEntry('video1', 'zh-CN')
    expect(result).toBeNull()
  })

  it('stores and retrieves a cache entry', async () => {
    const { getCacheEntry, setCacheEntry } = useCache()
    const entry: LocalCacheEntry = {
      videoId: 'abc123',
      targetLanguage: 'zh-CN',
      jobId: 'job_xyz',
      selectedMode: 'bilingual',
      lastSyncedAt: '2026-04-25T00:00:00Z',
    }
    await setCacheEntry(entry)
    const result = await getCacheEntry('abc123', 'zh-CN')
    expect(result).toEqual(entry)
  })

  it('stores and retrieves user preferences', async () => {
    const { getPreferences, setPreferences } = useCache()
    const prefs: UserPreferences = {
      videoId: 'abc123',
      targetLanguage: 'zh-CN',
      selectedMode: 'translated',
    }
    await setPreferences(prefs)
    const result = await getPreferences('abc123')
    expect(result).toEqual(prefs)
  })

  it('returns null for missing preferences', async () => {
    const { getPreferences } = useCache()
    const result = await getPreferences('nonexistent')
    expect(result).toBeNull()
  })
})
