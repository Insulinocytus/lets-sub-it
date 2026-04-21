import { useEffect, useState } from 'react';

import { ModeToggle } from '../../src/components/ModeToggle';
import { ProgressList } from '../../src/components/ProgressList';
import { createJob, getJobs, getSubtitleAsset } from '../../src/lib/api';
import { DEFAULT_SUBTITLE_MODE, isTerminalJob } from '../../src/lib/subtitleState';
import { getCache, saveCacheEntry, setSelectedMode } from '../../src/lib/storage';
import { extractVideoId } from '../../src/lib/youtube';
import type { Job, LocalCacheEntry, SubtitleMode } from '../../src/types';

function sortJobs(cache: Record<string, LocalCacheEntry>): Job[] {
  return Object.values(cache)
    .sort((left, right) => right.lastSyncedAt.localeCompare(left.lastSyncedAt))
    .flatMap((entry) => (entry.recentJob ? [entry.recentJob] : []));
}

export default function App() {
  const [youtubeUrl, setYoutubeUrl] = useState('');
  const [targetLanguage, setTargetLanguage] = useState('zh-CN');
  const [jobs, setJobs] = useState<Job[]>([]);
  const [cache, setCache] = useState<Record<string, LocalCacheEntry>>({});
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function loadCachedJobs() {
      const cachedEntries = await getCache();

      if (!cancelled) {
        setCache(cachedEntries);
        setJobs(sortJobs(cachedEntries));
      }
    }

    void loadCachedJobs();

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    const activeJobs = jobs.filter((job) => !isTerminalJob(job));
    if (activeJobs.length === 0) {
      return undefined;
    }

    const id = window.setInterval(async () => {
      try {
        const refreshedJobs = await getJobs(activeJobs.map((job) => job.id));
        const subtitleAssets = await Promise.all(
          refreshedJobs.map(async (job) => {
            if (job.status !== 'completed') {
              return undefined;
            }

            try {
              return await getSubtitleAsset(job.videoId);
            } catch (assetError) {
              const message =
                assetError instanceof Error
                  ? assetError.message
                  : 'Failed to load subtitle asset.';
              setError(message);
              return undefined;
            }
          }),
        );

        setCache((current) => {
          const nextCache = { ...current };
          const refreshedAt = new Date().toISOString();

          for (const [index, job] of refreshedJobs.entries()) {
            const existing = nextCache[job.videoId];
            const subtitleAsset = subtitleAssets[index];
            const nextEntry: LocalCacheEntry = {
              videoId: job.videoId,
              jobId: job.id,
              selectedMode: existing?.selectedMode ?? DEFAULT_SUBTITLE_MODE,
              lastSyncedAt: refreshedAt,
              subtitleUrls: subtitleAsset?.subtitleUrls ?? existing?.subtitleUrls,
              recentJob: job,
            };

            nextCache[job.videoId] = nextEntry;
            void saveCacheEntry(nextEntry);
          }

          setJobs(sortJobs(nextCache));
          return nextCache;
        });
      } catch (pollError) {
        const message =
          pollError instanceof Error ? pollError.message : 'Failed to refresh jobs.';
        setError(message);
      }
    }, 2000);

    return () => window.clearInterval(id);
  }, [jobs]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError('');

    const videoId = extractVideoId(youtubeUrl);
    if (!videoId) {
      setError('Enter a valid YouTube watch URL.');
      return;
    }

    setIsSubmitting(true);

    try {
      const job = await createJob({ youtubeUrl, targetLanguage, videoId });
      const nextEntry: LocalCacheEntry = {
        videoId,
        jobId: job.id,
        selectedMode: DEFAULT_SUBTITLE_MODE,
        lastSyncedAt: new Date().toISOString(),
        recentJob: job,
      };
      await saveCacheEntry(nextEntry);
      setCache((current) => ({
        ...current,
        [videoId]: nextEntry,
      }));
      setJobs((current) => [job, ...current]);
      setYoutubeUrl('');
    } catch (submitError) {
      const message =
        submitError instanceof Error ? submitError.message : 'Failed to create job.';
      setError(message);
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleModeChange(videoId: string, mode: SubtitleMode) {
    const updated = await setSelectedMode(videoId, mode);
    if (!updated) {
      return;
    }

    setCache((current) => ({
      ...current,
      [videoId]: updated,
    }));

    if (chrome.runtime?.sendMessage) {
      void chrome.runtime.sendMessage({
        type: 'subtitle-cache:set-mode',
        videoId,
        mode,
      });
    }
  }

  return (
    <main>
      <form onSubmit={handleSubmit}>
        <label>
          YouTube URL
          <input
            value={youtubeUrl}
            onChange={(event) => setYoutubeUrl(event.target.value)}
            placeholder="https://www.youtube.com/watch?v="
          />
        </label>
        <label>
          Target language
          <input
            value={targetLanguage}
            onChange={(event) => setTargetLanguage(event.target.value)}
          />
        </label>
        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Submitting...' : 'Process'}
        </button>
      </form>
      {error ? <p>{error}</p> : null}
      <ProgressList jobs={jobs} />
      {Object.values(cache)
        .filter((entry) => entry.recentJob)
        .map((entry) => (
          <section key={entry.videoId}>
            <p>Mode for {entry.videoId}</p>
            <ModeToggle
              mode={entry.selectedMode}
              onChange={(mode) => void handleModeChange(entry.videoId, mode)}
            />
          </section>
        ))}
    </main>
  );
}
