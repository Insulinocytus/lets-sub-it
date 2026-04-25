export default defineContentScript({
  matches: ['https://www.youtube.com/watch*'],
  main() {
    console.info('Lets Sub It YouTube content script loaded')
  },
})
