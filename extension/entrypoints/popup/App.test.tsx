// @vitest-environment jsdom

import { act } from 'react';
import ReactDOM from 'react-dom/client';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import App from './App';
import { createJob, getJobs, getSubtitleAsset } from '../../src/lib/api';

vi.mock('../../src/lib/api', () => ({
  createJob: vi.fn(),
  getJobs: vi.fn(),
  getSubtitleAsset: vi.fn(),
}));

function createChromeStorage(initialCache: Record<string, unknown> = {}) {
  let store = { ...initialCache };

  return {
    runtime: {
      sendMessage: vi.fn().mockResolvedValue(undefined),
    },
    storage: {
      local: {
        get: vi.fn(async (key: string) => ({ [key]: store[key] })),
        set: vi.fn(async (value: Record<string, unknown>) => {
          store = { ...store, ...value };
        }),
      },
    },
  };
}

function setInputValue(input: HTMLInputElement, value: string) {
  const descriptor = Object.getOwnPropertyDescriptor(
    HTMLInputElement.prototype,
    'value',
  );
  descriptor?.set?.call(input, value);
  input.dispatchEvent(new Event('input', { bubbles: true }));
}

describe('popup App cache integration', () => {
  let container: HTMLDivElement;
  let root: ReactDOM.Root;

  beforeEach(() => {
    vi.stubGlobal('IS_REACT_ACT_ENVIRONMENT', true);
    container = document.createElement('div');
    document.body.appendChild(container);
    root = ReactDOM.createRoot(container);
  });

  afterEach(async () => {
    await act(async () => {
      root.unmount();
    });
    container.remove();
    vi.useRealTimers();
    vi.clearAllMocks();
    vi.unstubAllGlobals();
  });

  it('polls active jobs and refreshes the popup list', async () => {
    vi.useFakeTimers();
    const chromeMock = createChromeStorage({
      'subtitle-cache': {
        abc123xyz00: {
          videoId: 'abc123xyz00',
          jobId: 'job-1',
          selectedMode: 'translated',
          lastSyncedAt: '2026-04-20T09:00:00.000Z',
          recentJob: {
            id: 'job-1',
            videoId: 'abc123xyz00',
            youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
            targetLanguage: 'zh-CN',
            status: 'running',
            stage: 'transcribing',
            progress: 25,
            errorMessage: '',
          },
        },
      },
    });
    vi.stubGlobal('chrome', chromeMock);
    vi.mocked(getJobs).mockResolvedValue([
      {
        id: 'job-1',
        videoId: 'abc123xyz00',
        youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
        targetLanguage: 'zh-CN',
        status: 'completed',
        stage: 'completed',
        progress: 100,
        errorMessage: '',
      },
    ]);
    vi.mocked(getSubtitleAsset).mockResolvedValue({
      jobId: 'job-1',
      videoId: 'abc123xyz00',
      sourceLanguage: 'en',
      targetLanguage: 'zh-CN',
      subtitleUrls: {
        translated: 'http://localhost:8080/assets/abc123xyz00/translated.vtt',
        bilingual: 'http://localhost:8080/assets/abc123xyz00/bilingual.vtt',
      },
    });

    await act(async () => {
      root.render(<App />);
    });

    await act(async () => {
      vi.advanceTimersByTime(2000);
      await Promise.resolve();
    });

    expect(getJobs).toHaveBeenCalledWith(['job-1']);
    expect(getSubtitleAsset).toHaveBeenCalledWith('abc123xyz00');
    expect(chromeMock.runtime.sendMessage).toHaveBeenCalledWith({
      type: 'subtitle-cache:sync-entry',
      payload: {
        videoId: 'abc123xyz00',
        jobId: 'job-1',
        selectedMode: 'translated',
        lastSyncedAt: expect.any(String),
        subtitleUrls: {
          translated: 'http://localhost:8080/assets/abc123xyz00/translated.vtt',
          bilingual: 'http://localhost:8080/assets/abc123xyz00/bilingual.vtt',
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
      },
    });
    expect(container.textContent).toContain('completed');
    expect(container.textContent).toContain('100%');

    vi.useRealTimers();
  });

  it('shows an asset lookup error and keeps the completed job visible', async () => {
    vi.useFakeTimers();
    vi.stubGlobal(
      'chrome',
      createChromeStorage({
        'subtitle-cache': {
          abc123xyz00: {
            videoId: 'abc123xyz00',
            jobId: 'job-1',
            selectedMode: 'translated',
            lastSyncedAt: '2026-04-20T09:00:00.000Z',
            recentJob: {
              id: 'job-1',
              videoId: 'abc123xyz00',
              youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
              targetLanguage: 'zh-CN',
              status: 'running',
              stage: 'transcribing',
              progress: 25,
              errorMessage: '',
            },
          },
        },
      }),
    );
    vi.mocked(getJobs).mockResolvedValue([
      {
        id: 'job-1',
        videoId: 'abc123xyz00',
        youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
        targetLanguage: 'zh-CN',
        status: 'completed',
        stage: 'completed',
        progress: 100,
        errorMessage: '',
      },
    ]);
    vi.mocked(getSubtitleAsset).mockRejectedValue(
      new Error('Failed to load subtitle asset: not found'),
    );

    await act(async () => {
      root.render(<App />);
    });

    await act(async () => {
      vi.advanceTimersByTime(2000);
      await Promise.resolve();
    });

    expect(container.textContent).toContain('Failed to load subtitle asset: not found');
    expect(container.textContent).toContain('completed');
    expect(container.textContent).toContain('100%');

    vi.useRealTimers();
  });

  it('loads cached recent jobs on startup', async () => {
    vi.stubGlobal(
      'chrome',
      createChromeStorage({
        'subtitle-cache': {
          abc123xyz00: {
            videoId: 'abc123xyz00',
            jobId: 'job-1',
            selectedMode: 'translated',
            lastSyncedAt: '2026-04-20T09:00:00.000Z',
            recentJob: {
              id: 'job-1',
              videoId: 'abc123xyz00',
              youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
              targetLanguage: 'zh-CN',
              status: 'queued',
              stage: 'queued',
              progress: 0,
              errorMessage: '',
            },
          },
        },
      }),
    );

    await act(async () => {
      root.render(<App />);
    });

    expect(container.textContent).toContain('abc123xyz00');
    expect(container.textContent).toContain('queued');
    expect(container.textContent).not.toContain('No jobs yet.');
  });

  it('shows a polling error instead of leaving an unhandled rejection', async () => {
    vi.useFakeTimers();
    vi.stubGlobal(
      'chrome',
      createChromeStorage({
        'subtitle-cache': {
          abc123xyz00: {
            videoId: 'abc123xyz00',
            jobId: 'job-1',
            selectedMode: 'translated',
            lastSyncedAt: '2026-04-20T09:00:00.000Z',
            recentJob: {
              id: 'job-1',
              videoId: 'abc123xyz00',
              youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
              targetLanguage: 'zh-CN',
              status: 'running',
              stage: 'transcribing',
              progress: 25,
              errorMessage: '',
            },
          },
        },
      }),
    );
    vi.mocked(getJobs).mockRejectedValue(new Error('Failed to load job: Network down'));

    await act(async () => {
      root.render(<App />);
    });

    await act(async () => {
      vi.advanceTimersByTime(2000);
      await Promise.resolve();
    });

    expect(container.textContent).toContain('Failed to load job: Network down');

    vi.useRealTimers();
  });

  it('stores new jobs in cache after successful submission', async () => {
    const chromeMock = createChromeStorage({ 'subtitle-cache': {} });
    vi.stubGlobal('chrome', chromeMock);
    vi.mocked(createJob).mockResolvedValue({
      id: 'job-2',
      videoId: 'def456uvw99',
      youtubeUrl: 'https://www.youtube.com/watch?v=def456uvw99',
      targetLanguage: 'zh-CN',
      status: 'queued',
      stage: 'queued',
      progress: 0,
      errorMessage: '',
    });

    await act(async () => {
      root.render(<App />);
    });

    const [urlInput, languageInput] = Array.from(container.querySelectorAll('input'));
    const form = container.querySelector('form');

    if (!urlInput || !languageInput || !form) {
      throw new Error('Popup form elements not found');
    }

    await act(async () => {
      setInputValue(urlInput, 'https://www.youtube.com/watch?v=def456uvw99');
      setInputValue(languageInput, 'zh-CN');
    });

    await act(async () => {
      form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    });

    expect(createJob).toHaveBeenCalledWith({
      youtubeUrl: 'https://www.youtube.com/watch?v=def456uvw99',
      targetLanguage: 'zh-CN',
      videoId: 'def456uvw99',
    });
    expect(chromeMock.storage.local.set).toHaveBeenCalledTimes(1);
    expect(chromeMock.storage.local.set).toHaveBeenCalledWith({
      'subtitle-cache': {
        def456uvw99: {
          videoId: 'def456uvw99',
          jobId: 'job-2',
          selectedMode: 'translated',
          lastSyncedAt: expect.any(String),
          recentJob: {
            id: 'job-2',
            videoId: 'def456uvw99',
            youtubeUrl: 'https://www.youtube.com/watch?v=def456uvw99',
            targetLanguage: 'zh-CN',
            status: 'queued',
            stage: 'queued',
            progress: 0,
            errorMessage: '',
          },
        },
      },
    });
    expect(container.textContent).toContain('def456uvw99');
  });

  it('persists mode changes for cached videos', async () => {
    const chromeMock = createChromeStorage({
      'subtitle-cache': {
        abc123xyz00: {
          videoId: 'abc123xyz00',
          jobId: 'job-1',
          selectedMode: 'translated',
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
            status: 'completed',
            stage: 'completed',
            progress: 100,
            errorMessage: '',
          },
        },
      },
    });
    vi.stubGlobal('chrome', chromeMock);

    await act(async () => {
      root.render(<App />);
    });

    const bilingualButton = Array.from(container.querySelectorAll('button')).find(
      (button) => button.textContent === 'Bilingual',
    );

    if (!bilingualButton) {
      throw new Error('Mode toggle button not found');
    }

    await act(async () => {
      bilingualButton.dispatchEvent(new MouseEvent('click', { bubbles: true }));
    });

    expect(chromeMock.storage.local.set).toHaveBeenCalledWith({
      'subtitle-cache': {
        abc123xyz00: {
          videoId: 'abc123xyz00',
          jobId: 'job-1',
          selectedMode: 'bilingual',
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
            status: 'completed',
            stage: 'completed',
            progress: 100,
            errorMessage: '',
          },
        },
      },
    });
  });
});
