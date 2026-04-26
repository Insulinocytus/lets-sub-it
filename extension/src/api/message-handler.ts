import { BackendClientError, createBackendClient } from './backend-client'
import {
  assertDifferentLanguages,
  type ExtensionMessage,
  type MessageError,
  type MessageResult,
} from './messages'
import { getSettings, updateSettings } from '@/storage/settings'
import {
  getCachedSubtitleAsset,
  getVideoPreference,
  setCachedSubtitleAsset,
  updateCachedSubtitleMode,
} from '@/storage/subtitle-cache'
import { startJobMonitor } from './job-monitor'

export type MessageHandlerDeps = {
  fetchImpl?: typeof fetch
  now?: () => string
  startJobMonitor?: typeof startJobMonitor
}

export async function handleExtensionMessage(
  message: ExtensionMessage,
  deps: MessageHandlerDeps = {},
): Promise<MessageResult<unknown>> {
  try {
    const fetchImpl = deps.fetchImpl ?? fetch
    const now = deps.now ?? (() => new Date().toISOString())

    switch (message.type) {
      case 'settings:get':
        return ok(await getSettings())
      case 'settings:update':
        return ok(await updateSettings(message.payload))
      case 'job:create': {
        assertDifferentLanguages(
          message.payload.sourceLanguage,
          message.payload.targetLanguage,
        )
        const settings = await updateSettings({
          sourceLanguage: message.payload.sourceLanguage,
          targetLanguage: message.payload.targetLanguage,
        })
        const client = createBackendClient(settings.backendBaseUrl, fetchImpl)
        const response = await client.createJob(message.payload)
        const monitor = deps.startJobMonitor ?? startJobMonitor
        void Promise.resolve(
          monitor(response.job, {
            backendBaseUrl: settings.backendBaseUrl,
            client,
            now,
          }),
        ).catch(() => {})
        return ok(response)
      }
      case 'job:get': {
        const client = await clientFromSettings(fetchImpl)
        return ok(await client.getJob(message.payload.jobId))
      }
      case 'subtitle:resolve': {
        const settings = await getSettings()
        const preference = await getVideoPreference(
          settings.backendBaseUrl,
          message.payload.videoId,
        )
        const targetLanguage = preference?.targetLanguage ?? settings.targetLanguage
        const cached = await getCachedSubtitleAsset(
          settings.backendBaseUrl,
          message.payload.videoId,
          targetLanguage,
        )
        if (cached) {
          return ok(cached)
        }

        const client = createBackendClient(settings.backendBaseUrl, fetchImpl)
        const response = await client.getSubtitleAsset(
          message.payload.videoId,
          targetLanguage,
        )
        if (!response.asset) {
          return ok(null)
        }

        const entry = await setCachedSubtitleAsset(
          response.asset,
          preference?.selectedMode ?? 'translated',
          now(),
          settings.backendBaseUrl,
        )
        return ok(entry)
      }
      case 'subtitle:fetch-file': {
        const client = await clientFromSettings(fetchImpl)
        return ok(
          await client.fetchSubtitleFile(message.payload.jobId, message.payload.mode),
        )
      }
      case 'subtitle:update-mode':
        return ok(await updateSubtitleMode(message, deps))
    }
  } catch (error) {
    return { ok: false, error: errorToMessage(error) }
  }
}

async function clientFromSettings(fetchImpl: typeof fetch) {
  const settings = await getSettings()
  return createBackendClient(settings.backendBaseUrl, fetchImpl)
}

async function updateSubtitleMode(
  message: Extract<ExtensionMessage, { type: 'subtitle:update-mode' }>,
  _deps: MessageHandlerDeps,
) {
  const settings = await getSettings()
  return updateCachedSubtitleMode(
    settings.backendBaseUrl,
    message.payload.videoId,
    message.payload.targetLanguage,
    message.payload.mode,
  )
}

function ok<T>(data: T): MessageResult<T> {
  return { ok: true, data }
}

function errorToMessage(error: unknown): MessageError {
  if (error instanceof BackendClientError) {
    return { code: error.code, message: error.message }
  }
  if (
    error instanceof Error &&
    error.message === 'sourceLanguage and targetLanguage must be different'
  ) {
    return {
      code: 'invalid_language_pair',
      message: 'sourceLanguage and targetLanguage must be different',
    }
  }
  if (error instanceof Error) {
    return { code: 'internal_error', message: error.message }
  }
  return { code: 'internal_error', message: 'Unknown extension error' }
}
