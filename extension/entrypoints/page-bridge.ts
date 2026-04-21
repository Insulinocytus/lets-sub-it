import type { SubtitleLoadPayload } from '../src/types';

export const LOAD_EVENT = 'LETS_SUB_IT_LOAD';
export const READY_EVENT = 'LETS_SUB_IT_BRIDGE_READY';
export const TRACK_ID = 'lets-sub-it-track';
export const BRIDGE_READY_ATTRIBUTE = 'data-lets-sub-it-bridge-ready';

export function mountSubtitleTrack(payload: SubtitleLoadPayload, root: ParentNode = document) {
  const video = root.querySelector('video');
  if (!video) {
    return false;
  }

  const existing =
    root instanceof Document ? root.getElementById(TRACK_ID) : root.querySelector(`#${TRACK_ID}`);
  existing?.remove();

  const ownerDocument = root instanceof Document ? root : root.ownerDocument ?? document;
  const track = ownerDocument.createElement('track');
  track.id = TRACK_ID;
  track.kind = 'subtitles';
  track.label = payload.mode;
  track.src = payload.subtitleUrl;
  track.default = true;
  video.append(track);
  return true;
}

export function initializePageBridge(
  win: Window = window,
  doc: Document = document,
  mountRetryDelayMs = 100,
) {
  let pendingPayload: SubtitleLoadPayload | undefined;
  let retryTimeoutId: number | undefined;

  const flushPendingPayload = () => {
    if (!pendingPayload) {
      return;
    }

    if (mountSubtitleTrack(pendingPayload, doc)) {
      pendingPayload = undefined;
      if (retryTimeoutId !== undefined) {
        win.clearTimeout(retryTimeoutId);
        retryTimeoutId = undefined;
      }
      return;
    }

    if (retryTimeoutId !== undefined) {
      return;
    }

    retryTimeoutId = win.setTimeout(() => {
      retryTimeoutId = undefined;
      flushPendingPayload();
    }, mountRetryDelayMs);
  };

  doc.documentElement.setAttribute(BRIDGE_READY_ATTRIBUTE, 'true');
  win.postMessage({ type: READY_EVENT }, win.location.origin);

  win.addEventListener('message', (event) => {
    if (event.source !== win || event.data?.type !== LOAD_EVENT) {
      return;
    }

    pendingPayload = event.data.payload as SubtitleLoadPayload;
    flushPendingPayload();
  });
}

export default defineUnlistedScript(() => {
  initializePageBridge();
});
