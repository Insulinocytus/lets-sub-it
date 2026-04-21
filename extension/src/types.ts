export type JobStatus = 'queued' | 'running' | 'completed' | 'failed';

export type JobStage =
  | 'queued'
  | 'downloading'
  | 'transcribing'
  | 'translating'
  | 'packaging'
  | 'completed'
  | 'failed';

export type SubtitleMode = 'translated' | 'bilingual';

export interface SubtitleUrls {
  translated?: string;
  bilingual?: string;
}

export interface CreateJobInput {
  youtubeUrl: string;
  targetLanguage: string;
  videoId: string;
}

export interface Job {
  id: string;
  videoId: string;
  youtubeUrl: string;
  targetLanguage: string;
  status: JobStatus;
  stage: JobStage;
  progress: number;
  errorMessage: string;
}

export interface LocalCacheEntry {
  videoId: string;
  jobId: string;
  selectedMode: SubtitleMode;
  lastSyncedAt: string;
  subtitleUrls?: SubtitleUrls;
  recentJob?: Job;
}

export interface SubtitleLoadPayload {
  videoId: string;
  mode: SubtitleMode;
  subtitleUrl: string;
}
