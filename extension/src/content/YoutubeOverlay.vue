<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { browser } from 'wxt/browser'
import {
  SUBTITLE_MODES,
  type Settings,
  type SubtitleMode,
  sendExtensionMessage,
} from '@/api/messages'
import type { SubtitleAssetCacheEntry } from '@/storage/subtitle-cache'
import { findActiveCue } from '@/subtitles/active-cue'
import { parseVtt, type VttCue } from '@/subtitles/vtt'
import {
  getCurrentVideoId,
  watchVideoIdChanges,
} from '@/youtube/page-watch'

const enabled = ref(true)
const status = ref('查找字幕')
const currentVideoId = ref<string | null>(null)
const currentAsset = ref<SubtitleAssetCacheEntry | null>(null)
const selectedMode = ref<SubtitleMode>('translated')
const settingsSubtitleMode = ref<SubtitleMode>('translated')
const subtitleFontSizePx = ref(20)
const cues = ref<VttCue[]>([])
const activeText = ref('')

let removeVideoListeners: (() => void) | null = null
let removeVideoIdWatch: (() => void) | null = null
let removeRuntimeListener: (() => void) | null = null
let removeToggleListener: (() => void) | null = null
let isMounted = false
let requestToken = 0

const hasSubtitle = computed(() => cues.value.length > 0)
const isWatchPage = computed(() => currentVideoId.value !== null)
const subtitleStyle = computed(() => ({
  fontSize: `${subtitleFontSizePx.value}px`,
}))

onMounted(async () => {
  isMounted = true
  await loadSettings()
  if (!isMounted) {
    return
  }

  const videoId = getCurrentVideoId()
  currentVideoId.value = videoId
  void loadForVideo(videoId)

  removeVideoIdWatch = watchVideoIdChanges((nextVideoId) => {
    currentVideoId.value = nextVideoId
    void loadForVideo(nextVideoId)
  })

  const handleRuntimeMessage = (message: unknown) => {
    if (isSettingsUpdatedMessage(message)) {
      void handleSettingsUpdated(message.settings)
      return
    }

    if (!isSubtitleUpdatedMessage(message)) {
      return
    }

    if (message.videoId !== currentVideoId.value) {
      return
    }

    void loadForVideo(currentVideoId.value)
  }
  browser.runtime.onMessage.addListener(handleRuntimeMessage)
  removeRuntimeListener = () => {
    browser.runtime.onMessage.removeListener(handleRuntimeMessage)
  }

  const handleToggleSubtitles = () => {
    enabled.value = !enabled.value
  }
  window.addEventListener('lets-sub-it:toggle-subtitles', handleToggleSubtitles)
  removeToggleListener = () => {
    window.removeEventListener('lets-sub-it:toggle-subtitles', handleToggleSubtitles)
  }
})

onUnmounted(() => {
  isMounted = false
  requestToken += 1
  cleanupVideoListeners()
  removeVideoIdWatch?.()
  removeVideoIdWatch = null
  removeRuntimeListener?.()
  removeRuntimeListener = null
  removeToggleListener?.()
  removeToggleListener = null
})

async function loadSettings() {
  let result
  try {
    result = await sendExtensionMessage<Settings>({ type: 'settings:get' })
  } catch (error) {
    status.value = readableError(error)
    return
  }

  if (!result.ok) {
    status.value = result.error.message
    return
  }

  applySettings(result.data)
}

async function handleSettingsUpdated(settings: Settings) {
  const previousMode = settingsSubtitleMode.value
  applySettings(settings)

  if (settingsSubtitleMode.value !== previousMode && currentAsset.value) {
    await changeMode(settingsSubtitleMode.value)
  }
}

function applySettings(settings: Settings) {
  subtitleFontSizePx.value = settings.subtitleFontSizePx
  settingsSubtitleMode.value = settings.subtitleMode
}

async function loadForVideo(videoId: string | null) {
  const token = ++requestToken
  resetLoadedSubtitles()
  currentVideoId.value = videoId
  currentAsset.value = null

  if (!videoId) {
    status.value = '非 watch 页面'
    return
  }

  status.value = '查找字幕'
  let result
  try {
    result = await sendExtensionMessage<SubtitleAssetCacheEntry | null>({
      type: 'subtitle:resolve',
      payload: { videoId },
    })
  } catch (error) {
    if (!canUpdate(token)) {
      return
    }
    status.value = readableError(error)
    return
  }
  if (!canUpdate(token) || currentVideoId.value !== videoId) {
    return
  }

  if (!result.ok) {
    status.value = result.error.message
    return
  }

  if (!result.data) {
    status.value = '未找到字幕'
    return
  }

  currentAsset.value = result.data
  selectedMode.value = settingsSubtitleMode.value
  await loadVtt(token)
}

async function loadVtt(token = requestToken): Promise<boolean> {
  const asset = currentAsset.value
  if (!asset || !canUpdate(token)) {
    return false
  }
  const jobId = asset.jobId
  const mode = selectedMode.value

  resetLoadedSubtitles()
  status.value = '加载字幕'
  let result
  try {
    result = await sendExtensionMessage<string>({
      type: 'subtitle:fetch-file',
      payload: { jobId, mode },
    })
  } catch (error) {
    if (!canUpdate(token)) {
      return false
    }
    status.value = readableError(error)
    return false
  }
  if (
    !canUpdate(token) ||
    currentAsset.value?.jobId !== jobId ||
    selectedMode.value !== mode
  ) {
    return false
  }

  if (!result.ok) {
    status.value = result.error.message
    return false
  }

  try {
    cues.value = parseVtt(result.data)
  } catch {
    resetLoadedSubtitles()
    status.value = '字幕解析失败'
    return false
  }

  status.value = '字幕已加载'
  bindVideo(token)
  return true
}

async function changeMode(mode: SubtitleMode) {
  if (selectedMode.value === mode) {
    return
  }

  const token = requestToken
  const previousMode = selectedMode.value
  const previousCues = cues.value
  const previousActiveText = activeText.value
  const asset = currentAsset.value
  if (!asset) {
    return
  }
  const videoId = asset.videoId
  const jobId = asset.jobId
  const targetLanguage = asset.targetLanguage

  selectedMode.value = mode
  resetLoadedSubtitles()
  let result
  try {
    result = await sendExtensionMessage<SubtitleAssetCacheEntry | null>({
      type: 'subtitle:update-mode',
      payload: {
        videoId,
        targetLanguage,
        mode,
      },
    })
  } catch (error) {
    if (
      canUpdate(token) &&
      currentVideoId.value === videoId &&
      currentAsset.value?.jobId === jobId &&
      selectedMode.value === mode
    ) {
      selectedMode.value = previousMode
      restoreDisplayedSubtitles(token, previousCues, previousActiveText)
      status.value = readableError(error)
    }
    return
  }

  if (
    !canUpdate(token) ||
    currentVideoId.value !== videoId ||
    currentAsset.value?.jobId !== jobId ||
    selectedMode.value !== mode
  ) {
    return
  }

  if (!result.ok) {
    selectedMode.value = previousMode
    restoreDisplayedSubtitles(token, previousCues, previousActiveText)
    status.value = result.error.message
    return
  }

  if (!result.data) {
    selectedMode.value = previousMode
    restoreDisplayedSubtitles(token, previousCues, previousActiveText)
    status.value = '字幕模式切换失败'
    return
  }

  currentAsset.value = result.data
  const loaded = await loadVtt(token)
  if (
    !loaded &&
    canUpdate(token) &&
    currentVideoId.value === videoId &&
    currentAsset.value?.jobId === jobId &&
    selectedMode.value === mode
  ) {
    selectedMode.value = previousMode
    if (currentAsset.value) {
      currentAsset.value = {
        ...currentAsset.value,
        selectedMode: previousMode,
      }
    }
    try {
      const rollbackResult = await sendExtensionMessage<SubtitleAssetCacheEntry | null>({
        type: 'subtitle:update-mode',
        payload: {
          videoId,
          targetLanguage,
          mode: previousMode,
        },
      })
      if (
        !canUpdate(token) ||
        currentVideoId.value !== videoId ||
        currentAsset.value?.jobId !== jobId ||
        selectedMode.value !== previousMode
      ) {
        return
      }
      if (!rollbackResult.ok) {
        restoreDisplayedSubtitles(token, previousCues, previousActiveText)
        status.value = rollbackResult.error.message
        return
      }
      if (!rollbackResult.data) {
        restoreDisplayedSubtitles(token, previousCues, previousActiveText)
        status.value = '字幕模式回滚失败'
        return
      }
      currentAsset.value = rollbackResult.data
    } catch (error) {
      if (
        canUpdate(token) &&
        currentVideoId.value === videoId &&
        currentAsset.value?.jobId === jobId &&
        selectedMode.value === previousMode
      ) {
        restoreDisplayedSubtitles(token, previousCues, previousActiveText)
        status.value = readableError(error)
      }
      return
    }
    await loadVtt(token)
  }
}

function bindVideo(token: number) {
  if (!canUpdate(token)) {
    return
  }

  cleanupVideoListeners()

  const video = document.querySelector('video')
  if (!video) {
    status.value = '未找到视频'
    return
  }

  const updateActiveCue = () => {
    if (!canUpdate(token)) {
      return
    }
    activeText.value = findActiveCue(cues.value, video.currentTime)?.text ?? ''
  }

  video.addEventListener('timeupdate', updateActiveCue)
  video.addEventListener('seeked', updateActiveCue)
  updateActiveCue()

  removeVideoListeners = () => {
    video.removeEventListener('timeupdate', updateActiveCue)
    video.removeEventListener('seeked', updateActiveCue)
  }
}

function canUpdate(token: number) {
  return isMounted && token === requestToken
}

function cleanupVideoListeners() {
  removeVideoListeners?.()
  removeVideoListeners = null
}

function resetLoadedSubtitles() {
  cleanupVideoListeners()
  cues.value = []
  activeText.value = ''
}

function restoreDisplayedSubtitles(
  token: number,
  previousCues: VttCue[],
  previousActiveText: string,
) {
  if (!canUpdate(token)) {
    return
  }

  cues.value = previousCues
  activeText.value = previousActiveText

  if (previousCues.length > 0) {
    bindVideo(token)
  }
}

function readableError(error: unknown) {
  return error instanceof Error ? error.message : '字幕操作失败'
}

type SubtitleUpdatedMessage = {
  type: 'lets-sub-it:subtitle-updated'
  videoId: string
}

function isSubtitleUpdatedMessage(message: unknown): message is SubtitleUpdatedMessage {
  if (typeof message !== 'object' || message === null) {
    return false
  }

  const candidate = message as Partial<SubtitleUpdatedMessage>
  return candidate.type === 'lets-sub-it:subtitle-updated' && typeof candidate.videoId === 'string'
}

type SettingsUpdatedMessage = {
  type: 'lets-sub-it:settings-updated'
  settings: Settings
}

function isSettingsUpdatedMessage(message: unknown): message is SettingsUpdatedMessage {
  if (typeof message !== 'object' || message === null) {
    return false
  }

  const candidate = message as Partial<SettingsUpdatedMessage>
  return candidate.type === 'lets-sub-it:settings-updated' && isSettings(candidate.settings)
}

function isSettings(value: unknown): value is Settings {
  if (typeof value !== 'object' || value === null) {
    return false
  }

  const candidate = value as Partial<Settings>
  return (
    typeof candidate.subtitleFontSizePx === 'number' &&
    candidate.subtitleFontSizePx > 0 &&
    SUBTITLE_MODES.includes(candidate.subtitleMode as SubtitleMode)
  )
}
</script>

<template>
  <div
    v-if="isWatchPage && enabled && hasSubtitle && activeText"
    class="lets-sub-it-subtitle-layer"
  >
    <div
      class="lets-sub-it-subtitle-text"
      :style="subtitleStyle"
    >
      {{ activeText }}
    </div>
  </div>
</template>
