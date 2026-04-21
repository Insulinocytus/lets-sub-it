// @vitest-environment jsdom

import { afterEach, describe, expect, it, vi } from 'vitest';

describe('page bridge', () => {
  afterEach(() => {
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
});
