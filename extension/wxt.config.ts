import { defineConfig } from 'wxt';

export default defineConfig({
  manifest: {
    permissions: ['storage', 'tabs'],
    host_permissions: ['http://localhost:8080/*', 'https://www.youtube.com/*'],
    web_accessible_resources: [
      {
        resources: ['page-bridge.js'],
        matches: ['https://www.youtube.com/*'],
      },
    ],
  },
});
