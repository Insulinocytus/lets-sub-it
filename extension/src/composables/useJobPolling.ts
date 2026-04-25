import { ref, onUnmounted } from 'vue'
import { useApi } from './useApi'
import type { JobResponse, JobStatus } from '@/types'

export function useJobPolling() {
  const jobId = ref<string | null>(null)
  const status = ref<JobStatus | null>(null)
  const job = ref<JobResponse | null>(null)
  const error = ref<string | null>(null)
  const isPolling = ref(false)

  const { getJob } = useApi()

  let intervalId: ReturnType<typeof setInterval> | null = null

  const terminalStatuses: JobStatus[] = ['completed', 'failed']

  function start(id: string, intervalMs = 1000) {
    jobId.value = id
    isPolling.value = true
    error.value = null
    poll()
    intervalId = setInterval(poll, intervalMs)
  }

  function stop() {
    if (intervalId !== null) {
      clearInterval(intervalId)
      intervalId = null
    }
    isPolling.value = false
  }

  async function poll() {
    if (!jobId.value) return
    try {
      const result = await getJob(jobId.value)
      job.value = result
      status.value = result.status as JobStatus
      if (terminalStatuses.includes(result.status as JobStatus)) {
        stop()
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'polling failed'
      stop()
    }
  }

  function reset() {
    stop()
    jobId.value = null
    status.value = null
    job.value = null
    error.value = null
  }

  onUnmounted(() => {
    stop()
  })

  return { jobId, status, job, error, isPolling, start, stop, reset }
}
