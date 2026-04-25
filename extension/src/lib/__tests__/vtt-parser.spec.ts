import { describe, it, expect } from 'vitest'
import { parseVtt } from '../vtt-parser'

describe('parseVtt', () => {
  it('parses WEBVTT header and cues', () => {
    const vtt = `WEBVTT

00:00:01.000 --> 00:00:03.500
Hello world

00:00:04.000 --> 00:00:06.000
Second cue
`
    const result = parseVtt(vtt)
    expect(result.cues).toHaveLength(2)
    expect(result.cues[0]).toEqual({ start: 1, end: 3.5, text: 'Hello world' })
    expect(result.cues[1]).toEqual({ start: 4, end: 6, text: 'Second cue' })
  })

  it('handles cue text with multiple lines', () => {
    const vtt = `WEBVTT

00:00:00.000 --> 00:00:02.000
Line one
Line two
`
    const result = parseVtt(vtt)
    expect(result.cues).toHaveLength(1)
    expect(result.cues[0].text).toBe('Line one\nLine two')
  })

  it('returns empty cues for empty VTT', () => {
    const result = parseVtt('WEBVTT\n\n')
    expect(result.cues).toHaveLength(0)
  })

  it('ignores cue identifiers', () => {
    const vtt = `WEBVTT

1
00:00:00.000 --> 00:00:01.000
Text
`
    const result = parseVtt(vtt)
    expect(result.cues).toHaveLength(1)
    expect(result.cues[0].text).toBe('Text')
  })

  it('parses timestamps with hours', () => {
    const vtt = `WEBVTT

01:02:03.000 --> 01:02:05.500
Long video
`
    const result = parseVtt(vtt)
    expect(result.cues[0].start).toBe(3723)
    expect(result.cues[0].end).toBe(3725.5)
  })

  it('parses bilingual VTT with two text lines per cue', () => {
    const vtt = `WEBVTT

00:00:00.000 --> 00:00:02.000
source text
翻译文本
`
    const result = parseVtt(vtt)
    expect(result.cues).toHaveLength(1)
    expect(result.cues[0].text).toBe('source text\n翻译文本')
  })
})
