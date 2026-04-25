import '@/style.css'
import { createApp } from 'vue'
import { createShadowRootUi } from 'wxt/utils/content-script-ui/shadow-root'
import YoutubeOverlay from '@/content/YoutubeOverlay.vue'

export default defineContentScript({
  matches: ['https://www.youtube.com/*'],
  cssInjectionMode: 'ui',
  async main(ctx) {
    const ui = await createShadowRootUi(ctx, {
      name: 'lets-sub-it-youtube-ui',
      position: 'inline',
      anchor: 'body',
      onMount: (container) => {
        const app = createApp(YoutubeOverlay)
        app.mount(container)
        return app
      },
      onRemove: (app) => {
        app?.unmount()
      },
    })

    ui.mount()
  },
})
