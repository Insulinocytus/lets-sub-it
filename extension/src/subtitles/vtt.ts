export type VttCue = {
  start: number
  end: number
  text: string
}

const TIMING_RE =
  /^(?<start>\d{2}:\d{2}:\d{2}\.\d{3})\s+-->\s+(?<end>\d{2}:\d{2}:\d{2}\.\d{3})/

export function parseVtt(input: string): VttCue[] {
  const lines = input.replace(/\r\n/g, '\n').split('\n')
  const cues: VttCue[] = []
  let index = 0

  while (index < lines.length) {
    const line = lines[index].trim()
    const match = TIMING_RE.exec(line)

    if (!match?.groups) {
      index += 1
      continue
    }

    const textLines: string[] = []
    index += 1
    while (index < lines.length && lines[index].trim() !== '') {
      textLines.push(lines[index])
      index += 1
    }

    const text = textLines.join('\n').trim()
    if (text) {
      cues.push({
        start: parseTimestamp(match.groups.start),
        end: parseTimestamp(match.groups.end),
        text,
      })
    }
  }

  if (cues.length === 0) {
    throw new Error('VTT contains no cues')
  }

  return cues
}

function parseTimestamp(value: string): number {
  const [hours, minutes, secondsWithMillis] = value.split(':')
  const [seconds, millis] = secondsWithMillis.split('.')
  return (
    Number(hours) * 3600 +
    Number(minutes) * 60 +
    Number(seconds) +
    Number(millis) / 1000
  )
}
