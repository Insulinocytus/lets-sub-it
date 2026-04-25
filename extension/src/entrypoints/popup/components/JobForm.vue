<script setup lang="ts">
import { ref, computed } from 'vue'
import { Button } from '@/components/ui/button'
import type { CreateJobParams } from '@/types'

const props = defineProps<{
  submitJob: (params: CreateJobParams) => Promise<void>
}>()

const youtubeUrl = ref('')
const sourceLanguage = ref('ja')
const targetLanguage = ref('zh-CN')
const isLoading = ref(false)
const error = ref<string | null>(null)

const isValidUrl = computed(() => {
  return /^(https?:\/\/)?(www\.)?(youtube\.com\/watch\?v=|youtu\.be\/)[\w-]{11}/.test(youtubeUrl.value)
})

const languages = [
  { code: 'ja', label: '日本語' },
  { code: 'en', label: 'English' },
  { code: 'ko', label: '한국어' },
  { code: 'zh-CN', label: '简体中文' },
  { code: 'zh-TW', label: '繁体中文' },
]

async function handleSubmit() {
  if (!isValidUrl.value || isLoading.value) return
  isLoading.value = true
  error.value = null

  try {
    await props.submitJob({
      youtubeUrl: youtubeUrl.value,
      sourceLanguage: sourceLanguage.value,
      targetLanguage: targetLanguage.value,
    })
  } catch (err) {
    error.value = err instanceof Error ? err.message : '无法连接服务器，请检查服务是否启动'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <h1 class="text-lg font-semibold">Let's Sub It</h1>

    <div class="flex flex-col gap-2">
      <label class="text-sm font-medium">YouTube URL</label>
      <input
        v-model="youtubeUrl"
        type="url"
        placeholder="https://www.youtube.com/watch?v=..."
        class="px-3 py-2 border rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-ring"
        :class="{ 'border-destructive': youtubeUrl && !isValidUrl }"
      />
      <p v-if="youtubeUrl && !isValidUrl" class="text-xs text-destructive">
        请输入有效的 YouTube URL
      </p>
    </div>

    <div class="flex gap-3">
      <div class="flex flex-col gap-2 flex-1">
        <label class="text-sm font-medium">源语言</label>
        <select
          v-model="sourceLanguage"
          class="px-3 py-2 border rounded-md text-sm bg-background"
        >
          <option v-for="lang in languages" :key="lang.code" :value="lang.code">
            {{ lang.label }}
          </option>
        </select>
      </div>

      <div class="flex flex-col gap-2 flex-1">
        <label class="text-sm font-medium">目标语言</label>
        <select
          v-model="targetLanguage"
          class="px-3 py-2 border rounded-md text-sm bg-background"
        >
          <option v-for="lang in languages" :key="lang.code" :value="lang.code">
            {{ lang.label }}
          </option>
        </select>
      </div>
    </div>

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <Button
      :disabled="!isValidUrl || isLoading"
      @click="handleSubmit"
    >
      {{ isLoading ? '提交中...' : '生成字幕' }}
    </Button>
  </div>
</template>
