import { beforeEach, describe, expect, it } from 'vitest'
import { fakeBrowser } from 'wxt/testing/fake-browser'
import { DEFAULT_SETTINGS, getSettings, updateSettings } from './settings'

describe('settings storage', () => {
  beforeEach(() => {
    fakeBrowser.reset()
  })

  it('returns default settings when storage is empty', async () => {
    await expect(getSettings()).resolves.toEqual(DEFAULT_SETTINGS)
    expect(DEFAULT_SETTINGS.targetLanguage).toBe('zh')
  })

  it('includes default subtitle display settings', async () => {
    await expect(getSettings()).resolves.toEqual({
      backendBaseUrl: 'http://127.0.0.1:8080',
      sourceLanguage: 'en',
      targetLanguage: 'zh',
      subtitleFontSizePx: 20,
      subtitleMode: 'translated',
    })
  })

  it('merges subtitle defaults into previously stored settings', async () => {
    await fakeBrowser.storage.local.set({
      settings: {
        backendBaseUrl: 'http://localhost:9090',
        sourceLanguage: 'zh',
        targetLanguage: 'en',
      },
    })

    await expect(getSettings()).resolves.toEqual({
      backendBaseUrl: 'http://localhost:9090',
      sourceLanguage: 'zh',
      targetLanguage: 'en',
      subtitleFontSizePx: 20,
      subtitleMode: 'translated',
    })
  })

  it('does not let callers mutate the default settings fallback', async () => {
    const first = await getSettings()
    first.backendBaseUrl = 'http://localhost:9090'

    await expect(getSettings()).resolves.toEqual(DEFAULT_SETTINGS)
  })

  it('updates backend URL and language settings', async () => {
    const settings = await updateSettings({
      backendBaseUrl: 'http://localhost:9090',
      sourceLanguage: 'zh',
      targetLanguage: 'en',
    })

    expect(settings).toEqual({
      ...DEFAULT_SETTINGS,
      backendBaseUrl: 'http://localhost:9090',
      sourceLanguage: 'zh',
      targetLanguage: 'en',
    })
    await expect(getSettings()).resolves.toEqual(settings)
  })

  it('normalizes backend URL before persisting settings', async () => {
    const settings = await updateSettings({
      backendBaseUrl: 'http://localhost:9090/',
    })

    expect(settings.backendBaseUrl).toBe('http://localhost:9090')
    await expect(getSettings()).resolves.toEqual({
      ...DEFAULT_SETTINGS,
      backendBaseUrl: 'http://localhost:9090',
    })
  })

  it('updates subtitle display settings without imposing an upper bound', async () => {
    const settings = await updateSettings({
      subtitleFontSizePx: 240,
      subtitleMode: 'bilingual',
    })

    expect(settings).toEqual({
      ...DEFAULT_SETTINGS,
      subtitleFontSizePx: 240,
      subtitleMode: 'bilingual',
    })
    await expect(getSettings()).resolves.toEqual(settings)
  })

  it('rejects non-positive subtitle font sizes without persisting settings', async () => {
    await expect(updateSettings({ subtitleFontSizePx: 0 })).rejects.toThrow(
      'subtitleFontSizePx must be a positive number',
    )
    await expect(updateSettings({ subtitleFontSizePx: -4 })).rejects.toThrow(
      'subtitleFontSizePx must be a positive number',
    )
    await expect(getSettings()).resolves.toEqual(DEFAULT_SETTINGS)
  })

  it('rejects same source and target languages', async () => {
    await expect(
      updateSettings({ sourceLanguage: 'en', targetLanguage: 'en' }),
    ).rejects.toThrow('sourceLanguage and targetLanguage must be different')
  })

  it('rejects invalid backend ports without persisting settings', async () => {
    await expect(
      updateSettings({ backendBaseUrl: 'http://localhost:65536' }),
    ).rejects.toMatchObject({
      code: 'invalid_backend_url',
      message: 'backendBaseUrl must be a localhost or 127.0.0.1 origin',
    })
    await expect(getSettings()).resolves.toEqual(DEFAULT_SETTINGS)
  })
})
