<script setup lang="ts">
import { watch } from 'vue'
import type { JobResponse, JobStatus as JobStatusType, SubtitleMode } from '@/types'

const props = defineProps<{
  job: JobResponse | null
  status: JobStatusType | null
  error: string | null
}>()

const emit = defineEmits<{
  complete: [job: JobResponse, mode: SubtitleMode]
  failed: [job: JobResponse]
}>()

const stageLabels: Record<string, string> = {
  queued: '等待处理',
  downloading: '准备媒体',
  transcribing: '转写中',
  translating: '翻译中',
  packaging: '生成字幕',
}

watch(
  () => props.status,
  (newStatus) => {
    if (newStatus === 'completed' && props.job) {
      emit('complete', props.job, 'bilingual')
    } else if (newStatus === 'failed' && props.job) {
      emit('failed', props.job)
    }
  },
)
</script>

<template>
  <div class="flex flex-col gap-4 items-center justify-center min-h-[250px]">
    <div v-if="error" class="text-center">
      <p class="text-destructive text-sm">{{ error }}</p>
    </div>

    <template v-else-if="job">
      <div class="animate-spin h-8 w-8 border-2 border-primary border-t-transparent rounded-full" />

      <div class="text-center">
        <p class="text-sm font-medium">
          {{ stageLabels[job.stage] || job.stage }}
        </p>
        <p class="text-xs text-muted-foreground mt-1">
          {{ job.progressText }}
        </p>
      </div>

      <!-- Progress bar -->
      <div class="w-full bg-secondary rounded-full h-2">
        <div
          class="bg-primary h-2 rounded-full transition-all duration-500"
          :style="{ width: job.stage === 'packaging' ? '90%' : job.stage === 'translating' ? '60%' : '30%  ' }"
        />
      </div>
    </template>

    <div v-else class="text-sm text-muted-foreground">
      准备中...
    </div>
  </div>
</template>
