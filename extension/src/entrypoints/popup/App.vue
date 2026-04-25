<script setup lang="ts">
import { ref } from 'vue'
import type { JobStatus, JobResponse, CreateJobParams, SubtitleMode } from '@/types'
import { useApi } from '@/composables/useApi'
import { useJobPolling } from '@/composables/useJobPolling'
import { useCache } from '@/composables/useCache'
import JobForm from './components/JobForm.vue'
import JobStatusView from './components/JobStatus.vue'
import JobResultView from './components/JobResult.vue'

type ViewState = 'form' | 'polling' | 'result'

const viewState = ref<ViewState>('form')
const completedJob = ref<JobResponse | null>(null)
const failedJob = ref<JobResponse | null>(null)

const api = useApi()
const polling = useJobPolling()
const cache = useCache()
let modeChangeWrite = Promise.resolve()

async function handleSubmit(params: CreateJobParams) {
  const result = await api.createJob(params)
  if (result.reused && result.job.status === 'completed') {
    await handleJobComplete(result.job, 'bilingual')
  } else if (result.job.status === 'failed') {
    handleJobFailed(result.job)
  } else {
    viewState.value = 'polling'
    polling.start(result.job.id)
  }
}

async function handleJobComplete(job: JobResponse, selectedMode: SubtitleMode) {
  await Promise.all([
    cache.setCacheEntry({
      videoId: job.videoId,
      targetLanguage: job.targetLanguage,
      jobId: job.id,
      selectedMode,
      lastSyncedAt: new Date().toISOString(),
    }),
    cache.setPreferences({
      videoId: job.videoId,
      targetLanguage: job.targetLanguage,
      selectedMode,
    }),
  ])
  completedJob.value = job
  viewState.value = 'result'
}

async function handlePollingComplete(job: JobResponse, selectedMode: SubtitleMode) {
  try {
    await handleJobComplete(job, selectedMode)
  } catch (err) {
    failedJob.value = {
      ...job,
      status: 'failed',
      errorMessage: err instanceof Error ? err.message : '无法保存字幕缓存',
    }
    viewState.value = 'result'
  }
}

function handleJobFailed(job: JobResponse) {
  failedJob.value = job
  viewState.value = 'result'
}

async function handleModeChange(selectedMode: SubtitleMode) {
  const job = completedJob.value
  if (!job) return

  const currentWrite = modeChangeWrite.catch(() => undefined).then(async () => {
    const { videoId, targetLanguage } = job
    const existingEntry = await cache.getCacheEntry(videoId, targetLanguage)
    await cache.setSubtitleSelection(
      {
        videoId,
        targetLanguage,
        jobId: job.id,
        lastSyncedAt: new Date().toISOString(),
        ...existingEntry,
        selectedMode,
      },
      {
        videoId,
        targetLanguage,
        selectedMode,
      },
    )
  })
  modeChangeWrite = currentWrite
  await currentWrite
}

function handleRetry() {
  completedJob.value = null
  failedJob.value = null
  polling.reset()
  viewState.value = 'form'
}
</script>

<template>
  <div class="w-[400px] min-h-[300px] p-4">
    <JobForm
      v-if="viewState === 'form'"
      :submit-job="handleSubmit"
    />
    <JobStatusView
      v-else-if="viewState === 'polling'"
      :job="polling.job.value"
      :status="polling.status.value"
      :error="polling.error.value"
      @complete="handlePollingComplete"
      @failed="handleJobFailed"
    />
    <JobResultView
      v-else
      :job="completedJob"
      :failed-job="failedJob"
      :update-mode="handleModeChange"
      @retry="handleRetry"
    />
  </div>
</template>
