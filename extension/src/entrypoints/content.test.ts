// @vitest-environment jsdom

import { afterEach, describe, expect, it, vi } from 'vitest';

function createCacheEntry() {
  return {
    videoId: 'abc123xyz00',
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
});
