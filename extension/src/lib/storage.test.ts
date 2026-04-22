import { afterEach, describe, expect, it, vi } from 'vitest';

import { getCache, saveCacheEntry, setSelectedMode } from './storage';

function createChromeStorage(initialCache: Record<string, unknown> = {}) {
  let store = structuredClone(initialCache);

  return {
    storage: {
      local: {
        get: vi.fn(async (key: string) => {
          if (!(key in store)) {
            return {};
          }

          return { [key]: structuredClone(store[key]) };
        }),
        set: vi.fn(async (value: Record<string, unknown>) => {
          store = {
            ...store,
            ...structuredClone(value),
          };
        }),
      },
    },
  };
}

describe('storage helpers', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('returns an empty cache when nothing is stored', async () => {
    vi.stubGlobal('chrome', {
      storage: {
        local: {
          get: vi.fn().mockResolvedValue({}),
          set: vi.fn(),
        },
      },
    });

    await expect(getCache()).resolves.toEqual({});
  });

  it('stores selectedMode for a video', async () => {
    const chromeMock = createChromeStorage({ 'subtitle-cache': {} });

    vi.stubGlobal('chrome', chromeMock);

    await saveCacheEntry({
      videoId: 'abc123xyz00',
      jobId: 'job-1',
      selectedMode: 'bilingual',
      lastSyncedAt: '2026-04-20T00:00:00Z',
      recentJob: {
        id: 'job-1',
        videoId: 'abc123xyz00',
        youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
        targetLanguage: 'zh-CN',
        status: 'completed',
        stage: 'completed',
        progress: 100,
        errorMessage: '',
      },
    });

    const cache = await getCache();

    expect(chromeMock.storage.local.set).toHaveBeenCalledWith({
      'subtitle-cache': {
        abc123xyz00: {
          videoId: 'abc123xyz00',
          jobId: 'job-1',
          selectedMode: 'bilingual',
          lastSyncedAt: '2026-04-20T00:00:00Z',
          recentJob: {
            id: 'job-1',
            videoId: 'abc123xyz00',
            youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
            targetLanguage: 'zh-CN',
            status: 'completed',
            stage: 'completed',
            progress: 100,
            errorMessage: '',
          },
        },
      },
    });
    expect(cache.abc123xyz00.selectedMode).toBe('bilingual');
  });

  it('merges new cache data with an existing entry', async () => {
    const chromeMock = createChromeStorage({
      'subtitle-cache': {
        abc123xyz00: {
          videoId: 'abc123xyz00',
          jobId: 'job-1',
          selectedMode: 'translated',
          lastSyncedAt: '2026-04-20T00:00:00Z',
          subtitleUrls: {
            translated: 'http://localhost:8080/assets/translated.vtt',
            bilingual: 'http://localhost:8080/assets/bilingual.vtt',
          },
        },
      },
    });

    vi.stubGlobal('chrome', chromeMock);

    await saveCacheEntry({
      videoId: 'abc123xyz00',
      jobId: 'job-1',
      selectedMode: 'translated',
      lastSyncedAt: '2026-04-20T01:00:00Z',
      recentJob: {
        id: 'job-1',
        videoId: 'abc123xyz00',
        youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
        targetLanguage: 'zh-CN',
        status: 'running',
        stage: 'transcribing',
        progress: 55,
        errorMessage: '',
      },
    });

    const cache = await getCache();

    expect(cache.abc123xyz00.subtitleUrls).toEqual({
      translated: 'http://localhost:8080/assets/translated.vtt',
      bilingual: 'http://localhost:8080/assets/bilingual.vtt',
    });
    expect(cache.abc123xyz00.recentJob?.stage).toBe('transcribing');
    expect(cache.abc123xyz00.lastSyncedAt).toBe('2026-04-20T01:00:00Z');
  });

  it('updates selectedMode without dropping cached subtitle urls', async () => {
    const chromeMock = createChromeStorage({
      'subtitle-cache': {
        abc123xyz00: {
          videoId: 'abc123xyz00',
          jobId: 'job-1',
          selectedMode: 'translated',
          lastSyncedAt: '2026-04-20T00:00:00Z',
          subtitleUrls: {
            translated: 'http://localhost:8080/assets/translated.vtt',
            bilingual: 'http://localhost:8080/assets/bilingual.vtt',
          },
        },
      },
    });

    vi.stubGlobal('chrome', chromeMock);

    await setSelectedMode('abc123xyz00', 'bilingual');

    const cache = await getCache();

    expect(cache.abc123xyz00.selectedMode).toBe('bilingual');
    expect(cache.abc123xyz00.subtitleUrls).toEqual({
      translated: 'http://localhost:8080/assets/translated.vtt',
      bilingual: 'http://localhost:8080/assets/bilingual.vtt',
    });
  });

  it('preserves selectedMode when a stale cache refresh finishes after a mode update', async () => {
    let store = structuredClone({
      'subtitle-cache': {
        abc123xyz00: {
          videoId: 'abc123xyz00',
          jobId: 'job-1',
          selectedMode: 'translated',
          lastSyncedAt: '2026-04-20T00:00:00Z',
          subtitleUrls: {
            translated: 'http://localhost:8080/assets/translated.vtt',
            bilingual: 'http://localhost:8080/assets/bilingual.vtt',
          },
          recentJob: {
            id: 'job-1',
            videoId: 'abc123xyz00',
            youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
            targetLanguage: 'zh-CN',
            status: 'running',
            stage: 'transcribing',
            progress: 55,
            errorMessage: '',
          },
        },
      },
    });
    let firstSetValue: Record<string, unknown> | undefined;
    let releaseFirstSet = () => {};
    const firstSetReleased = new Promise<void>((resolve) => {
      releaseFirstSet = () => {
        store = {
          ...store,
          ...structuredClone(firstSetValue ?? {}),
        };
        resolve();
      };
    });
    let setCalls = 0;

    vi.stubGlobal('chrome', {
      storage: {
        local: {
          get: vi.fn(async (key: string) => {
            if (!(key in store)) {
              return {};
            }

            return { [key]: structuredClone(store[key]) };
          }),
          set: vi.fn(async (value: Record<string, unknown>) => {
            setCalls += 1;
            if (setCalls === 1) {
              firstSetValue = structuredClone(value);
              await firstSetReleased;
              return;
            }

            store = {
              ...store,
              ...structuredClone(value),
            };
          }),
        },
      },
    });

    const staleRefreshWrite = saveCacheEntry({
      videoId: 'abc123xyz00',
      jobId: 'job-1',
      selectedMode: 'translated',
      lastSyncedAt: '2026-04-20T01:00:00Z',
      subtitleUrls: {
        translated: 'http://localhost:8080/assets/translated.vtt',
        bilingual: 'http://localhost:8080/assets/bilingual.vtt',
      },
      recentJob: {
        id: 'job-1',
        videoId: 'abc123xyz00',
        youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
        targetLanguage: 'zh-CN',
        status: 'completed',
        stage: 'completed',
        progress: 100,
        errorMessage: '',
      },
    });

    await vi.waitFor(() => {
      expect(setCalls).toBe(1);
    });

    let modeUpdateResolved = false;
    const modeUpdate = setSelectedMode('abc123xyz00', 'bilingual').then(() => {
      modeUpdateResolved = true;
    });

    await Promise.resolve();
    expect(modeUpdateResolved).toBe(false);

    releaseFirstSet();
    await modeUpdate;
    await staleRefreshWrite;

    const cache = await getCache();

    expect(cache.abc123xyz00.selectedMode).toBe('bilingual');
    expect(cache.abc123xyz00.lastSyncedAt).toBe('2026-04-20T01:00:00Z');
    expect(cache.abc123xyz00.recentJob?.status).toBe('completed');
  });
});
