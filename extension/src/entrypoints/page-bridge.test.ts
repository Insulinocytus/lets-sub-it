// @vitest-environment jsdom

import { afterEach, describe, expect, it, vi } from 'vitest';

describe('page bridge', () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
    document.body.innerHTML = '';
    document.documentElement.removeAttribute('data-lets-sub-it-bridge-ready');
  });

  it('announces readiness when initialized', async () => {
    vi.stubGlobal('defineUnlistedScript', (main: () => void) => main);

    const postMessage = vi.spyOn(window, 'postMessage');
    const { initializePageBridge, READY_EVENT, BRIDGE_READY_ATTRIBUTE } = await import(
      '../../entrypoints/page-bridge'
    );

    initializePageBridge(window, document);

    expect(document.documentElement.getAttribute(BRIDGE_READY_ATTRIBUTE)).toBe('true');
    expect(postMessage).toHaveBeenCalledWith({ type: READY_EVENT }, window.location.origin);
  });

  it('retries pending subtitle mounts until the video element exists', async () => {
    vi.useFakeTimers();
    vi.stubGlobal('defineUnlistedScript', (main: () => void) => main);

    const { initializePageBridge, LOAD_EVENT, TRACK_ID } = await import('../../entrypoints/page-bridge');

    initializePageBridge(window, document, 25);

    window.dispatchEvent(
      new MessageEvent('message', {
        source: window,
        data: {
          type: LOAD_EVENT,
          payload: {
            videoId: 'abc123xyz00',
            mode: 'translated',
            subtitleUrl: 'http://localhost:8080/assets/translated.vtt',
          },
        },
      }),
    );

    expect(document.getElementById(TRACK_ID)).toBeNull();

    document.body.innerHTML = '<video></video>';
    vi.advanceTimersByTime(25);

    const track = document.getElementById(TRACK_ID) as HTMLTrackElement | null;
    expect(track?.src).toBe('http://localhost:8080/assets/translated.vtt');
  });

  it('removes the existing subtitle track when asked to clear it', async () => {
    vi.stubGlobal('defineUnlistedScript', (main: () => void) => main);

    const { initializePageBridge, CLEAR_EVENT, TRACK_ID } = await import('../../entrypoints/page-bridge');

    document.body.innerHTML = '<video><track id="lets-sub-it-track" kind="subtitles"></video>';
    initializePageBridge(window, document);

    expect(document.getElementById(TRACK_ID)).not.toBeNull();

    window.dispatchEvent(
      new MessageEvent('message', {
        source: window,
        data: { type: CLEAR_EVENT },
      }),
    );

    expect(document.getElementById(TRACK_ID)).toBeNull();
  });
});
