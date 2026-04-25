import type { LanguageCode } from '@/api/messages'

export type CreateJobForm = {
  backendBaseUrl: string
  youtubeUrl: string
  sourceLanguage: LanguageCode
  targetLanguage: LanguageCode
}

export function validateCreateJobForm(form: CreateJobForm): string | null {
  if (!form.backendBaseUrl.trim()) {
    return '请输入 backend URL'
  }
  if (!form.youtubeUrl.trim()) {
    return '请输入 YouTube URL'
  }
  if (form.sourceLanguage === form.targetLanguage) {
    return '源语言和目标语言不能相同'
  }
  return null
}
