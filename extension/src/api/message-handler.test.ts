import { beforeEach, describe, expect, it, vi } from 'vitest'
import { fakeBrowser } from 'wxt/testing/fake-browser'
import { getCachedSubtitleAsset } from '@/storage/subtitle-cache'
import { handleExtensionMessage } from './message-handler'

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
          job: {
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
          },
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
      { fetchImpl, now: () => '2026-04-25T00:00:00Z' },
    )

    expect(result.ok).toBe(true)
    expect(fetchImpl).toHaveBeenCalledOnce()
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
    await expect(getCachedSubtitleAsset('video_123', 'zh-CN')).resolves.toMatchObject({
      jobId: 'job_123',
      selectedMode: 'translated',
    })
  })
})
