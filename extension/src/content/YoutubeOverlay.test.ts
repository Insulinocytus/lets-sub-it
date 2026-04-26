import { flushPromises, mount } from '@vue/test-utils'
import type { VueWrapper } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
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
  sendExtensionMessage,
}))

vi.mock('@/youtube/page-watch', () => ({
  getCurrentVideoId,
  watchVideoIdChanges,
}))

vi.mock('@/components/ui/button', () => ({
  Button: {
    name: 'Button',
    template: '<button v-bind="$attrs" @click="$emit(\'click\')"><slot /></button>',
  },
}))

vi.mock('@/components/ui/badge', () => ({
  Badge: {
    name: 'Badge',
    template: '<div v-bind="$attrs"><slot /></div>',
  },
}))

const asset = {
  jobId: 'job_123',
  videoId: 'video_123',
  sourceLanguage: 'en',
  targetLanguage: 'zh-CN',
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

function getButtonByText(wrapper: VueWrapper, label: string) {
  const button = wrapper
    .findAll('button')
    .find((candidate) => candidate.text() === label)

  if (!button) {
    throw new Error(`button not found: ${label}`)
  }

  return button
}

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

    getCurrentVideoId.mockReturnValue('video_123')
    watchVideoIdChanges.mockReturnValue(() => {})

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

    await getButtonByText(wrapper, '双语').trigger('click')
    await flushPromises()

    expect(getMessagesByType('subtitle:update-mode')).toEqual(
      expect.arrayContaining([
        {
          type: 'subtitle:update-mode',
          payload: {
            videoId: 'video_123',
            targetLanguage: 'zh-CN',
            mode: 'bilingual',
          },
        },
        {
          type: 'subtitle:update-mode',
          payload: {
            videoId: 'video_123',
            targetLanguage: 'zh-CN',
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
    expect(wrapper.text()).toContain('字幕已加载')
    expect(getButtonByText(wrapper, '翻译').attributes('variant')).toBe('secondary')
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

    await getButtonByText(wrapper, '双语').trigger('click')
    await flushPromises()

    expect(getMessagesByType('subtitle:update-mode')).toEqual(
      expect.arrayContaining([
        {
          type: 'subtitle:update-mode',
          payload: {
            videoId: 'video_123',
            targetLanguage: 'zh-CN',
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
    expect(wrapper.text()).toContain('回滚模式失败')
    expect(wrapper.text()).not.toContain('字幕已加载')
    expect(getButtonByText(wrapper, '翻译').attributes('variant')).toBe('secondary')
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

    await getButtonByText(wrapper, '双语').trigger('click')
    await flushPromises()

    expect(getMessagesByType('subtitle:update-mode')).toEqual(
      expect.arrayContaining([
        {
          type: 'subtitle:update-mode',
          payload: {
            videoId: 'video_123',
            targetLanguage: 'zh-CN',
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
    expect(wrapper.text()).toContain('字幕模式回滚失败')
    expect(wrapper.text()).not.toContain('字幕已加载')
    expect(getButtonByText(wrapper, '翻译').attributes('variant')).toBe('secondary')
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

    await getButtonByText(wrapper, '双语').trigger('click')
    await flushPromises()

    expect(getMessagesByType('subtitle:update-mode')).toEqual(
      expect.arrayContaining([
        {
          type: 'subtitle:update-mode',
          payload: {
            videoId: 'video_123',
            targetLanguage: 'zh-CN',
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
    expect(wrapper.text()).toContain('切换模式失败')
    expect(wrapper.text()).not.toContain('字幕已加载')
    expect(getButtonByText(wrapper, '翻译').attributes('variant')).toBe('secondary')
  })

  it('restores previous subtitles when the first mode update throws', async () => {
    sendExtensionMessage
      .mockResolvedValueOnce({ ok: true, data: asset })
      .mockResolvedValueOnce({ ok: true, data: validVtt })
      .mockRejectedValueOnce(new Error('切换模式异常'))

    const wrapper = mount(YoutubeOverlay)
    await flushPromises()

    await getButtonByText(wrapper, '双语').trigger('click')
    await flushPromises()

    expect(getMessagesByType('subtitle:update-mode')).toEqual(
      expect.arrayContaining([
        {
          type: 'subtitle:update-mode',
          payload: {
            videoId: 'video_123',
            targetLanguage: 'zh-CN',
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
    expect(wrapper.text()).toContain('切换模式异常')
    expect(wrapper.text()).not.toContain('字幕已加载')
    expect(getButtonByText(wrapper, '翻译').attributes('variant')).toBe('secondary')
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

    await getButtonByText(wrapper, '双语').trigger('click')
    await flushPromises()

    expect(getMessagesByType('subtitle:update-mode')).toEqual(
      expect.arrayContaining([
        {
          type: 'subtitle:update-mode',
          payload: {
            videoId: 'video_123',
            targetLanguage: 'zh-CN',
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
    expect(wrapper.text()).toContain('字幕模式切换失败')
    expect(wrapper.text()).not.toContain('字幕已加载')
    expect(getButtonByText(wrapper, '翻译').attributes('variant')).toBe('secondary')
  })
})
