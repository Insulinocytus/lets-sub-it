import { storage } from 'wxt/utils/storage'
import { normalizeBackendBaseUrl } from '@/api/backend-client'
import type { LanguageCode, SubtitleAsset, SubtitleMode } from '@/api/messages'

export type VideoPreference = {
  videoId: string
  targetLanguage: LanguageCode
  selectedMode: SubtitleMode
}

export type SubtitleAssetCacheEntry = SubtitleAsset & {
  selectedMode: SubtitleMode
  lastSyncedAt: string
}

function subtitleAssetKey(
  backendBaseUrl: string,
  videoId: string,
  targetLanguage: LanguageCode,
) {
  const backendOrigin = normalizeBackendBaseUrl(backendBaseUrl)
  return `local:subtitleAssets:${encodeURIComponent(backendOrigin)}:${encodeURIComponent(videoId)}:${targetLanguage}` as const
}

function videoPreferenceKey(backendBaseUrl: string, videoId: string) {
  const backendOrigin = normalizeBackendBaseUrl(backendBaseUrl)
  return `local:videoPreferences:${encodeURIComponent(backendOrigin)}:${encodeURIComponent(videoId)}` as const
}

export async function getCachedSubtitleAsset(
  backendBaseUrl: string,
  videoId: string,
  targetLanguage: LanguageCode,
): Promise<SubtitleAssetCacheEntry | null> {
  const item = storage.defineItem<SubtitleAssetCacheEntry>(
    subtitleAssetKey(backendBaseUrl, videoId, targetLanguage),
  )
  return item.getValue()
}

export async function setCachedSubtitleAsset(
  asset: SubtitleAsset,
  selectedMode: SubtitleMode,
  lastSyncedAt: string,
  backendBaseUrl: string,
): Promise<SubtitleAssetCacheEntry> {
  const entry: SubtitleAssetCacheEntry = {
    ...asset,
    selectedMode,
    lastSyncedAt,
  }
  const item = storage.defineItem<SubtitleAssetCacheEntry>(
    subtitleAssetKey(backendBaseUrl, asset.videoId, asset.targetLanguage),
  )
  await item.setValue(entry)
  await setVideoPreference({
    backendBaseUrl,
    videoId: asset.videoId,
    targetLanguage: asset.targetLanguage,
    selectedMode,
  })
  return entry
}

export async function getVideoPreference(
  backendBaseUrl: string,
  videoId: string,
): Promise<VideoPreference | null> {
  const item = storage.defineItem<VideoPreference>(
    videoPreferenceKey(backendBaseUrl, videoId),
  )
  return item.getValue()
}

export async function setVideoPreference(
  preference: VideoPreference & { backendBaseUrl: string },
): Promise<VideoPreference> {
  const { backendBaseUrl, ...storedPreference } = preference
  const item = storage.defineItem<VideoPreference>(
    videoPreferenceKey(backendBaseUrl, preference.videoId),
  )
  await item.setValue(storedPreference)
  return storedPreference
}

export async function updateCachedSubtitleMode(
  backendBaseUrl: string,
  videoId: string,
  targetLanguage: LanguageCode,
  mode: SubtitleMode,
): Promise<SubtitleAssetCacheEntry | null> {
  const current = await getCachedSubtitleAsset(
    backendBaseUrl,
    videoId,
    targetLanguage,
  )
  if (!current) {
    return null
  }

  const next: SubtitleAssetCacheEntry = {
    ...current,
    selectedMode: mode,
  }
  const item = storage.defineItem<SubtitleAssetCacheEntry>(
    subtitleAssetKey(backendBaseUrl, videoId, targetLanguage),
  )
  await item.setValue(next)
  await setVideoPreference({
    backendBaseUrl,
    videoId,
    targetLanguage,
    selectedMode: mode,
  })
  return next
}
