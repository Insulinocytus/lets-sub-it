import { storage } from 'wxt/utils/storage'
import { normalizeBackendBaseUrl } from '@/api/backend-client'
import {
  SUBTITLE_MODES,
  assertDifferentLanguages,
  type LanguageCode,
  type Settings,
  type SubtitleMode,
} from '@/api/messages'

export const DEFAULT_SETTINGS: Settings = Object.freeze({
  backendBaseUrl: 'http://127.0.0.1:8080',
  sourceLanguage: 'en',
  targetLanguage: 'zh',
  subtitleFontSizePx: 20,
  subtitleMode: 'translated',
})

const settingsItem = storage.defineItem<Settings>('local:settings', {
  fallback: DEFAULT_SETTINGS,
})

export async function getSettings(): Promise<Settings> {
  const settings = await settingsItem.getValue()
  return { ...DEFAULT_SETTINGS, ...settings }
}

export async function updateSettings(patch: Partial<Settings>): Promise<Settings> {
  const current = await getSettings()
  const next: Settings = {
    ...current,
    ...patch,
  }

  assertDifferentLanguages(next.sourceLanguage, next.targetLanguage)
  assertPositiveSubtitleFontSize(next.subtitleFontSizePx)
  assertSubtitleMode(next.subtitleMode)
  const normalizedNext: Settings = {
    ...next,
    backendBaseUrl: normalizeBackendBaseUrl(next.backendBaseUrl),
  }

  await settingsItem.setValue(normalizedNext)
  return normalizedNext
}

export function createLanguagePair(
  sourceLanguage: LanguageCode,
  targetLanguage: LanguageCode,
) {
  assertDifferentLanguages(sourceLanguage, targetLanguage)
  return { sourceLanguage, targetLanguage }
}

function assertPositiveSubtitleFontSize(value: number) {
  if (!Number.isFinite(value) || value <= 0) {
    throw new Error('subtitleFontSizePx must be a positive number')
  }
}

function assertSubtitleMode(value: SubtitleMode) {
  if (!SUBTITLE_MODES.includes(value)) {
    throw new Error('subtitleMode must be translated or bilingual')
  }
}
