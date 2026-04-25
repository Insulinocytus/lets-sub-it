import { describe, expect, it, vi } from 'vitest'
import {
  BackendClientError,
  createBackendClient,
  normalizeBackendBaseUrl,
} from './backend-client'

describe('createBackendClient', () => {
  it('creates a job through POST /jobs', async () => {
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
    const client = createBackendClient('http://127.0.0.1:8080/', fetchImpl)

    const response = await client.createJob({
      youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
      sourceLanguage: 'en',
      targetLanguage: 'zh-CN',
    })

    expect(response.reused).toBe(false)
    expect(response.job.id).toBe('job_123')
    expect(fetchImpl).toHaveBeenCalledWith('http://127.0.0.1:8080/jobs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
        sourceLanguage: 'en',
        targetLanguage: 'zh-CN',
      }),
    })
  })

  it('fetches VTT text from subtitle file endpoint', async () => {
    const fetchImpl = vi.fn(async () => new Response('WEBVTT\n', { status: 200 }))
    const client = createBackendClient('http://localhost:8080', fetchImpl)

    await expect(client.fetchSubtitleFile('job_123', 'translated')).resolves.toBe(
      'WEBVTT\n',
    )
  })

  it('fetches subtitle asset metadata with video and language query', async () => {
    const fetchImpl = vi.fn(async () =>
      new Response(
        JSON.stringify({
          asset: {
            jobId: 'job_123',
            videoId: 'video_123',
            targetLanguage: 'zh-CN',
            sourceLanguage: 'en',
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
    const client = createBackendClient('http://localhost:8080/', fetchImpl)

    const response = await client.getSubtitleAsset('video_123', 'zh-CN')

    expect(response.asset?.jobId).toBe('job_123')
    expect(fetchImpl).toHaveBeenCalledWith(
      'http://localhost:8080/subtitle-assets?videoId=video_123&targetLanguage=zh-CN',
      undefined,
    )
  })

  it('normalizes localhost backend URLs to their origin', () => {
    expect(normalizeBackendBaseUrl('http://localhost:8080/')).toBe(
      'http://localhost:8080',
    )
  })

  it('converts backend JSON errors into BackendClientError', async () => {
    const fetchImpl = vi.fn(async () =>
      new Response(
        JSON.stringify({
          error: { code: 'invalid_request', message: 'bad input' },
        }),
        { status: 400, headers: { 'Content-Type': 'application/json' } },
      ),
    )
    const client = createBackendClient('http://127.0.0.1:8080', fetchImpl)

    await expect(client.getJob('missing')).rejects.toMatchObject({
      code: 'invalid_request',
      message: 'bad input',
    })
  })

  it('rejects non-local backend URLs', () => {
    expect(() => createBackendClient('https://api.example.com', fetch)).toThrow(
      'backendBaseUrl must be a localhost or 127.0.0.1 origin',
    )
  })

  it('rejects backend URLs with path, query, hash, or credentials', () => {
    const invalidUrls = [
      'http://localhost',
      'http://127.0.0.1',
      'http://localhost:8080/api',
      'http://localhost:8080/?debug=1',
      'http://localhost:8080/#x',
      'http://user@localhost:8080',
    ]

    for (const url of invalidUrls) {
      expect(() => createBackendClient(url, fetch)).toThrow(BackendClientError)
      expect(() => createBackendClient(url, fetch)).toThrow(
        'backendBaseUrl must be a localhost or 127.0.0.1 origin',
      )
    }
  })

  it('uses network_error for failed fetch calls', async () => {
    const fetchImpl = vi.fn(async () => {
      throw new TypeError('failed to fetch')
    })
    const client = createBackendClient('http://127.0.0.1:8080', fetchImpl)

    const promise = client.getJob('job_123')

    await expect(promise).rejects.toBeInstanceOf(BackendClientError)
    await expect(promise).rejects.toMatchObject({
      code: 'network_error',
    })
  })
})
