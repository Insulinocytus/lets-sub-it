import { afterEach, describe, expect, it, vi } from 'vitest';

import { createJob, getJob, getJobs, getSubtitleAsset } from './api';

function createJsonResponse(body: unknown, ok = true, status = 200) {
  return {
    ok,
    status,
    json: vi.fn().mockResolvedValue(body),
  };
}

describe('api client', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('posts createJob requests to the backend', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      createJsonResponse({
        id: 'job-1',
        videoId: 'abc123xyz00',
        youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
        targetLanguage: 'zh-CN',
        status: 'queued',
        stage: 'queued',
        progress: 0,
        errorMessage: '',
      }),
    );

    vi.stubGlobal('fetch', fetchMock);

    const job = await createJob({
      youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
      targetLanguage: 'zh-CN',
      videoId: 'abc123xyz00',
    });

    expect(fetchMock).toHaveBeenCalledWith('http://localhost:8080/api/jobs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
        targetLanguage: 'zh-CN',
        videoId: 'abc123xyz00',
      }),
    });
    expect(job.id).toBe('job-1');
  });

  it('throws the backend error for non-2xx createJob responses', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      createJsonResponse(
        {
          error: 'video already queued',
        },
        false,
        409,
      ),
    );

    vi.stubGlobal('fetch', fetchMock);

    await expect(
      createJob({
        youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
        targetLanguage: 'zh-CN',
        videoId: 'abc123xyz00',
      }),
    ).rejects.toThrow('Failed to create job: video already queued');
  });

  it('throws a stable message when createJob hits a network error', async () => {
    const fetchMock = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'));
    vi.stubGlobal('fetch', fetchMock);

    await expect(
      createJob({
        youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
        targetLanguage: 'zh-CN',
        videoId: 'abc123xyz00',
      }),
    ).rejects.toThrow('Failed to create job: Failed to fetch');
  });

  it('fetches jobs by id', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      createJsonResponse({
        id: 'job-1',
        videoId: 'abc123xyz00',
        youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
        targetLanguage: 'zh-CN',
        status: 'queued',
        stage: 'queued',
        progress: 0,
        errorMessage: '',
      }),
    );

    vi.stubGlobal('fetch', fetchMock);

    const job = await getJob('job-1');

    expect(fetchMock).toHaveBeenCalledWith('http://localhost:8080/api/jobs/job-1');
    expect(job.videoId).toBe('abc123xyz00');
  });

  it('throws the backend error for non-2xx getJob responses', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      createJsonResponse(
        {
          error: 'job not found',
        },
        false,
        404,
      ),
    );

    vi.stubGlobal('fetch', fetchMock);

    await expect(getJob('missing-job')).rejects.toThrow(
      'Failed to load job: job not found',
    );
  });

  it('throws a stable message when getJob hits a network error', async () => {
    const fetchMock = vi.fn().mockRejectedValue(new TypeError('Network down'));
    vi.stubGlobal('fetch', fetchMock);

    await expect(getJob('job-1')).rejects.toThrow('Failed to load job: Network down');
  });

  it('polls multiple jobs and returns them in order', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        createJsonResponse({
          id: 'job-1',
          videoId: 'abc123xyz00',
          youtubeUrl: 'https://www.youtube.com/watch?v=abc123xyz00',
          targetLanguage: 'zh-CN',
          status: 'running',
          stage: 'transcribing',
          progress: 55,
          errorMessage: '',
        }),
      )
      .mockResolvedValueOnce(
        createJsonResponse({
          id: 'job-2',
          videoId: 'def456uvw99',
          youtubeUrl: 'https://www.youtube.com/watch?v=def456uvw99',
          targetLanguage: 'zh-CN',
          status: 'completed',
          stage: 'completed',
          progress: 100,
          errorMessage: '',
        }),
      );

    vi.stubGlobal('fetch', fetchMock);

    await expect(getJobs(['job-1', 'job-2'])).resolves.toEqual([
      expect.objectContaining({ id: 'job-1', progress: 55 }),
      expect.objectContaining({ id: 'job-2', progress: 100 }),
    ]);

    expect(fetchMock).toHaveBeenNthCalledWith(1, 'http://localhost:8080/api/jobs/job-1');
    expect(fetchMock).toHaveBeenNthCalledWith(2, 'http://localhost:8080/api/jobs/job-2');
  });

  it('fetches subtitle assets by video id', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      createJsonResponse({
        videoId: 'abc123xyz00',
        targetLanguage: 'zh-CN',
        subtitleUrls: {
          translated: 'http://localhost:8080/assets/abc123xyz00/translated.vtt',
          bilingual: 'http://localhost:8080/assets/abc123xyz00/bilingual.vtt',
        },
      }),
    );

    vi.stubGlobal('fetch', fetchMock);

    const asset = await getSubtitleAsset('abc123xyz00');

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/videos/abc123xyz00/subtitles',
    );
    expect(asset.subtitleUrls.translated).toBe(
      'http://localhost:8080/assets/abc123xyz00/translated.vtt',
    );
  });
});
