import type {
  CreateJobInput,
  Job,
  LanguageCode,
  SubtitleAsset,
  SubtitleMode,
} from './messages'

export type CreateJobResponse = {
  job: Job
  reused: boolean
}

export type GetJobResponse = {
  job: Job
}

export type GetSubtitleAssetResponse = {
  asset: SubtitleAsset | null
}

type FetchLike = typeof fetch

export class BackendClientError extends Error {
  constructor(
    readonly code: string,
    message: string,
  ) {
    super(message)
    this.name = 'BackendClientError'
  }
}

export type BackendClient = {
  createJob(input: CreateJobInput): Promise<CreateJobResponse>
  getJob(jobId: string): Promise<GetJobResponse>
  getSubtitleAsset(
    videoId: string,
    targetLanguage: LanguageCode,
  ): Promise<GetSubtitleAssetResponse>
  fetchSubtitleFile(jobId: string, mode: SubtitleMode): Promise<string>
}

export function createBackendClient(
  backendBaseUrl: string,
  fetchImpl: FetchLike = fetch,
): BackendClient {
  const baseUrl = normalizeBackendBaseUrl(backendBaseUrl)

  return {
    createJob(input) {
      return requestJson<CreateJobResponse>(fetchImpl, `${baseUrl}/jobs`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(input),
      })
    },
    getJob(jobId) {
      return requestJson<GetJobResponse>(
        fetchImpl,
        `${baseUrl}/jobs/${encodeURIComponent(jobId)}`,
      )
    },
    getSubtitleAsset(videoId, targetLanguage) {
      const params = new URLSearchParams({ videoId, targetLanguage })
      return requestJson<GetSubtitleAssetResponse>(
        fetchImpl,
        `${baseUrl}/subtitle-assets?${params.toString()}`,
      )
    },
    async fetchSubtitleFile(jobId, mode) {
      const response = await request(
        fetchImpl,
        `${baseUrl}/subtitle-files/${encodeURIComponent(jobId)}/${mode}`,
      )
      return response.text()
    },
  }
}

function normalizeBackendBaseUrl(value: string): string {
  let url: URL
  try {
    url = new URL(value)
  } catch {
    throw new BackendClientError('invalid_backend_url', 'backendBaseUrl is invalid')
  }

  const isLocalHost = url.hostname === '127.0.0.1' || url.hostname === 'localhost'
  if (!isLocalHost || url.protocol !== 'http:') {
    throw new BackendClientError(
      'invalid_backend_url',
      'backendBaseUrl must use localhost or 127.0.0.1',
    )
  }

  return url.toString().replace(/\/$/, '')
}

async function requestJson<T>(
  fetchImpl: FetchLike,
  url: string,
  init?: RequestInit,
): Promise<T> {
  const response = await request(fetchImpl, url, init)
  return response.json() as Promise<T>
}

async function request(
  fetchImpl: FetchLike,
  url: string,
  init?: RequestInit,
): Promise<Response> {
  let response: Response
  try {
    response = await fetchImpl(url, init)
  } catch {
    throw new BackendClientError('network_error', 'Cannot connect to backend')
  }

  if (!response.ok) {
    throw await errorFromResponse(response)
  }

  return response
}

async function errorFromResponse(response: Response): Promise<BackendClientError> {
  try {
    const payload = (await response.json()) as {
      error?: { code?: string; message?: string }
    }
    if (payload.error?.code && payload.error.message) {
      return new BackendClientError(payload.error.code, payload.error.message)
    }
  } catch {
    return new BackendClientError('backend_error', `Backend returned ${response.status}`)
  }

  return new BackendClientError('backend_error', `Backend returned ${response.status}`)
}
