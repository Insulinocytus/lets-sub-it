import { describe, expect, it } from 'vitest'
import { parseVtt } from './vtt'

describe('parseVtt', () => {
  it('parses WebVTT cues', () => {
    const cues = parseVtt(`WEBVTT

00:00:00.000 --> 00:00:02.000
Hello

00:00:02.000 --> 00:00:04.500
World
`)

    expect(cues).toEqual([
      { start: 0, end: 2, text: 'Hello' },
      { start: 2, end: 4.5, text: 'World' },
    ])
  })

  it('keeps multiline cue text', () => {
    const cues = parseVtt(`WEBVTT

00:00:00.000 --> 00:00:02.000
Hello
你好
`)

    expect(cues[0].text).toBe('Hello\n你好')
  })

  it('throws when no cues are present', () => {
    expect(() => parseVtt('WEBVTT\n')).toThrow('VTT contains no cues')
  })
})
