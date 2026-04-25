// --- Backend API response types (matching backend mock API) ---

export interface JobResponse {
  id: string
  videoId: string
  youtubeUrl: string
  sourceLanguage: string
  targetLanguage: string
  status: JobStatus
  stage: string
  progressText: string
  errorMessage: string | null
  createdAt: string
  updatedAt: string
}

export type JobStatus =
  | 'queued'
  | 'downloading'
  | 'transcribing'
  | 'translating'
  | 'packaging'
  | 'completed'
  | 'failed'

export type SubtitleMode = 'translated' | 'bilingual'

export interface CreateJobParams {
  youtubeUrl: string
  sourceLanguage: string
  targetLanguage: string
}

export interface CreateJobResult {
  job: JobResponse
  reused: boolean
}

export interface SubtitleAssetResponse {
  asset: {
    jobId: string
    videoId: string
    targetLanguage: string
    sourceLanguage: string
    files: {
      source: string
      translated: string
      bilingual: string
    }
    createdAt: string
  } | null
}

// --- Extension storage types ---

export interface LocalCacheEntry {
  videoId: string
  targetLanguage: string
  jobId: string
  selectedMode: SubtitleMode
  lastSyncedAt: string
}

export interface UserPreferences {
  videoId: string
  targetLanguage: string | null
  selectedMode: SubtitleMode
}

// --- Message types for popup <-> background communication ---

export type MessageType =
  | 'CREATE_JOB'
  | 'GET_JOB'
  | 'GET_SUBTITLE_ASSETS'
  | 'GET_SUBTITLE_FILE'

export interface ExtensionMessage {
  type: MessageType
  payload: Record<string, unknown>
}

export interface ExtensionResponse {
  success: boolean
  data?: unknown
  error?: string
}

// --- Content script types ---

export interface SubtitleCue {
  start: number  // seconds
  end: number    // seconds
  text: string
}

export interface SubtitleData {
  videoId: string
  cues: SubtitleCue[]
  mode: SubtitleMode
  vttUrl: string
}

// --- VTT parser ---

export interface VttParseResult {
  cues: SubtitleCue[]
}
