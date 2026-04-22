import { describe, expect, it, vi } from 'vitest';

describe('background mode rebroadcast', () => {
  it('rebroadcasts updated mode to youtube watch tabs', async () => {
    vi.stubGlobal('defineBackground', (main: () => void) => main);

    const entry = {
      videoId: 'abc123xyz00',
      jobId: 'job-1',
      selectedMode: 'bilingual' as const,
      lastSyncedAt: '2026-04-20T09:00:00.000Z',
      subtitleUrls: {
        translated: 'http://localhost:8080/assets/translated.vtt',
        bilingual: 'http://localhost:8080/assets/bilingual.vtt',
      },
    };

    const deps = {
      getCacheEntry: vi.fn(),
      setSelectedMode: vi.fn().mockResolvedValue(entry),
      tabs: {
        query: vi.fn().mockResolvedValue([{ id: 10 }, { id: 11 }]),
        sendMessage: vi.fn().mockResolvedValue(undefined),
      },
    };

    const { createBackgroundMessageHandler } = await import('../../entrypoints/background');
    const handler = createBackgroundMessageHandler(deps);
    const sendResponse = vi.fn();

    expect(
      handler(
        {
          type: 'subtitle-cache:set-mode',
          videoId: 'abc123xyz00',
          mode: 'bilingual',
        },
        {} as chrome.runtime.MessageSender,
        sendResponse,
      ),
    ).toBe(true);

    await vi.waitFor(() => {
      expect(deps.tabs.sendMessage).toHaveBeenCalledTimes(2);
    });

    expect(deps.tabs.sendMessage).toHaveBeenNthCalledWith(1, 10, {
      type: 'subtitle-mode-changed',
      payload: entry,
    });
    expect(deps.tabs.sendMessage).toHaveBeenNthCalledWith(2, 11, {
      type: 'subtitle-mode-changed',
      payload: entry,
    });
    expect(sendResponse).toHaveBeenCalledWith(entry);
  });

  it('rebroadcasts refreshed subtitle cache entries to youtube watch tabs', async () => {
    vi.stubGlobal('defineBackground', (main: () => void) => main);

    const entry = {
      videoId: 'abc123xyz00',
      jobId: 'job-1',
      selectedMode: 'translated' as const,
      lastSyncedAt: '2026-04-20T09:00:00.000Z',
      subtitleUrls: {
        translated: 'http://localhost:8080/assets/translated.vtt',
        bilingual: 'http://localhost:8080/assets/bilingual.vtt',
      },
      recentJob: {
        id: 'job-1',
        videoId: 'abc123xyz00',
        youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
        targetLanguage: 'zh-CN',
        status: 'completed' as const,
        stage: 'completed' as const,
        progress: 100,
        errorMessage: '',
      },
    };

    const deps = {
      getCacheEntry: vi.fn(),
      setSelectedMode: vi.fn(),
      tabs: {
        query: vi.fn().mockResolvedValue([{ id: 10 }, { id: 11 }]),
        sendMessage: vi.fn().mockResolvedValue(undefined),
      },
    };

    const { createBackgroundMessageHandler } = await import('../../entrypoints/background');
    const handler = createBackgroundMessageHandler(deps);
    const sendResponse = vi.fn();

    expect(
      handler(
        {
          type: 'subtitle-cache:sync-entry',
          payload: entry,
        } as never,
        {} as chrome.runtime.MessageSender,
        sendResponse,
      ),
    ).toBe(true);

    await vi.waitFor(() => {
      expect(deps.tabs.sendMessage).toHaveBeenCalledTimes(2);
    });

    expect(deps.tabs.sendMessage).toHaveBeenNthCalledWith(1, 10, {
      type: 'subtitle-mode-changed',
      payload: entry,
    });
    expect(deps.tabs.sendMessage).toHaveBeenNthCalledWith(2, 11, {
      type: 'subtitle-mode-changed',
      payload: entry,
    });
    expect(sendResponse).toHaveBeenCalledWith(entry);
  });
});
