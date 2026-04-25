import { describe, expect, it } from 'vitest'
import { validateCreateJobForm } from './form-validation'

describe('validateCreateJobForm', () => {
  it('accepts a valid form', () => {
    expect(
      validateCreateJobForm({
        backendBaseUrl: 'http://127.0.0.1:8080',
        youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
        sourceLanguage: 'en',
        targetLanguage: 'zh-CN',
      }),
    ).toBeNull()
  })

  it('rejects empty YouTube URL', () => {
    expect(
      validateCreateJobForm({
        backendBaseUrl: 'http://127.0.0.1:8080',
        youtubeUrl: '',
        sourceLanguage: 'en',
        targetLanguage: 'zh-CN',
      }),
    ).toBe('请输入 YouTube URL')
  })

  it('rejects empty backend URL', () => {
    expect(
      validateCreateJobForm({
        backendBaseUrl: '',
        youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
        sourceLanguage: 'en',
        targetLanguage: 'zh-CN',
      }),
    ).toBe('请输入 backend URL')
  })

  it('rejects equal source and target languages', () => {
    expect(
      validateCreateJobForm({
        backendBaseUrl: 'http://127.0.0.1:8080',
        youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
        sourceLanguage: 'en',
        targetLanguage: 'en',
      }),
    ).toBe('源语言和目标语言不能相同')
  })
})
