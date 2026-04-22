import type { LocalCacheEntry, SubtitleMode } from '../types';

const KEY = 'subtitle-cache';
let cacheWriteQueue: Promise<void> = Promise.resolve();

async function readCache(): Promise<Record<string, LocalCacheEntry>> {
  const result = await chrome.storage.local.get(KEY);
  return normalizeCacheEntries(
    (result[KEY] as Record<string, LocalCacheEntry> | undefined) ?? {},
  );
}

function normalizeCacheEntries(
  cache: Record<string, LocalCacheEntry>,
): Record<string, LocalCacheEntry> {
  const normalized: Record<string, LocalCacheEntry> = {};
  for (const [key, entry] of Object.entries(cache)) {
    const cacheKey = entry.jobId || key;
    normalized[cacheKey] = entry;
  }

  return normalized;
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

export async function getCacheEntry(jobId: string): Promise<LocalCacheEntry | undefined> {
  const cache = await getCache();
  return cache[jobId];
}

export async function getLatestCacheEntryForVideo(
  videoId: string,
): Promise<LocalCacheEntry | undefined> {
  const cache = await getCache();
  return Object.values(cache)
    .filter((entry) => entry.videoId === videoId)
    .sort((left, right) => right.lastSyncedAt.localeCompare(left.lastSyncedAt))[0];
}

export async function saveCacheEntry(entry: LocalCacheEntry): Promise<void> {
  await enqueueCacheWrite(async (current) => {
    const nextCache = { ...current };
    nextCache[entry.jobId] = buildMergedEntry(nextCache[entry.jobId], entry);
    await chrome.storage.local.set({ [KEY]: nextCache });
  });
}

export async function setSelectedMode(
  jobId: string,
  selectedMode: SubtitleMode,
): Promise<LocalCacheEntry | undefined> {
  return enqueueCacheWrite(async (current) => {
    const existing = current[jobId];
    if (!existing) {
      return undefined;
    }

    const nextCache = { ...current };
    const updatedEntry: LocalCacheEntry = {
      ...existing,
      selectedMode,
    };

    nextCache[jobId] = buildMergedEntry(existing, updatedEntry);
    await chrome.storage.local.set({ [KEY]: nextCache });
    return nextCache[jobId];
  });
}
