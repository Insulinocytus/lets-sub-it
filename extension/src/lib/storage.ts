import type { LocalCacheEntry, SubtitleMode } from '../types';

const KEY = 'subtitle-cache';

export async function getCache(): Promise<Record<string, LocalCacheEntry>> {
  const result = await chrome.storage.local.get(KEY);
  return (result[KEY] as Record<string, LocalCacheEntry> | undefined) ?? {};
}

export async function getCacheEntry(videoId: string): Promise<LocalCacheEntry | undefined> {
  const cache = await getCache();
  return cache[videoId];
}

export async function saveCacheEntry(entry: LocalCacheEntry): Promise<void> {
  const current = await getCache();
  const existing = current[entry.videoId];
  const subtitleUrls =
    existing?.subtitleUrls || entry.subtitleUrls
      ? {
          ...existing?.subtitleUrls,
          ...entry.subtitleUrls,
        }
      : undefined;

  current[entry.videoId] = {
    ...existing,
    ...entry,
    ...(subtitleUrls ? { subtitleUrls } : {}),
  };

  await chrome.storage.local.set({ [KEY]: current });
}

export async function setSelectedMode(
  videoId: string,
  selectedMode: SubtitleMode,
): Promise<LocalCacheEntry | undefined> {
  const existing = await getCacheEntry(videoId);
  if (!existing) {
    return undefined;
  }

  const updatedEntry: LocalCacheEntry = {
    ...existing,
    selectedMode,
  };

  await saveCacheEntry(updatedEntry);

  return updatedEntry;
}
