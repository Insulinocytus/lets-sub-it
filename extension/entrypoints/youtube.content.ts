import '@/style.css'
import { createApp } from 'vue'
import { browser } from 'wxt/browser'
import YoutubeOverlay from '@/content/YoutubeOverlay.vue'
import { injectPlayerButton, removePlayerButton } from '@/content/player-button'

export default defineContentScript({
  matches: ['https://www.youtube.com/watch*'],
  cssInjectionMode: 'ui',
  async main(ctx) {
    let app: ReturnType<typeof createApp> | null = null
    let shadowHost: HTMLDivElement | null = null

    function mountOverlay() {
      const playerContainer = document.querySelector('#movie_player')
      if (!playerContainer) {
        setTimeout(mountOverlay, 500)
        return
      }

      const existing = document.getElementById('lsi-subtitle-host')
      if (existing) return

      shadowHost = document.createElement('div')
      shadowHost.id = 'lsi-subtitle-host'
      shadowHost.style.cssText = `
        position: absolute;
        inset: 0;
        pointer-events: none;
        z-index: 20;
        overflow: visible;
      `
      const shadowRoot = shadowHost.attachShadow({ mode: 'open' })

      const container = document.createElement('div')
      container.style.cssText = `
        position: absolute;
        inset: 0;
        pointer-events: none;
        overflow: visible;
      `
      shadowRoot.appendChild(container)

      app = createApp(YoutubeOverlay)
      app.mount(container)

      playerContainer.appendChild(shadowHost)
    }

    function unmountOverlay() {
      app?.unmount()
      app = null
      shadowHost?.remove()
      shadowHost = null
      removePlayerButton()
    }

    function toggleSubtitle() {
      if (!app) return
      const instance = app._instance
      if (!instance?.exposed?.toggleEnabled) return
      instance.exposed.toggleEnabled()
    }

    mountOverlay()
    injectPlayerButton(toggleSubtitle)

    window.addEventListener('yt-navigate-finish', () => {
      unmountOverlay()
      mountOverlay()
      injectPlayerButton(toggleSubtitle)
    })

    ctx.onInvalidated(() => {
      unmountOverlay()
    })
  },
})
