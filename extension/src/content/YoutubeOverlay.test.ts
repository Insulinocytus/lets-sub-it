import { flushPromises, mount } from '@vue/test-utils'
import type { VueWrapper } from '@vue/test-utils'
import type { Settings } from '@/api/messages'
import type { SubtitleAssetCacheEntry } from '@/storage/subtitle-cache'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import YoutubeOverlay from './YoutubeOverlay.vue'

const {
  sendExtensionMessage,
  getCurrentVideoId,
  watchVideoIdChanges,
  addRuntimeListener,
  removeRuntimeListener,
} = vi.hoisted(() => ({
  sendExtensionMessage: vi.fn(),
  getCurrentVideoId: vi.fn(),
  watchVideoIdChanges: vi.fn(),
  addRuntimeListener: vi.fn(),
  removeRuntimeListener: vi.fn(),
}))

vi.mock('wxt/browser', () => ({
  browser: {
    runtime: {
      onMessage: {
        addListener: addRuntimeListener,
        removeListener: removeRuntimeListener,
      },
    },
  },
}))

vi.mock('@/api/messages', () => ({
  SUBTITLE_MODES: ['translated', 'bilingual'],
  sendExtensionMessage,
}))

vi.mock('@/youtube/page-watch', () => ({
  getCurrentVideoId,
  watchVideoIdChanges,
}))

const settings: Settings = {
  backendBaseUrl: 'http://127.0.0.1:8080',
  sourceLanguage: 'en' as const,
  targetLanguage: 'zh' as const,
  subtitleFontSizePx: 20,
  subtitleMode: 'translated' as const,
}

const asset: SubtitleAssetCacheEntry = {
  jobId: 'job_123',
  videoId: 'video_123',
  sourceLanguage: 'en',
  targetLanguage: 'zh',
  files: {
    source: '/subtitle-files/job_123/source',
    translated: '/subtitle-files/job_123/translated',
    bilingual: '/subtitle-files/job_123/bilingual',
  },
  createdAt: '2026-04-26T00:00:00Z',
  lastSyncedAt: '2026-04-26T00:00:00Z',
  selectedMode: 'translated' as const,
}

const validVtt = `WEBVTT

00:00:00.000 --> 00:00:01.000
hello
`

let wrapper: VueWrapper | null = null

function mockInitialLoad(options: {
  settingsOverride?: Partial<Settings>
  assetOverride?: Partial<SubtitleAssetCacheEntry>
  vtt?: string
} = {}) {
  sendExtensionMessage
    .mockResolvedValueOnce({ ok: true, data: { ...settings, ...options.settingsOverride } })
    .mockResolvedValueOnce({ ok: true, data: { ...asset, ...options.assetOverride } })
    .mockResolvedValueOnce({ ok: true, data: options.vtt ?? validVtt })
}

function getSentMessages() {
  return sendExtensionMessage.mock.calls.map(([message]) => message)
}

function getMessagesByType(type: string) {
  return getSentMessages().filter((message) => message?.type === type)
}

function sendSettingsUpdated(nextSettings: unknown) {
  const listener = addRuntimeListener.mock.calls[0]?.[0]
  if (!listener) {
    throw new Error('runtime listener not registered')
  }
  listener({ type: 'lets-sub-it:settings-updated', settings: nextSettings })
}

async function mountLoadedOverlay() {
  mockInitialLoad()
  wrapper = mount(YoutubeOverlay)
  await flushPromises()
  return wrapper
}

function mountOverlay() {
  wrapper = mount(YoutubeOverlay)
  return wrapper
}

function sendVideoIdChange(nextVideoId: string | null) {
  const listener = watchVideoIdChanges.mock.calls[0]?.[0]
  if (!listener) {
    throw new Error('video id listener not registered')
  }
  listener(nextVideoId)
}

function expectOldControlsHidden(wrapper: VueWrapper) {
  expect(wrapper.text()).not.toContain('字幕开')
  expect(wrapper.text()).not.toContain('翻译')
  expect(wrapper.text()).not.toContain('双语')
  expect(wrapper.text()).not.toContain('字幕已加载')
}

describe('YoutubeOverlay', () => {
  beforeEach(() => {
    sendExtensionMessage.mockReset()
    getCurrentVideoId.mockReset()
    watchVideoIdChanges.mockReset()
    addRuntimeListener.mockReset()
    removeRuntimeListener.mockReset()

    getCurrentVideoId.mockReturnValue('video_123')
    watchVideoIdChanges.mockReturnValue(() => {})

    document.body.innerHTML = ''
    document.body.appendChild(document.createElement('video'))
  })

  afterEach(() => {
    wrapper?.unmount()
    wrapper = null
  })

  it('renders only subtitle text without the old floating controls or status', async () => {
    const wrapper = await mountLoadedOverlay()

    expect(wrapper.text()).toContain('hello')
    expectOldControlsHidden(wrapper)
  })

  it('applies updated settings and reloads subtitles in the updated mode', async () => {
    const wrapper = await mountLoadedOverlay()
    sendExtensionMessage
      .mockResolvedValueOnce({
        ok: true,
        data: { ...asset, selectedMode: 'bilingual' as const },
      })
      .mockResolvedValueOnce({ ok: true, data: validVtt })

    sendSettingsUpdated({ ...settings, subtitleFontSizePx: 32, subtitleMode: 'bilingual' })
    await flushPromises()

    expect(wrapper.find('.lets-sub-it-subtitle-text').attributes('style')).toContain(
      'font-size: 32px',
    )
    expect(getMessagesByType('subtitle:update-mode')).toEqual([
      {
        type: 'subtitle:update-mode',
        payload: {
          videoId: 'video_123',
          targetLanguage: 'zh',
          mode: 'bilingual',
        },
      },
    ])
    expect(getMessagesByType('subtitle:fetch-file').at(-1)).toEqual({
      type: 'subtitle:fetch-file',
      payload: { jobId: 'job_123', mode: 'bilingual' },
    })
    expectOldControlsHidden(wrapper)
  })

  it('uses the initial settings mode for the first VTT request', async () => {
    mockInitialLoad({ settingsOverride: { subtitleMode: 'bilingual' } })

    mountOverlay()
    await flushPromises()

    expect(getMessagesByType('subtitle:fetch-file')[0]).toEqual({
      type: 'subtitle:fetch-file',
      payload: { jobId: 'job_123', mode: 'bilingual' },
    })
  })

  it('toggles subtitle visibility from the YouTube player window event', async () => {
    const wrapper = await mountLoadedOverlay()

    window.dispatchEvent(new CustomEvent('lets-sub-it:toggle-subtitles'))
    await flushPromises()
    expect(wrapper.text()).not.toContain('hello')

    window.dispatchEvent(new CustomEvent('lets-sub-it:toggle-subtitles'))
    await flushPromises()
    expect(wrapper.text()).toContain('hello')
  })

  it('ignores settings update messages with non-object settings payloads', async () => {
    const wrapper = await mountLoadedOverlay()

    sendSettingsUpdated('bilingual')
    await flushPromises()

    expect(getMessagesByType('subtitle:update-mode')).toHaveLength(0)
    expect(wrapper.find('.lets-sub-it-subtitle-text').attributes('style')).toContain(
      'font-size: 20px',
    )
  })

  it('ignores settings update messages with non-finite subtitle font sizes', async () => {
    const wrapper = await mountLoadedOverlay()

    sendSettingsUpdated({
      ...settings,
      subtitleFontSizePx: Infinity,
      subtitleMode: 'bilingual',
    })
    await flushPromises()

    expect(getMessagesByType('subtitle:update-mode')).toHaveLength(0)
    expect(wrapper.find('.lets-sub-it-subtitle-text').attributes('style')).toContain(
      'font-size: 20px',
    )
    expect(wrapper.find('.lets-sub-it-subtitle-text').attributes('style')).not.toContain(
      'Infinitypx',
    )
  })

  it('retries the same settings mode after a failed mode change rollback', async () => {
    const wrapper = await mountLoadedOverlay()
    sendExtensionMessage
      .mockResolvedValueOnce({
        ok: true,
        data: { ...asset, selectedMode: 'bilingual' as const },
      })
      .mockResolvedValueOnce({
        ok: false,
        error: {
          code: 'subtitle_file_missing',
          message: '字幕文件不存在',
        },
      })
      .mockResolvedValueOnce({
        ok: true,
        data: { ...asset, selectedMode: 'translated' as const },
      })
      .mockResolvedValueOnce({ ok: true, data: validVtt })
      .mockResolvedValueOnce({
        ok: true,
        data: { ...asset, selectedMode: 'bilingual' as const },
      })
      .mockResolvedValueOnce({ ok: true, data: validVtt })

    sendSettingsUpdated({ ...settings, subtitleMode: 'bilingual' })
    await flushPromises()
    sendSettingsUpdated({ ...settings, subtitleMode: 'bilingual' })
    await flushPromises()

    expect(
      getMessagesByType('subtitle:update-mode').filter(
        (message) => message?.payload?.mode === 'bilingual',
      ),
    ).toHaveLength(2)
    expect(
      getMessagesByType('subtitle:fetch-file').filter(
        (message) => message?.payload?.mode === 'bilingual',
      ),
    ).toHaveLength(2)
    expect(wrapper.text()).toContain('hello')
  })

  it('keeps the latest settings mode when an in-flight mode change is aborted by video navigation', async () => {
    const initialWrapper = await mountLoadedOverlay()
    let resolveModeUpdate: (value: unknown) => void = () => {}
    sendExtensionMessage.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveModeUpdate = resolve
      }),
    )
    sendExtensionMessage
      .mockResolvedValueOnce({
        ok: true,
        data: { ...asset, jobId: 'job_456', videoId: 'video_456' },
      })
      .mockResolvedValueOnce({ ok: true, data: validVtt })
      .mockResolvedValueOnce({
        ok: true,
        data: { ...asset, jobId: 'job_789', videoId: 'video_789' },
      })
      .mockResolvedValueOnce({ ok: true, data: validVtt })

    sendSettingsUpdated({ ...settings, subtitleMode: 'bilingual' })
    sendVideoIdChange('video_456')
    await flushPromises()
    resolveModeUpdate({
      ok: true,
      data: { ...asset, selectedMode: 'bilingual' as const },
    })
    await flushPromises()
    sendVideoIdChange('video_789')
    await flushPromises()

    expect(initialWrapper.text()).toContain('hello')
    expect(getMessagesByType('subtitle:fetch-file')).toEqual(
      expect.arrayContaining([
        {
          type: 'subtitle:fetch-file',
          payload: { jobId: 'job_456', mode: 'bilingual' },
        },
        {
          type: 'subtitle:fetch-file',
          payload: { jobId: 'job_789', mode: 'bilingual' },
        },
      ]),
    )
  })

  it('removes the window subtitle toggle listener on unmount', async () => {
    const removeSpy = vi.spyOn(window, 'removeEventListener')
    const mounted = await mountLoadedOverlay()

    mounted.unmount()
    wrapper = null

    expect(removeSpy).toHaveBeenCalledWith(
      'lets-sub-it:toggle-subtitles',
      expect.any(Function),
    )
    removeSpy.mockRestore()
  })

  it('does not register listeners after unmounting during settings load', async () => {
    let resolveSettings: (value: unknown) => void = () => {}
    sendExtensionMessage.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveSettings = resolve
      }),
    )

    const mounted = mountOverlay()
    mounted.unmount()
    wrapper = null
    resolveSettings({ ok: true, data: settings })
    await flushPromises()

    expect(getCurrentVideoId).not.toHaveBeenCalled()
    expect(watchVideoIdChanges).not.toHaveBeenCalled()
    expect(addRuntimeListener).not.toHaveBeenCalled()
  })

  it('writes previous mode back to storage and reloads previous subtitles when rollback succeeds', async () => {
    const wrapper = await mountLoadedOverlay()
    sendExtensionMessage
      .mockResolvedValueOnce({
        ok: true,
        data: { ...asset, selectedMode: 'bilingual' as const },
      })
      .mockResolvedValueOnce({
        ok: false,
        error: {
          code: 'subtitle_file_missing',
          message: '字幕文件不存在',
        },
      })
      .mockResolvedValueOnce({
        ok: true,
        data: { ...asset, selectedMode: 'translated' as const },
      })
      .mockResolvedValueOnce({ ok: true, data: validVtt })

    sendSettingsUpdated({ ...settings, subtitleMode: 'bilingual' })
    await flushPromises()

    expect(getMessagesByType('subtitle:update-mode')).toEqual(
      expect.arrayContaining([
        {
          type: 'subtitle:update-mode',
          payload: {
            videoId: 'video_123',
            targetLanguage: 'zh',
            mode: 'bilingual',
          },
        },
        {
          type: 'subtitle:update-mode',
          payload: {
            videoId: 'video_123',
            targetLanguage: 'zh',
            mode: 'translated',
          },
        },
      ]),
    )
    expect(
      getMessagesByType('subtitle:fetch-file').filter(
        (message) => message?.payload?.mode === 'translated',
      ),
    ).toHaveLength(2)
    expect(wrapper.text()).toContain('hello')
    expectOldControlsHidden(wrapper)
  })

  it('stops rollback reload when persistence fails', async () => {
    const wrapper = await mountLoadedOverlay()
    sendExtensionMessage
      .mockResolvedValueOnce({
        ok: true,
        data: { ...asset, selectedMode: 'bilingual' as const },
      })
      .mockResolvedValueOnce({
        ok: false,
        error: {
          code: 'subtitle_file_missing',
          message: '字幕文件不存在',
        },
      })
      .mockResolvedValueOnce({
        ok: false,
        error: {
          code: 'rollback_failed',
          message: '回滚模式失败',
        },
      })

    sendSettingsUpdated({ ...settings, subtitleMode: 'bilingual' })
    await flushPromises()

    expect(getMessagesByType('subtitle:update-mode')).toEqual(
      expect.arrayContaining([
        {
          type: 'subtitle:update-mode',
          payload: {
            videoId: 'video_123',
            targetLanguage: 'zh',
            mode: 'translated',
          },
        },
      ]),
    )
    expect(
      getMessagesByType('subtitle:fetch-file').filter(
        (message) => message?.payload?.mode === 'translated',
      ),
    ).toHaveLength(1)
    expect(wrapper.text()).toContain('hello')
    expect(wrapper.text()).not.toContain('回滚模式失败')
    expectOldControlsHidden(wrapper)
  })

  it('treats rollback success without data as a failure and stops reloading subtitles', async () => {
    const wrapper = await mountLoadedOverlay()
    sendExtensionMessage
      .mockResolvedValueOnce({
        ok: true,
        data: { ...asset, selectedMode: 'bilingual' as const },
      })
      .mockResolvedValueOnce({
        ok: false,
        error: {
          code: 'subtitle_file_missing',
          message: '字幕文件不存在',
        },
      })
      .mockResolvedValueOnce({
        ok: true,
        data: null,
      })

    sendSettingsUpdated({ ...settings, subtitleMode: 'bilingual' })
    await flushPromises()

    expect(getMessagesByType('subtitle:update-mode')).toEqual(
      expect.arrayContaining([
        {
          type: 'subtitle:update-mode',
          payload: {
            videoId: 'video_123',
            targetLanguage: 'zh',
            mode: 'translated',
          },
        },
      ]),
    )
    expect(
      getMessagesByType('subtitle:fetch-file').filter(
        (message) => message?.payload?.mode === 'translated',
      ),
    ).toHaveLength(1)
    expect(wrapper.text()).toContain('hello')
    expectOldControlsHidden(wrapper)
  })

  it('restores previous subtitles when the first mode update returns an error', async () => {
    const wrapper = await mountLoadedOverlay()
    sendExtensionMessage.mockResolvedValueOnce({
      ok: false,
      error: {
        code: 'update_mode_failed',
        message: '切换模式失败',
      },
    })

    sendSettingsUpdated({ ...settings, subtitleMode: 'bilingual' })
    await flushPromises()

    expect(getMessagesByType('subtitle:update-mode')).toEqual(
      expect.arrayContaining([
        {
          type: 'subtitle:update-mode',
          payload: {
            videoId: 'video_123',
            targetLanguage: 'zh',
            mode: 'bilingual',
          },
        },
      ]),
    )
    expect(
      getMessagesByType('subtitle:fetch-file').filter(
        (message) => message?.payload?.mode === 'translated',
      ),
    ).toHaveLength(1)
    expect(wrapper.text()).toContain('hello')
    expect(wrapper.text()).not.toContain('切换模式失败')
    expectOldControlsHidden(wrapper)
  })

  it('restores previous subtitles when the first mode update throws', async () => {
    const wrapper = await mountLoadedOverlay()
    sendExtensionMessage.mockRejectedValueOnce(new Error('切换模式异常'))

    sendSettingsUpdated({ ...settings, subtitleMode: 'bilingual' })
    await flushPromises()

    expect(getMessagesByType('subtitle:update-mode')).toEqual(
      expect.arrayContaining([
        {
          type: 'subtitle:update-mode',
          payload: {
            videoId: 'video_123',
            targetLanguage: 'zh',
            mode: 'bilingual',
          },
        },
      ]),
    )
    expect(
      getMessagesByType('subtitle:fetch-file').filter(
        (message) => message?.payload?.mode === 'translated',
      ),
    ).toHaveLength(1)
    expect(wrapper.text()).toContain('hello')
    expect(wrapper.text()).not.toContain('切换模式异常')
    expectOldControlsHidden(wrapper)
  })

  it('treats the first mode update success without data as a failure', async () => {
    const wrapper = await mountLoadedOverlay()
    sendExtensionMessage.mockResolvedValueOnce({
      ok: true,
      data: null,
    })

    sendSettingsUpdated({ ...settings, subtitleMode: 'bilingual' })
    await flushPromises()

    expect(getMessagesByType('subtitle:update-mode')).toEqual(
      expect.arrayContaining([
        {
          type: 'subtitle:update-mode',
          payload: {
            videoId: 'video_123',
            targetLanguage: 'zh',
            mode: 'bilingual',
          },
        },
      ]),
    )
    expect(
      getMessagesByType('subtitle:fetch-file').filter(
        (message) => message?.payload?.mode === 'translated',
      ),
    ).toHaveLength(1)
    expect(
      getMessagesByType('subtitle:fetch-file').filter(
        (message) => message?.payload?.mode === 'bilingual',
      ),
    ).toHaveLength(0)
    expect(wrapper.text()).toContain('hello')
    expect(wrapper.text()).not.toContain('字幕模式切换失败')
    expectOldControlsHidden(wrapper)
  })
})
