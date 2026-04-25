/**
 * @vitest-environment jsdom
 */
import { mount, flushPromises } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import JobResult from '../components/JobResult.vue'
import type { JobResponse } from '@/types'

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

function mountJobResult(updateMode = vi.fn().mockResolvedValue(undefined)) {
  return mount(JobResult, {
    props: {
      job,
      failedJob: null,
      updateMode,
    },
    global: {
      stubs: {
        Button: {
          props: ['variant'],
          template: '<button type="button" :data-variant="variant" :disabled="$attrs.disabled" @click="$emit(\'click\', $event)"><slot /></button>',
        },
      },
    },
  })
}

function deferred<T = void>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise
  })
  return { promise, resolve }
}

describe('JobResult', () => {
  it('rolls back the selected mode and shows an error when mode persistence fails', async () => {
    const updateMode = vi.fn().mockRejectedValue(new Error('storage 写入失败'))
    const wrapper = mountJobResult(updateMode)

    await wrapper.findAll('button').find((button) => button.text() === '仅翻译')!.trigger('click')
    await flushPromises()

    const buttons = wrapper.findAll('button')
    expect(updateMode).toHaveBeenCalledWith('translated')
    expect(buttons.find((button) => button.text() === '仅翻译')!.attributes('data-variant')).toBe('outline')
    expect(buttons.find((button) => button.text() === '双语')!.attributes('data-variant')).toBe('default')
    expect(wrapper.text()).toContain('storage 写入失败')
  })

  it('does not start another mode update while a previous update is pending', async () => {
    const pendingUpdate = deferred()
    const updateMode = vi.fn().mockReturnValue(pendingUpdate.promise)
    const wrapper = mountJobResult(updateMode)

    const translatedButton = wrapper.findAll('button').find((button) => button.text() === '仅翻译')!
    const bilingualButton = wrapper.findAll('button').find((button) => button.text() === '双语')!

    await translatedButton.trigger('click')
    await bilingualButton.trigger('click')

    expect(updateMode).toHaveBeenCalledOnce()
    expect(updateMode).toHaveBeenCalledWith('translated')
    expect(translatedButton.attributes('disabled')).toBeDefined()
    expect(bilingualButton.attributes('disabled')).toBeDefined()

    pendingUpdate.resolve()
    await flushPromises()
  })
})
