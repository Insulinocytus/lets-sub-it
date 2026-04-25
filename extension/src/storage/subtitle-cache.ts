import { storage } from 'wxt/utils/storage'
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

function subtitleAssetKey(videoId: string, targetLanguage: LanguageCode) {
  return `local:subtitleAssets:${videoId}:${targetLanguage}` as const
}

function videoPreferenceKey(videoId: string) {
  return `local:videoPreferences:${videoId}` as const
}

export async function getCachedSubtitleAsset(
  videoId: string,
  targetLanguage: LanguageCode,
): Promise<SubtitleAssetCacheEntry | null> {
  const item = storage.defineItem<SubtitleAssetCacheEntry>(
    subtitleAssetKey(videoId, targetLanguage),
  )
  return item.getValue()
}

export async function setCachedSubtitleAsset(
  asset: SubtitleAsset,
  selectedMode: SubtitleMode,
  lastSyncedAt: string,
): Promise<SubtitleAssetCacheEntry> {
  const entry: SubtitleAssetCacheEntry = {
    ...asset,
    selectedMode,
    lastSyncedAt,
  }
  const item = storage.defineItem<SubtitleAssetCacheEntry>(
    subtitleAssetKey(asset.videoId, asset.targetLanguage),
  )
  await item.setValue(entry)
  await setVideoPreference({
    videoId: asset.videoId,
    targetLanguage: asset.targetLanguage,
    selectedMode,
  })
  return entry
}

export async function getVideoPreference(
  videoId: string,
): Promise<VideoPreference | null> {
  const item = storage.defineItem<VideoPreference>(videoPreferenceKey(videoId))
  return item.getValue()
}

export async function setVideoPreference(
  preference: VideoPreference,
): Promise<VideoPreference> {
  const item = storage.defineItem<VideoPreference>(
    videoPreferenceKey(preference.videoId),
  )
  await item.setValue(preference)
  return preference
}

export async function updateCachedSubtitleMode(
  videoId: string,
  targetLanguage: LanguageCode,
  mode: SubtitleMode,
): Promise<SubtitleAssetCacheEntry | null> {
  const current = await getCachedSubtitleAsset(videoId, targetLanguage)
  if (!current) {
    return null
  }

  const next: SubtitleAssetCacheEntry = {
    ...current,
    selectedMode: mode,
  }
  const item = storage.defineItem<SubtitleAssetCacheEntry>(
    subtitleAssetKey(videoId, targetLanguage),
  )
  await item.setValue(next)
  await setVideoPreference({ videoId, targetLanguage, selectedMode: mode })
  return next
}
