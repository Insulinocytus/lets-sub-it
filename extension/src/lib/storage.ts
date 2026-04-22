import type { LocalCacheEntry, SubtitleMode } from '../types';

const KEY = 'subtitle-cache';
let cacheWriteQueue: Promise<void> = Promise.resolve();

async function readCache(): Promise<Record<string, LocalCacheEntry>> {
  const result = await chrome.storage.local.get(KEY);
  return (result[KEY] as Record<string, LocalCacheEntry> | undefined) ?? {};
}

function buildMergedEntry(
  existing: LocalCacheEntry | undefined,
  entry: LocalCacheEntry,
): LocalCacheEntry {
  const subtitleUrls =
    existing?.subtitleUrls || entry.subtitleUrls
      ? {
          ...existing?.subtitleUrls,
          ...entry.subtitleUrls,
        }
      : undefined;

  return {
    ...existing,
    ...entry,
    ...(subtitleUrls ? { subtitleUrls } : {}),
  };
}

function enqueueCacheWrite<T>(
  update: (current: Record<string, LocalCacheEntry>) => Promise<T> | T,
): Promise<T> {
  const run = cacheWriteQueue.then(async () => update(await readCache()));
  cacheWriteQueue = run.then(
    () => undefined,
    () => undefined,
  );
  return run;
}

export async function getCache(): Promise<Record<string, LocalCacheEntry>> {
  await cacheWriteQueue;
  return readCache();
}

export async function getCacheEntry(videoId: string): Promise<LocalCacheEntry | undefined> {
  const cache = await getCache();
  return cache[videoId];
}

export async function saveCacheEntry(entry: LocalCacheEntry): Promise<void> {
  await enqueueCacheWrite(async (current) => {
    const nextCache = { ...current };
    nextCache[entry.videoId] = buildMergedEntry(nextCache[entry.videoId], entry);
    await chrome.storage.local.set({ [KEY]: nextCache });
  });
}

export async function setSelectedMode(
  videoId: string,
  selectedMode: SubtitleMode,
): Promise<LocalCacheEntry | undefined> {
  return enqueueCacheWrite(async (current) => {
    const existing = current[videoId];
    if (!existing) {
      return undefined;
    }

    const nextCache = { ...current };
    const updatedEntry: LocalCacheEntry = {
      ...existing,
      selectedMode,
    };

    nextCache[videoId] = buildMergedEntry(existing, updatedEntry);
    await chrome.storage.local.set({ [KEY]: nextCache });
    return nextCache[videoId];
  });
}
