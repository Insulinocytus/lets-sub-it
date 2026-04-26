import { browser, type Browser } from 'wxt/browser'
import { storage } from 'wxt/utils/storage'
import {
  createBackendClient,
  normalizeBackendBaseUrl,
  type BackendClient,
} from './backend-client'
import type { Job } from './messages'
import { setCachedSubtitleAsset as setCachedSubtitleAssetInStorage } from '@/storage/subtitle-cache'
import { getSettings } from '@/storage/settings'

const DEFAULT_POLL_INTERVAL_MS = 1500
export const MIN_JOB_MONITOR_ALARM_DELAY_MS = 30_000
const JOB_MONITOR_ALARM_PREFIX = 'lets-sub-it:job-monitor:'

type PersistedJobMonitor = {
  jobId: string
  backendBaseUrl: string
}

const persistedJobMonitorsItem = storage.defineItem<PersistedJobMonitor[]>(
  'local:jobMonitors',
  {
    fallback: [],
  },
)
let persistedJobMonitorsMutationQueue: Promise<void> = Promise.resolve()

export type JobMonitorDeps = {
  backendBaseUrl?: string
  client?: Pick<BackendClient, 'getJob' | 'getSubtitleAsset'>
  fetchImpl?: typeof fetch
  now?: () => string
  pollIntervalMs?: number
  setCachedSubtitleAsset?: typeof setCachedSubtitleAssetInStorage
  notifySubtitleUpdated?: (videoId: string) => Promise<void>
}

export async function startJobMonitor(
  job: Job,
  deps: JobMonitorDeps = {},
): Promise<boolean> {
  const backendBaseUrl = await getMonitorBackendBaseUrl(deps)
  const monitor = { jobId: job.id, backendBaseUrl }
  const added = await addPersistedJobMonitor(monitor)
  await ensureJobMonitorAlarm(monitor, deps)
  if (added) {
    await pollJob(monitorKey(monitor), { ...deps, backendBaseUrl })
  }
  return added
}

export async function ensurePersistedJobMonitors(
  deps: JobMonitorDeps = {},
): Promise<void> {
  const monitors = await getPersistedJobMonitors()
  await Promise.all(monitors.map((monitor) => ensureJobMonitorAlarm(monitor, deps)))
}

export async function handleJobMonitorAlarm(
  alarm: Browser.alarms.Alarm,
  deps: JobMonitorDeps = {},
): Promise<void> {
  const key = getMonitorKeyFromAlarmName(alarm.name)
  if (!key) {
    return
  }
  if (!(await hasPersistedJobMonitor(key))) {
    await browser.alarms.clear(alarm.name)
    return
  }

  await pollJob(key, deps)
}

export async function resetJobMonitorsForTest(): Promise<void> {
  await updatePersistedJobMonitors(async () => ({
    result: undefined,
    monitors: [],
  }))
  await browser.alarms.clearAll()
}

export function getJobMonitorAlarmName(backendBaseUrl: string, jobId: string) {
  return `${JOB_MONITOR_ALARM_PREFIX}${createMonitorKey(backendBaseUrl, jobId)}`
}

async function pollJob(key: string, deps: JobMonitorDeps) {
  const monitor = await getPersistedJobMonitor(key)
  if (!monitor) {
    return
  }
  const monitorDeps = {
    ...deps,
    backendBaseUrl: monitor.backendBaseUrl,
  }

  try {
    const client = await getMonitorClient(monitorDeps)
    const response = await client.getJob(monitor.jobId)
    if (!(await hasPersistedJobMonitor(key))) {
      return
    }

    if (response.job.status === 'completed') {
      await cacheAndNotify(response.job, client, monitorDeps)
      await stopJobMonitor(monitor)
      return
    }

    if (response.job.status === 'failed') {
      await stopJobMonitor(monitor)
      return
    }

    await ensureJobMonitorAlarm(monitor, deps)
  } catch {
    await ensureJobMonitorAlarm(monitor, deps)
  }
}

async function cacheAndNotify(
  job: Job,
  client: Pick<BackendClient, 'getSubtitleAsset'>,
  deps: JobMonitorDeps,
) {
  const response = await client.getSubtitleAsset(job.videoId, job.targetLanguage)
  if (!response.asset) {
    return
  }

  const cache = deps.setCachedSubtitleAsset ?? setCachedSubtitleAssetInStorage
  const backendBaseUrl =
    deps.backendBaseUrl ?? (await getSettings()).backendBaseUrl
  await cache(
    response.asset,
    'translated',
    (deps.now ?? defaultNow)(),
    backendBaseUrl,
  )
  await (deps.notifySubtitleUpdated ?? notifyYoutubeWatchTabs)(job.videoId)
}

async function stopJobMonitor(monitor: PersistedJobMonitor) {
  await removePersistedJobMonitor(monitorKey(monitor))
  await browser.alarms.clear(getJobMonitorAlarmName(monitor.backendBaseUrl, monitor.jobId))
}

async function getMonitorClient(
  deps: JobMonitorDeps,
): Promise<Pick<BackendClient, 'getJob' | 'getSubtitleAsset'>> {
  if (deps.client) {
    return deps.client
  }

  return createBackendClient(
    await getMonitorBackendBaseUrl(deps),
    deps.fetchImpl ?? fetch,
  )
}

async function getPersistedJobMonitors(): Promise<PersistedJobMonitor[]> {
  await persistedJobMonitorsMutationQueue
  return getPersistedJobMonitorsSnapshot()
}

async function hasPersistedJobMonitor(key: string): Promise<boolean> {
  const monitors = await getPersistedJobMonitors()
  return monitors.some((monitor) => monitorKey(monitor) === key)
}

async function getPersistedJobMonitor(
  key: string,
): Promise<PersistedJobMonitor | null> {
  const monitors = await getPersistedJobMonitors()
  return monitors.find((monitor) => monitorKey(monitor) === key) ?? null
}

async function addPersistedJobMonitor(
  monitor: PersistedJobMonitor,
): Promise<boolean> {
  return updatePersistedJobMonitors(async (monitors) => {
    if (monitors.some((current) => monitorKey(current) === monitorKey(monitor))) {
      return { result: false, monitors }
    }

    return {
      result: true,
      monitors: [...monitors, monitor],
    }
  })
}

async function removePersistedJobMonitor(key: string): Promise<void> {
  await updatePersistedJobMonitors(async (monitors) => ({
    result: undefined,
    monitors: monitors.filter((monitor) => monitorKey(monitor) !== key),
  }))
}

async function ensureJobMonitorAlarm(
  monitor: PersistedJobMonitor,
  deps: JobMonitorDeps,
) {
  const name = getJobMonitorAlarmName(monitor.backendBaseUrl, monitor.jobId)
  const existing = await browser.alarms.get(name)
  if (existing) {
    return
  }

  const delayMs = Math.max(
    deps.pollIntervalMs ?? DEFAULT_POLL_INTERVAL_MS,
    MIN_JOB_MONITOR_ALARM_DELAY_MS,
  )
  browser.alarms.create(name, {
    delayInMinutes: delayMs / 60000,
  })
}

function getMonitorKeyFromAlarmName(name: string): string | null {
  if (!name.startsWith(JOB_MONITOR_ALARM_PREFIX)) {
    return null
  }

  return name.slice(JOB_MONITOR_ALARM_PREFIX.length) || null
}

function defaultNow() {
  return new Date().toISOString()
}

async function getMonitorBackendBaseUrl(deps: JobMonitorDeps): Promise<string> {
  if (deps.backendBaseUrl) {
    return normalizeBackendBaseUrl(deps.backendBaseUrl)
  }

  const settings = await getSettings()
  return normalizeBackendBaseUrl(settings.backendBaseUrl)
}

function createMonitorKey(backendBaseUrl: string, jobId: string) {
  return `${encodeURIComponent(normalizeBackendBaseUrl(backendBaseUrl))}:${encodeURIComponent(jobId)}`
}

function monitorKey(monitor: PersistedJobMonitor) {
  return createMonitorKey(monitor.backendBaseUrl, monitor.jobId)
}

async function getPersistedJobMonitorsSnapshot(): Promise<PersistedJobMonitor[]> {
  return (await persistedJobMonitorsItem.getValue()) ?? []
}

function updatePersistedJobMonitors<T>(
  update: (
    monitors: PersistedJobMonitor[],
  ) => Promise<{ result: T; monitors: PersistedJobMonitor[] }>,
): Promise<T> {
  const run = async () => {
    const current = await getPersistedJobMonitorsSnapshot()
    const next = await update(current)
    await persistedJobMonitorsItem.setValue(next.monitors)
    return next.result
  }

  const result = persistedJobMonitorsMutationQueue.then(run, run)
  persistedJobMonitorsMutationQueue = result.then(
    () => undefined,
    () => undefined,
  )
  return result
}

async function notifyYoutubeWatchTabs(videoId: string) {
  const tabs = await browser.tabs.query({
    url: ['https://www.youtube.com/watch*'],
  })

  await Promise.allSettled(
    tabs.map(async (tab) => {
      if (!tab.id) {
        return
      }

      await browser.tabs.sendMessage(tab.id, {
        type: 'lets-sub-it:subtitle-updated',
        videoId,
      })
    }),
  )
}
