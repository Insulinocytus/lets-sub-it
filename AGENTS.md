# AGENTS.md

## Project Overview

Lets Sub It is a self-hosted YouTube subtitle generation and translation project. The backend downloads YouTube audio, sends audio to a local Whisper HTTP service for WebVTT transcription, translates subtitles through an OpenAI-compatible chat completions API, persists jobs and assets in SQLite, and serves results to a Chrome extension.

Architecture:

```text
Chrome MV3 extension -> Go backend HTTP API -> yt-dlp + ffmpeg
                                  |          -> Whisper HTTP service
                                  |          -> OpenAI-compatible LLM
                                  v
                           SQLite + WebVTT files
```

Main modules:

| Path | Purpose | Stack |
| --- | --- | --- |
| `backend/` | HTTP API, job lifecycle, SQLite persistence, audio download, Whisper/LLM orchestration | Go 1.22, GORM, SQLite |
| `whisper/` | Local transcription HTTP service that accepts uploaded audio and returns validated VTT | Python 3.12, FastAPI, uvicorn, faster-whisper, uv |
| `extension/` | Chrome Manifest V3 extension for job submission and YouTube watch-page subtitle rendering | WXT, Vue 3, TypeScript, Tailwind CSS, shadcn-vue, Vitest |

This repository is not a package-manager monorepo. Use the root `Taskfile.yml` as the canonical cross-module command interface. Tool versions are pinned in `mise.toml`.

## Tooling Rules

- Run commands from the repository root unless a command explicitly says to run from a module directory.
- Prefer `task <name>` for repeatable project workflows.
- If `task` is not on `PATH`, run the same command through mise, for example `mise exec -- task <name>`.
- Use `mise exec -- <command>` for direct module commands so the pinned toolchain is used.
- Run `mise trust` once if mise refuses to use this repository's config.
- Do not edit generated or local artifact directories such as `extension/.output/`, `whisper/.venv/`, `whisper/.pytest_cache/`, `whisper/dist/`, `backend/server`, or Python `__pycache__/` directories unless the task explicitly asks for release artifacts.
- Do not commit `.env` or real credentials. Start from `.env.example` for local Docker/backend configuration.

Pinned tools in `mise.toml`:

| Tool | Version |
| --- | --- |
| Python | `3.12` |
| Go | `1.22` |
| Node.js | `22` |
| uv | `latest` |
| Task | `latest` |
| actionlint | `latest` |

## Setup Commands

- Install pinned tools and all module dependencies: `task setup`
- Install only pinned tools: `task tools`
- Install backend dependencies: `task deps:backend`
- Install Whisper service dependencies: `task deps:whisper`
- Install extension dependencies: `task deps:extension`

Manual equivalents:

```bash
mise install
cd backend && mise exec -- go mod download
cd whisper && mise exec -- uv sync --dev
cd extension && mise exec -- npm install
```

For Docker/self-hosted runs, create local environment configuration:

```bash
cp .env.example .env
```

At minimum, set real translation credentials before running full translation jobs:

```env
LSI_LLM_API_KEY=sk-your-key-here
LSI_LLM_MODEL=gpt-4.1-mini
```

## Development Workflow

Start local services:

- Start Whisper HTTP service: `task dev:whisper`
- Start backend API: `task dev:backend`
- Start extension dev server: `task dev:extension`

Local backend development expects the Whisper service to be running at `LSI_WHISPER_BASE_URL`, which defaults to `http://127.0.0.1:8081`. Run `task dev:whisper` and `task dev:backend` in separate terminals for the local non-Docker stack.

Local backend runtime also needs `yt-dlp` and `ffmpeg` available on `PATH`. Docker is the simplest way to run the full backend plus Whisper stack if those system dependencies are missing.

Useful local commands:

- Submit a smoke-test job to the local backend: `task api:smoke`
- Override smoke-test backend URL: `LSI_BACKEND_URL=http://127.0.0.1:8080 task api:smoke`
- List all Taskfile commands: `task --list`

Docker workflow:

- Start backend and Whisper services: `task docker:up`
- Rebuild and start backend and Whisper services: `task docker:build`
- Follow service logs: `task docker:logs`
- Stop services: `task docker:down`

Docker Compose uses `.env`, starts `backend` on port `8080`, starts `whisper` on port `8081` inside the Compose network, and requires `LSI_DOCKER_BIND_HOST` to be set. Volumes are `lsi-data` for backend data, `lsi-hf-cache` for Hugging Face model cache, and `lsi-whisper-data` for Whisper transcription work data.

## Runtime Configuration

Backend environment variables:

| Variable | Default | Notes |
| --- | --- | --- |
| `LSI_ADDR` | `127.0.0.1:8080` | Backend listen address; Docker sets `0.0.0.0:8080` |
| `LSI_DB_PATH` | `./data/backend.sqlite3` | SQLite database path |
| `LSI_WORK_DIR` | `./data/jobs` | Backend job file workspace |
| `LSI_LOG_LEVEL` | `info` | Backend log level |
| `LSI_DOWNLOAD_TIMEOUT` | `10m` | Audio download timeout |
| `LSI_WHISPER_BASE_URL` | `http://127.0.0.1:8081` | Whisper service base URL; Docker uses `http://whisper:8081` |
| `LSI_WHISPER_MODEL` | `small` | faster-whisper model name |
| `LSI_WHISPER_COMPUTE_TYPE` | `default` | faster-whisper compute type |
| `LSI_WHISPER_TIMEOUT` | `30m` | Whisper request timeout |
| `LSI_WHISPER_POLL_INTERVAL` | `2s` | Backend polling interval for Whisper tasks |
| `LSI_LLM_BASE_URL` | `https://api.openai.com` | OpenAI-compatible API base URL |
| `LSI_LLM_API_KEY` | empty | Required for real translation flows |
| `LSI_LLM_MODEL` | empty | Required by the translator |
| `LSI_LLM_TIMEOUT` | `2m` | LLM request timeout |

Whisper service environment variables:

| Variable | Default | Notes |
| --- | --- | --- |
| `LSI_WHISPER_WORK_DIR` | `/data/transcriptions` | Transcription task data directory; `task dev:whisper` uses `whisper/.data/transcriptions` |
| `HF_HOME` | tool default | Hugging Face cache directory; Docker sets `/huggingface` |
| `HF_TOKEN` | empty | Optional token to raise Hugging Face Hub download limits |

Extension backend URL behavior:

- The default backend URL is `http://127.0.0.1:8080`.
- The manifest only grants host permissions for `http://127.0.0.1:*/*` and `http://localhost:*/*`.
- Extension URL validation only accepts `localhost` or `127.0.0.1` HTTP origins. Do not assume arbitrary remote backend URLs work without manifest and validation changes.

## Testing Instructions

Run all tests and checks from the repository root:

- Run all module tests: `task test`
- Run all tests plus extension typecheck: `task check`
- Run backend tests: `task test:backend`
- Run Whisper tests: `task test:whisper`
- Run extension tests: `task test:extension`
- Run extension TypeScript typecheck: `task typecheck`

Direct module commands:

```bash
cd backend && mise exec -- go test ./...
cd whisper && mise exec -- uv run pytest
cd extension && mise exec -- npm run test
cd extension && mise exec -- npm run typecheck
```

Focused test examples:

```bash
cd backend && mise exec -- go test ./internal/api -run TestName
cd whisper && mise exec -- uv run pytest tests/test_server.py -k transcriptions
cd extension && mise exec -- npx vitest run src/subtitles/vtt.test.ts
```

Test locations and conventions:

- Go tests live next to implementation files and use `*_test.go`.
- Python tests live under `whisper/tests/` and use `test_*.py`.
- Extension tests live next to source files or entrypoints and use `*.test.ts`.
- Vitest uses `jsdom`, global test APIs, and WXT's Vitest plugin from `extension/vitest.config.ts`.
- Add or update tests for behavior changes, especially API routes, persistence, job lifecycle, downloader/transcriber/translator orchestration, VTT parsing/rendering, extension messaging, storage, and YouTube page integration.
- Run the narrowest relevant test during development, then run `task check` before finalizing broad or cross-module changes.

## Build Commands

- Build all modules: `task build`
- Build backend Go packages: `task build:backend`
- Build Whisper Python package: `task build:whisper`
- Build Chrome MV3 extension: `task build:extension`

Build outputs:

| Module | Output |
| --- | --- |
| Backend | Go package build via `go build ./...`; Docker image built from `backend/Dockerfile` |
| Whisper | Python package artifacts under `whisper/dist/` from `uv build` |
| Extension | Chrome MV3 output under `extension/.output/chrome-mv3` |

Use `task docker:build` to build and run the Docker backend plus Whisper services locally.

## Code Style Guidelines

General:

- Make small, surgical changes that directly serve the task.
- Preserve module boundaries: backend/API code in `backend/`, transcription service code in `whisper/`, browser-extension code in `extension/`.
- Prefer existing patterns and names over new abstractions.
- Do not add backward-compatibility paths unless persisted data, shipped behavior, external consumers, or explicit requirements make them necessary.
- Keep documentation and commands current when behavior, configuration, or workflows change.

Backend Go:

- Keep executable entrypoints under `backend/cmd/` and reusable packages under `backend/internal/`.
- Keep HTTP routing and response behavior in `backend/internal/api/`.
- Keep configuration and app assembly in `backend/internal/app/`.
- Keep SQLite models, migrations, and persistence behavior in `backend/internal/store/`.
- Keep downloader, HTTP transcriber, translator, VTT cue handling, and job runner orchestration in `backend/internal/runner/`.
- Use standard Go formatting. Run `gofmt` on changed Go files.
- Keep API route behavior covered by tests under `backend/internal/api` and store behavior covered under `backend/internal/store`.

Whisper Python:

- Package source lives under `whisper/src/whisper_cli/`.
- Keep FastAPI service behavior in `server.py`, transcription behavior in `transcribe.py`, and WebVTT rendering helpers in `vtt.py` unless a change clearly requires a different split.
- Use `uv` for dependency sync, command execution, tests, and builds.
- Keep service tests in `whisper/tests/test_server.py`, transcription tests in `whisper/tests/test_transcribe.py`, and VTT tests in `whisper/tests/test_vtt.py`.

Extension Vue/TypeScript:

- WXT is configured with `srcDir: 'src'` and `entrypointsDir: '../entrypoints'`; extension entrypoints live in `extension/entrypoints/`.
- Use `@` imports for `extension/src` where appropriate, matching `extension/wxt.config.ts`.
- Shared shadcn-vue UI primitives live under `extension/src/components/ui/`.
- For extension UI, prefer shadcn-vue components whenever they provide the needed primitive. Only create custom UI components when shadcn-vue has no suitable component.
- Before adding or changing shadcn-vue usage, read the current official documentation at `https://www.shadcn-vue.com/llms.txt`.
- Preserve Chrome MV3 constraints and host permissions unless the task explicitly requires manifest changes.
- Run `npm run test` or a focused Vitest command after logic changes, and run `npm run typecheck` after TypeScript or Vue changes.

## API Reference For Agents

Backend HTTP endpoints:

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/jobs` | Create or reuse a subtitle job |
| `GET` | `/jobs/{jobId}` | Get job status |
| `GET` | `/jobs/active?videoId=...&targetLanguage=...` | Get latest active job for a video and target language |
| `GET` | `/subtitle-assets?videoId=...&targetLanguage=...` | List generated subtitle assets |
| `GET` | `/subtitle-files/{jobId}/{mode}` | Download `source`, `translated`, or `bilingual` VTT |

Whisper HTTP endpoints:

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/healthz` | Health check |
| `POST` | `/transcriptions` | Create transcription task with multipart `audio`, `model`, `language`, optional `computeType`, optional `jobId` |
| `GET` | `/transcriptions/{task_id}` | Get transcription task status |
| `GET` | `/transcriptions/{task_id}/vtt` | Fetch completed VTT |
| `DELETE` | `/transcriptions/{task_id}` | Delete transcription task data |

Job states used by the backend include `queued`, `downloading`, `transcribing`, `translating`, `packaging`, `completed`, and `failed`.

## Security And Secrets

- `.env` is local-only and ignored by git. Never commit real API keys, Hugging Face tokens, database files, or generated job data.
- `LSI_LLM_API_KEY` is required for real translation flows. Tests should use fakes or local test servers, not real LLM calls.
- `HF_TOKEN` is optional and only used to increase Hugging Face Hub download limits.
- The extension intentionally restricts backend origins to localhost. Treat any request to support remote origins as a security-sensitive manifest and validation change.
- Backend subtitle files and SQLite state are local artifacts. Do not add them to fixtures unless the task explicitly requires sanitized test fixtures.

## CI/CD

- `.github/workflows/backend-ci.yml` runs Go tests for backend changes on pull requests, pushes to `main`, and manual dispatch.
- `.github/workflows/whisper-ci.yml` runs `uv sync --dev` and `uv run pytest` for Whisper changes on pull requests, pushes to `main`, and manual dispatch.
- `.github/workflows/extension-ci.yml` runs `npm ci`, `npm run test`, and `npm run typecheck` for extension changes on pull requests, pushes to `main`, and manual dispatch.
- `.github/workflows/extension-build.yml` builds the Chrome MV3 extension on `main` pushes affecting extension files and uploads `extension/.output/chrome-mv3` as an artifact named `lets-sub-it-extension-chrome-<run-number>`.
- `.github/workflows/backend-image.yml` builds and pushes GHCR Docker images for both `backend/Dockerfile` and `whisper/Dockerfile` on `main` pushes affecting Docker/backend/whisper paths. Images are named `ghcr.io/<owner>/<repo>-api` and `ghcr.io/<owner>/<repo>-whisper`.
- Before changing workflow files, validate YAML and GitHub Actions syntax where practical. `actionlint` is available through mise.

## Pull Request Guidelines

- Use `.github/pull_request_template.md` for PR descriptions.
- Include references, summary, file-level notes, validation commands, and review focus areas when relevant.
- Run the narrowest relevant checks while developing and include the commands run in the PR.
- Run `task check` before finalizing changes that affect code across modules.
- For Docker/backend/whisper deployment changes, also run the relevant build command such as `task build:backend`, `task build:whisper`, or `task docker:build` when feasible.
- Do not commit generated artifacts unless the task explicitly asks for release artifacts.

## Common Gotchas

- `task dev:backend` does not start Whisper. Start `task dev:whisper` separately or use Docker Compose.
- Local backend runs need `yt-dlp` and `ffmpeg` on `PATH`; Docker images include the needed runtime tools.
- First transcription can be slow because faster-whisper downloads the selected model. Docker persists this cache in `lsi-hf-cache`.
- Docker Compose fails early if `LSI_DOCKER_BIND_HOST` is missing from `.env`.
- Extension tests and typechecking call `wxt prepare`; use `task test:extension`, `task typecheck`, or the package scripts rather than invoking `vue-tsc` directly.
- The extension only supports localhost backend origins unless both WXT manifest host permissions and URL validation are changed.
- `whisper/uv.lock` and `extension/package-lock.json` are committed lockfiles. Keep them updated when dependencies change.
