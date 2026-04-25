import type { ExtensionMessage, ExtensionResponse, JobResponse } from '@/types'

const BACKEND_BASE = 'http://127.0.0.1:8080'

function extractJobResponse(data: unknown): JobResponse | null {
  if (!data || typeof data !== 'object') {
    return null
  }

  const job = (data as { job?: unknown }).job
  if (!job || typeof job !== 'object') {
    return null
  }

  const record = job as Record<string, unknown>
  if (typeof record.id !== 'string' || typeof record.status !== 'string') {
    return null
  }

  return job as JobResponse
}

export default defineBackground(() => {
  // In-memory state cache for popup recovery
  const stateCache = new Map<string, unknown>()

  chrome.runtime.onMessage.addListener(
    (message: ExtensionMessage, _sender, sendResponse: (response: ExtensionResponse) => void) => {
      switch (message.type) {
        case 'CREATE_JOB':
          handleCreateJob(message.payload, sendResponse)
          return true // keep channel open for async
        case 'GET_JOB':
          handleGetJob(message.payload, sendResponse)
          return true
        case 'GET_SUBTITLE_ASSETS':
          handleGetSubtitleAssets(message.payload, sendResponse)
          return true
        case 'GET_SUBTITLE_FILE':
          handleGetSubtitleFile(message.payload, sendResponse)
          return true
        default:
          sendResponse({ success: false, error: `unknown message type: ${message.type}` })
          return false
      }
    },
  )

  async function handleCreateJob(
    payload: Record<string, unknown>,
    sendResponse: (response: ExtensionResponse) => void,
  ) {
    try {
      const res = await fetch(`${BACKEND_BASE}/jobs`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          youtubeUrl: payload.youtubeUrl,
          sourceLanguage: payload.sourceLanguage,
          targetLanguage: payload.targetLanguage,
        }),
      })
      const data = await res.json()
      if (!res.ok) {
        sendResponse({ success: false, error: data.error?.message ?? 'create job failed' })
        return
      }
      const job = data.job
      if (!job || typeof job.id !== 'string') {
        sendResponse({ success: false, error: 'invalid job response: missing job.id' })
        return
      }
      stateCache.set(job.id, data)
      sendResponse({ success: true, data })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'create job failed'
      sendResponse({ success: false, error: message })
    }
  }

  async function handleGetJob(
    payload: Record<string, unknown>,
    sendResponse: (response: ExtensionResponse) => void,
  ) {
    const jobId = payload.jobId as string
    if (!jobId) {
      sendResponse({ success: false, error: 'jobId is required' })
      return
    }

    async function tryFetch(): Promise<{ ok: boolean; data?: unknown; error?: string }> {
      try {
        const res = await fetch(`${BACKEND_BASE}/jobs/${jobId}`)
        const data = await res.json()
        if (!res.ok) {
          return { ok: false, error: data.error?.message ?? 'get job failed' }
        }
        const job = extractJobResponse(data)
        if (!job) {
          return { ok: false, error: 'invalid job response: missing job' }
        }
        return { ok: true, data: job }
      } catch {
        return { ok: false }
      }
    }

    const first = await tryFetch()
    if (first.ok) {
      stateCache.set(jobId, first.data)
      sendResponse({ success: true, data: first.data })
      return
    }
    if (first.error) {
      // Application-level error, not a network failure — don't retry
      sendResponse({ success: false, error: first.error })
      return
    }

    // Network error — retry once
    const second = await tryFetch()
    if (second.ok) {
      stateCache.set(jobId, second.data)
      sendResponse({ success: true, data: second.data })
    } else {
      sendResponse({ success: false, error: second.error ?? 'get job failed' })
    }
  }

  async function handleGetSubtitleAssets(
    payload: Record<string, unknown>,
    sendResponse: (response: ExtensionResponse) => void,
  ) {
    const videoId = payload.videoId as string
    const targetLanguage = payload.targetLanguage as string
    if (!videoId || !targetLanguage) {
      sendResponse({ success: false, error: 'videoId and targetLanguage are required' })
      return
    }
    try {
      const url = `${BACKEND_BASE}/subtitle-assets?videoId=${encodeURIComponent(videoId)}&targetLanguage=${encodeURIComponent(targetLanguage)}`
      const res = await fetch(url)
      const data = await res.json()
      if (!res.ok) {
        sendResponse({ success: false, error: data.error?.message ?? 'get subtitle assets failed' })
        return
      }
      sendResponse({ success: true, data })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'get subtitle assets failed'
      sendResponse({ success: false, error: message })
    }
  }

  async function handleGetSubtitleFile(
    payload: Record<string, unknown>,
    sendResponse: (response: ExtensionResponse) => void,
  ) {
    const jobId = payload.jobId
    const mode = payload.mode as string
    if (
      typeof jobId !== 'string' ||
      jobId.length === 0 ||
      (mode !== 'translated' && mode !== 'bilingual')
    ) {
      sendResponse({ success: false, error: 'jobId and valid mode are required' })
      return
    }

    try {
      const res = await fetch(
        `${BACKEND_BASE}/subtitle-files/${encodeURIComponent(jobId)}/${mode}`,
      )
      const text = await res.text()
      if (!res.ok) {
        sendResponse({
          success: false,
          error: `get subtitle file failed: HTTP ${res.status}`,
        })
        return
      }
      sendResponse({ success: true, data: text })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'get subtitle file failed'
      sendResponse({ success: false, error: message })
    }
  }
})
