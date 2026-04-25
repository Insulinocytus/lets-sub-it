import { describe, expect, it } from 'vitest'
import { findActiveCue } from './active-cue'
import type { VttCue } from './vtt'

const cues: VttCue[] = [
  { start: 0, end: 2, text: 'first' },
  { start: 2, end: 4, text: 'second' },
]

describe('findActiveCue', () => {
  it('returns the cue that contains the current time', () => {
    expect(findActiveCue(cues, 1)?.text).toBe('first')
  })

  it('treats cue end time as exclusive', () => {
    expect(findActiveCue(cues, 2)?.text).toBe('second')
  })

  it('returns null when no cue matches', () => {
    expect(findActiveCue(cues, 5)).toBeNull()
  })
})
