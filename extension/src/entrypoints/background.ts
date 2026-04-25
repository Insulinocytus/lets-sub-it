import type { ExtensionMessage, ExtensionResponse } from '@/types'

const BACKEND_BASE = 'http://127.0.0.1:8080'

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
      const jobId = (data.job as { id: string }).id
      stateCache.set(jobId, data)
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
    try {
      const res = await fetch(`${BACKEND_BASE}/jobs/${jobId}`)
      const data = await res.json()
      if (!res.ok) {
        sendResponse({ success: false, error: data.error?.message ?? 'get job failed' })
        return
      }
      stateCache.set(jobId, data)
      sendResponse({ success: true, data })
    } catch (err) {
      // Retry once on network error
      try {
        const res = await fetch(`${BACKEND_BASE}/jobs/${jobId}`)
        const data = await res.json()
        if (!res.ok) {
          sendResponse({ success: false, error: data.error?.message ?? 'get job failed' })
          return
        }
        stateCache.set(jobId, data)
        sendResponse({ success: true, data })
      } catch (err2) {
        const message = err2 instanceof Error ? err2.message : 'get job failed'
        sendResponse({ success: false, error: message })
      }
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
})
