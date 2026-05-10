import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import YoutubeOverlay from './YoutubeOverlay.vue'

const {
  sendExtensionMessage,
  getCurrentVideoId,
  watchVideoIdChanges,
  addRuntimeListener,
  removeRuntimeListener,
  getSettings,
} = vi.hoisted(() => ({
  sendExtensionMessage: vi.fn(),
  getCurrentVideoId: vi.fn(),
  watchVideoIdChanges: vi.fn(),
  addRuntimeListener: vi.fn(),
  removeRuntimeListener: vi.fn(),
  getSettings: vi.fn(),
}))

vi.mock('@/storage/settings', () => ({
  getSettings,
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
  sendExtensionMessage,
}))

vi.mock('@/youtube/page-watch', () => ({
  getCurrentVideoId,
  watchVideoIdChanges,
}))

const asset = {
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

function getSentMessages() {
  return sendExtensionMessage.mock.calls.map(([message]) => message)
}

function getMessagesByType(type: string) {
  return getSentMessages().filter((message) => message?.type === type)
}

describe('YoutubeOverlay', () => {
  beforeEach(() => {
    sendExtensionMessage.mockReset()
    getCurrentVideoId.mockReset()
    watchVideoIdChanges.mockReset()
    addRuntimeListener.mockReset()
    removeRuntimeListener.mockReset()
    getSettings.mockReset()

    getCurrentVideoId.mockReturnValue('video_123')
    watchVideoIdChanges.mockReturnValue(() => {})
    getSettings.mockResolvedValue({ subtitleFontSize: 20 })

    document.body.innerHTML = ''
    document.body.appendChild(document.createElement('video'))
  })

  it('writes previous mode back to storage and reloads previous subtitles when rollback succeeds', async () => {
    sendExtensionMessage
      .mockResolvedValueOnce({ ok: true, data: asset })
      .mockResolvedValueOnce({ ok: true, data: validVtt })
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

    const wrapper = mount(YoutubeOverlay)
    await flushPromises()

    await (wrapper.vm as any).changeMode('bilingual')
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
    expect((wrapper.vm as any).status).toBe('字幕已加载')
  })

  it('stops and surfaces rollback persistence failures instead of pretending rollback succeeded', async () => {
    sendExtensionMessage
      .mockResolvedValueOnce({ ok: true, data: asset })
      .mockResolvedValueOnce({ ok: true, data: validVtt })
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

    const wrapper = mount(YoutubeOverlay)
    await flushPromises()

    await (wrapper.vm as any).changeMode('bilingual')
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
    expect((wrapper.vm as any).status).toBe('回滚模式失败')
    expect((wrapper.vm as any).status).not.toBe('字幕已加载')
  })

  it('treats rollback success without data as a failure and stops reloading subtitles', async () => {
    sendExtensionMessage
      .mockResolvedValueOnce({ ok: true, data: asset })
      .mockResolvedValueOnce({ ok: true, data: validVtt })
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

    const wrapper = mount(YoutubeOverlay)
    await flushPromises()

    await (wrapper.vm as any).changeMode('bilingual')
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
    expect((wrapper.vm as any).status).toBe('字幕模式回滚失败')
    expect((wrapper.vm as any).status).not.toBe('字幕已加载')
  })

  it('restores previous subtitles when the first mode update returns an error', async () => {
    sendExtensionMessage
      .mockResolvedValueOnce({ ok: true, data: asset })
      .mockResolvedValueOnce({ ok: true, data: validVtt })
      .mockResolvedValueOnce({
        ok: false,
        error: {
          code: 'update_mode_failed',
          message: '切换模式失败',
        },
      })

    const wrapper = mount(YoutubeOverlay)
    await flushPromises()

    await (wrapper.vm as any).changeMode('bilingual')
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
    expect((wrapper.vm as any).status).toBe('切换模式失败')
    expect((wrapper.vm as any).status).not.toBe('字幕已加载')
  })

  it('restores previous subtitles when the first mode update throws', async () => {
    sendExtensionMessage
      .mockResolvedValueOnce({ ok: true, data: asset })
      .mockResolvedValueOnce({ ok: true, data: validVtt })
      .mockRejectedValueOnce(new Error('切换模式异常'))

    const wrapper = mount(YoutubeOverlay)
    await flushPromises()

    await (wrapper.vm as any).changeMode('bilingual')
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
    expect((wrapper.vm as any).status).toBe('切换模式异常')
    expect((wrapper.vm as any).status).not.toBe('字幕已加载')
  })

  it('initializes fontSize from settings', async () => {
    getSettings.mockResolvedValue({ subtitleFontSize: 24 })
    sendExtensionMessage
      .mockResolvedValueOnce({ ok: true, data: asset })
      .mockResolvedValueOnce({ ok: true, data: validVtt })

    const wrapper = mount(YoutubeOverlay)
    await flushPromises()

    const subtitleDiv = wrapper.find('[style*="font-size"]')
    expect(subtitleDiv.attributes('style')).toContain('font-size: 24px')
  })

  it('updates fontSize when subtitle:settings-changed message arrives', async () => {
    getSettings.mockResolvedValue({ subtitleFontSize: 20 })
    sendExtensionMessage
      .mockResolvedValueOnce({ ok: true, data: asset })
      .mockResolvedValueOnce({ ok: true, data: validVtt })

    const wrapper = mount(YoutubeOverlay)
    await flushPromises()

    const messageHandler = addRuntimeListener.mock.calls[0]?.[0]
    if (messageHandler) {
      messageHandler({
        type: 'subtitle:settings-changed',
        payload: { fontSize: 36 },
      })
      await flushPromises()

      const subtitleDiv = wrapper.find('[style*="font-size"]')
      expect(subtitleDiv.attributes('style')).toContain('font-size: 36px')
    }
  })

  it('exposes toggleEnabled to toggle subtitle visibility', async () => {
    getSettings.mockResolvedValue({ subtitleFontSize: 20 })
    const assetWithEmptyFiles = {
      jobId: 'job_123',
      videoId: 'video_123',
      sourceLanguage: 'en',
      targetLanguage: 'zh',
      files: { source: '', translated: '', bilingual: '' },
      createdAt: '2026-04-26T00:00:00Z',
      lastSyncedAt: '2026-04-26T00:00:00Z',
      selectedMode: 'translated' as const,
    }

    sendExtensionMessage
      .mockResolvedValueOnce({ ok: true, data: assetWithEmptyFiles })
      .mockResolvedValueOnce({ ok: true, data: validVtt })

    const wrapper = mount(YoutubeOverlay)
    await flushPromises()

    expect(wrapper.find('[style*="font-size"]').exists()).toBe(true)
    expect((wrapper.vm as any).enabled).toBe(true)

    ;(wrapper.vm as any).toggleEnabled()
    await flushPromises()

    expect((wrapper.vm as any).enabled).toBe(false)
    expect(wrapper.find('[style*="font-size"]').exists()).toBe(false)
  })

  it('treats the first mode update success without data as a failure', async () => {
    sendExtensionMessage
      .mockResolvedValueOnce({ ok: true, data: asset })
      .mockResolvedValueOnce({ ok: true, data: validVtt })
      .mockResolvedValueOnce({
        ok: true,
        data: null,
      })

    const wrapper = mount(YoutubeOverlay)
    await flushPromises()

    await (wrapper.vm as any).changeMode('bilingual')
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
    expect((wrapper.vm as any).status).toBe('字幕模式切换失败')
    expect((wrapper.vm as any).status).not.toBe('字幕已加载')
  })
})
