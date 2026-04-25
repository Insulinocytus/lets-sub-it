import { browser } from 'wxt/browser'
import { handleExtensionMessage } from '@/api/message-handler'
import type { ExtensionMessage } from '@/api/messages'

export default defineBackground(() => {
  browser.runtime.onMessage.addListener((message, _sender, sendResponse) => {
    handleExtensionMessage(message as ExtensionMessage).then(sendResponse)
    return true
  })
})
