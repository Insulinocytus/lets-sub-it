import { beforeEach, describe, expect, it, vi } from 'vitest'
import { browser, type Browser } from 'wxt/browser'
import { fakeBrowser } from 'wxt/testing/fake-browser'
import { storage } from 'wxt/utils/storage'
import { updateSettings } from '@/storage/settings'
import type { BackendClient } from './backend-client'
import type { Job, SubtitleAsset } from './messages'
import {
  ensurePersistedJobMonitors,
  getJobMonitorAlarmName,
  handleJobMonitorAlarm,
  MIN_JOB_MONITOR_ALARM_DELAY_MS,
  resetJobMonitorsForTest,
  startJobMonitor,
} from './job-monitor'

const queuedJob: Job = {
  id: 'job_123',
  videoId: 'video_123',
  youtubeUrl: 'https://www.youtube.com/watch?v=video_123',
  sourceLanguage: 'en',
  targetLanguage: 'zh-CN',
  status: 'queued',
  stage: 'queued',
  progressText: '等待处理',
  errorMessage: null,
  createdAt: '2026-04-25T00:00:00Z',
  updatedAt: '2026-04-25T00:00:00Z',
}

const completedJob: Job = {
  ...queuedJob,
  status: 'completed',
  stage: 'completed',
  progressText: '完成',
  updatedAt: '2026-04-25T00:01:00Z',
}

const secondQueuedJob: Job = {
  ...queuedJob,
  id: 'job_456',
  videoId: 'video_456',
  youtubeUrl: 'https://www.youtube.com/watch?v=video_456',
}

const sameIdSecondBackendJob: Job = {
  ...queuedJob,
  videoId: 'video_backend_b',
  youtubeUrl: 'https://www.youtube.com/watch?v=video_backend_b',
}

const asset: SubtitleAsset = {
  jobId: 'job_123',
  videoId: 'video_123',
  sourceLanguage: 'en',
  targetLanguage: 'zh-CN',
  files: {
    source: '/subtitle-files/job_123/source',
    translated: '/subtitle-files/job_123/translated',
    bilingual: '/subtitle-files/job_123/bilingual',
  },
  createdAt: '2026-04-25T00:01:00Z',
}

const defaultBackendBaseUrl = 'http://127.0.0.1:8080'

describe('job monitor', () => {
  beforeEach(() => {
    vi.useRealTimers()
    fakeBrowser.reset()
    return resetJobMonitorsForTest()
  })

  it('caches subtitle assets and notifies the current tab when a job completes', async () => {
    const client = fakeClient({
      getJob: vi.fn(async () => ({ job: completedJob })),
      getSubtitleAsset: vi.fn(async () => ({ asset })),
    })
    const setCachedSubtitleAsset = vi.fn()
    const notifySubtitleUpdated = vi.fn()

    await startJobMonitor(queuedJob, {
      backendBaseUrl: defaultBackendBaseUrl,
      client,
      now: () => '2026-04-25T00:02:00Z',
      setCachedSubtitleAsset,
      notifySubtitleUpdated,
    })

    expect(client.getJob).toHaveBeenCalledWith('job_123')
    await vi.waitFor(() => {
      expect(client.getSubtitleAsset).toHaveBeenCalledWith('video_123', 'zh-CN')
      expect(setCachedSubtitleAsset).toHaveBeenCalledWith(
        asset,
        'translated',
        '2026-04-25T00:02:00Z',
        defaultBackendBaseUrl,
      )
    })
    expect(notifySubtitleUpdated).toHaveBeenCalledWith('video_123')
  })

  it('does not start duplicate monitors for the same job', async () => {
    const client = fakeClient({
      getJob: vi.fn(async () => ({ job: queuedJob })),
    })

    const firstStarted = await startJobMonitor(queuedJob, {
      client,
      pollIntervalMs: 1000,
    })
    const secondStarted = await startJobMonitor(queuedJob, {
      client,
      pollIntervalMs: 1000,
    })

    expect(firstStarted).toBe(true)
    expect(secondStarted).toBe(false)
    expect(client.getJob).toHaveBeenCalledTimes(1)
  })

  it('allows monitors with the same jobId on different backend origins to coexist', async () => {
    const client = fakeClient({
      getJob: vi.fn(async () => ({ job: queuedJob })),
    })
    const persistedJobMonitorsItem = storage.defineItem<
      Array<{ jobId: string; backendBaseUrl: string }>
    >('local:jobMonitors', {
      fallback: [],
    })

    const firstStarted = await startJobMonitor(queuedJob, {
      backendBaseUrl: 'http://127.0.0.1:8080',
      client,
      pollIntervalMs: 1000,
    })
    const secondStarted = await startJobMonitor(sameIdSecondBackendJob, {
      backendBaseUrl: 'http://127.0.0.1:9090',
      client,
      pollIntervalMs: 1000,
    })

    expect(firstStarted).toBe(true)
    expect(secondStarted).toBe(true)
    await expect(persistedJobMonitorsItem.getValue()).resolves.toEqual([
      { jobId: 'job_123', backendBaseUrl: 'http://127.0.0.1:8080' },
      { jobId: 'job_123', backendBaseUrl: 'http://127.0.0.1:9090' },
    ])
  })

  it('persists monitors and continues processing after alarm-based recovery', async () => {
    const client = fakeClient({
      getJob: vi
        .fn()
        .mockResolvedValueOnce({ job: queuedJob })
        .mockResolvedValueOnce({ job: completedJob }),
      getSubtitleAsset: vi.fn(async () => ({ asset })),
    })
    const setCachedSubtitleAsset = vi.fn()
    const notifySubtitleUpdated = vi.fn()
    const alarmName = getJobMonitorAlarmName(defaultBackendBaseUrl, queuedJob.id)

    const started = await startJobMonitor(queuedJob, {
      backendBaseUrl: defaultBackendBaseUrl,
      client,
      now: () => '2026-04-25T00:02:00Z',
      pollIntervalMs: 1000,
      setCachedSubtitleAsset,
      notifySubtitleUpdated,
    })

    expect(started).toBe(true)
    expect(client.getJob).toHaveBeenCalledTimes(1)
    await expect(browser.alarms.get(alarmName)).resolves.toMatchObject({
      name: alarmName,
    })

    await browser.alarms.clear(alarmName)
    await expect(browser.alarms.get(alarmName)).resolves.toBeUndefined()

    await ensurePersistedJobMonitors({ pollIntervalMs: 1000 })

    await expect(browser.alarms.get(alarmName)).resolves.toMatchObject({
      name: alarmName,
    })

    await handleJobMonitorAlarm(
      { name: alarmName, scheduledTime: Date.now() } as Browser.alarms.Alarm,
      {
        backendBaseUrl: defaultBackendBaseUrl,
        client,
        now: () => '2026-04-25T00:02:00Z',
        pollIntervalMs: 1000,
        setCachedSubtitleAsset,
        notifySubtitleUpdated,
      },
    )

    expect(client.getJob).toHaveBeenCalledTimes(2)
    expect(client.getSubtitleAsset).toHaveBeenCalledWith('video_123', 'zh-CN')
    expect(setCachedSubtitleAsset).toHaveBeenCalledWith(
      asset,
      'translated',
      '2026-04-25T00:02:00Z',
      defaultBackendBaseUrl,
    )
    expect(notifySubtitleUpdated).toHaveBeenCalledWith('video_123')
    await expect(browser.alarms.get(alarmName)).resolves.toBeUndefined()
  })

  it('keeps using the persisted backend origin after settings switch during alarm recovery', async () => {
    let jobPollCount = 0
    const fetchImpl = vi.fn(async (input: string | URL | Request) => {
      const url = String(input)

      if (url === 'http://127.0.0.1:8080/jobs/job_123') {
        jobPollCount += 1
        return new Response(
          JSON.stringify({ job: jobPollCount === 1 ? queuedJob : completedJob }),
          {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          },
        )
      }

      if (
        url ===
        'http://127.0.0.1:8080/subtitle-assets?videoId=video_123&targetLanguage=zh-CN'
      ) {
        return new Response(JSON.stringify({ asset }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        })
      }

      return new Response('unexpected url', { status: 500 })
    })
    const setCachedSubtitleAsset = vi.fn()
    const alarmName = getJobMonitorAlarmName('http://127.0.0.1:8080', queuedJob.id)

    await startJobMonitor(queuedJob, {
      backendBaseUrl: 'http://127.0.0.1:8080',
      fetchImpl,
      now: () => '2026-04-25T00:02:00Z',
      pollIntervalMs: 1000,
      setCachedSubtitleAsset,
    })

    await updateSettings({
      backendBaseUrl: 'http://127.0.0.1:9090',
    })
    await browser.alarms.clear(alarmName)
    await ensurePersistedJobMonitors({ pollIntervalMs: 1000 })

    await handleJobMonitorAlarm(
      { name: alarmName, scheduledTime: Date.now() } as Browser.alarms.Alarm,
      {
        fetchImpl,
        now: () => '2026-04-25T00:02:00Z',
        pollIntervalMs: 1000,
        setCachedSubtitleAsset,
      },
    )

    expect(fetchImpl).toHaveBeenCalledWith(
      'http://127.0.0.1:8080/jobs/job_123',
      undefined,
    )
    expect(setCachedSubtitleAsset).toHaveBeenCalledWith(
      asset,
      'translated',
      '2026-04-25T00:02:00Z',
      'http://127.0.0.1:8080',
    )
  })

  it('routes alarm recovery to the persisted monitor matching that backend origin', async () => {
    const backendAAsset: SubtitleAsset = {
      ...asset,
      videoId: 'video_backend_a',
      files: {
        source: '/subtitle-files/job_123_a/source',
        translated: '/subtitle-files/job_123_a/translated',
        bilingual: '/subtitle-files/job_123_a/bilingual',
      },
    }
    const backendBAsset: SubtitleAsset = {
      ...asset,
      videoId: 'video_backend_b',
      files: {
        source: '/subtitle-files/job_123_b/source',
        translated: '/subtitle-files/job_123_b/translated',
        bilingual: '/subtitle-files/job_123_b/bilingual',
      },
    }
    const completedBackendAJob: Job = {
      ...completedJob,
      videoId: 'video_backend_a',
      youtubeUrl: 'https://www.youtube.com/watch?v=video_backend_a',
    }
    const completedBackendBJob: Job = {
      ...completedJob,
      videoId: 'video_backend_b',
      youtubeUrl: 'https://www.youtube.com/watch?v=video_backend_b',
    }
    let backendAJobPollCount = 0
    let backendBJobPollCount = 0
    const fetchImpl = vi.fn(async (input: string | URL | Request) => {
      const url = String(input)

      if (url === 'http://127.0.0.1:8080/jobs/job_123') {
        backendAJobPollCount += 1
        return new Response(
          JSON.stringify({
            job: backendAJobPollCount === 1 ? queuedJob : completedBackendAJob,
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (url === 'http://127.0.0.1:9090/jobs/job_123') {
        backendBJobPollCount += 1
        return new Response(
          JSON.stringify({
            job:
              backendBJobPollCount === 1
                ? sameIdSecondBackendJob
                : completedBackendBJob,
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        )
      }

      if (
        url ===
        'http://127.0.0.1:8080/subtitle-assets?videoId=video_backend_a&targetLanguage=zh-CN'
      ) {
        return new Response(JSON.stringify({ asset: backendAAsset }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        })
      }

      if (
        url ===
        'http://127.0.0.1:9090/subtitle-assets?videoId=video_backend_b&targetLanguage=zh-CN'
      ) {
        return new Response(JSON.stringify({ asset: backendBAsset }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        })
      }

      return new Response('unexpected url', { status: 500 })
    })
    const setCachedSubtitleAsset = vi.fn()
    const backendAAlarmName = getJobMonitorAlarmName('http://127.0.0.1:8080', 'job_123')
    const backendBAlarmName = getJobMonitorAlarmName('http://127.0.0.1:9090', 'job_123')

    await startJobMonitor(queuedJob, {
      backendBaseUrl: 'http://127.0.0.1:8080',
      fetchImpl,
      now: () => '2026-04-25T00:02:00Z',
      pollIntervalMs: 1000,
      setCachedSubtitleAsset,
    })
    await startJobMonitor(sameIdSecondBackendJob, {
      backendBaseUrl: 'http://127.0.0.1:9090',
      fetchImpl,
      now: () => '2026-04-25T00:02:00Z',
      pollIntervalMs: 1000,
      setCachedSubtitleAsset,
    })

    await browser.alarms.clear(backendAAlarmName)
    await browser.alarms.clear(backendBAlarmName)
    await ensurePersistedJobMonitors({ pollIntervalMs: 1000 })

    await handleJobMonitorAlarm(
      { name: backendBAlarmName, scheduledTime: Date.now() } as Browser.alarms.Alarm,
      {
        fetchImpl,
        now: () => '2026-04-25T00:02:00Z',
        pollIntervalMs: 1000,
        setCachedSubtitleAsset,
      },
    )

    expect(fetchImpl).toHaveBeenCalledWith(
      'http://127.0.0.1:9090/jobs/job_123',
      undefined,
    )
    expect(setCachedSubtitleAsset).toHaveBeenCalledWith(
      backendBAsset,
      'translated',
      '2026-04-25T00:02:00Z',
      'http://127.0.0.1:9090',
    )
    expect(setCachedSubtitleAsset).not.toHaveBeenCalledWith(
      backendAAsset,
      'translated',
      '2026-04-25T00:02:00Z',
      'http://127.0.0.1:8080',
    )
  })

  it('broadcasts subtitle updates to injected youtube tabs and ignores per-tab delivery failures', async () => {
    const client = fakeClient({
      getJob: vi.fn(async () => ({ job: completedJob })),
      getSubtitleAsset: vi.fn(async () => ({ asset })),
    })
    const setCachedSubtitleAsset = vi.fn()
    const queryTabs = vi.spyOn(browser.tabs, 'query').mockImplementation(async () => [
        { id: 11, url: 'https://www.youtube.com/watch?v=video_123' },
        { id: 22, url: 'https://www.youtube.com/watch?v=another' },
      ] as Browser.tabs.Tab[])
    const sendMessage = vi
      .spyOn(browser.tabs, 'sendMessage')
      .mockRejectedValueOnce(new Error('content script not ready'))
      .mockResolvedValueOnce(undefined)

    await startJobMonitor(queuedJob, {
      backendBaseUrl: defaultBackendBaseUrl,
      client,
      now: () => '2026-04-25T00:02:00Z',
      setCachedSubtitleAsset,
    })

    await vi.waitFor(() => {
      expect(setCachedSubtitleAsset).toHaveBeenCalledWith(
        asset,
        'translated',
        '2026-04-25T00:02:00Z',
        defaultBackendBaseUrl,
      )
    })

    expect(queryTabs).toHaveBeenCalledWith({
      url: ['https://www.youtube.com/watch*'],
    })
    expect(sendMessage).toHaveBeenCalledTimes(2)
    expect(sendMessage).toHaveBeenNthCalledWith(1, 11, {
      type: 'lets-sub-it:subtitle-updated',
      videoId: 'video_123',
    })
    expect(sendMessage).toHaveBeenNthCalledWith(2, 22, {
      type: 'lets-sub-it:subtitle-updated',
      videoId: 'video_123',
    })
  })

  it('clamps alarm delay to the MV3 minimum granularity', async () => {
    const client = fakeClient({
      getJob: vi.fn(async () => ({ job: queuedJob })),
    })
    const alarmName = getJobMonitorAlarmName(defaultBackendBaseUrl, queuedJob.id)
    const startTime = Date.now()

    await startJobMonitor(queuedJob, {
      client,
      pollIntervalMs: 1000,
    })

    const alarm = await browser.alarms.get(alarmName)

    expect(alarm).toMatchObject({ name: alarmName })
    expect((alarm?.scheduledTime ?? 0) - startTime).toBeGreaterThanOrEqual(
      MIN_JOB_MONITOR_ALARM_DELAY_MS,
    )
  })

  it('keeps both persisted monitors when two jobs start concurrently', async () => {
    const client = fakeClient({
      getJob: vi.fn(async () => ({ job: queuedJob })),
    })
    const persistedJobMonitorsItem = storage.defineItem<
      Array<{ jobId: string; backendBaseUrl: string }>
    >(
      'local:jobMonitors',
      {
        fallback: [],
      },
    )

    await Promise.all([
      startJobMonitor(queuedJob, { client, pollIntervalMs: 1000 }),
      startJobMonitor(secondQueuedJob, { client, pollIntervalMs: 1000 }),
    ])

    await expect(persistedJobMonitorsItem.getValue()).resolves.toEqual([
      { jobId: 'job_123', backendBaseUrl: defaultBackendBaseUrl },
      { jobId: 'job_456', backendBaseUrl: defaultBackendBaseUrl },
    ])
  })
})

function fakeClient(overrides: Partial<BackendClient>): BackendClient {
  return {
    createJob: vi.fn(),
    getJob: vi.fn(),
    getSubtitleAsset: vi.fn(),
    fetchSubtitleFile: vi.fn(),
    ...overrides,
  } as BackendClient
}
