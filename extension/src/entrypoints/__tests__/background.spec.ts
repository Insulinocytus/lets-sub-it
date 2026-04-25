import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { ExtensionResponse, JobResponse } from '@/types'

type OnMessageListener = (
  message: { type: string; payload: Record<string, unknown> },
  sender: unknown,
  sendResponse: (response: ExtensionResponse) => void,
) => boolean | void

const job: JobResponse = {
  id: 'job_123',
  videoId: 'video_123',
  youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
  sourceLanguage: 'ja',
  targetLanguage: 'zh-Hans',
  status: 'queued',
  stage: 'downloading',
  progressText: 'Downloading audio',
  errorMessage: null,
  createdAt: '2026-04-25T00:00:00Z',
  updatedAt: '2026-04-25T00:00:00Z',
}

let listener: OnMessageListener | null = null

async function loadBackground() {
  await import('../background')
}

async function sendMessage(type: string, payload: Record<string, unknown>) {
  if (!listener) {
    throw new Error('background listener was not registered')
  }

  return await new Promise<ExtensionResponse>((resolve) => {
    listener!({ type, payload }, {}, resolve)
  })
}

async function sendGetJobResponse(payload: Record<string, unknown>) {
  return await sendMessage('GET_JOB', payload)
}

beforeEach(() => {
  vi.resetModules()
  listener = null

  vi.stubGlobal('defineBackground', (setup: () => unknown) => setup())
  vi.stubGlobal('chrome', {
    runtime: {
      onMessage: {
        addListener: vi.fn((registeredListener: OnMessageListener) => {
          listener = registeredListener
        }),
      },
    },
  })
})

describe('background GET_SUBTITLE_FILE', () => {
  it('returns subtitle file text from the backend', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      text: vi.fn().mockResolvedValue('WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nHello'),
    })
    vi.stubGlobal('fetch', fetchMock)

    await loadBackground()

    const response = await sendMessage('GET_SUBTITLE_FILE', {
      jobId: 'job_123',
      mode: 'bilingual',
    })

    expect(fetchMock).toHaveBeenCalledWith(
      'http://127.0.0.1:8080/subtitle-files/job_123/bilingual',
    )
    expect(response).toEqual({
      success: true,
      data: 'WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nHello',
    })
  })

  it('returns an error when the backend rejects subtitle file requests', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      status: 404,
      text: vi.fn().mockResolvedValue('not found'),
    })
    vi.stubGlobal('fetch', fetchMock)

    await loadBackground()

    const response = await sendMessage('GET_SUBTITLE_FILE', {
      jobId: 'missing-job',
      mode: 'translated',
    })

    expect(response).toEqual({
      success: false,
      error: 'get subtitle file failed: HTTP 404',
    })
  })

  it('rejects non-string jobId without calling the backend', async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)

    await loadBackground()

    const response = await sendMessage('GET_SUBTITLE_FILE', {
      jobId: { id: 'job_123' },
      mode: 'bilingual',
    })

    expect(fetchMock).not.toHaveBeenCalled()
    expect(response).toEqual({
      success: false,
      error: 'jobId and valid mode are required',
    })
  })
})

describe('background GET_JOB', () => {
  it('unwraps the job object from the backend envelope', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue({ job }),
    })
    vi.stubGlobal('fetch', fetchMock)

    await loadBackground()

    const response = await sendGetJobResponse({ jobId: 'job_123' })

    expect(response).toEqual({ success: true, data: job })
    expect(response.data).not.toHaveProperty('job')
  })

  it('rejects successful responses without a valid job envelope', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue({ job: null }),
    })
    vi.stubGlobal('fetch', fetchMock)

    await loadBackground()

    const response = await sendGetJobResponse({ jobId: 'job_123' })

    expect(response.success).toBe(false)
    expect(response.error).toContain('invalid job response')
  })

  it('does not retry application-level GET_JOB errors', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      json: vi.fn().mockResolvedValue({
        error: { message: 'job not found' },
      }),
    })
    vi.stubGlobal('fetch', fetchMock)

    await loadBackground()

    const response = await sendGetJobResponse({ jobId: 'missing-job' })

    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(response).toEqual({ success: false, error: 'job not found' })
  })
})
