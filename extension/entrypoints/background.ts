import { getCacheEntry, setSelectedMode } from '../src/lib/storage';

export function createBackgroundMessageHandler(
  deps: {
    getCacheEntry: typeof getCacheEntry;
    setSelectedMode: typeof setSelectedMode;
    tabs: Pick<typeof chrome.tabs, 'query' | 'sendMessage'>;
  } = {
    getCacheEntry,
    setSelectedMode,
    tabs: chrome.tabs,
  },
) {
  return (
    message: { type?: string; videoId?: string; mode?: 'translated' | 'bilingual' },
    _sender: chrome.runtime.MessageSender,
    sendResponse: (response?: unknown) => void,
  ) => {
    if (message?.type === 'subtitle-cache:get' && message.videoId) {
      void deps.getCacheEntry(message.videoId).then(sendResponse);
      return true;
    }

    if (message?.type === 'subtitle-cache:set-mode' && message.videoId && message.mode) {
      void deps.setSelectedMode(message.videoId, message.mode).then(async (entry) => {
        if (entry) {
          const tabs = await deps.tabs.query({
            url: ['https://www.youtube.com/watch*'],
          });

          await Promise.all(
            tabs
              .filter((tab) => typeof tab.id === 'number')
              .map((tab) =>
                deps.tabs
                  .sendMessage(tab.id as number, {
                    type: 'subtitle-mode-changed',
                    payload: entry,
                  })
                  .catch(() => undefined),
              ),
          );
        }

        sendResponse(entry);
      });

      return true;
    }

    return false;
  };
}

export default defineBackground(() => {
  chrome.runtime.onMessage.addListener(createBackgroundMessageHandler());
});
