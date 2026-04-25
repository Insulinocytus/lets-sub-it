/**
 * @vitest-environment jsdom
 */
import { mount, flushPromises } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import JobForm from '../components/JobForm.vue'

function mountJobForm(submitJob: ReturnType<typeof vi.fn>) {
  return mount(JobForm, {
    props: {
      submitJob,
    },
    global: {
      stubs: {
        Button: {
          template: '<button type="button" :disabled="$attrs.disabled" @click="$emit(\'click\', $event)"><slot /></button>',
        },
      },
    },
  })
}

async function fillValidUrl(wrapper: ReturnType<typeof mount>) {
  await wrapper.get('input').setValue('https://www.youtube.com/watch?v=dQw4w9WgXcQ')
}

describe('JobForm', () => {
  it('shows the submit error message and restores loading state when submit fails', async () => {
    const submitJob = vi.fn().mockRejectedValue(new Error('backend 不可达'))
    const wrapper = mountJobForm(submitJob)

    await fillValidUrl(wrapper)
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(submitJob).toHaveBeenCalledWith({
      youtubeUrl: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
      sourceLanguage: 'ja',
      targetLanguage: 'zh-CN',
    })
    expect(wrapper.text()).toContain('backend 不可达')
    expect(wrapper.get('button').text()).toBe('生成字幕')
    expect(wrapper.get('button').attributes('disabled')).toBeUndefined()
  })

  it('calls submit logic and restores loading state when submit succeeds', async () => {
    const submitJob = vi.fn().mockResolvedValue(undefined)
    const wrapper = mountJobForm(submitJob)

    await fillValidUrl(wrapper)
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(submitJob).toHaveBeenCalledOnce()
    expect(wrapper.text()).not.toContain('无法连接服务器，请检查服务是否启动')
    expect(wrapper.get('button').text()).toBe('生成字幕')
    expect(wrapper.get('button').attributes('disabled')).toBeUndefined()
  })
})
