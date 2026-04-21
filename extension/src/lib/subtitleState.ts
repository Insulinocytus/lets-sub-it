import type { Job, LocalCacheEntry, SubtitleMode } from '../types';

export const DEFAULT_SUBTITLE_MODE: SubtitleMode = 'translated';

export function isTerminalJob(job: Job): boolean {
  return job.status === 'completed' || job.status === 'failed';
}

export function getSelectedMode(entry?: LocalCacheEntry): SubtitleMode {
  return entry?.selectedMode ?? DEFAULT_SUBTITLE_MODE;
}

export function getSubtitleUrlForMode(
  entry: LocalCacheEntry,
  mode = getSelectedMode(entry),
): string {
  return entry.subtitleUrls?.[mode] ?? entry.subtitleUrls?.translated ?? entry.subtitleUrls?.bilingual ?? '';
}
