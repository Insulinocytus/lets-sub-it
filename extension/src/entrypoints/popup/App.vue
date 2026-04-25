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

async function handleSubmit(params: CreateJobParams) {
  try {
    const result = await api.createJob(params)
    if (result.reused && result.job.status === 'completed') {
      handleJobComplete(result.job, 'bilingual')
    } else if (result.job.status === 'failed') {
      handleJobFailed(result.job)
    } else {
      viewState.value = 'polling'
      polling.start(result.job.id)
    }
  } catch (err) {
    throw err
  }
}

function handleJobComplete(job: JobResponse, selectedMode: SubtitleMode) {
  completedJob.value = job
  cache.setCacheEntry({
    videoId: job.videoId,
    targetLanguage: job.targetLanguage,
    jobId: job.id,
    selectedMode,
    lastSyncedAt: new Date().toISOString(),
  })
  cache.setPreferences({
    videoId: job.videoId,
    targetLanguage: job.targetLanguage,
    selectedMode,
  })
  viewState.value = 'result'
}

function handleJobFailed(job: JobResponse) {
  failedJob.value = job
  viewState.value = 'result'
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
      @submit="handleSubmit"
    />
    <JobStatusView
      v-else-if="viewState === 'polling'"
      :job="polling.job.value"
      :status="polling.status.value"
      :error="polling.error.value"
      @complete="handleJobComplete"
      @failed="handleJobFailed"
    />
    <JobResultView
      v-else
      :job="completedJob"
      :failed-job="failedJob"
      @retry="handleRetry"
    />
  </div>
</template>
