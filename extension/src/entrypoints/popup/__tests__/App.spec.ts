/**
 * @vitest-environment jsdom
 */
import { mount, flushPromises } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick, ref } from 'vue'
import App from '../App.vue'
import type { JobResponse, JobStatus, LocalCacheEntry, UserPreferences } from '@/types'

const job: JobResponse = {
  id: 'job_123',
  videoId: 'video_123',
  youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
  sourceLanguage: 'ja',
  targetLanguage: 'zh-Hans',
  status: 'completed',
  stage: 'completed',
  progressText: 'Done',
  errorMessage: null,
  createdAt: '2026-04-25T00:00:00Z',
  updatedAt: '2026-04-25T00:00:00Z',
}

const createJob = vi.fn()
const getCacheEntry = vi.fn()
const setCacheEntry = vi.fn()
const setPreferences = vi.fn()
const setSubtitleSelection = vi.fn()
const pollingStatus = ref<JobStatus | null>(null)
const pollingJob = ref<JobResponse | null>(null)
const pollingError = ref<string | null>(null)
const pollingStart = vi.fn()
const pollingReset = vi.fn()

function deferred<T = void>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise
  })
  return { promise, resolve }
}

function buttonStub() {
  return {
    template: '<button type="button" :disabled="$attrs.disabled" @click="$emit(\'click\', $event)"><slot /></button>',
  }
}

async function submitRealJobForm(wrapper: ReturnType<typeof mount>) {
  await wrapper.get('input').setValue('https://www.youtube.com/watch?v=dQw4w9WgXcQ')
  await wrapper.get('button').trigger('click')
  await flushPromises()
}

vi.mock('@/composables/useApi', () => ({
  useApi: () => ({
    createJob,
  }),
}))

vi.mock('@/composables/useJobPolling', () => ({
  useJobPolling: () => ({
    jobId: ref(null),
    status: pollingStatus,
    job: pollingJob,
    error: pollingError,
    isPolling: ref(false),
    start: pollingStart,
    stop: vi.fn(),
    reset: pollingReset,
  }),
}))

vi.mock('@/composables/useCache', () => ({
  useCache: () => ({
    getCacheEntry,
    setCacheEntry,
    setPreferences,
    setSubtitleSelection,
  }),
}))

beforeEach(() => {
  createJob.mockReset()
  getCacheEntry.mockReset()
  setCacheEntry.mockReset()
  setPreferences.mockReset()
  setSubtitleSelection.mockReset()
  pollingStatus.value = null
  pollingJob.value = null
  pollingError.value = null
  pollingStart.mockReset()
  pollingReset.mockReset()
})

describe('popup App', () => {
  it('shows the form error when creating a job fails through the real JobForm binding', async () => {
    createJob.mockRejectedValue(new Error('backend 不可达'))

    const wrapper = mount(App, {
      global: {
        stubs: {
          JobStatusView: true,
          Button: buttonStub(),
        },
      },
    })

    await submitRealJobForm(wrapper)

    expect(wrapper.text()).toContain('backend 不可达')
    expect(wrapper.get('button').text()).toBe('生成字幕')
  })

  it('does not retain a completed job when cache persistence fails before result view', async () => {
    createJob.mockResolvedValue({ reused: true, job })
    setCacheEntry.mockRejectedValue(new Error('storage 写入失败'))
    setPreferences.mockResolvedValue(undefined)

    const wrapper = mount(App, {
      global: {
        stubs: {
          JobStatusView: true,
          Button: buttonStub(),
        },
      },
    })

    await submitRealJobForm(wrapper)

    expect(wrapper.text()).toContain('storage 写入失败')
    expect(wrapper.text()).not.toContain('字幕已生成')
    expect((wrapper.vm as unknown as { completedJob: JobResponse | null }).completedJob).toBeNull()
  })

  it('shows a retry path when cache persistence fails after polling completes', async () => {
    createJob.mockResolvedValue({ reused: false, job: { ...job, status: 'queued', stage: 'queued' } })
    setCacheEntry.mockRejectedValue(new Error('storage 写入失败'))
    setPreferences.mockResolvedValue(undefined)

    const wrapper = mount(App, {
      global: {
        stubs: {
          Button: buttonStub(),
        },
      },
    })

    await submitRealJobForm(wrapper)
    pollingJob.value = job
    pollingStatus.value = 'completed'
    await nextTick()
    await flushPromises()

    expect(wrapper.text()).toContain('处理失败')
    expect(wrapper.text()).toContain('storage 写入失败')
    expect(wrapper.findAll('button').some((button) => button.text() === '重新提交')).toBe(true)
    expect(wrapper.text()).not.toContain('准备中')
  })

  it('persists subtitle mode changes from the job result view', async () => {
    const existingEntry: LocalCacheEntry = {
      videoId: job.videoId,
      targetLanguage: job.targetLanguage,
      jobId: job.id,
      selectedMode: 'bilingual',
      lastSyncedAt: '2026-04-25T00:00:00Z',
    }
    createJob.mockResolvedValue({ reused: true, job })
    getCacheEntry.mockResolvedValue(existingEntry)
    setCacheEntry.mockResolvedValue(undefined)
    setPreferences.mockResolvedValue(undefined)
    setSubtitleSelection.mockResolvedValue(undefined)

    const wrapper = mount(App, {
      global: {
        stubs: {
          JobForm: {
            props: ['submitJob'],
            template: '<button type="button" @click="submitJob(params)">submit</button>',
            data() {
              return {
                params: {
                  youtubeUrl: job.youtubeUrl,
                  sourceLanguage: job.sourceLanguage,
                  targetLanguage: job.targetLanguage,
                },
              }
            },
          },
          JobStatusView: true,
          Button: {
            template: '<button type="button" @click="$emit(\'click\', $event)"><slot /></button>',
          },
        },
      },
    })

    await wrapper.get('button').trigger('click')
    await flushPromises()
    await wrapper.findAll('button').find((button) => button.text() === '仅翻译')!.trigger('click')
    await flushPromises()

    expect(setSubtitleSelection).toHaveBeenLastCalledWith({
      ...existingEntry,
      selectedMode: 'translated',
    }, {
      videoId: job.videoId,
      targetLanguage: job.targetLanguage,
      selectedMode: 'translated',
    })
    expect(setCacheEntry).toHaveBeenCalledTimes(1)
    expect(setPreferences).toHaveBeenCalledTimes(1)
  })

  it('creates a matching cache entry when mode changes without an existing cache entry', async () => {
    createJob.mockResolvedValue({ reused: true, job })
    getCacheEntry.mockResolvedValue(null)
    setCacheEntry.mockResolvedValue(undefined)
    setPreferences.mockResolvedValue(undefined)
    setSubtitleSelection.mockResolvedValue(undefined)

    const wrapper = mount(App, {
      global: {
        stubs: {
          JobForm: {
            props: ['submitJob'],
            template: '<button type="button" @click="submitJob(params)">submit</button>',
            data() {
              return {
                params: {
                  youtubeUrl: job.youtubeUrl,
                  sourceLanguage: job.sourceLanguage,
                  targetLanguage: job.targetLanguage,
                },
              }
            },
          },
          JobStatusView: true,
          Button: {
            template: '<button type="button" @click="$emit(\'click\', $event)"><slot /></button>',
          },
        },
      },
    })

    await wrapper.get('button').trigger('click')
    await flushPromises()
    await wrapper.findAll('button').find((button) => button.text() === '仅翻译')!.trigger('click')
    await flushPromises()

    expect(setSubtitleSelection).toHaveBeenLastCalledWith({
      videoId: job.videoId,
      targetLanguage: job.targetLanguage,
      jobId: job.id,
      selectedMode: 'translated',
      lastSyncedAt: expect.any(String),
    }, {
      videoId: job.videoId,
      targetLanguage: job.targetLanguage,
      selectedMode: 'translated',
    })
    expect(setCacheEntry).toHaveBeenCalledTimes(1)
    expect(setPreferences).toHaveBeenCalledTimes(1)
  })

  it('keeps storage on the pending mode when a rapid second mode click is ignored', async () => {
    createJob.mockResolvedValue({ reused: true, job })
    getCacheEntry.mockResolvedValue(null)
    setCacheEntry.mockResolvedValue(undefined)
    setPreferences.mockResolvedValue(undefined)
    setSubtitleSelection.mockResolvedValue(undefined)

    const wrapper = mount(App, {
      global: {
        stubs: {
          JobForm: {
            props: ['submitJob'],
            template: '<button type="button" @click="submitJob(params)">submit</button>',
            data() {
              return {
                params: {
                  youtubeUrl: job.youtubeUrl,
                  sourceLanguage: job.sourceLanguage,
                  targetLanguage: job.targetLanguage,
                },
              }
            },
          },
          JobStatusView: true,
          Button: {
            template: '<button type="button" @click="$emit(\'click\', $event)"><slot /></button>',
          },
        },
      },
    })

    await wrapper.get('button').trigger('click')
    await flushPromises()

    const translatedCacheWrite = deferred()
    const translatedPrefsWrite = deferred()
    let storedCache: LocalCacheEntry | null = null
    let storedPrefs: UserPreferences | null = null

    setSubtitleSelection.mockImplementation((entry: LocalCacheEntry, prefs: UserPreferences) => {
      if (entry.selectedMode === 'translated') {
        return Promise.all([translatedCacheWrite.promise, translatedPrefsWrite.promise]).then(() => {
          storedCache = entry
          storedPrefs = prefs
        })
      }
      storedCache = entry
      storedPrefs = prefs
      return Promise.resolve()
    })

    const buttons = wrapper.findAll('button')
    await buttons.find((button) => button.text() === '仅翻译')!.trigger('click')
    await buttons.find((button) => button.text() === '双语')!.trigger('click')
    await flushPromises()

    translatedCacheWrite.resolve()
    translatedPrefsWrite.resolve()
    await flushPromises()

    expect(storedCache?.selectedMode).toBe('translated')
    expect(storedPrefs?.selectedMode).toBe('translated')
  })
})
