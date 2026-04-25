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

  it('rejects same source and target languages', async () => {
    await expect(
      updateSettings({ sourceLanguage: 'en', targetLanguage: 'en' }),
    ).rejects.toThrow('sourceLanguage and targetLanguage must be different')
  })
})
