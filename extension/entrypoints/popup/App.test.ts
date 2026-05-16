import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App.vue'

const { sendExtensionMessage, queryTabs, sendTabMessage } = vi.hoisted(() => ({
  sendExtensionMessage: vi.fn(),
  queryTabs: vi.fn(),
  sendTabMessage: vi.fn(),
}))

vi.mock('wxt/browser', () => ({
  browser: {
    tabs: {
      query: queryTabs,
      sendMessage: sendTabMessage,
    },
  },
}))

vi.mock('@/api/messages', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/api/messages')>()),
  sendExtensionMessage,
}))

vi.mock('@/components/ui/alert', () => ({
  Alert: {
    name: 'Alert',
    template: '<div><slot /></div>',
  },
  AlertDescription: {
    name: 'AlertDescription',
    template: '<div><slot /></div>',
  },
}))

vi.mock('@/components/ui/badge', () => ({
  Badge: {
    name: 'Badge',
    template: '<span><slot /></span>',
  },
}))

vi.mock('@/components/ui/button', () => ({
  Button: {
    name: 'Button',
    template: '<button v-bind="$attrs"><slot /></button>',
  },
}))

vi.mock('@/components/ui/card', () => ({
  Card: {
    name: 'Card',
    template: '<section><slot /></section>',
  },
  CardContent: {
    name: 'CardContent',
    template: '<div><slot /></div>',
  },
  CardDescription: {
    name: 'CardDescription',
    template: '<p><slot /></p>',
  },
  CardHeader: {
    name: 'CardHeader',
    template: '<header><slot /></header>',
  },
  CardTitle: {
    name: 'CardTitle',
    template: '<h1><slot /></h1>',
  },
}))

vi.mock('@/components/ui/input', () => ({
  Input: {
    name: 'Input',
    props: ['modelValue'],
    emits: ['update:modelValue'],
    template: '<input :value="modelValue" v-bind="$attrs" @input="$emit(\'update:modelValue\', $event.target.value)" />',
  },
}))

vi.mock('@/components/ui/select', () => ({
  Select: {
    name: 'Select',
    template: '<div><slot /></div>',
  },
  SelectContent: {
    name: 'SelectContent',
    template: '<div><slot /></div>',
  },
  SelectItem: {
    name: 'SelectItem',
    template: '<div><slot /></div>',
  },
  SelectTrigger: {
    name: 'SelectTrigger',
    template: '<button type="button"><slot /></button>',
  },
  SelectValue: {
    name: 'SelectValue',
    template: '<span />',
  },
}))

const settings = {
  backendBaseUrl: 'http://127.0.0.1:8080',
  sourceLanguage: 'en',
  targetLanguage: 'zh',
  subtitleFontSizePx: 20,
  subtitleMode: 'translated' as const,
}

const activeJob = {
  id: 'job_123',
  videoId: 'video_123',
  youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
  sourceLanguage: 'en',
  targetLanguage: 'zh',
  status: 'transcribing',
  stage: 'transcribing',
  progressText: '转写中',
  errorMessage: null,
  createdAt: '2026-04-25T00:00:00Z',
  updatedAt: '2026-04-25T00:01:00Z',
}

async function selectSettingsTab(wrapper: ReturnType<typeof mount<typeof App>>) {
  const settingsTab = wrapper.get('button[data-testid="subtitle-settings-tab"]')
  await settingsTab.trigger('mousedown', { button: 0, ctrlKey: false })
  await settingsTab.trigger('click')
}

describe('popup App', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    sendExtensionMessage.mockReset()
    queryTabs.mockReset()
    sendTabMessage.mockReset()
    queryTabs.mockResolvedValue([
      { id: 7, url: 'https://www.youtube.com/watch?v=video_123' },
    ])
    sendExtensionMessage.mockImplementation(async (message) => {
      if (message.type === 'settings:get') {
        return { ok: true, data: settings }
      }
      if (message.type === 'job:active') {
        return { ok: true, data: { job: activeJob } }
      }
      if (message.type === 'job:get') {
        return { ok: true, data: { job: activeJob } }
      }
      if (message.type === 'settings:update') {
        return { ok: true, data: { ...settings, ...message.payload } }
      }
      return { ok: true, data: null }
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('restores the active job for the current YouTube tab when opened', async () => {
    const wrapper = mount(App)

    await flushPromises()

    expect(sendExtensionMessage).toHaveBeenCalledWith({
      type: 'job:active',
      payload: {
        videoId: 'video_123',
        targetLanguage: 'zh',
      },
    })
    expect(wrapper.text()).toContain('转写中')
    expect(wrapper.text()).toContain('任务状态')
    expect(wrapper.text()).not.toContain('转写已用')
  })

  it('shows generation and subtitle settings tabs', async () => {
    const wrapper = mount(App)

    await flushPromises()

    expect(wrapper.text()).toContain('生成字幕')
    expect(wrapper.text()).toContain('字幕设置')
    expect(wrapper.text()).toContain('backend URL')

    await selectSettingsTab(wrapper)

    expect(wrapper.text()).toContain('字体大小')
    expect(wrapper.text()).toContain('显示模式')
    expect(wrapper.text()).toContain('翻译 only')
    expect(wrapper.text()).toContain('双语')
  })

  it('uses tabs semantics for popup views', async () => {
    const wrapper = mount(App)

    await flushPromises()

    expect(wrapper.find('[data-testid="popup-tabs"]').exists()).toBe(true)

    const viewGroup = wrapper.get('[aria-label="功能切换"]')
    const generateTab = wrapper.get('button[data-testid="generate-tab"]')
    const settingsTab = wrapper.get('button[data-testid="subtitle-settings-tab"]')

    expect(viewGroup.attributes('role')).toBe('tablist')
    expect(generateTab.attributes('role')).toBe('tab')
    expect(settingsTab.attributes('role')).toBe('tab')

    await settingsTab.trigger('mousedown', { button: 0, ctrlKey: false })
    await settingsTab.trigger('click')

    expect(wrapper.text()).toContain('字体大小')
  })

  it('exposes the pressed state for subtitle display mode buttons', async () => {
    const wrapper = mount(App)

    await flushPromises()
    await selectSettingsTab(wrapper)

    const modeGroup = wrapper.get('[aria-label="字幕显示模式"]')
    const translatedMode = wrapper.get('button[data-testid="subtitle-mode-translated"]')
    const bilingualMode = wrapper.get('button[data-testid="subtitle-mode-bilingual"]')

    expect(modeGroup.attributes('role')).toBe('group')
    expect(translatedMode.attributes('aria-pressed')).toBe('true')
    expect(bilingualMode.attributes('aria-pressed')).toBe('false')

    await bilingualMode.trigger('click')

    expect(translatedMode.attributes('aria-pressed')).toBe('false')
    expect(bilingualMode.attributes('aria-pressed')).toBe('true')
  })

  it('saves subtitle settings and notifies the current YouTube tab', async () => {
    const wrapper = mount(App)

    await flushPromises()
    await selectSettingsTab(wrapper)
    await wrapper.get('input[data-testid="subtitle-font-size-input"]').setValue('32')
    await wrapper.get('button[data-testid="subtitle-mode-bilingual"]').trigger('click')
    await wrapper.get('button[data-testid="save-subtitle-settings"]').trigger('click')
    await flushPromises()

    expect(sendExtensionMessage).toHaveBeenCalledWith({
      type: 'settings:update',
      payload: { subtitleFontSizePx: 32, subtitleMode: 'bilingual' },
    })
    expect(sendTabMessage).toHaveBeenCalledWith(7, {
      type: 'lets-sub-it:settings-updated',
      settings: {
        ...settings,
        subtitleFontSizePx: 32,
        subtitleMode: 'bilingual',
      },
    })
  })

  it('rejects invalid subtitle font size input before saving', async () => {
    const wrapper = mount(App)

    await flushPromises()
    await selectSettingsTab(wrapper)
    await wrapper.get('input[data-testid="subtitle-font-size-input"]').setValue('0')
    await wrapper.get('button[data-testid="save-subtitle-settings"]').trigger('click')

    expect(wrapper.text()).toContain('字幕字体大小必须是正数')
    expect(sendExtensionMessage).not.toHaveBeenCalledWith({
      type: 'settings:update',
      payload: expect.objectContaining({ subtitleFontSizePx: 0 }),
    })
  })
})
