# Implementation Plan: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE] | **Spec**: [link]
**Input**: Feature specification from `/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

[Extract from feature spec: primary requirement + technical approach from research]

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: [e.g., Python 3.11, Swift 5.9, Rust 1.75 or NEEDS CLARIFICATION]
**Primary Dependencies**: [e.g., FastAPI, UIKit, LLVM or NEEDS CLARIFICATION]
**Storage**: [if applicable, e.g., PostgreSQL, CoreData, files or N/A]
**Testing**: [e.g., pytest, XCTest, cargo test or NEEDS CLARIFICATION]
**Target Platform**: [e.g., Linux server, iOS 15+, WASM or NEEDS CLARIFICATION]
**Project Type**: [e.g., library/cli/web-service/mobile-app/compiler/desktop-app or NEEDS CLARIFICATION]
**Performance Goals**: [domain-specific, e.g., 1000 req/s, 10k lines/sec, 60 fps or NEEDS CLARIFICATION]
**Constraints**: [domain-specific, e.g., <200ms p95, <100MB memory, offline-capable or NEEDS CLARIFICATION]
**Scale/Scope**: [domain-specific, e.g., 10k users, 1M LOC, 50 screens or NEEDS CLARIFICATION]

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Local-first scope**: Plan keeps the feature within single-user local
  self-hosting for public YouTube videos. Any private video, login, cookie,
  auth-bypass, multi-tenant, or public-internet exposure change is explicitly
  rejected or documented as a constitution amendment.
- **Pipeline and module contracts**: Plan preserves the `backend/`, `whisper/`,
  and `extension/` boundaries. Content scripts do not call the backend directly,
  and API responses do not expose local absolute file paths.
- **Offline tests**: Plan lists deterministic tests for behavior changes and
  avoids real YouTube, model downloads, GPU, external LLM calls, provider keys,
  private local data, or long-lived network dependencies in unit tests.
- **Security and secrets**: Plan keeps provider keys in backend configuration,
  preserves localhost-only extension backend origins with explicit ports, and
  maintains safe subtitle file serving boundaries.
- **Simplicity**: Plan justifies any new dependency, framework, queue/background
  system, broad language support, provider abstraction, or cross-module rewrite;
  otherwise the change follows existing local patterns.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
backend/
├── cmd/server/
└── internal/
    ├── api/
    ├── app/
    ├── runner/
    └── store/

whisper/
├── src/whisper_cli/
└── tests/

extension/
├── entrypoints/
└── src/
    ├── api/
    ├── popup/
    ├── storage/
    ├── subtitles/
    └── youtube/

docs/
└── [Chinese project docs and plans when affected]
```

**Structure Decision**: [Document the selected structure and reference the real
directories captured above]

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
