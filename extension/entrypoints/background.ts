import { getLatestCacheEntryForVideo, setSelectedMode } from '../src/lib/storage';
import type { LocalCacheEntry, SubtitleMode } from '../src/types';

type BackgroundMessage =
  | { type?: 'subtitle-cache:get'; videoId?: string }
  | { type?: 'subtitle-cache:set-mode'; jobId?: string; mode?: SubtitleMode }
  | { type?: 'subtitle-cache:sync-entry'; payload?: LocalCacheEntry };

async function rebroadcastEntry(
  entry: LocalCacheEntry,
  tabs: Pick<typeof chrome.tabs, 'query' | 'sendMessage'>,
) {
  const youtubeTabs = await tabs.query({
    url: ['https://www.youtube.com/watch*'],
  });

  await Promise.all(
    youtubeTabs
      .filter((tab) => typeof tab.id === 'number')
      .map((tab) =>
        tabs
          .sendMessage(tab.id as number, {
            type: 'subtitle-mode-changed',
            payload: entry,
          })
          .catch(() => undefined),
      ),
  );
}

export function createBackgroundMessageHandler(
  deps: {
    getLatestCacheEntryForVideo: typeof getLatestCacheEntryForVideo;
    setSelectedMode: typeof setSelectedMode;
    tabs: Pick<typeof chrome.tabs, 'query' | 'sendMessage'>;
  } = {
    getLatestCacheEntryForVideo,
    setSelectedMode,
    tabs: chrome.tabs,
  },
) {
  return (
    message: BackgroundMessage,
    _sender: chrome.runtime.MessageSender,
    sendResponse: (response?: unknown) => void,
  ) => {
    if (message?.type === 'subtitle-cache:get' && message.videoId) {
      void deps.getLatestCacheEntryForVideo(message.videoId).then(sendResponse);
      return true;
    }

    if (message?.type === 'subtitle-cache:set-mode' && message.jobId && message.mode) {
      void deps.setSelectedMode(message.jobId, message.mode).then(async (entry) => {
        if (entry) {
          await rebroadcastEntry(entry, deps.tabs);
        }

        sendResponse(entry);
      });

      return true;
    }

    if (message?.type === 'subtitle-cache:sync-entry' && message.payload) {
      void rebroadcastEntry(message.payload, deps.tabs).then(() => {
        sendResponse(message.payload);
      });

      return true;
    }

    return false;
  };
}

export default defineBackground(() => {
  chrome.runtime.onMessage.addListener(createBackgroundMessageHandler());
});
