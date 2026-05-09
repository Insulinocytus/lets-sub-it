# AGENTS.md

## Project Overview

Lets Sub It is a self-hosted YouTube subtitle generation and translation tool. A user submits a public YouTube URL, the backend creates or reuses a job, downloads audio, transcribes locally, translates each cue, writes WebVTT files, and the Chrome extension renders subtitles on the YouTube watch page.

This repository is a multi-module MVP, not a single-package workspace:

- `backend/` — Go 1.22 HTTP API server with SQLite/GORM persistence, job deduplication, mock runner by default, optional real runner (`yt-dlp` + `whisper-cli` + LLM), and VTT file serving.
- `whisper/` — Python 3.12 `whisper-cli` package wrapping `faster-whisper`, producing validated WebVTT and JSON summaries.
- `extension/` — Chrome MV3 extension using WXT, Vue, TypeScript, Vitest, Tailwind/shadcn-vue, background API gateway, storage cache, and YouTube subtitle overlay.
- `docs/` — Chinese PRD, specs, and implementation plans. The nested `docs/AGENTS.md` takes precedence for all files under `docs/`.

Default behavior is offline-friendly: `backend` uses `LSI_RUNNER_MODE=mock`, so it does not contact YouTube, download Whisper models, or call an LLM. `LSI_RUNNER_MODE=real` enables actual download, transcription, translation, and packaging.

## Agent Operating Rules

- Communicate with users in **Simplified Chinese**.
- Run `git status --short` before and after edits. Do not overwrite or revert unrelated user changes.
- Keep changes surgical. Every changed line should trace to the request.
- Prefer existing module boundaries and local patterns over new abstractions.
- Add or update tests for behavior changes in the touched module.
- Unit tests must stay offline and repeatable: no real YouTube, model downloads, GPU, external LLM, private local data, or provider keys.
- Do not commit build artifacts, SQLite databases, model files, download caches, real audio samples, `.env`, API keys, or LLM request logs.

## Dev Environment Tips

- Agent shells do **not** auto-activate `mise`; prefix tool commands with `mise exec --`.
- Install pinned toolchains from the repository root with `mise install`.
- This is not a unified package-manager monorepo. Enter the relevant module before running commands.
- Tool versions are pinned in `mise.toml`: Go 1.22, Python 3.12, Node.js 22, and `uv`.
- Use `rg` / `rg --files` for search unless unavailable.
- For architecture or cross-module questions, read `graphify-out/GRAPH_REPORT.md` first. If `graphify-out/wiki/index.md` exists, navigate that before raw files.

## Setup Commands

Install local toolchains:

```bash
mise install
```

Install module dependencies:

```bash
cd backend && mise exec -- go mod download
cd ../whisper && mise exec -- uv sync --dev
cd ../extension && mise exec -- npm install
```

Docker backend setup for real runner:

```bash
cp .env.example .env
# edit .env: set LSI_LLM_API_KEY and LSI_LLM_MODEL at minimum
docker compose up -d
```

Docker builds the Go backend, Python `whisper-cli`, `yt-dlp`, and `ffmpeg` into one backend image. Runtime data persists in the `lsi-data` Docker volume, and Hugging Face model cache persists in the separate `lsi-hf-cache` volume.

## Development Workflow

Start backend with the default mock runner:

```bash
cd backend
LSI_ADDR=127.0.0.1:8080 mise exec -- go run ./cmd/server
```

Smoke test the API:

```bash
curl -X POST "http://127.0.0.1:8080/jobs" \
  -H "Content-Type: application/json" \
  -d '{"youtubeUrl":"https://www.youtube.com/watch?v=dQw4w9WgXcQ","sourceLanguage":"en","targetLanguage":"zh"}'
```

Start backend with the real runner:

```bash
cd whisper && mise exec -- uv sync --dev && cd ../backend
PATH="$PWD/../whisper/.venv/bin:$PATH" \
LSI_RUNNER_MODE=real \
LSI_DOWNLOAD_TIMEOUT=10m \
LSI_WHISPER_MODEL=small \
LSI_WHISPER_COMPUTE_TYPE=int8 \
LSI_LLM_BASE_URL=https://api.openai.com \
LSI_LLM_API_KEY="$OPENAI_API_KEY" \
LSI_LLM_MODEL=gpt-4.1-mini \
LSI_LLM_TIMEOUT=2m \
LSI_ADDR=127.0.0.1:8080 \
mise exec -- go run ./cmd/server
```

Start the extension dev server:

```bash
cd extension
mise exec -- npm run dev
```

Load `extension/.output/chrome-mv3` in Chrome extension developer mode. The popup defaults to `http://127.0.0.1:8080`; only `http://localhost:<port>` and `http://127.0.0.1:<port>` are valid backend origins.

Run `whisper-cli` directly:

```bash
cd whisper
mise exec -- uv run whisper-cli \
  --input /path/to/audio.mp3 \
  --output /tmp/source.vtt \
  --model small \
  --language ja
```

Real transcription can download models and may require GPU. Tests use fake models and must remain offline.

## Testing Instructions

Run all tests for each module:

| Module | Command |
| --- | --- |
| backend | `cd backend && mise exec -- go test ./...` |
| whisper | `cd whisper && mise exec -- uv run pytest` |
| extension | `cd extension && mise exec -- npm run test` |

Focused commands:

| Scope | Command |
| --- | --- |
| backend package | `cd backend && mise exec -- go test ./internal/api` |
| backend named test | `cd backend && mise exec -- go test ./internal/runner -run TestRealRunnerCompletesJob` |
| whisper file | `cd whisper && mise exec -- uv run pytest tests/test_vtt.py` |
| whisper pattern | `cd whisper && mise exec -- uv run pytest -k "vtt"` |
| extension file | `cd extension && mise exec -- npx vitest run src/api/backend-client.test.ts` |
| extension typecheck | `cd extension && mise exec -- npm run typecheck` |

Test locations and conventions:

- `backend/internal/**/*_test.go` lives beside the tested Go package.
- `whisper/tests/test_*.py` uses pytest with `pythonpath = ["src"]`.
- `extension/src/**/*.test.ts` uses Vitest + jsdom + WXT prepare.

When changing behavior:

- `backend/internal/*` changes need same-package Go tests.
- `whisper/src/whisper_cli/*` changes need pytest coverage.
- `extension/src/*` or `extension/entrypoints/*` changes need Vitest coverage, plus `npm run typecheck` for type-facing changes.

## Build Verification

No automated CI is configured. Verify affected modules manually:

| Module | Build command | Generated artifacts |
| --- | --- | --- |
| backend | `cd backend && mise exec -- go build ./...` | Go build outputs; do not commit binaries |
| whisper | `cd whisper && mise exec -- uv build` | `whisper/dist/` |
| extension | `cd extension && mise exec -- npm run build` | `extension/.output/`, `extension/.wxt/` |

Use Docker for backend deployment verification:

```bash
docker compose up -d --build
docker compose logs -f
docker compose down
```

## Code Style

### Go backend

- Use standard `gofmt`; do not introduce a formatter beyond Go tooling.
- Keep HTTP parsing, response structs, routing, and CORS in `backend/internal/api/`.
- Keep SQLite/GORM persistence in `backend/internal/store/`; schema is initialized with GORM `AutoMigrate`.
- Keep job execution in `backend/internal/runner/`; `MockRunner` simulates all stages and `RealRunner` calls external tools.
- Do not add frameworks, queue systems, or background task systems unless explicitly requested.
- API responses must never expose local absolute file paths. Frontend should use `/subtitle-files/:jobId/:mode`.
- File serving must stay contained within the job work directory; preserve path traversal and symlink escape protections.

### Python whisper-cli

- Use Python 3.12 syntax with type annotations.
- Follow the existing style: `from __future__ import annotations`, 4-space indent, concise functions.
- CLI entry point: `whisper/src/whisper_cli/cli.py`, exposed as `whisper-cli` in `pyproject.toml`.
- Transcription adapter: `whisper/src/whisper_cli/transcribe.py`; WebVTT validation/rendering: `whisper/src/whisper_cli/vtt.py`.
- Do not create single-use abstractions or add dependencies unless required.
- No formatter/linter is configured; do not reformat entire files unprompted.
- `uv.lock` is committed. Update dependencies through `uv`, never by manually editing the lockfile.

### Chrome extension

- Use TypeScript, Vue SFC, WXT, npm, Tailwind/shadcn-vue, and the `@/*` alias for `extension/src/*`.
- Popup entry lives in `extension/entrypoints/popup/`; business validation belongs in `extension/src/popup/`.
- Background message protocol and HTTP client belong in `extension/src/api/`.
- Storage logic belongs in `extension/src/storage/`.
- WebVTT parsing and cue matching belong in `extension/src/subtitles/` and must stay independently testable.
- YouTube watch detection and SPA navigation belong in `extension/src/youtube/`.
- shadcn-vue components live under `extension/src/components/ui/`; only add components that are actually used.
- Content scripts must not call the Go backend directly. All network requests go through the background service worker.
- Never store provider keys or long-lived secrets in the extension.
- `package-lock.json` is committed. Update dependencies through `npm`, never by manually editing the lockfile.

## Key Entry Points

| Area | Path | Purpose |
| --- | --- | --- |
| backend server | `backend/cmd/server/main.go` | HTTP server entry point |
| backend config/app | `backend/internal/app/` | env config, app wiring, real-mode tool checks |
| backend API | `backend/internal/api/` | routes, handlers, CORS, response mapping |
| backend store | `backend/internal/store/` | GORM models and SQLite persistence |
| backend runner | `backend/internal/runner/` | mock/real runner, download, translation, VTT packaging |
| whisper CLI | `whisper/src/whisper_cli/cli.py` | command-line contract and exit codes |
| whisper transcribe | `whisper/src/whisper_cli/transcribe.py` | faster-whisper adapter |
| whisper VTT | `whisper/src/whisper_cli/vtt.py` | cue validation and WebVTT rendering |
| extension background | `extension/entrypoints/background.ts` | runtime message gateway |
| extension content | `extension/entrypoints/youtube.content.ts` | YouTube page integration |
| extension popup | `extension/entrypoints/popup/` | popup UI entry |

## Runtime Contracts

### Backend configuration

| Env var | Default | Description |
| --- | --- | --- |
| `LSI_ADDR` | `127.0.0.1:8080` | HTTP listen address; Docker sets `0.0.0.0:8080` inside the container |
| `LSI_DB_PATH` | `./data/backend.sqlite3` | SQLite database path |
| `LSI_WORK_DIR` | `./data/jobs` | job work directory root |
| `LSI_RUNNER_MODE` | `mock` | `mock` or `real` |
| `LSI_LOG_LEVEL` | `info` | backend structured log level: `debug`, `info`, `warn`, or `error` |
| `LSI_DOWNLOAD_TIMEOUT` | `10m` | download timeout in real mode |
| `LSI_WHISPER_MODEL` | `small` | `faster-whisper` model or local CTranslate2 model directory passed to `whisper-cli --model` |
| `LSI_WHISPER_COMPUTE_TYPE` | `default` | faster-whisper compute type passed to `whisper-cli --compute-type`; use `int8` to reduce CPU memory use |
| `HF_TOKEN` | empty | optional Hugging Face token for higher model download limits; read by backend container tooling |
| `LSI_LLM_BASE_URL` | `https://api.openai.com` | OpenAI-compatible API origin |
| `LSI_LLM_API_KEY` | empty | required for the OpenAI default endpoint; backend-only |
| `LSI_LLM_MODEL` | empty | required in real mode for translation |
| `LSI_LLM_TIMEOUT` | `2m` | per-cue translation timeout |

### API reference

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/jobs` | create or reuse a subtitle generation job |
| `GET` | `/jobs/:id` | query job status |
| `GET` | `/jobs/active?videoId=...&targetLanguage=...` | query the latest job for a video/language pair so the popup can restore progress |
| `GET` | `/subtitle-assets?videoId=...&targetLanguage=...` | query completed subtitle assets |
| `GET` | `/subtitle-files/:jobId/:mode` | serve VTT file; mode is `source`, `translated`, or `bilingual` |

Job state flow: `queued` -> `downloading` -> `transcribing` -> `translating` -> `packaging` -> `completed`. On failure, state is `failed` and `errorMessage` contains an error summary.

### Whisper CLI contract

Input: local audio file. Output: valid WebVTT at `--output` plus a JSON summary on stdout.

| Param | Required | Description |
| --- | --- | --- |
| `--input` | yes | local audio file path |
| `--output` | yes | output `.vtt` path; must differ from input |
| `--model` | yes | faster-whisper model name, e.g. `small` |
| `--language` | yes | language code, e.g. `ja` or `en` |

Exit codes: `0` success, `2` input validation failure, `3` transcription failure, `4` output validation failure.

### Extension contract

- Background service worker is the only HTTP API gateway.
- Popup and content script communicate via runtime messages.
- Supported languages are `en` and `zh`; `sourceLanguage` must not equal `targetLanguage`.
- YouTube watch page subtitle modes are `translated` and `bilingual` only, although backend also serves `source`.
- Backend URL must be a localhost HTTP origin with explicit port.
- Manifest host permissions are limited to `http://127.0.0.1:*/*` and `http://localhost:*/*`.

## Documentation Rules

- All prose under `docs/` must be Chinese. Code, commands, paths, config keys, API names, and required references may remain English.
- The closest `AGENTS.md` takes precedence; read `docs/AGENTS.md` before editing `docs/`.
- When changing behavioral contracts, commands, APIs, exit codes, directory structure, supported languages, mock/real boundaries, or security boundaries, check whether `README.md`, `backend/README.md`, `whisper/README.md`, `extension/README.md`, and related docs also need updates.
- Keep root `README.md` human-facing. Keep `AGENTS.md` focused on instructions for coding agents.

## Security & Data Boundaries

- Process **YouTube public videos only**. Do not add private video support, login sessions, cookie import, or auth bypass.
- Backend has no user auth and is intended for single-user local self-hosting. Do not describe it as production-ready for the public internet.
- Translation provider keys belong in backend/server config only.
- Extension must not store secrets or call translation providers directly.
- Do not relax localhost-only extension backend URLs unless explicitly requested, and update manifest permissions plus specs if that changes.
- Never log, commit, or preserve LLM request logs containing private data or keys.

## Pull Request Guidelines

- PR label/title prefix should indicate scope: `[backend]`, `[whisper]`, `[extension]`, or `[docs]`.
- Follow `.github/pull_request_template.md`: reference, summary, close issue, per-file explanation, verification commands, and review focus.
- No CI pipeline exists. Run relevant checks manually and list commands/results in the PR description.

Pre-merge verification by scope:

| Scope | Required | Additional when relevant |
| --- | --- | --- |
| `backend/` behavior | `cd backend && mise exec -- go test ./...` | `cd backend && mise exec -- go build ./...` for packaging/dependency changes |
| `whisper/` behavior | `cd whisper && mise exec -- uv run pytest` | `cd whisper && mise exec -- uv build` for packaging/dependency changes |
| `extension/` behavior | `cd extension && mise exec -- npm run test` | `cd extension && mise exec -- npm run typecheck`; `npm run build` for packaging changes |
| docs only | `git diff --check` | verify links/commands touched by the docs |

## Common Pitfalls

- Backend default is `mock`; only `LSI_RUNNER_MODE=real` performs real download/transcription/translation.
- Docker backend runs real mode by default and binds host port `127.0.0.1:8080` unless `LSI_DOCKER_BIND_HOST` changes.
- Extension MVP supports only `en` and `zh`; do not promise a broad language list.
- CORS and backend URL validation are localhost-only with explicit ports.
- Unit tests must stay offline.
- Lockfiles are managed by tooling only: `go.sum` via Go, `uv.lock` via `uv`, `package-lock.json` via npm.
- Real transcription can download large models; never trigger it from unit tests.

## graphify

This project has a graphify knowledge graph at `graphify-out/`.

- Before answering architecture or codebase questions, read `graphify-out/GRAPH_REPORT.md` for god nodes and community structure.
- If `graphify-out/wiki/index.md` exists, navigate it instead of reading raw files.
- For cross-module questions, prefer `graphify query "<question>"`, `graphify path "<A>" "<B>"`, or `graphify explain "<concept>"` over grep because these traverse extracted and inferred edges.
- After modifying code files in this session, run `graphify update .` to keep the graph current. This is AST-only and has no API cost.

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
<!-- SPECKIT END -->
