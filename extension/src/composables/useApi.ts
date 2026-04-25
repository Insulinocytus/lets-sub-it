import type {
  CreateJobParams,
  CreateJobResult,
  ExtensionMessage,
  ExtensionResponse,
  JobResponse,
  SubtitleAssetResponse,
} from '@/types'

let inflightMap = new Map<string, Promise<unknown>>()

async function sendMessage(type: string, payload: Record<string, unknown>): Promise<unknown> {
  const key = `${type}:${JSON.stringify(payload)}`
  const inflight = inflightMap.get(key)
  if (inflight) return inflight

  const promise = (async () => {
    try {
      const response: ExtensionResponse = await chrome.runtime.sendMessage({
        type,
        payload,
      } as ExtensionMessage)
      if (!response.success) {
        throw new Error(response.error ?? 'unknown error')
      }
      return response.data
    } finally {
      inflightMap.delete(key)
    }
  })()

  inflightMap.set(key, promise)
  return promise
}

export function useApi() {
  async function createJob(params: CreateJobParams): Promise<CreateJobResult> {
    return sendMessage('CREATE_JOB', params as unknown as Record<string, unknown>) as Promise<CreateJobResult>
  }

  async function getJob(jobId: string): Promise<JobResponse> {
    return sendMessage('GET_JOB', { jobId }) as Promise<JobResponse>
  }

  async function getSubtitleAssets(videoId: string, targetLanguage: string): Promise<SubtitleAssetResponse> {
    return sendMessage('GET_SUBTITLE_ASSETS', { videoId, targetLanguage }) as Promise<SubtitleAssetResponse>
  }

  return { createJob, getJob, getSubtitleAssets }
}
