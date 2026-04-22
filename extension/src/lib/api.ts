import type { CreateJobInput, Job, SubtitleUrls } from '../types';

const API_BASE = 'http://localhost:8080';

async function parseJobResponse(response: Response): Promise<Job> {
  if (response.ok) {
    return response.json() as Promise<Job>;
  }

  let message = `Request failed with status ${response.status}.`;

  try {
    const payload = (await response.json()) as { error?: string; message?: string };
    message = payload.error ?? payload.message ?? message;
  } catch {
    // Fall back to the HTTP status message when the error body is not JSON.
  }

  throw new Error(message);
}

export async function createJob(input: CreateJobInput): Promise<Job> {
  try {
    const response = await fetch(`${API_BASE}/api/jobs`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input),
    });

    return await parseJobResponse(response);
  } catch (error) {
    if (error instanceof Error) {
      throw new Error(`Failed to create job: ${error.message}`);
    }

    throw new Error('Failed to create job.');
  }
}

export async function getJob(id: string): Promise<Job> {
  try {
    const response = await fetch(`${API_BASE}/api/jobs/${id}`);
    return await parseJobResponse(response);
  } catch (error) {
    if (error instanceof Error) {
      throw new Error(`Failed to load job: ${error.message}`);
    }

    throw new Error('Failed to load job.');
  }
}

export async function getJobs(ids: string[]): Promise<Job[]> {
  return Promise.all(ids.map((id) => getJob(id)));
}

export interface SubtitleAsset {
  jobId: string;
  videoId: string;
  sourceLanguage: string;
  targetLanguage: string;
  subtitleUrls: SubtitleUrls;
}

export async function getSubtitleAsset(jobId: string): Promise<SubtitleAsset> {
  try {
    const response = await fetch(`${API_BASE}/api/jobs/${jobId}/subtitles`);
    return (await parseJobResponse(response)) as SubtitleAsset;
  } catch (error) {
    if (error instanceof Error) {
      throw new Error(`Failed to load subtitle asset: ${error.message}`);
    }

    throw new Error('Failed to load subtitle asset.');
  }
}
