import type { Job } from '../types';

export function ProgressList({ jobs }: { jobs: Job[] }) {
  if (jobs.length === 0) {
    return <p>No jobs yet.</p>;
  }

  return (
    <ul>
      {jobs.map((job) => (
        <li key={job.id}>
          {job.videoId} · {job.stage} · {job.progress}%
        </li>
      ))}
    </ul>
  );
}
