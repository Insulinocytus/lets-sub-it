import { beforeEach, describe, expect, it } from 'vitest'
import { fakeBrowser } from 'wxt/testing/fake-browser'
import { DEFAULT_SETTINGS, getSettings, updateSettings } from './settings'

describe('settings storage', () => {
  beforeEach(() => {
    fakeBrowser.reset()
  })

  it('returns default settings when storage is empty', async () => {
    await expect(getSettings()).resolves.toEqual(DEFAULT_SETTINGS)
  })

  it('does not let callers mutate the default settings fallback', async () => {
    const first = await getSettings()
    first.backendBaseUrl = 'http://localhost:9090'

    await expect(getSettings()).resolves.toEqual(DEFAULT_SETTINGS)
  })

  it('updates backend URL and language settings', async () => {
    const settings = await updateSettings({
      backendBaseUrl: 'http://localhost:9090',
      sourceLanguage: 'zh-CN',
      targetLanguage: 'en',
    })

    expect(settings).toEqual({
      backendBaseUrl: 'http://localhost:9090',
      sourceLanguage: 'zh-CN',
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
