<div align="center">

# Lets Sub It

**Self-hosted YouTube subtitle generation and translation**

**English** | [简体中文](README.zh-CN.md)

</div>

Lets Sub It downloads the audio of a YouTube video, transcribes it with local Whisper or an OpenAI-compatible STT service, then translates it through an OpenAI-compatible API. A Chrome extension submits jobs and shows source-language, translated or bilingual subtitles on the video's watch page.

## What it does

- Transcribes locally with the bundled `hwdsl2/whisper-server` (faster-whisper) and produces `source.vtt`, `translated.vtt` and `bilingual.vtt`.
- Shows subtitles on the YouTube page in step with playback; after the popup closes, the background keeps tracking whether the job has finished.
- Stores jobs and subtitle assets in SQLite and reuses completed results.
- Runs the Go backend together with the Whisper transcription container via Docker Compose.

> [!IMPORTANT]
> The extension can currently connect only to a backend **on the same machine**, at `http://127.0.0.1` or `http://localhost` with an explicit port. Docker also binds to the local address by default; exposing the container port to your LAN does not let an extension on another machine connect remotely.

## Quick start

You need Docker (with Compose), Chrome/Chromium and a working OpenAI-compatible Chat Completions API. Run the following commands from the repository root.

1. Copy `.env.example` to `.env`. Replace the example values with a real `LSI_LLM_API_KEY` and `LSI_LLM_MODEL`; change `LSI_LOCAL_STT_API_KEY` to your own random key (shared by the local whisper container and the backend); set `LSI_LLM_BASE_URL` if you use another translation service. Keep `LSI_DOCKER_BIND_HOST=127.0.0.1` so the backend is reachable only from this machine.
2. Start the services:

   ```bash
   docker compose up -d --build
   ```

   The backend listens on `http://127.0.0.1:8080`. Run `docker compose ps` to check status; the whisper container downloads its model on first start (`small` is about 465 MB), which takes a while.

3. Install the extension: open a run of the **Extension Build** workflow in GitHub Actions, then download and unzip the `lets-sub-it-extension-chrome-<run-number>` artifact; or install [mise](https://mise.jdx.dev/) and build from the repository root:

   ```bash
   mise trust
   mise install
   task deps:extension
   task build:extension
   ```

   In `chrome://extensions`, enable Developer mode, click **Load unpacked** and select `extension/.output/chrome-mv3` (or the directory in the downloaded artifact that contains `manifest.json`).
4. Open a YouTube video, click the extension icon, choose the languages in the popup and submit a job. Once the job completes you can switch between source, translated and bilingual subtitles.

> [!NOTE]
> Submitting a job really downloads the video's audio, runs the local model and calls the translation API; the latter may incur costs. `.env` contains secrets; do not commit it.

Stop the services with `docker compose down`; SQLite data, subtitle files and the model cache are kept in Compose volumes.

## Local development

Use Go 1.22 and Node 22 from `mise.toml` (local transcription is provided by Docker). Run `mise trust`, `mise install` and `task setup`, then start the transcription container and the backend in one terminal and the extension in another:

```bash
export LSI_STT_API_KEY='replace-with-your-own-local-key'
docker run -d --rm -p 127.0.0.1:9000:9000 -v whisper-data:/var/lib/whisper \
  -e WHISPER_MODEL=small -e WHISPER_API_KEY="$LSI_STT_API_KEY" hwdsl2/whisper-server
LSI_STT_API_KEY="$LSI_STT_API_KEY" task dev:backend   # http://127.0.0.1:8080
# In another terminal: task dev:extension
```

The local backend also needs `yt-dlp` and `ffmpeg` on `PATH`. `task dev:backend` does **not** start whisper; it connects to `http://127.0.0.1:9000/v1` by default (`http://whisper:9000/v1` inside Compose). When running locally, provide `LSI_LLM_API_KEY` and `LSI_LLM_MODEL` in the backend process's environment; copying `.env` alone does not load configuration for a local process. The Taskfile's development commands use POSIX-style shell environment variable assignments.

## Using the HTTP API

The extension normally handles submission, polling and subtitle loading; you can also submit jobs directly to the local backend:

```bash
curl -X POST http://127.0.0.1:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{"youtubeUrl":"https://www.youtube.com/watch?v=dQw4w9WgXcQ","sourceLanguage":"en","targetLanguage":"zh"}'
```

Use the response's `job.id` to query progress. Replace `JOB_ID` below with the actual value, and download the VTT once the status is `completed`:

```bash
curl http://127.0.0.1:8080/jobs/JOB_ID
curl http://127.0.0.1:8080/subtitle-files/JOB_ID/bilingual
```

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/jobs` | Create or reuse a job; returns `job` and `reused` |
| `GET` | `/jobs/{jobId}` | Query status; includes the error message when `failed` |
| `GET` | `/jobs/active?videoId=...&targetLanguage=...` | Query the latest job for that video and language |
| `GET` | `/subtitle-assets?videoId=...&targetLanguage=...` | Query the generated subtitle asset |
| `GET` | `/subtitle-files/{jobId}/{mode}` | Download the `source`, `translated` or `bilingual` VTT |
| `DELETE` | `/subtitle-results?videoId=...&targetLanguage=...` | Delete the Subtitle Result for that video and language; returns `204` (also when there is nothing to delete) |

A job moves through `queued` → `downloading` → `transcribing` → `translating` → `packaging` → `completed`; an error at any stage moves it to `failed`.

Deletion is logical: job records stay in the database as history, but are no longer reused and no longer appear in any query; the working directory (audio and VTT) is removed. A job still running at deletion time finishes its current step and discards its output at its next status update. Submitting the same video and language afterwards runs the pipeline again.

`task api:smoke` runs one real end-to-end check against a running backend: it first deletes the `en → zh` Subtitle Result of the test video (by default the 19-second `jNQXAC9IVRw`, overridable with `LSI_SMOKE_VIDEO_ID`; set the backend address with `LSI_BACKEND_URL`), then submits a job, waits for it to complete, and prints every stage and the beginning of all three VTT files; it exits non-zero on failure or if the job has not completed within 10 minutes. It calls the real YouTube, STT and LLM, and wipes that video's existing subtitles.

## Architecture and configuration

```text
Chrome extension (popup / background / YouTube content script)
             │ submit jobs, poll status, fetch VTT
             ▼
Go HTTP API ── SQLite (jobs, subtitle assets)
      │
      ├── yt-dlp / ffmpeg download audio
      ├── Whisper HTTP service transcribes (docker-whisper) → source.vtt generated locally
      └── OpenAI-compatible translation API → translated.vtt / bilingual.vtt
```

| Directory | Responsibility |
| --- | --- |
| [`backend/`](backend/) | Go API, SQLite state, download/transcription/translation job orchestration, WebVTT generation |
| [`extension/`](extension/) | WXT/Vue Chrome MV3 extension, job monitoring and YouTube subtitle rendering |

Backend jobs do not depend on the originating HTTP connection; after a service restart, unfinished jobs are marked failed and must be resubmitted. Transcription is a single synchronous `POST /audio/transcriptions` call (`response_format=verbose_json`); the backend requires valid `segments` (each with `start`, `end`, `text`) before writing `source.vtt`. There is no chunking, no queue and no fallback to another service. After the extension popup closes, the background keeps monitoring jobs with browser alarms.

Common settings (see [`.env.example`](.env.example) for the full list and Docker defaults; see [`backend/internal/app/config.go`](backend/internal/app/config.go) for non-Docker backend defaults):

| Variable | Purpose |
| --- | --- |
| `LSI_LLM_API_KEY`, `LSI_LLM_MODEL` | Translation service credentials and model; required for real translation |
| `LSI_LLM_BASE_URL` | OpenAI-compatible service URL; defaults to `https://api.openai.com` |
| `LSI_LOCAL_STT_API_KEY` | Key for the whisper container in Compose; the backend also falls back to it when `LSI_STT_API_KEY` is not set |
| `WHISPER_MODEL`, `WHISPER_COMPUTE_TYPE` | Model and compute type of the local whisper container (default `small` / `int8`) |
| `LSI_STT_BASE_URL`, `LSI_STT_API_KEY`, `LSI_STT_MODEL`, `LSI_STT_TIMEOUT` | Backend transcription service settings; leave empty to use the whisper container in Compose, or point at any OpenAI-compatible STT service |
| `LSI_DOCKER_BIND_HOST` | Bind address for the Compose backend port; `127.0.0.1` in the example |
| `LSI_DB_PATH`, `LSI_WORK_DIR` | Backend database and job file paths |

Transcription and LLM settings are independent: without `LSI_STT_*` the backend uses the bundled whisper container; to use an external service (such as OpenRouter), set a non-empty `LSI_STT_BASE_URL` and `LSI_STT_API_KEY` in `.env`, and optionally `LSI_STT_MODEL`. The external key is injected only into the backend container; it is not passed to the local whisper container and does not affect the `LSI_LLM_*` translation settings.

## Testing and building

The root [`Taskfile.yml`](Taskfile.yml) defines cross-module commands; if `task` is not on `PATH`, use `mise exec -- task <name>`.

| Command | Purpose |
| --- | --- |
| `task check` | All tests + extension typecheck |
| `task test:backend` / `task test:extension` | Run Go / Vitest alone |
| `task typecheck` | Extension TypeScript/Vue typecheck |
| `task build` | Build all modules |
| `task build:extension` | Output to `extension/.output/chrome-mv3` |
| `task --list` | List other tasks, including Docker start/stop and logs |

Go tests sit next to the source; extension tests sit next to their components. Automated tests use substitutes and need no real model, YouTube download or LLM key.

## Troubleshooting

- **Backend fails to start:** install `yt-dlp` and `ffmpeg` locally and add them to `PATH`, or use the Docker image.
- **Transcription does not start:** check that the whisper container has passed its health check (`docker compose ps`); when running locally, check that `LSI_STT_BASE_URL` points at a running service. The model download on first start may be slow.
- **Translation fails:** check the running process's `LSI_LLM_API_KEY`, `LSI_LLM_MODEL` and compatible API URL.
- **Extension cannot connect:** the backend URL must be a local HTTP address with a port; neither the current manifest permissions nor the client validation support remote hosts.
