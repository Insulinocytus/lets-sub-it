import { defineConfig } from 'wxt'

export default defineConfig({
  srcDir: 'src',
  extensionApi: 'chrome',
  modules: ['@wxt-dev/module-vue'],
  manifest: {
    name: "Let's Sub It",
    version: '0.1.0',
    permissions: ['storage', 'tabs'],
    host_permissions: [
      'http://localhost:8080/*',
      'http://127.0.0.1:8080/*',
      'https://www.youtube.com/*',
    ],
    action: {
      default_title: "Let's Sub It",
      default_popup: 'entrypoints/popup/index.html',
    },
  },
})
