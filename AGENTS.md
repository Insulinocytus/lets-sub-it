# Repository Guidelines

## Project Overview

Lets Sub It is a self-hosted YouTube subtitle generator and translator. A Chrome MV3 extension submits jobs and renders subtitles; a Go API downloads audio, coordinates a Whisper transcription container (`hwdsl2/whisper-server`, an OpenAI-compatible `/v1/audio/transcriptions` service) and an OpenAI-compatible translation API, and stores jobs in SQLite and subtitles as WebVTT files.

## Architecture & Data Flow

`extension/entrypoints/popup/` → background messages → `backend/internal/api/` → `backend/internal/runner/` (`yt-dlp`/`ffmpeg` → synchronous Whisper `verbose_json` request → LLM) → SQLite and `source`, `translated`, `bilingual` VTT → extension cache/content-script overlay. `backend/internal/store/` owns persistent jobs and assets; the backend renders `source.vtt` itself from the returned segments, and transcription has no queue, chunking or fallback. The backend runs jobs independently of the originating HTTP request and marks interrupted jobs failed on restart. The popup polls while open; `extension/src/api/job-monitor.ts` uses browser alarms to continue monitoring after it closes. The YouTube content script loads VTT and selects cues from player time.

## Key Directories

- `backend/cmd/server/`: executable; `backend/internal/app/`: configuration and assembly; `api/`: routes and responses; `store/`: GORM/SQLite; `runner/`: download, transcription, translation, VTT packaging.
- `extension/entrypoints/`: WXT popup, background and YouTube content-script entry points; `extension/src/api/`, `storage/`, `subtitles/`, `content/`, `youtube/`: messaging/network, persistence, VTT parsing, overlay and navigation. `extension/src/components/ui/` contains shared shadcn-vue primitives.
- `.github/workflows/`: module-specific CI and build/image publishing; see `README.md` for API/deployment details and `extension/README.md` for installing the extension.

## Development Commands

Run root tasks from the repository root (`Taskfile.yml` is authoritative):

| Purpose | Command |
| --- | --- |
| Install toolchains/dependencies | `task setup` |
| Run services locally, in separate terminals | `task dev:backend`, `task dev:extension` (transcription is a Docker service) |
| Run all tests and extension typecheck | `task check` |
| Run a module's tests | `task test:backend`, `task test:extension` |
| Typecheck extension / build all modules | `task typecheck` / `task build` |
| Run Compose stack | Copy `.env.example` to `.env`, set `LSI_DOCKER_BIND_HOST` and `LSI_LOCAL_STT_API_KEY`, then `task docker:build`; stop with `task docker:down` |

`task api:smoke` only submits a real YouTube job to a running backend; it does not wait for or verify VTT output. There is no `task lint`; CI enforces `gofmt -l` and `go vet` (`backend-ci.yml`), extension tests and typecheck (`extension-ci.yml`), and `actionlint` plus `docker compose config` against `.env.example` (`config-ci.yml`). Run the matching command locally before pushing.

## Code Conventions & Common Patterns

- Go: keep wiring/config in `app`, HTTP validation and `{error:{code,message}}` responses in `api`, persistence in `store`, and pipeline/state transitions in `runner`. Constructors inject small dependencies (`NewHTTPHandler`, `NewRealRunner`, `Transcriber`, `Translator`); preserve the existing error/status handling instead of adding parallel pathways.
- Transcription: the backend calls an OpenAI-compatible `POST {base}/audio/transcriptions` with `response_format=verbose_json` and requires valid `segments` (`start`, `end`, `text`) before writing `source.vtt`. There is no audio chunking or fallback to another provider.
- Extension: use the `@` alias for `extension/src` imports. Background messaging uses discriminated `type` requests and `{ok,data}` / `{ok:false,error}` results; browser-local settings/cache and alarm-backed job monitors outlive the popup. In the overlay, invalidate outstanding asynchronous loads on video changes and clean up listeners on unmount. Prefer existing shadcn-vue primitives for UI; consult `https://www.shadcn-vue.com/llms.txt` before changing their usage.
- Cross-component changes: update both sides of API/message contracts (`backend/internal/api/`, `extension/src/api/`) and trace VTT changes through `backend/internal/runner/vtt_cue.go` and `extension/src/subtitles/vtt.ts`.

## Important Files

- `backend/cmd/server/main.go`, `backend/internal/app/app.go`, `backend/internal/api/routes.go`, `backend/internal/runner/real_runner.go`, `backend/internal/store/models.go`: startup, HTTP surface, job stages and persisted state.
- `extension/entrypoints/{background.ts,youtube.content.ts,popup/App.vue}`: message/alarm handling, page injection and popup; `extension/src/api/{backend-client.ts,job-monitor.ts}`: API contract and persistent monitoring; `extension/src/content/YoutubeOverlay.vue`: rendering.
- `Taskfile.yml`, `mise.toml`, `docker-compose.yml`, `.env.example`, `extension/wxt.config.ts`, `extension/package.json`: commands, versions, deployment, MV3 permissions and scripts.

## Runtime/Tooling Preferences

- Use `mise` for Go 1.22 and Node 22; `task` and `actionlint` are set to `latest` in `mise.toml`. If `task` is unavailable, use `mise exec -- task <name>`; run direct module commands through `mise exec --`. Go uses modules and the extension uses npm/`package-lock.json` (not a single package-manager workspace).
- The backend's SQLite driver (`go-sqlite3`) needs cgo: building or testing the backend requires `gcc` on `PATH` (Windows: `scoop install gcc`). Without it, Go silently sets `CGO_ENABLED=0` and every store-backed test fails with "requires cgo".
- Local backend startup requires `yt-dlp` and `ffmpeg` on `PATH`; `task dev:backend` does not start transcription — run the `hwdsl2/whisper-server` image (Docker Compose or `docker run`) and point `LSI_STT_BASE_URL` at it (Docker: `http://whisper:9000/v1`, local: `http://127.0.0.1:9000/v1`). Real translation requires `LSI_LLM_API_KEY` and `LSI_LLM_MODEL`; keep `.env`, credentials, SQLite/work files and generated build outputs out of commits. Local transcription and external STT settings are independent of the LLM config.
- Extension backend URLs require HTTP `localhost` or `127.0.0.1` with an explicit port: both `extension/wxt.config.ts` host permissions and `extension/src/api/backend-client.ts` validation enforce this. Change both when deliberately supporting another origin. WXT `npm run test` and `npm run typecheck` run `wxt prepare` first.

## Testing & QA

- Go: colocated `*_test.go` using `testing`, `httptest`, temporary SQLite and injected downloader/transcriber/translator substitutes; cover API routes, persistence, failures and job transitions.
- Extension: colocated `*.test.ts` using Vitest, jsdom, WXT fake browser and Vue Test Utils; cover messaging, storage/alarm recovery, navigation and rendered subtitles.
- Run the narrowest relevant suite first, then `task check` for cross-module changes. Examples: from `backend/`, `mise exec -- go test ./internal/api`; from `extension/`, `mise exec -- npm run test -- src/api/job-monitor.test.ts`. There is no configured coverage threshold or dedicated e2e runner. For changes to real behavior, smoke the affected path in addition to tests; do not use real LLM/Whisper/YouTube in automated tests.

## Agent skills

### Issue tracker

Issues and specs are tracked in this repo's GitHub Issues. See `docs/agents/issue-tracker.md`.

### Triage labels

The five canonical triage roles use same-named labels. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `CONTEXT.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
