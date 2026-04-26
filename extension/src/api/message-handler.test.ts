import { beforeEach, describe, expect, it, vi } from 'vitest'
import { fakeBrowser } from 'wxt/testing/fake-browser'
import { DEFAULT_SETTINGS, getSettings } from '@/storage/settings'
import { getCachedSubtitleAsset } from '@/storage/subtitle-cache'
import type { Job } from './messages'
import { handleExtensionMessage } from './message-handler'

const queuedJob: Job = {
  id: 'job_123',
  videoId: 'video_123',
  youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
  sourceLanguage: 'en',
  targetLanguage: 'zh-CN',
  status: 'queued',
  stage: 'queued',
  progressText: '等待处理',
  errorMessage: null,
  createdAt: '2026-04-25T00:00:00Z',
  updatedAt: '2026-04-25T00:00:00Z',
}

describe('handleExtensionMessage', () => {
  beforeEach(() => {
    fakeBrowser.reset()
  })

  it('rejects job creation when source and target languages are equal', async () => {
    const fetchImpl = vi.fn()

    const result = await handleExtensionMessage(
      {
        type: 'job:create',
        payload: {
          youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
          sourceLanguage: 'en',
          targetLanguage: 'en',
        },
      },
      { fetchImpl, now: () => '2026-04-25T00:00:00Z' },
    )

    expect(result).toEqual({
      ok: false,
      error: {
        code: 'invalid_language_pair',
        message: 'sourceLanguage and targetLanguage must be different',
      },
    })
    expect(fetchImpl).not.toHaveBeenCalled()
  })

  it('creates a job through the backend client', async () => {
    const fetchImpl = vi.fn(async () =>
      new Response(
        JSON.stringify({
          job: queuedJob,
          reused: false,
        }),
        { status: 201, headers: { 'Content-Type': 'application/json' } },
      ),
    )

    const result = await handleExtensionMessage(
      {
        type: 'job:create',
        payload: {
          youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
          sourceLanguage: 'en',
          targetLanguage: 'zh-CN',
        },
      },
      {
        fetchImpl,
        now: () => '2026-04-25T00:00:00Z',
        startJobMonitor: vi.fn(),
      },
    )

    expect(result.ok).toBe(true)
    expect(fetchImpl).toHaveBeenCalledOnce()
  })

  it('starts a job monitor after creating a job', async () => {
    await handleExtensionMessage({
      type: 'settings:update',
      payload: {
        backendBaseUrl: 'http://127.0.0.1:9090/',
      },
    })
    const fetchImpl = vi.fn(async () =>
      new Response(JSON.stringify({ job: queuedJob, reused: false }), {
        status: 201,
        headers: { 'Content-Type': 'application/json' },
      }),
    )
    const startJobMonitor = vi.fn()

    const result = await handleExtensionMessage(
      {
        type: 'job:create',
        payload: {
          youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
          sourceLanguage: 'en',
          targetLanguage: 'zh-CN',
        },
      },
      { fetchImpl, now: () => '2026-04-25T00:00:00Z', startJobMonitor },
    )

    expect(result.ok).toBe(true)
    expect(startJobMonitor).toHaveBeenCalledWith(
      queuedJob,
      expect.objectContaining({
        backendBaseUrl: 'http://127.0.0.1:9090',
        now: expect.any(Function),
      }),
    )
  })

  it('returns create success even when starting the job monitor fails', async () => {
    const fetchImpl = vi.fn(async () =>
      new Response(JSON.stringify({ job: queuedJob, reused: false }), {
        status: 201,
        headers: { 'Content-Type': 'application/json' },
      }),
    )
    const startJobMonitor = vi.fn(
      async () => Promise.reject(new Error('monitor bootstrap failed')),
    )

    const result = await handleExtensionMessage(
      {
        type: 'job:create',
        payload: {
          youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
          sourceLanguage: 'en',
          targetLanguage: 'zh-CN',
        },
      },
      { fetchImpl, now: () => '2026-04-25T00:00:00Z', startJobMonitor },
    )

    expect(result).toEqual({
      ok: true,
      data: {
        job: queuedJob,
        reused: false,
      },
    })
    expect(startJobMonitor).toHaveBeenCalledWith(
      queuedJob,
      expect.objectContaining({
        now: expect.any(Function),
      }),
    )
  })

  it('rejects invalid backend URLs without persisting settings', async () => {
    const result = await handleExtensionMessage({
      type: 'settings:update',
      payload: {
        backendBaseUrl: 'http://localhost:8080/api',
      },
    })

    expect(result).toEqual({
      ok: false,
      error: {
        code: 'invalid_backend_url',
        message: 'backendBaseUrl must be a localhost or 127.0.0.1 origin',
      },
    })
    await expect(getSettings()).resolves.toEqual(DEFAULT_SETTINGS)
  })

  it('resolves and caches a subtitle asset from backend when local cache is empty', async () => {
    const fetchImpl = vi.fn(async () =>
      new Response(
        JSON.stringify({
          asset: {
            jobId: 'job_123',
            videoId: 'video_123',
            sourceLanguage: 'en',
            targetLanguage: 'zh-CN',
            files: {
              source: '/subtitle-files/job_123/source',
              translated: '/subtitle-files/job_123/translated',
              bilingual: '/subtitle-files/job_123/bilingual',
            },
            createdAt: '2026-04-25T00:00:00Z',
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    )

    const result = await handleExtensionMessage(
      { type: 'subtitle:resolve', payload: { videoId: 'video_123' } },
      { fetchImpl, now: () => '2026-04-25T00:01:00Z' },
    )

    expect(result.ok).toBe(true)
    await expect(
      getCachedSubtitleAsset(DEFAULT_SETTINGS.backendBaseUrl, 'video_123', 'zh-CN'),
    ).resolves.toMatchObject({
      jobId: 'job_123',
      selectedMode: 'translated',
    })
  })

  it('does not reuse cached subtitles from a different backend origin', async () => {
    await handleExtensionMessage({
      type: 'settings:update',
      payload: {
        backendBaseUrl: 'http://127.0.0.1:8080',
      },
    })

    const oldBackendFetch = vi.fn(async () =>
      new Response(
        JSON.stringify({
          asset: {
            jobId: 'job_old',
            videoId: 'video_123',
            sourceLanguage: 'en',
            targetLanguage: 'zh-CN',
            files: {
              source: '/subtitle-files/job_old/source',
              translated: '/subtitle-files/job_old/translated',
              bilingual: '/subtitle-files/job_old/bilingual',
            },
            createdAt: '2026-04-25T00:00:00Z',
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    )

    await handleExtensionMessage(
      { type: 'subtitle:resolve', payload: { videoId: 'video_123' } },
      { fetchImpl: oldBackendFetch, now: () => '2026-04-25T00:01:00Z' },
    )

    await handleExtensionMessage({
      type: 'settings:update',
      payload: {
        backendBaseUrl: 'http://127.0.0.1:9090/',
      },
    })

    const newBackendFetch = vi.fn(async () =>
      new Response(
        JSON.stringify({
          asset: {
            jobId: 'job_new',
            videoId: 'video_123',
            sourceLanguage: 'en',
            targetLanguage: 'zh-CN',
            files: {
              source: '/subtitle-files/job_new/source',
              translated: '/subtitle-files/job_new/translated',
              bilingual: '/subtitle-files/job_new/bilingual',
            },
            createdAt: '2026-04-25T00:02:00Z',
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    )

    const result = await handleExtensionMessage(
      { type: 'subtitle:resolve', payload: { videoId: 'video_123' } },
      { fetchImpl: newBackendFetch, now: () => '2026-04-25T00:03:00Z' },
    )

    expect(result).toEqual({
      ok: true,
      data: expect.objectContaining({
        jobId: 'job_new',
      }),
    })
    expect(newBackendFetch).toHaveBeenCalledOnce()
  })
})
