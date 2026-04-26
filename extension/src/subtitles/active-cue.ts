import type { VttCue } from './vtt'

export function findActiveCue(
  cues: VttCue[],
  currentTime: number,
): VttCue | null {
  return cues.find((cue) => cue.start <= currentTime && currentTime < cue.end) ?? null
}
