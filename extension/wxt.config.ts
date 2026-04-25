import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'wxt'

export default defineConfig({
  srcDir: 'src',
  entrypointsDir: '../entrypoints',
  manifest: {
    name: 'Lets Sub It',
    description: 'Self-hosted YouTube subtitle generation and translation helper.',
    version: '0.1.0',
    permissions: ['storage', 'activeTab'],
    host_permissions: ['http://127.0.0.1:*/*', 'http://localhost:*/*'],
  },
  vite: () => ({
    plugins: [vue(), tailwindcss()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
  }),
})
