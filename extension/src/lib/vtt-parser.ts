import type { SubtitleCue, VttParseResult } from '@/types'

function parseTimestamp(ts: string): number {
  const parts = ts.split(':')
  if (parts.length === 3) {
    const [h, m, s] = parts
    return parseInt(h) * 3600 + parseInt(m) * 60 + parseFloat(s)
  }
  // mm:ss.xxx
  return parseInt(parts[0]) * 60 + parseFloat(parts[1])
}

export function parseVtt(vttContent: string): VttParseResult {
  const lines = vttContent.split('\n')
  const cues: SubtitleCue[] = []

  let i = 0
  // Skip WEBVTT header and blank lines
  while (i < lines.length && !lines[i].includes('-->')) {
    i++
  }

  while (i < lines.length) {
    // Find a timing line
    const timingLine = lines[i]
    if (!timingLine.includes('-->')) {
      i++
      continue
    }

    const [startStr, endStr] = timingLine.split(' --> ')
    const start = parseTimestamp(startStr.trim())
    const end = parseTimestamp(endStr.trim().split(/\s+/)[0])

    i++
    // Collect cue text lines (until blank line or end)
    const textLines: string[] = []
    while (i < lines.length && lines[i].trim() !== '') {
      textLines.push(lines[i].trim())
      i++
    }

    if (textLines.length > 0) {
      cues.push({
        start,
        end,
        text: textLines.join('\n'),
      })
    }

    i++
  }

  return { cues }
}

export function findCueAtTime(cues: SubtitleCue[], time: number): SubtitleCue | null {
  return cues.find((cue) => time >= cue.start && time < cue.end) ?? null
}
