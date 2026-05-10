const BUTTON_ATTR = 'data-lsi-subtitle-toggle'

export function injectPlayerButton(onToggle: () => void): void {
  const controls = document.querySelector('.ytp-right-controls')
  if (!controls) return

  const existing = controls.querySelector(`button[${BUTTON_ATTR}]`)
  if (existing) return

  const button = document.createElement('button')
  button.setAttribute(BUTTON_ATTR, '')
  button.className = 'ytp-button'
  button.title = 'Lets Sub It'
  button.setAttribute('aria-label', 'Lets Sub It 字幕')
  button.textContent = 'LS'
  button.style.cssText = `
    width: auto;
    min-width: 40px;
    font-size: 13px;
    font-weight: 600;
    color: #fff;
    text-shadow: 0 0 2px rgba(0,0,0,0.5);
  `
  button.addEventListener('click', (e) => {
    e.stopPropagation()
    e.preventDefault()
    onToggle()
  })

  controls.appendChild(button)
}

export function removePlayerButton(): void {
  const button = document.querySelector(`button[${BUTTON_ATTR}]`)
  button?.remove()
}
