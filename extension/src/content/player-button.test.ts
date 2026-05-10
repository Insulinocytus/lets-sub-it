import { beforeEach, describe, expect, it, vi } from 'vitest'
import { injectPlayerButton, removePlayerButton } from './player-button'

describe('player-button', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
  })

  it('injects a button into ytp-right-controls', () => {
    const controls = document.createElement('div')
    controls.className = 'ytp-right-controls'
    document.body.appendChild(controls)

    injectPlayerButton(() => {})

    const button = controls.querySelector('button[data-lsi-subtitle-toggle]')
    expect(button).not.toBeNull()
    expect(button?.textContent).toBe('LS')
  })

  it('does not inject a second button when one already exists', () => {
    const existing = document.createElement('button')
    existing.setAttribute('data-lsi-subtitle-toggle', '')
    const controls = document.createElement('div')
    controls.className = 'ytp-right-controls'
    controls.appendChild(existing)
    document.body.appendChild(controls)

    injectPlayerButton(() => {})

    expect(controls.querySelectorAll('button[data-lsi-subtitle-toggle]').length).toBe(1)
  })

  it('calls the toggle callback on click', () => {
    const controls = document.createElement('div')
    controls.className = 'ytp-right-controls'
    document.body.appendChild(controls)

    const toggle = vi.fn()
    injectPlayerButton(toggle)

    const button = controls.querySelector('button[data-lsi-subtitle-toggle]')!
    ;(button as HTMLButtonElement).click()
    expect(toggle).toHaveBeenCalledOnce()
  })

  it('removePlayerButton removes the injected button', () => {
    const controls = document.createElement('div')
    controls.className = 'ytp-right-controls'
    document.body.appendChild(controls)

    injectPlayerButton(() => {})
    expect(controls.querySelector('button[data-lsi-subtitle-toggle]')).not.toBeNull()

    removePlayerButton()
    expect(document.body.querySelector('button[data-lsi-subtitle-toggle]')).toBeNull()
  })

  it('injectPlayerButton is a no-op when controls bar is missing', () => {
    expect(() => injectPlayerButton(() => {})).not.toThrow()
  })
})
