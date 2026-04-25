import { describe, it, expect } from 'vitest'
import { findCueAtTime } from '../vtt-parser'
import type { SubtitleCue } from '@/types'

describe('findCueAtTime', () => {
  const cues: SubtitleCue[] = [
    { start: 0, end: 2, text: 'First' },
    { start: 2, end: 5, text: 'Second' },
    { start: 5, end: 10, text: 'Third' },
  ]

  it('finds cue by current time', () => {
    expect(findCueAtTime(cues, 0)).toEqual(cues[0])
    expect(findCueAtTime(cues, 1.5)).toEqual(cues[0])
    expect(findCueAtTime(cues, 2)).toEqual(cues[1])
    expect(findCueAtTime(cues, 7)).toEqual(cues[2])
  })

  it('returns null for time before first cue', () => {
    expect(findCueAtTime(cues, -1)).toBeNull()
  })

  it('returns null for time after last cue', () => {
    expect(findCueAtTime(cues, 10)).toBeNull()
  })

  it('returns null for empty cues array', () => {
    expect(findCueAtTime([], 5)).toBeNull()
  })
})
