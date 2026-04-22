// @vitest-environment jsdom

import { afterEach, describe, expect, it, vi } from 'vitest';

function createCacheEntry(videoId = 'abc123xyz00') {
  return {
    videoId,
    jobId: 'job-1',
    selectedMode: 'translated' as const,
    lastSyncedAt: '2026-04-20T09:00:00.000Z',
    subtitleUrls: {
      translated: 'http://localhost:8080/assets/translated.vtt',
      bilingual: 'http://localhost:8080/assets/bilingual.vtt',
    },
  };
}

describe('content subtitle loading', () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
    document.documentElement.removeAttribute('data-lets-sub-it-bridge-ready');
    document.head.innerHTML = '';
    document.body.innerHTML = '';
  });

  it('delivers cached subtitle load after the bridge announces ready', async () => {
    vi.stubGlobal('defineContentScript', (definition: unknown) => definition);
    vi.stubGlobal('defineUnlistedScript', (main: () => void) => main);

    const postMessage = vi.spyOn(window, 'postMessage');
    const runtime = {
      sendMessage: vi.fn().mockResolvedValue(createCacheEntry()),
    };

    const { createBridgeMessenger, loadCachedSubtitle, READY_EVENT } = await import(
      '../../entrypoints/content/index'
    );

    const bridgeMessenger = createBridgeMessenger(window, document);
    await loadCachedSubtitle('abc123xyz00', runtime, bridgeMessenger);

    expect(postMessage).not.toHaveBeenCalled();

    window.dispatchEvent(
      new MessageEvent('message', {
        source: window,
        data: { type: READY_EVENT },
      }),
    );

    expect(postMessage).toHaveBeenCalledWith(
      {
        type: 'LETS_SUB_IT_LOAD',
        payload: {
          videoId: 'abc123xyz00',
          mode: 'translated',
          subtitleUrl: 'http://localhost:8080/assets/translated.vtt',
        },
      },
      window.location.origin,
    );
  });

  it('autoloads immediately when the bridge is already ready', async () => {
    vi.stubGlobal('defineContentScript', (definition: unknown) => definition);
    vi.stubGlobal('defineUnlistedScript', (main: () => void) => main);
    document.documentElement.setAttribute('data-lets-sub-it-bridge-ready', 'true');

    const postMessage = vi.spyOn(window, 'postMessage');
    const runtime = {
      sendMessage: vi.fn().mockResolvedValue(createCacheEntry()),
    };

    const { createBridgeMessenger, loadCachedSubtitle } = await import(
      '../../entrypoints/content/index'
    );

    const bridgeMessenger = createBridgeMessenger(window, document);
    await loadCachedSubtitle('abc123xyz00', runtime, bridgeMessenger);

    expect(postMessage).toHaveBeenCalledWith(
      {
        type: 'LETS_SUB_IT_LOAD',
        payload: {
          videoId: 'abc123xyz00',
          mode: 'translated',
          subtitleUrl: 'http://localhost:8080/assets/translated.vtt',
        },
      },
      window.location.origin,
    );
  });

  it('reloads subtitle cache and mode listeners after YouTube SPA navigation', async () => {
    vi.useFakeTimers();
    vi.stubGlobal('defineContentScript', (definition: unknown) => definition);
    vi.stubGlobal('defineUnlistedScript', (main: () => void) => main);

    const listeners = new Map<string, EventListener[]>();
    const fakeWindow = {
      location: {
        href: 'https://www.youtube.com/watch?v=abc123xyz00',
        origin: 'https://www.youtube.com',
      },
      postMessage: vi.fn(),
      addEventListener: vi.fn((type: string, listener: EventListener) => {
        listeners.set(type, [...(listeners.get(type) ?? []), listener]);
      }),
      removeEventListener: vi.fn((type: string, listener: EventListener) => {
        listeners.set(
          type,
          (listeners.get(type) ?? []).filter((registeredListener) => registeredListener !== listener),
        );
      }),
      setInterval: window.setInterval.bind(window),
      clearInterval: window.clearInterval.bind(window),
    } as unknown as Window;

    const runtime = {
      getURL: vi.fn().mockReturnValue('chrome-extension://test/page-bridge.js'),
      sendMessage: vi
        .fn()
        .mockResolvedValueOnce(createCacheEntry('abc123xyz00'))
        .mockResolvedValueOnce(createCacheEntry('new456video9')),
      onMessage: {
        addListener: vi.fn(),
        removeListener: vi.fn(),
      },
    };

    const { initializeContentScript } = await import('../../entrypoints/content/index');

    const controller = initializeContentScript(fakeWindow, document, runtime, 25);

    await vi.waitFor(() => {
      expect(runtime.sendMessage).toHaveBeenCalledWith({
        type: 'subtitle-cache:get',
        videoId: 'abc123xyz00',
      });
      expect(runtime.onMessage.addListener).toHaveBeenCalledTimes(1);
    });

    fakeWindow.location.href = 'https://www.youtube.com/watch?v=new456video9';
    vi.advanceTimersByTime(25);

    await vi.waitFor(() => {
      expect(runtime.sendMessage).toHaveBeenCalledWith({
        type: 'subtitle-cache:get',
        videoId: 'new456video9',
      });
      expect(runtime.onMessage.removeListener).toHaveBeenCalledTimes(1);
      expect(runtime.onMessage.addListener).toHaveBeenCalledTimes(2);
    });

    controller.dispose();
  });

  it('clears stale subtitle tracks after navigating to an uncached video', async () => {
    vi.useFakeTimers();
    vi.stubGlobal('defineContentScript', (definition: unknown) => definition);
    vi.stubGlobal('defineUnlistedScript', (main: () => void) => main);
    document.documentElement.setAttribute('data-lets-sub-it-bridge-ready', 'true');

    const fakeWindow = {
      location: {
        href: 'https://www.youtube.com/watch?v=abc123xyz00',
        origin: 'https://www.youtube.com',
      },
      postMessage: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      setInterval: window.setInterval.bind(window),
      clearInterval: window.clearInterval.bind(window),
    } as unknown as Window;

    const runtime = {
      getURL: vi.fn().mockReturnValue('chrome-extension://test/page-bridge.js'),
      sendMessage: vi
        .fn()
        .mockResolvedValueOnce(createCacheEntry('abc123xyz00'))
        .mockResolvedValueOnce(undefined),
      onMessage: {
        addListener: vi.fn(),
        removeListener: vi.fn(),
      },
    };

    const { CLEAR_EVENT, LOAD_EVENT, initializeContentScript } = await import(
      '../../entrypoints/content/index'
    );

    const controller = initializeContentScript(fakeWindow, document, runtime, 25);

    await vi.waitFor(() => {
      expect(fakeWindow.postMessage).toHaveBeenCalledWith(
        {
          type: LOAD_EVENT,
          payload: {
            videoId: 'abc123xyz00',
            mode: 'translated',
            subtitleUrl: 'http://localhost:8080/assets/translated.vtt',
          },
        },
        'https://www.youtube.com',
      );
    });

    fakeWindow.location.href = 'https://www.youtube.com/watch?v=new456video9';
    vi.advanceTimersByTime(25);

    await vi.waitFor(() => {
      expect(fakeWindow.postMessage).toHaveBeenCalledWith(
        { type: CLEAR_EVENT },
        'https://www.youtube.com',
      );
    });

    controller.dispose();
  });
});
