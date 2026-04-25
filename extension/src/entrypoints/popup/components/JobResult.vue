<script setup lang="ts">
import { ref } from 'vue'
import type { JobResponse, SubtitleMode } from '@/types'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  job: JobResponse | null
  failedJob: JobResponse | null
}>()

const emit = defineEmits<{
  retry: []
}>()

const selectedMode = ref<SubtitleMode>('bilingual')

const isFailed = props.failedJob !== null
const displayJob = props.failedJob ?? props.job
</script>

<template>
  <div class="flex flex-col gap-4 min-h-[250px]">
    <div v-if="isFailed" class="text-center">
      <p class="text-destructive font-medium">处理失败</p>
      <p class="text-sm text-muted-foreground mt-1">
        阶段：{{ displayJob?.stage }}
      </p>
      <p class="text-xs text-muted-foreground mt-1" v-if="displayJob?.errorMessage">
        {{ displayJob.errorMessage }}
      </p>
      <Button class="mt-4" variant="outline" @click="emit('retry')">
        重新提交
      </Button>
    </div>

    <div v-else class="flex flex-col gap-4">
      <div class="text-center">
        <p class="text-green-600 font-medium">字幕已生成</p>
      </div>

      <div class="flex flex-col gap-2">
        <label class="text-sm font-medium">字幕模式</label>
        <div class="flex gap-2">
          <Button
            :variant="selectedMode === 'translated' ? 'default' : 'outline'"
            size="sm"
            @click="selectedMode = 'translated'"
          >
            仅翻译
          </Button>
          <Button
            :variant="selectedMode === 'bilingual' ? 'default' : 'outline'"
            size="sm"
            @click="selectedMode = 'bilingual'"
          >
            双语
          </Button>
        </div>
      </div>

      <div class="text-xs text-muted-foreground">
        <p>视频 ID: {{ displayJob?.videoId }}</p>
        <p>{{ displayJob?.sourceLanguage }} → {{ displayJob?.targetLanguage }}</p>
      </div>

      <p class="text-xs text-muted-foreground">
        打开 YouTube 视频页面，字幕将自动显示。
      </p>
    </div>
  </div>
</template>
