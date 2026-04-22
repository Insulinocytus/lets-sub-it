import {
  CLEAR_EVENT,
  BRIDGE_READY_ATTRIBUTE,
  LOAD_EVENT,
  READY_EVENT,
} from '../page-bridge';
import { getSelectedMode, getSubtitleUrlForMode } from '../../src/lib/subtitleState';
import { extractVideoId } from '../../src/lib/youtube';
import type { LocalCacheEntry, SubtitleLoadPayload } from '../../src/types';

export { BRIDGE_READY_ATTRIBUTE, CLEAR_EVENT, LOAD_EVENT, READY_EVENT } from '../page-bridge';

type RuntimeMessaging = {
  sendMessage: typeof chrome.runtime.sendMessage;
  onMessage: Pick<typeof chrome.runtime.onMessage, 'addListener' | 'removeListener'>;
};

type ContentRuntime = RuntimeMessaging & Pick<typeof chrome.runtime, 'getURL'>;

export function injectPageBridge(
  doc: Document = document,
  runtime: Pick<typeof chrome.runtime, 'getURL'> = chrome.runtime,
) {
  const existing = doc.querySelector('script[data-lets-sub-it-bridge]');
  if (existing) {
    return existing as HTMLScriptElement;
  }

  const script = doc.createElement('script');
  script.dataset.letsSubItBridge = 'true';
  script.src = runtime.getURL('/page-bridge.js');
  script.onload = () => {
    script.remove();
  };
  (doc.head ?? doc.documentElement).append(script);
  return script;
}

export function buildSubtitleLoadPayload(
  entry: LocalCacheEntry,
  videoId: string,
): SubtitleLoadPayload | undefined {
  const mode = getSelectedMode(entry);
  const subtitleUrl = getSubtitleUrlForMode(entry, mode);
  if (!subtitleUrl) {
    return undefined;
  }

  return {
    videoId,
    mode,
    subtitleUrl,
    targetLanguage: entry.recentJob?.targetLanguage ?? 'und',
  };
}

export function createBridgeMessenger(
  win: Window = window,
  doc: Document = document,
) {
  let pendingEvent:
    | { type: 'clear' }
    | { type: 'load'; payload: SubtitleLoadPayload }
    | undefined;
  let bridgeReady = doc.documentElement.getAttribute(BRIDGE_READY_ATTRIBUTE) === 'true';

  function postPayload(payload: SubtitleLoadPayload) {
    win.postMessage({ type: LOAD_EVENT, payload }, win.location.origin);
  }

  function postClear() {
    win.postMessage({ type: CLEAR_EVENT }, win.location.origin);
  }

  function flushPending() {
    if (!bridgeReady || !pendingEvent) {
      return;
    }

    if (pendingEvent.type === 'clear') {
      postClear();
    } else {
      postPayload(pendingEvent.payload);
    }

    pendingEvent = undefined;
  }

  function markReady() {
    bridgeReady = true;
    doc.documentElement.setAttribute(BRIDGE_READY_ATTRIBUTE, 'true');
    flushPending();
  }

  function queuePayload(payload: SubtitleLoadPayload) {
    pendingEvent = { type: 'load', payload };
    flushPending();
  }

  function clear() {
    pendingEvent = { type: 'clear' };
    flushPending();
  }

  function handleWindowMessage(event: MessageEvent) {
    if (event.source !== win || event.data?.type !== READY_EVENT) {
      return;
    }

    markReady();
  }

  win.addEventListener('message', handleWindowMessage);

  return {
    clear,
    markReady,
    queuePayload,
    dispose() {
      win.removeEventListener('message', handleWindowMessage);
    },
  };
}

export function loadSubtitle(
  entry: LocalCacheEntry,
  videoId: string,
  bridgeMessenger: Pick<ReturnType<typeof createBridgeMessenger>, 'queuePayload'>,
) {
  const payload = buildSubtitleLoadPayload(entry, videoId);
  if (!payload) {
    return;
  }

  bridgeMessenger.queuePayload(payload);
}

export async function loadCachedSubtitle(
  videoId: string,
  runtime: Pick<RuntimeMessaging, 'sendMessage'>,
  bridgeMessenger?: Pick<ReturnType<typeof createBridgeMessenger>, 'queuePayload'>,
): Promise<LocalCacheEntry | undefined> {
  const entry = (await runtime.sendMessage({
    type: 'subtitle-cache:get',
    videoId,
  })) as LocalCacheEntry | undefined;

  if (entry && bridgeMessenger) {
    loadSubtitle(entry, videoId, bridgeMessenger);
  }

  return entry;
}

export function addModeChangeListener(
  videoId: string,
  runtime: Pick<RuntimeMessaging, 'onMessage'>,
  bridgeMessenger: Pick<ReturnType<typeof createBridgeMessenger>, 'queuePayload'>,
) {
  const listener = (message: unknown) => {
    const typedMessage = message as { type?: string; payload?: LocalCacheEntry };
    if (typedMessage.type !== 'subtitle-mode-changed') {
      return;
    }

    const entry = typedMessage.payload;
    if (!entry || entry.videoId !== videoId) {
      return;
    }

    loadSubtitle(entry, videoId, bridgeMessenger);
  };

  runtime.onMessage.addListener(listener);
  return () => {
    runtime.onMessage.removeListener(listener);
  };
}

export function observeVideoPageNavigation(
  onNavigate: () => void | Promise<void>,
  win: Window = window,
  navigationPollIntervalMs = 500,
) {
  let lastHref = win.location.href;

  const checkForNavigation = () => {
    const nextHref = win.location.href;
    if (nextHref === lastHref) {
      return;
    }

    lastHref = nextHref;
    void onNavigate();
  };

  win.addEventListener('yt-navigate-finish', checkForNavigation);
  win.addEventListener('popstate', checkForNavigation);

  const intervalId = win.setInterval(checkForNavigation, navigationPollIntervalMs);

  return () => {
    win.removeEventListener('yt-navigate-finish', checkForNavigation);
    win.removeEventListener('popstate', checkForNavigation);
    win.clearInterval(intervalId);
  };
}

export function initializeContentScript(
  win: Window = window,
  doc: Document = document,
  runtime: ContentRuntime = chrome.runtime,
  navigationPollIntervalMs = 500,
) {
  injectPageBridge(doc, runtime);
  const bridgeMessenger = createBridgeMessenger(win, doc);
  let currentVideoId = '';
  let removeModeChangeListener: (() => void) | undefined;
  let navigationVersion = 0;

  const syncToCurrentVideo = async () => {
    const nextVideoId = extractVideoId(win.location.href);
    if (nextVideoId === currentVideoId) {
      return;
    }

    if (currentVideoId) {
      bridgeMessenger.clear();
    }

    currentVideoId = nextVideoId;
    removeModeChangeListener?.();
    removeModeChangeListener = undefined;
    navigationVersion += 1;
    const currentNavigationVersion = navigationVersion;

    if (!nextVideoId) {
      return;
    }

    removeModeChangeListener = addModeChangeListener(nextVideoId, runtime, bridgeMessenger);
    const entry = await loadCachedSubtitle(nextVideoId, runtime);
    if (currentNavigationVersion !== navigationVersion || currentVideoId !== nextVideoId) {
      return;
    }
    if (entry) {
      loadSubtitle(entry, nextVideoId, bridgeMessenger);
    }
  };

  void syncToCurrentVideo();
  const stopObservingNavigation = observeVideoPageNavigation(
    syncToCurrentVideo,
    win,
    navigationPollIntervalMs,
  );

  return {
    dispose() {
      stopObservingNavigation();
      removeModeChangeListener?.();
      bridgeMessenger.dispose();
    },
  };
}

export default defineContentScript({
  matches: ['https://www.youtube.com/watch*'],
  runAt: 'document_idle',
  main() {
    initializeContentScript();
  },
});
