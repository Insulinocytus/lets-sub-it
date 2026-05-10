import '@/youtube/youtube-content.css'
import { createApp } from 'vue'
import YoutubeOverlay from '@/content/YoutubeOverlay.vue'
import {
  cleanupPlayerOverlayHost,
  ensurePlayerOverlayHost,
  findYouTubePlayer,
  mountSubtitleToggleButton,
  removeSubtitleToggleButton,
  type PlayerOverlayHost,
} from '@/youtube/player-ui'

export default defineContentScript({
  matches: ['https://www.youtube.com/watch*'],
  cssInjectionMode: 'manifest',
  main(ctx) {
    let mountedHost: PlayerOverlayHost | null = null
    let observer: MutationObserver | null = null

    const dispatchToggle = () => {
      window.dispatchEvent(new CustomEvent('lets-sub-it:toggle-subtitles'))
    }

    const mount = () => {
      const host = ensurePlayerOverlayHost(findYouTubePlayer())
      if (host && host !== mountedHost) {
        cleanupPlayerOverlayHost(mountedHost)
        mountedHost = host
        const app = createApp(YoutubeOverlay)
        app.mount(host)
        host.__letsSubItCleanup = () => app.unmount()
      }
      mountSubtitleToggleButton(true, dispatchToggle)
    }

    mount()

    observer = new MutationObserver(() => mount())
    observer.observe(document.documentElement, { childList: true, subtree: true })

    ctx.onInvalidated(() => {
      observer?.disconnect()
      observer = null
      removeSubtitleToggleButton()
      cleanupPlayerOverlayHost(mountedHost)
      mountedHost = null
    })
  },
})
