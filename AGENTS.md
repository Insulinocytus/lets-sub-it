# Repository Guidelines

## Project Overview

Lets Sub It is a self-hosted YouTube subtitle generator and translator. A Chrome MV3 extension submits jobs and renders subtitles; a Go API downloads audio, coordinates a local Python Whisper service and an OpenAI-compatible translation API, and stores jobs in SQLite and subtitles as WebVTT files.

## Architecture & Data Flow

`extension/entrypoints/popup/` → background messages → `backend/internal/api/` → `backend/internal/runner/` (`yt-dlp`/`ffmpeg` → Whisper HTTP → LLM) → SQLite and `source`, `translated`, `bilingual` VTT → extension cache/content-script overlay. `backend/internal/store/` owns persistent jobs and assets; Whisper's task queue is in memory. The backend runs jobs independently of the originating HTTP request and marks interrupted jobs failed on restart. The popup polls while open; `extension/src/api/job-monitor.ts` uses browser alarms to continue monitoring after it closes. The YouTube content script loads VTT and selects cues from player time.

## Key Directories

- `backend/cmd/server/`: executable; `backend/internal/app/`: configuration and assembly; `api/`: routes and responses; `store/`: GORM/SQLite; `runner/`: download, transcription, translation, VTT packaging.
- `whisper/src/whisper_cli/`: FastAPI service (`server.py`), faster-whisper transcription (`transcribe.py`), WebVTT validation/rendering (`vtt.py`); `whisper/tests/`: pytest.
- `extension/entrypoints/`: WXT popup, background and YouTube content-script entry points; `extension/src/api/`, `storage/`, `subtitles/`, `content/`, `youtube/`: messaging/network, persistence, VTT parsing, overlay and navigation. `extension/src/components/ui/` contains shared shadcn-vue primitives.
- `.github/workflows/`: module-specific CI and build/image publishing; see `README.md` for API/deployment details and `extension/README.md` for installing the extension.

## Development Commands

Run root tasks from the repository root (`Taskfile.yml` is authoritative):

| Purpose | Command |
| --- | --- |
| Install toolchains/dependencies | `task setup` |
| Run services locally, in separate terminals | `task dev:whisper`, `task dev:backend`, `task dev:extension` |
| Run all tests and extension typecheck | `task check` |
| Run a module's tests | `task test:backend`, `task test:whisper`, `task test:extension` |
| Typecheck extension / build all modules | `task typecheck` / `task build` |
| Run Compose stack | Copy `.env.example` to `.env`, set `LSI_DOCKER_BIND_HOST`, then `task docker:build`; stop with `task docker:down` |

`task api:smoke` only submits a real YouTube job to a running backend; it does not wait for or verify VTT output. There is no configured `task lint` or npm lint script; use `gofmt` for changed Go files, `task typecheck` for extension types, and `actionlint` when editing workflows.

## Code Conventions & Common Patterns

- Go: follow standard `gofmt`; keep wiring/config in `app`, HTTP validation and `{error:{code,message}}` responses in `api`, persistence in `store`, and pipeline/state transitions in `runner`. Constructors inject small dependencies (`NewHTTPHandler`, `NewRealRunner`, `Transcriber`, `Translator`); preserve the existing error/status handling instead of adding parallel pathways.
- Python: `server.py` owns FastAPI routes and a single queued worker; inject the transcription callable through `TranscriptionService` for tests. `transcribe.py` handles model calls; `vtt.py` validates segments before emitting WebVTT.
- Extension: use the `@` alias for `extension/src` imports. Background messaging uses discriminated `type` requests and `{ok,data}` / `{ok:false,error}` results; browser-local settings/cache and alarm-backed job monitors outlive the popup. In the overlay, invalidate outstanding asynchronous loads on video changes and clean up listeners on unmount. Prefer existing shadcn-vue primitives for UI; consult `https://www.shadcn-vue.com/llms.txt` before changing their usage.
- Cross-component changes: update both sides of API/message contracts (`backend/internal/api/`, `extension/src/api/`) and trace VTT changes through `whisper/src/whisper_cli/vtt.py`, `backend/internal/runner/vtt_cue.go`, and `extension/src/subtitles/vtt.ts`.

## Important Files

- `backend/cmd/server/main.go`, `backend/internal/app/app.go`, `backend/internal/api/routes.go`, `backend/internal/runner/real_runner.go`, `backend/internal/store/models.go`: startup, HTTP surface, job stages and persisted state.
- `whisper/src/whisper_cli/server.py`: `/transcriptions` task lifecycle and VTT endpoint.
- `extension/entrypoints/{background.ts,youtube.content.ts,popup/App.vue}`: message/alarm handling, page injection and popup; `extension/src/api/{backend-client.ts,job-monitor.ts}`: API contract and persistent monitoring; `extension/src/content/YoutubeOverlay.vue`: rendering.
- `Taskfile.yml`, `mise.toml`, `docker-compose.yml`, `.env.example`, `extension/wxt.config.ts`, `extension/package.json`: commands, versions, deployment, MV3 permissions and scripts.

## Runtime/Tooling Preferences

- Use `mise` for Go 1.22, Python 3.12 and Node 22; `uv`, `task` and `actionlint` are set to `latest` in `mise.toml`. If `task` is unavailable, use `mise exec -- task <name>`; run direct module commands through `mise exec --`. Go uses modules, Whisper uses `uv`/`uv.lock`, and the extension uses npm/`package-lock.json` (not a single package-manager workspace).
- Local backend startup requires `yt-dlp` and `ffmpeg` on `PATH`; `task dev:backend` does not start Whisper. Docker images include those binaries. Real translation requires `LSI_LLM_API_KEY` and `LSI_LLM_MODEL`; keep `.env`, credentials, SQLite/work files and generated build outputs out of commits.
- Extension backend URLs require HTTP `localhost` or `127.0.0.1` with an explicit port: both `extension/wxt.config.ts` host permissions and `extension/src/api/backend-client.ts` validation enforce this. Change both when deliberately supporting another origin. WXT `npm run test` and `npm run typecheck` run `wxt prepare` first.

## Testing & QA

- Go: colocated `*_test.go` using `testing`, `httptest`, temporary SQLite and injected downloader/transcriber/translator substitutes; cover API routes, persistence, failures and job transitions.
- Python: `whisper/tests/test_*.py` with pytest, FastAPI `TestClient`, `tmp_path` and a fake transcriber/model; cover queue lifecycle and VTT validation without downloading models.
- Extension: colocated `*.test.ts` using Vitest, jsdom, WXT fake browser and Vue Test Utils; cover messaging, storage/alarm recovery, navigation and rendered subtitles.
- Run the narrowest relevant suite first, then `task check` for cross-module changes. Examples: from `backend/`, `mise exec -- go test ./internal/api`; from `whisper/`, `mise exec -- uv run pytest tests/test_server.py`; from `extension/`, `mise exec -- npm run test -- src/api/job-monitor.test.ts`. There is no configured coverage threshold or dedicated e2e runner. For changes to real behavior, smoke the affected path in addition to tests; do not use real LLM/Whisper/YouTube in automated tests.
