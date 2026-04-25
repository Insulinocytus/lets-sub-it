<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
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
const cues = ref<VttCue[]>([])
const activeText = ref('')

let removeVideoListeners: (() => void) | null = null
let removeVideoIdWatch: (() => void) | null = null

const hasSubtitle = computed(() => cues.value.length > 0)

onMounted(() => {
  const videoId = getCurrentVideoId()
  currentVideoId.value = videoId
  void loadForVideo(videoId)

  removeVideoIdWatch = watchVideoIdChanges((nextVideoId) => {
    currentVideoId.value = nextVideoId
    void loadForVideo(nextVideoId)
  })
})

onUnmounted(() => {
  cleanupVideoListeners()
  removeVideoIdWatch?.()
  removeVideoIdWatch = null
})

async function loadForVideo(videoId: string | null) {
  cleanupVideoListeners()
  currentVideoId.value = videoId
  currentAsset.value = null
  cues.value = []
  activeText.value = ''

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
    status.value = readableError(error)
    return
  }
  if (currentVideoId.value !== videoId) {
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
  selectedMode.value = result.data.selectedMode
  await loadVtt()
}

async function loadVtt() {
  const asset = currentAsset.value
  if (!asset) {
    return
  }

  status.value = '加载字幕'
  let result
  try {
    result = await sendExtensionMessage<string>({
      type: 'subtitle:fetch-file',
      payload: { jobId: asset.jobId, mode: selectedMode.value },
    })
  } catch (error) {
    status.value = readableError(error)
    return
  }
  if (currentAsset.value?.jobId !== asset.jobId) {
    return
  }

  if (!result.ok) {
    status.value = result.error.message
    return
  }

  try {
    cues.value = parseVtt(result.data)
  } catch {
    cues.value = []
    activeText.value = ''
    status.value = '字幕解析失败'
    cleanupVideoListeners()
    return
  }

  status.value = '字幕已加载'
  bindVideo()
}

async function changeMode(mode: SubtitleMode) {
  if (selectedMode.value === mode) {
    return
  }

  selectedMode.value = mode
  const asset = currentAsset.value
  if (!asset) {
    return
  }

  const result = await sendExtensionMessage<SubtitleAssetCacheEntry | null>({
    type: 'subtitle:update-mode',
    payload: {
      videoId: asset.videoId,
      targetLanguage: asset.targetLanguage,
      mode,
    },
  })

  if (!result.ok) {
    status.value = result.error.message
    return
  }

  if (result.data) {
    currentAsset.value = result.data
  }
  await loadVtt()
}

function bindVideo() {
  cleanupVideoListeners()

  const video = document.querySelector('video')
  if (!video) {
    status.value = '未找到视频'
    return
  }

  const updateActiveCue = () => {
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

function cleanupVideoListeners() {
  removeVideoListeners?.()
  removeVideoListeners = null
}

function handleModeClick(mode: SubtitleMode) {
  void changeMode(mode).catch((error: unknown) => {
    status.value = readableError(error)
  })
}

function readableError(error: unknown) {
  return error instanceof Error ? error.message : '字幕操作失败'
}
</script>

<template>
  <div class="pointer-events-none fixed inset-x-0 bottom-7 z-[2147483647] flex justify-center px-4">
    <div class="flex max-w-[min(720px,calc(100vw-32px))] flex-col items-center gap-2">
      <div class="pointer-events-auto flex items-center gap-1.5 rounded-md border border-white/15 bg-black/70 px-2 py-1.5 text-white shadow-lg backdrop-blur">
        <Button
          type="button"
          size="sm"
          :variant="enabled ? 'secondary' : 'ghost'"
          class="h-7 px-2 text-xs"
          @click="enabled = !enabled"
        >
          {{ enabled ? '字幕开' : '字幕关' }}
        </Button>
        <Button
          type="button"
          size="sm"
          :variant="selectedMode === 'translated' ? 'secondary' : 'ghost'"
          class="h-7 px-2 text-xs text-white hover:text-white"
          :disabled="!currentAsset"
          @click="handleModeClick('translated')"
        >
          翻译
        </Button>
        <Button
          type="button"
          size="sm"
          :variant="selectedMode === 'bilingual' ? 'secondary' : 'ghost'"
          class="h-7 px-2 text-xs text-white hover:text-white"
          :disabled="!currentAsset"
          @click="handleModeClick('bilingual')"
        >
          双语
        </Button>
        <Badge
          variant="outline"
          class="border-white/20 bg-white/10 text-[11px] leading-5 text-white"
        >
          {{ status }}
        </Badge>
      </div>

      <div
        v-if="enabled && hasSubtitle && activeText"
        class="pointer-events-auto max-w-full whitespace-pre-line rounded-md bg-black/78 px-4 py-2 text-center text-xl font-semibold leading-snug text-white shadow-lg [text-shadow:0_1px_2px_rgb(0_0_0/0.85)]"
      >
        {{ activeText }}
      </div>
    </div>
  </div>
</template>
