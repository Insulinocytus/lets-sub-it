# AGENTS.md

## Project Overview

Lets Sub It is a self-hosted YouTube subtitle generation and translation tool. The pipeline: submit a YouTube public video URL, download audio, transcribe locally, translate, generate subtitles, and render them on the YouTube watch page.

This is a multi-module monorepo at MVP stage:

- **`backend/`** — Go 1.22 API server with SQLite/GORM persistence, job deduplication, mock runner by default, optional real runner (yt-dlp + whisper-cli + LLM), and VTT file serving
- **`whisper/`** — Python 3.12 `whisper-cli` package wrapping `faster-whisper`, producing validated WebVTT output
- **`extension/`** — Chrome MV3 extension using WXT + Vue + TypeScript + Vitest + shadcn-vue, with popup submission, background API gateway, storage cache, and YouTube watch page subtitle overlay
- **`docs/`** — PRD, specs, and implementation plans. A nested `docs/AGENTS.md` requires Chinese for all prose under `docs/`

By default, the backend runs with a **mock runner** (no YouTube access, no Whisper, no LLM). Setting `LSI_RUNNER_MODE=real` enables actual download, transcription, and translation.

## Dev Environment Tips

- AI agent shells do **not** auto-activate `mise`. Always prefix tool commands with `mise exec --` (e.g. `mise exec -- go`, `mise exec -- uv`, `mise exec -- npm`).
- This is **not** a monorepo with a unified package manager. `cd` into the relevant subdirectory before running commands.
- Toolchain versions are pinned in root `mise.toml`: Go 1.22, Python 3.12, Node.js 22, uv.
- The closest `AGENTS.md` in the directory tree takes precedence. Currently `docs/AGENTS.md` mandates Chinese for `docs/` prose.
- User instructions always override this file. If a request conflicts with these guidelines, note the conflict and risk before following the user's explicit direction.
- This file targets coding agents. Human-facing project intro belongs in `README.md`.

## Setup Commands

Install toolchain (from repo root):

```bash
mise install
```

Install dependencies per module:

```bash
cd backend && mise exec -- go mod download
cd ../whisper && mise exec -- uv sync --dev
cd ../extension && mise exec -- npm install
```

## Development Workflow

### Start backend (mock runner)

```bash
cd backend
LSI_ADDR=127.0.0.1:8080 mise exec -- go run ./cmd/server
```

Quick smoke test:

```bash
curl -X POST "http://127.0.0.1:8080/jobs" \
  -H "Content-Type: application/json" \
  -d '{"youtubeUrl":"https://www.youtube.com/watch?v=dQw4w9WgXcQ","sourceLanguage":"en","targetLanguage":"zh"}'
```

### Start backend (real runner)

Requires `yt-dlp`, `ffmpeg`, and `whisper-cli` on `PATH`, plus a Chat Completions-compatible LLM:

```bash
cd whisper && mise exec -- uv sync --dev && cd ../backend
PATH="$PWD/../whisper/.venv/bin:$PATH" \
LSI_RUNNER_MODE=real \
LSI_DOWNLOAD_TIMEOUT=10m \
LSI_WHISPER_MODEL=small \
LSI_LLM_BASE_URL=https://api.openai.com \
LSI_LLM_API_KEY="$OPENAI_API_KEY" \
LSI_LLM_MODEL=gpt-4.1-mini \
LSI_LLM_TIMEOUT=2m \
LSI_ADDR=127.0.0.1:8080 \
mise exec -- go run ./cmd/server
```

### Start extension dev server

```bash
cd extension && mise exec -- npm run dev
```

Load `.output/chrome-mv3` in Chrome extension developer mode. Popup defaults to `http://127.0.0.1:8080` — only `http://localhost:<port>` or `http://127.0.0.1:<port>` are allowed.

### Run whisper-cli locally

```bash
cd whisper
mise exec -- uv run whisper-cli \
  --input /path/to/audio.mp3 \
  --output /tmp/source.vtt \
  --model small \
  --language ja
```

Real transcription triggers model downloads and may require GPU. Unit tests use a fake model and stay offline.

## Testing Instructions

### Run all tests per module

| Module | Command |
| --- | --- |
| backend | `cd backend && mise exec -- go test ./...` |
| whisper | `cd whisper && mise exec -- uv run pytest` |
| extension | `cd extension && mise exec -- npm run test` |

### Focused tests

| What | Command |
| --- | --- |
| backend single package | `cd backend && mise exec -- go test ./internal/api` |
| whisper single file | `cd whisper && mise exec -- uv run pytest tests/test_vtt.py` |
| whisper by name | `cd whisper && mise exec -- uv run pytest -k "vtt"` |
| extension single file | `cd extension && mise exec -- npx vitest run src/api/backend-client.test.ts` |
| extension type check | `cd extension && mise exec -- npm run typecheck` |

### Test patterns

- **backend**: `*_test.go` files live alongside the tested package in `backend/internal/*/`
- **whisper**: `test_*.py` in `whisper/tests/`; pytest config in `pyproject.toml` with `pythonpath = ["src"]`, `addopts = "-q"`
- **extension**: `src/**/*.test.ts`; Vitest + jsdom + WXT plugin

### Test rules

- When changing `backend/internal/*` behavior, add or update same-package Go tests first, then verify with `go test`.
- When changing `whisper/src/whisper_cli/` behavior, add or update adjacent pytest tests, then verify with `pytest`.
- When changing `extension/src/` or `extension/entrypoints/`, add or update adjacent Vitest tests, then verify with `npm run test`.
- **Unit tests must stay offline and repeatable** — no real YouTube, model downloads, GPU, external LLM, or local private data.

## Build Verification

No CI pipeline exists. Verify builds by entering each subdirectory:

| Module | Build command | Artifacts (do not commit) |
| --- | --- | --- |
| backend | `cd backend && mise exec -- go build ./...` | Go binary (gitignored) |
| whisper | `cd whisper && mise exec -- uv build` | `whisper/dist/` (gitignored) |
| extension | `cd extension && mise exec -- npm run build` | `extension/.output/`, `.wxt/` (gitignored) |

## Code Style

### Go backend

- Go 1.22 with standard `gofmt` style.
- Do not introduce new frameworks, queue systems, or background task systems unless the task explicitly asks for it.
- **`backend/internal/api/`** — HTTP layer: request parsing, response structs, routing, CORS. Keep these concerns inside this package.
- **`backend/internal/store/`** — SQLite/GORM persistence. Schema initialized via GORM `AutoMigrate`.
- **`backend/internal/runner/`** — `MockRunner` (all stages mocked) and `RealRunner` (calls yt-dlp, whisper-cli, LLM).
- Never expose local absolute file paths in API responses; frontend should use `/subtitle-files/:jobId/:mode`.
- File serving must be contained within the job work directory — prevent path traversal and symlink escapes.

### Python whisper-cli

- Python 3.12 syntax, source in `whisper/src/whisper_cli/`.
- Follow existing style: type annotations, `from __future__ import annotations`, 4-space indent, concise functions.
- CLI entry point: `whisper/src/whisper_cli/cli.py`, exposed as `whisper-cli` via `pyproject.toml` `[project.scripts]`.
- Transcription adapter: `whisper/src/whisper_cli/transcribe.py`; WebVTT validation: `whisper/src/whisper_cli/vtt.py`.
- Do not create single-use abstractions or introduce new dependencies unless the task requires it.
- No formatter/linter configured — do not reformat entire files unprompted.
- `uv.lock` is committed. Update Python deps through `uv`, never edit the lockfile manually.

### Chrome extension

- TypeScript, Vue SFC, WXT, npm. Path alias `@/*` → `extension/src/*`.
- **`extension/entrypoints/popup/`** — popup UI; business validation goes in `extension/src/popup/`.
- **`extension/src/api/`** — background message protocol and HTTP client.
- **`extension/src/storage/`** — extension storage logic.
- **`extension/src/subtitles/`** — WebVTT parsing and cue matching, must stay independently testable.
- **`extension/src/youtube/`** — YouTube watch URL detection and SPA navigation.
- **`extension/src/components/ui/`** — shadcn-vue local components. Only add components that are actually used.
- Content scripts must **not** call the Go backend directly; all network requests go through the background service worker.
- Never put translation provider keys or long-lived secrets in the extension.
- `package-lock.json` is committed. Update npm deps through `npm`, never edit the lockfile manually.

## Key Entry Points

| Module | Entry | Description |
| --- | --- | --- |
| backend server | `backend/cmd/server/main.go` | HTTP server entry |
| backend API | `backend/internal/api/` | Routes, handlers, CORS |
| backend store | `backend/internal/store/` | GORM models, SQLite persistence |
| backend runner | `backend/internal/runner/` | MockRunner, RealRunner |
| whisper CLI | `whisper/src/whisper_cli/cli.py` | CLI entry (`whisper-cli`) |
| whisper transcribe | `whisper/src/whisper_cli/transcribe.py` | faster-whisper adapter |
| whisper VTT | `whisper/src/whisper_cli/vtt.py` | WebVTT timeline & cue validation |
| extension entries | `extension/entrypoints/` | background.ts, youtube.content.ts, popup/ |
| extension logic | `extension/src/` | api, storage, subtitles, youtube, popup |

## Backend Configuration

| Env var | Default | Description |
| --- | --- | --- |
| `LSI_ADDR` | `127.0.0.1:8080` | HTTP listen address |
| `LSI_DB_PATH` | `./data/backend.sqlite3` | SQLite database path |
| `LSI_WORK_DIR` | `./data/jobs` | Job work directory root |
| `LSI_RUNNER_MODE` | `mock` | `mock` or `real` |
| `LSI_DOWNLOAD_TIMEOUT` | `10m` | Download timeout (real mode) |
| `LSI_WHISPER_MODEL` | `small` | faster-whisper model name (real mode) |
| `LSI_LLM_BASE_URL` | `https://api.openai.com` | OpenAI-compatible API origin |
| `LSI_LLM_API_KEY` | _(empty)_ | Required for OpenAI default; Bearer token, backend-only |
| `LSI_LLM_MODEL` | _(empty)_ | Required in real mode for translation |
| `LSI_LLM_TIMEOUT` | `2m` | Per-cue translation timeout |

## API Reference

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/jobs` | Create or reuse a subtitle generation job |
| `GET` | `/jobs/:id` | Query job status |
| `GET` | `/subtitle-assets?videoId=...&targetLanguage=...` | Query completed subtitle assets |
| `GET` | `/subtitle-files/:jobId/:mode` | Serve VTT file; mode is `source`, `translated`, or `bilingual` |

Job state flow: `queued` → `downloading` → `transcribing` → `translating` → `packaging` → `completed`. On failure, state is `failed` with `errorMessage` in the response.

## Whisper CLI Contract

Input: local audio file. Output: valid WebVTT to `--output`, JSON summary on stdout.

| Param | Required | Description |
| --- | --- | --- |
| `--input` | yes | Local audio file path |
| `--output` | yes | Output `.vtt` path (must differ from input) |
| `--model` | yes | faster-whisper model name, e.g. `small` |
| `--language` | yes | Language code, e.g. `ja`, `en` |

Exit codes: `0` success, `2` input validation failed, `3` transcription failed, `4` output validation failed.

## Extension Contract

- Stack: WXT + Vue + TypeScript + Vite + Vitest + shadcn-vue, npm for packages.
- Entry points: `extension/entrypoints/` — `background.ts`, `youtube.content.ts`, `popup/`.
- Background service worker is the sole HTTP API gateway; popup and content script communicate via runtime messages.
- Supported languages: `en` and `zh` only. `sourceLanguage` ≠ `targetLanguage`.
- Subtitle modes on watch page: `translated` and `bilingual` only (backend still serves `source`).
- Backend URL must be a localhost HTTP origin with explicit port (`http://localhost:<port>` or `http://127.0.0.1:<port>`).
- Manifest host permissions: `http://127.0.0.1:*/*` and `http://localhost:*/*` only.

## Collaboration Principles

- Communicate with users in **Simplified Chinese**.
- Make minimal viable changes — no unsolicited features, abstractions, or config.
- Surgical edits: only touch files and lines directly related to the task.
- If requirements are ambiguous, state assumptions; if high-risk uncertainty, ask first.
- Do not clean up unrelated code, restyle unrelated formatting, or refactor modules not in scope.
- Do not overwrite uncommitted changes by other agents. Run `git status --short` before and after work.
- Never commit build artifacts, databases, model files, download caches, real audio samples, `.env` files, API keys, or LLM request logs.

## Security & Data

- Process **YouTube public videos only** — no private video support, login sessions, cookie import, or auth bypass.
- Backend has no user auth; it is designed for single-user local self-hosting. Never describe it as production-ready for public internet.
- Never commit real audio samples, model files, download caches, SQLite databases, `.env` files, API keys, or LLM logs.
- Extension allows localhost backend URLs only — do not relax this to arbitrary remote hosts unless explicitly requested, and update the permissions spec accordingly.
- Translation provider keys belong in server-side config only; extension must not hold secrets.

## Pull Request Guidelines

- PR label format: `[backend]`, `[whisper]`, `[extension]`, `[docs]` to indicate scope.
- Follow `.github/pull_request_template.md` for structure: reference, summary, close issue, per-file explanation, verification steps, and review focus items.
- No automated CI exists. Run relevant tests manually before submitting and list the commands and results in the PR description.
- Pre-merge verification per scope:

| Scope | Test command | Additional |
| --- | --- | --- |
| `backend/` behavior | `cd backend && mise exec -- go test ./...` | `cd backend && mise exec -- go build ./...` if packaging/deps changed |
| `whisper/` behavior | `cd whisper && mise exec -- uv run pytest` | `cd whisper && mise exec -- uv build` if packaging/deps changed |
| `extension/` behavior | `cd extension && mise exec -- npm run test` | `cd extension && mise exec -- npm run typecheck` for type changes; `cd extension && mise exec -- npm run build` if packaging/deps changed |

## Documentation Rules

- All prose under `docs/` must be in Chinese. Code, commands, paths, config keys, API names, and required references may stay in English.
- Nested `AGENTS.md` takes precedence over root. Check `docs/AGENTS.md` for subproject-specific rules.
- When updating behavioral contracts, commands, APIs, exit codes, directory structure, supported languages, mock/real boundaries, or security boundaries, also check whether `README.md`, `backend/README.md`, `whisper/README.md`, `extension/README.md`, and relevant `docs/` files need updates.

## Common Pitfalls

- **Backend default is mock runner** — it stays offline. `LSI_RUNNER_MODE=real` triggers real download/transcription/translation.
- **Extension v1 only supports `en` and `zh`** — do not promise a complete language list in UI or docs.
- **CORS is localhost-only** with explicit port required.
- **Unit tests must be offline** — no real YouTube, model downloads, GPU, external LLM, or local private data.
- **Lockfiles**: update via tooling, never manual edits — `go.sum` via Go, `uv.lock` via uv, `package-lock.json` via npm.

## graphify

This project has a graphify knowledge graph at graphify-out/.

Rules:
- Before answering architecture or codebase questions, read graphify-out/GRAPH_REPORT.md for god nodes and community structure
- If graphify-out/wiki/index.md exists, navigate it instead of reading raw files
- For cross-module "how does X relate to Y" questions, prefer `graphify query "<question>"`, `graphify path "<A>" "<B>"`, or `graphify explain "<concept>"` over grep — these traverse the graph's EXTRACTED + INFERRED edges instead of scanning files
- After modifying code files in this session, run `graphify update .` to keep the graph current (AST-only, no API cost)
