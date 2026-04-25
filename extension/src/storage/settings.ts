import { storage } from 'wxt/utils/storage'
import {
  assertDifferentLanguages,
  type LanguageCode,
  type Settings,
} from '@/api/messages'

export const DEFAULT_SETTINGS: Settings = Object.freeze({
  backendBaseUrl: 'http://127.0.0.1:8080',
  sourceLanguage: 'en',
  targetLanguage: 'zh-CN',
})

const settingsItem = storage.defineItem<Settings>('local:settings', {
  fallback: DEFAULT_SETTINGS,
})

export async function getSettings(): Promise<Settings> {
  const settings = await settingsItem.getValue()
  return { ...settings }
}

export async function updateSettings(patch: Partial<Settings>): Promise<Settings> {
  const current = await getSettings()
  const next: Settings = {
    ...current,
    ...patch,
  }

  assertDifferentLanguages(next.sourceLanguage, next.targetLanguage)
  await settingsItem.setValue(next)
  return next
}

export function createLanguagePair(
  sourceLanguage: LanguageCode,
  targetLanguage: LanguageCode,
) {
  assertDifferentLanguages(sourceLanguage, targetLanguage)
  return { sourceLanguage, targetLanguage }
}
