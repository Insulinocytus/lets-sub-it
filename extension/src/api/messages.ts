import { browser } from 'wxt/browser'

export const SUPPORTED_LANGUAGES = ['en', 'zh-CN'] as const
export type LanguageCode = (typeof SUPPORTED_LANGUAGES)[number]

export const SUBTITLE_MODES = ['translated', 'bilingual'] as const
export type SubtitleMode = (typeof SUBTITLE_MODES)[number]

export type JobStatus =
  | 'queued'
  | 'downloading'
  | 'transcribing'
  | 'translating'
  | 'packaging'
  | 'completed'
  | 'failed'

export type Job = {
  id: string
  videoId: string
  youtubeUrl: string
  sourceLanguage: LanguageCode
  targetLanguage: LanguageCode
  status: JobStatus
  stage: JobStatus
  progressText: string
  errorMessage: string | null
  createdAt: string
  updatedAt: string
}

export type SubtitleAsset = {
  jobId: string
  videoId: string
  targetLanguage: LanguageCode
  sourceLanguage: LanguageCode
  files: {
    source: string
    translated: string
    bilingual: string
  }
  createdAt: string
}

export type Settings = {
  backendBaseUrl: string
  sourceLanguage: LanguageCode
  targetLanguage: LanguageCode
}

export type CreateJobInput = {
  youtubeUrl: string
  sourceLanguage: LanguageCode
  targetLanguage: LanguageCode
}

export type ExtensionMessage =
  | { type: 'settings:get' }
  | { type: 'settings:update'; payload: Partial<Settings> }
  | { type: 'job:create'; payload: CreateJobInput }
  | { type: 'job:get'; payload: { jobId: string } }
  | { type: 'subtitle:resolve'; payload: { videoId: string } }
  | {
      type: 'subtitle:fetch-file'
      payload: { jobId: string; mode: SubtitleMode }
    }
  | {
      type: 'subtitle:update-mode'
      payload: { videoId: string; targetLanguage: LanguageCode; mode: SubtitleMode }
    }

export type MessageError = {
  code: string
  message: string
}

export type MessageResult<T> =
  | { ok: true; data: T }
  | { ok: false; error: MessageError }

export function isSupportedLanguage(value: string): value is LanguageCode {
  return SUPPORTED_LANGUAGES.includes(value as LanguageCode)
}

export function assertDifferentLanguages(
  source: LanguageCode,
  target: LanguageCode,
) {
  if (source === target) {
    throw new Error('sourceLanguage and targetLanguage must be different')
  }
}

export async function sendExtensionMessage<T>(
  message: ExtensionMessage,
): Promise<MessageResult<T>> {
  return browser.runtime.sendMessage(message) as Promise<MessageResult<T>>
}
