import { browser } from 'wxt/browser'
import {
  ensurePersistedJobMonitors,
  handleJobMonitorAlarm,
} from '@/api/job-monitor'
import { handleExtensionMessage } from '@/api/message-handler'
import type { ExtensionMessage } from '@/api/messages'

export default defineBackground(() => {
  browser.alarms.onAlarm.addListener((alarm) => {
    void handleJobMonitorAlarm(alarm)
  })

  browser.runtime.onMessage.addListener((message, _sender, sendResponse) => {
    handleExtensionMessage(message as ExtensionMessage).then(sendResponse)
    return true
  })

  void ensurePersistedJobMonitors()
})
