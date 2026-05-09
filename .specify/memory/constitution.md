<!--
Sync Impact Report
Version change: template -> 1.0.0
Modified principles:
- Template principle 1 -> I. Local-First Self-Hosted Scope
- Template principle 2 -> II. Real Subtitle Pipeline Contracts
- Template principle 3 -> III. Offline Repeatable Test Discipline
- Template principle 4 -> IV. Security and Secret Boundaries
- Template principle 5 -> V. Simplicity and Module Discipline
Added sections:
- Operational Constraints
- Development Workflow and Quality Gates
Removed sections:
- None
Templates requiring updates:
- .specify/templates/plan-template.md: updated
- .specify/templates/spec-template.md: updated
- .specify/templates/tasks-template.md: updated
- .specify/templates/commands/*.md: not present in this project
Runtime guidance reviewed:
- README.md: reviewed, no change required
- AGENTS.md: reviewed, no change required
Follow-up TODOs:
- None
-->
# Lets Sub It Constitution

## Core Principles

### I. Local-First Self-Hosted Scope
Lets Sub It MUST remain a single-user, local self-hosted tool for public YouTube
videos. Features MUST NOT add private video access, login session handling, cookie
import, auth bypass, multi-tenant operation, or claims that the backend is ready
for public internet exposure. Any feature that changes the local-only deployment
model MUST include an explicit security design and constitution amendment.

Rationale: the repository has no user authentication boundary, and its value is a
controlled local subtitle workflow rather than a hosted public service.

### II. Real Subtitle Pipeline Contracts
The production backend MUST preserve the real job pipeline: create or reuse a job,
download audio with `yt-dlp` and `ffmpeg`, transcribe with `whisper-cli`, translate
through a backend-configured OpenAI-compatible LLM, and package WebVTT outputs.
Module boundaries MUST stay explicit: `backend/` owns HTTP, persistence, runner
orchestration, and safe VTT serving; `whisper/` owns local transcription and VTT
validation; `extension/` owns Chrome MV3 UI, storage, runtime messaging, and
YouTube overlay behavior. Content scripts MUST NOT call the backend directly.

Rationale: clear contracts keep the multi-module MVP debuggable and allow each
piece to be tested without invoking the full real-world pipeline.

### III. Offline Repeatable Test Discipline
Behavior changes MUST add or update tests in the touched module. Unit tests MUST
remain offline and repeatable: no real YouTube access, model downloads, GPU
requirements, external LLM calls, provider keys, private local data, or long-lived
network dependencies. Real transcription, download, and translation paths MUST be
covered through fakes, stubs, local test servers, or focused contract tests unless
the verification is explicitly documented as a manual smoke test.

Rationale: the real pipeline depends on heavyweight tools and paid or networked
services, so correctness must be protected by deterministic local tests.

### IV. Security and Secret Boundaries
Provider keys and translation credentials MUST stay in backend/server
configuration only. The extension MUST NOT store secrets or call translation
providers directly. API responses MUST NOT expose local absolute file paths.
Subtitle file serving MUST remain confined to each job work directory and preserve
path traversal and symlink escape protections. Extension backend origins MUST
remain localhost HTTP origins with explicit ports unless a constitution amendment
and manifest/security update are approved.

Rationale: local filesystem access, browser extension permissions, and LLM
credentials create the highest-impact failure modes in this project.

### V. Simplicity and Module Discipline
Changes MUST be surgical and follow existing module boundaries, local patterns,
and pinned toolchains. New frameworks, queue systems, background processing
systems, broad language support, provider abstractions, or cross-module rewrites
MUST be justified by a current requirement and documented in the implementation
plan. Single-use abstractions and speculative configurability MUST be avoided.

Rationale: the project is an MVP with three cooperating modules; small, direct
changes reduce regressions and keep the system understandable.

## Operational Constraints

Development and runtime work MUST respect the repository's pinned toolchain:
Go 1.22 for `backend/`, Python 3.12 with `uv` for `whisper/`, and Node.js 22 with
npm for `extension/`. Agent and developer shell commands SHOULD be run through
`mise exec --` where applicable.

The extension MVP supports only `en` and `zh` language pairs, and the source and
target language MUST differ. Backend URL validation and manifest host permissions
MUST remain limited to `http://127.0.0.1:*/*` and `http://localhost:*/*` unless
the security boundary is amended.

Generated artifacts, SQLite databases, model files, download caches, real audio
samples, `.env`, API keys, and LLM request logs MUST NOT be committed.

## Development Workflow and Quality Gates

Plans MUST identify the affected module or modules and state how each relevant
constitution principle is satisfied. Implementation tasks MUST preserve
independent user-story slices where Spec Kit is used, but they MUST also include
module-appropriate tests for behavior changes.

Before editing, contributors SHOULD inspect `git status --short` and avoid
overwriting unrelated work. Before completion, contributors MUST run the narrowest
meaningful verification for changed files and report any skipped checks with a
reason. Required verification by changed module is:

- `backend/`: `cd backend && mise exec -- go test ./...`
- `whisper/`: `cd whisper && mise exec -- uv run pytest`
- `extension/`: `cd extension && mise exec -- npm run test`
- Documentation-only changes: `git diff --check`

Behavioral contract changes MUST also check whether README files, specs, and
agent guidance need updates.

## Governance

This constitution supersedes conflicting local practices for Spec Kit planning,
task generation, and implementation review. Amendments MUST update this file,
include a Sync Impact Report, propagate required changes to dependent templates
or runtime guidance, and document any migration or manual follow-up.

Versioning follows semantic versioning:

- MAJOR: incompatible governance changes, principle removals, or redefinitions
  that invalidate existing plans.
- MINOR: new principles, new required sections, or materially expanded
  compliance gates.
- PATCH: clarifications, wording fixes, or non-semantic refinements.

Every implementation plan MUST pass the Constitution Check before Phase 0
research and re-check after Phase 1 design. Code review MUST treat unresolved
constitution violations as blockers unless the plan records an approved
exception and the simpler alternative that was rejected.

**Version**: 1.0.0 | **Ratified**: 2026-05-10 | **Last Amended**: 2026-05-10
