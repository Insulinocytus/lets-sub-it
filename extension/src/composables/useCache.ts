import type { LocalCacheEntry, SubtitleMode, UserPreferences } from '@/types'

function cacheKey(videoId: string, targetLanguage: string): string {
  return `cache:${videoId}:${targetLanguage}`
}

function prefsKey(videoId: string): string {
  return `prefs:${videoId}`
}

export function useCache() {
  async function getCacheEntry(videoId: string, targetLanguage: string): Promise<LocalCacheEntry | null> {
    const result = await chrome.storage.local.get(cacheKey(videoId, targetLanguage))
    return (result[cacheKey(videoId, targetLanguage)] as LocalCacheEntry) ?? null
  }

  async function setCacheEntry(entry: LocalCacheEntry): Promise<void> {
    await chrome.storage.local.set({ [cacheKey(entry.videoId, entry.targetLanguage)]: entry })
  }

  async function getPreferences(videoId: string): Promise<UserPreferences | null> {
    const result = await chrome.storage.local.get(prefsKey(videoId))
    return (result[prefsKey(videoId)] as UserPreferences) ?? null
  }

  async function setPreferences(prefs: UserPreferences): Promise<void> {
    await chrome.storage.local.set({ [prefsKey(prefs.videoId)]: prefs })
  }

  async function setSubtitleSelection(entry: LocalCacheEntry, prefs: UserPreferences): Promise<void> {
    await chrome.storage.local.set({
      [cacheKey(entry.videoId, entry.targetLanguage)]: entry,
      [prefsKey(prefs.videoId)]: prefs,
    })
  }

  return { getCacheEntry, setCacheEntry, getPreferences, setPreferences, setSubtitleSelection }
}
