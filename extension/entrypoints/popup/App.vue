<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { browser } from 'wxt/browser'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  sendExtensionMessage,
  type Job,
  type LanguageCode,
  type Settings,
  type SubtitleAsset,
} from '@/api/messages'
import {
  validateCreateJobForm,
  type CreateJobForm,
} from '@/popup/form-validation'

const languages = ['en', 'zh-CN'] satisfies LanguageCode[]
const languageLabels: Record<LanguageCode, string> = {
  en: 'English',
  'zh-CN': '简体中文',
}

const backendBaseUrl = ref('http://127.0.0.1:8080')
const youtubeUrl = ref('')
const sourceLanguage = ref<LanguageCode>('en')
const targetLanguage = ref<LanguageCode>('zh-CN')
const currentJob = ref<Job | null>(null)
const errorMessage = ref('')
const isSubmitting = ref(false)
const elapsedSeconds = ref(0)
const pollTimer = ref<number | null>(null)
const elapsedTimer = ref<number | null>(null)

const form = computed<CreateJobForm>(() => ({
  backendBaseUrl: backendBaseUrl.value,
  youtubeUrl: youtubeUrl.value,
  sourceLanguage: sourceLanguage.value,
  targetLanguage: targetLanguage.value,
}))

const formError = computed(() => validateCreateJobForm(form.value))
const alertMessage = computed(() => errorMessage.value || formError.value)
const isSubmitDisabled = computed(() => isSubmitting.value || formError.value !== null)

const statusLabel = computed(() => {
  if (!currentJob.value) {
    return ''
  }

  return {
    queued: '排队中',
    downloading: '下载中',
    transcribing: '转写中',
    translating: '翻译中',
    packaging: '打包中',
    completed: '已完成',
    failed: '失败',
  }[currentJob.value.status]
})

const statusBadgeVariant = computed(() => {
  if (currentJob.value?.status === 'failed') {
    return 'destructive'
  }
  if (currentJob.value?.status === 'completed') {
    return 'secondary'
  }
  return 'outline'
})

onMounted(async () => {
  const settingsResult = await sendExtensionMessage<Settings>({ type: 'settings:get' })
  if (settingsResult.ok) {
    backendBaseUrl.value = settingsResult.data.backendBaseUrl
    sourceLanguage.value = settingsResult.data.sourceLanguage
    targetLanguage.value = settingsResult.data.targetLanguage
  }

  const [tab] = await browser.tabs.query({ active: true, currentWindow: true })
  if (tab?.url?.startsWith('https://www.youtube.com/watch')) {
    youtubeUrl.value = tab.url
  }
})

onUnmounted(() => {
  clearTimers()
})

async function submitJob() {
  const validationError = validateCreateJobForm(form.value)
  if (validationError) {
    errorMessage.value = validationError
    return
  }

  isSubmitting.value = true
  errorMessage.value = ''
  currentJob.value = null
  elapsedSeconds.value = 0
  clearTimers()

  const settingsResult = await sendExtensionMessage<Settings>({
    type: 'settings:update',
    payload: {
      backendBaseUrl: backendBaseUrl.value.trim(),
      sourceLanguage: sourceLanguage.value,
      targetLanguage: targetLanguage.value,
    },
  })
  if (!settingsResult.ok) {
    stopWithError(settingsResult.error.message)
    return
  }

  const createResult = await sendExtensionMessage<{ job: Job; reused: boolean }>({
    type: 'job:create',
    payload: {
      youtubeUrl: youtubeUrl.value.trim(),
      sourceLanguage: sourceLanguage.value,
      targetLanguage: targetLanguage.value,
    },
  })
  if (!createResult.ok) {
    stopWithError(createResult.error.message)
    return
  }

  currentJob.value = createResult.data.job
  startPolling(createResult.data.job.id)
}

async function pollJob(jobId: string) {
  const result = await sendExtensionMessage<{ job: Job }>({
    type: 'job:get',
    payload: { jobId },
  })
  if (!result.ok) {
    stopWithError(result.error.message)
    return
  }

  currentJob.value = result.data.job
  if (result.data.job.status === 'completed') {
    const resolveResult = await sendExtensionMessage<SubtitleAsset | null>({
      type: 'subtitle:resolve',
      payload: { videoId: result.data.job.videoId },
    })
    clearTimers()
    isSubmitting.value = false
    if (!resolveResult.ok) {
      errorMessage.value = resolveResult.error.message
    }
    return
  }

  if (result.data.job.status === 'failed') {
    stopWithError(result.data.job.errorMessage ?? '任务失败')
  }
}

function startPolling(jobId: string) {
  void pollJob(jobId)
  pollTimer.value = window.setInterval(() => {
    void pollJob(jobId)
  }, 1500)
  elapsedTimer.value = window.setInterval(() => {
    if (currentJob.value?.status === 'transcribing') {
      elapsedSeconds.value += 1
    }
  }, 1000)
}

function clearTimers() {
  if (pollTimer.value !== null) {
    window.clearInterval(pollTimer.value)
    pollTimer.value = null
  }
  if (elapsedTimer.value !== null) {
    window.clearInterval(elapsedTimer.value)
    elapsedTimer.value = null
  }
}

function stopWithError(message: string) {
  errorMessage.value = message
  isSubmitting.value = false
  clearTimers()
}
</script>

<template>
  <main class="w-[380px] bg-background p-4 text-foreground">
    <Card class="gap-4 rounded-lg py-4 shadow-none">
      <CardHeader class="gap-1 px-4">
        <CardTitle class="text-base">Lets Sub It</CardTitle>
        <CardDescription class="text-xs">
          提交当前 YouTube 视频并生成双语字幕。
        </CardDescription>
      </CardHeader>

      <CardContent class="space-y-4 px-4">
        <form class="space-y-3" @submit.prevent="submitJob">
          <label class="grid gap-1.5 text-xs font-medium">
            backend URL
            <Input
              v-model="backendBaseUrl"
              class="h-8 text-xs"
              placeholder="http://127.0.0.1:8080"
            />
          </label>

          <label class="grid gap-1.5 text-xs font-medium">
            YouTube URL
            <Input
              v-model="youtubeUrl"
              class="h-8 text-xs"
              placeholder="https://www.youtube.com/watch?v=..."
            />
          </label>

          <div class="grid grid-cols-2 gap-3">
            <label class="grid min-w-0 gap-1.5 text-xs font-medium">
              源语言
              <Select v-model="sourceLanguage">
                <SelectTrigger class="h-8 w-full text-xs">
                  <SelectValue placeholder="源语言" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="language in languages"
                    :key="language"
                    :value="language"
                  >
                    {{ languageLabels[language] }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </label>

            <label class="grid min-w-0 gap-1.5 text-xs font-medium">
              目标语言
              <Select v-model="targetLanguage">
                <SelectTrigger class="h-8 w-full text-xs">
                  <SelectValue placeholder="目标语言" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="language in languages"
                    :key="language"
                    :value="language"
                  >
                    {{ languageLabels[language] }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </label>
          </div>

          <Alert v-if="alertMessage" variant="destructive" class="py-2">
            <AlertDescription class="text-xs">
              {{ alertMessage }}
            </AlertDescription>
          </Alert>

          <Button
            class="h-8 w-full text-xs"
            type="submit"
            :disabled="isSubmitDisabled"
          >
            {{ isSubmitting ? '处理中...' : '生成字幕' }}
          </Button>
        </form>

        <section
          v-if="currentJob"
          class="space-y-2 rounded-md border bg-muted/40 p-3 text-xs"
        >
          <div class="flex items-center justify-between gap-2">
            <span class="font-medium">任务状态</span>
            <Badge :variant="statusBadgeVariant">
              {{ statusLabel }}
            </Badge>
          </div>
          <p class="break-words text-muted-foreground">
            {{ currentJob.progressText }}
          </p>
          <p
            v-if="currentJob.status === 'transcribing'"
            class="text-muted-foreground"
          >
            转写已用 {{ elapsedSeconds }} 秒
          </p>
          <p
            v-if="currentJob.status === 'completed'"
            class="text-muted-foreground"
          >
            字幕已生成并写入本地缓存。
          </p>
        </section>
      </CardContent>
    </Card>
  </main>
</template>
