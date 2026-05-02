import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App.vue'

const { sendExtensionMessage, queryTabs } = vi.hoisted(() => ({
  sendExtensionMessage: vi.fn(),
  queryTabs: vi.fn(),
}))

vi.mock('wxt/browser', () => ({
  browser: {
    tabs: {
      query: queryTabs,
      sendMessage: vi.fn(),
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
    template: '<input v-bind="$attrs" />',
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

describe('popup App', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    sendExtensionMessage.mockReset()
    queryTabs.mockReset()
    queryTabs.mockResolvedValue([
      { url: 'https://www.youtube.com/watch?v=video_123' },
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
  })
})
