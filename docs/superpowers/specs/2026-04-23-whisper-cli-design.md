# Whisper CLI Design

## Context

This document defines the MVP design for the `fast-whisper cli` subtask from [PRD.md](/Users/demo/.codex/worktrees/789b/lets-sub-it/PRD.md).

The current implementation scope is intentionally narrow:

- Build a standalone local Python CLI under the repository root at `whisper/`
- Manage Python and `uv` versions with `mise`
- Manage Python dependencies and lockfile with `uv`
- Accept a local audio file path as input
- Produce a validated `source.vtt` file as output

The Go runner, API integration, download pipeline, and translation pipeline belong to later design documents.

## Goals

- Provide a stable local command that the Go runner can later invoke through `exec`
- Keep the contract minimal and explicit: input path in, `source.vtt` out, exit code indicates result
- Keep the project self-contained so setup, debugging, and dependency management stay local to `whisper/`
- Validate the output enough that downstream code can trust successful runs

## Deferred Scope

- Long-running service
- Queue or worker management
- YouTube download logic
- `ffmpeg` integration
- Result persistence beyond the generated VTT file
- Translation, bilingual packaging, and subtitle serving

## Constraints

- The project lives in a new top-level `whisper/` directory
- `mise` manages the local toolchain versions
- `uv` manages dependencies and lockfile
- The CLI must stay small enough for later `exec` integration from Go
- The implementation should match the PRD assumption that success can be determined from exit code and output artifact

## Recommended Approach

### Option A: Local Python CLI using `faster-whisper`

Recommended.

Implement a small Python package with a console entrypoint. The CLI loads `faster-whisper`, transcribes a local audio file, serializes the returned segments into WebVTT, validates the generated output, and exits with stable status codes.

Why this approach fits best:

- We control the command-line interface and output format
- We control error handling and exit code mapping
- We avoid external CLI version drift
- The contract is straightforward for future Go integration

### Option B: Python wrapper around a system `fast-whisper` command

Acceptable but weaker.

This reduces implementation time. It also delegates argument shape, stdout format, and error behavior to an external command whose version may drift across environments.

### Option C: Dedicated transcription service

Out of scope for MVP.

This creates more moving parts than needed for the current project stage.

## Project Layout

```text
whisper/
  .mise.toml
  pyproject.toml
  uv.lock
  README.md
  src/
    whisper_cli/
      __init__.py
      cli.py
      transcribe.py
      vtt.py
  tests/
    test_cli.py
    test_vtt.py
```

### File Responsibilities

- `whisper/.mise.toml`
  - Pins `python` and `uv`
- `whisper/pyproject.toml`
  - Defines project metadata, dependencies, and the `whisper-cli` script
- `whisper/uv.lock`
  - Locks dependency versions
- `whisper/src/whisper_cli/cli.py`
  - Parses arguments, runs the transcription flow, prints machine-readable success output, and maps exceptions to exit codes
- `whisper/src/whisper_cli/transcribe.py`
  - Wraps `faster-whisper` model loading and transcription calls
- `whisper/src/whisper_cli/vtt.py`
  - Serializes segments into standard WebVTT and validates the result
- `whisper/tests/test_cli.py`
  - Covers CLI contract and failure mapping
- `whisper/tests/test_vtt.py`
  - Covers WebVTT formatting and validation rules
- `whisper/README.md`
  - Documents local setup and execution

## CLI Contract

### Command

```bash
cd whisper
mise install
uv sync
uv run whisper-cli \
  --input /path/to/audio.mp3 \
  --output /path/to/source.vtt \
  --model small \
  --language ja
```

### Arguments

- `--input`
  - Required
  - Local audio file path
- `--output`
  - Required
  - Target VTT file path
- `--model`
  - Optional
  - Whisper model name
  - Default: `small`
- `--language`
  - Optional
  - Source language code
  - When omitted, the library performs language detection

The first version keeps the interface limited to these four options. This keeps the command stable and avoids speculative configurability.

## Execution Flow

1. Parse CLI arguments
2. Validate required arguments and input path existence
3. Ensure the output parent directory exists
4. Load the configured `faster-whisper` model
5. Transcribe the local audio file into ordered segments
6. Serialize segments into WebVTT
7. Validate the generated VTT structure
8. Write `source.vtt`
9. Print a one-line JSON success payload to stdout
10. Exit with code `0`

## Output Contract

### Success stdout

On success, the CLI writes one line of JSON to stdout.

Example:

```json
{"output":"/tmp/source.vtt","language":"ja","duration_seconds":123.4,"segments":87}
```

Fields:

- `output`
  - Absolute or resolved output path
- `language`
  - Detected or configured source language
- `duration_seconds`
  - Total transcribed duration as a number
- `segments`
  - Number of generated subtitle cues

### Error stderr

On failure, the CLI writes a concise one-line error summary to stderr.

Example:

```text
transcription failed: model download error
```

The CLI keeps stderr plain text, without progress bars, colors, or structured logging noise.

## Exit Codes

- `0`
  - Success
  - `source.vtt` exists and passes validation
- `2`
  - Argument or input validation error
  - Examples: missing required args, input path missing, invalid output path
- `3`
  - Transcription execution failure
  - Examples: model load failure, decode failure, runtime exception inside the transcription layer
- `4`
  - Output validation failure
  - Examples: empty segment list, invalid VTT structure, zero generated cues

These codes are intentionally small and stable so the future Go runner can map them directly to stage failures.

## WebVTT Rules

The generated output must satisfy all of the following:

- The file starts with `WEBVTT`
- Cue count is greater than zero
- Each cue preserves the original segment timing
- Cue timestamps are strictly non-decreasing
- Each cue contains subtitle text
- The output file is overwritten if it already exists
- Parent directories are created automatically when missing

The initial VTT writer keeps cues simple:

- cue identifier omitted
- one cue block per segment
- text content taken directly from the transcribed segment text after trimming surrounding whitespace

## Architecture

### `transcribe.py`

This module owns the library boundary. It converts CLI arguments into a `faster-whisper` call and returns a minimal in-memory transcription result:

- detected or provided language
- ordered segments
- aggregate duration metadata

This module returns transcription data and leaves exit code mapping to `cli.py`.

### `vtt.py`

This module owns VTT formatting and validation. It accepts normalized segments and returns serialized WebVTT content or raises a validation error.

This separation keeps formatting logic testable without loading a model.

### `cli.py`

This module owns the public command behavior:

- parse args
- call the transcription layer
- call the VTT layer
- write output
- print success JSON
- convert known failure classes into stable exit codes

## Error Handling

The first version uses a few explicit error classes:

- input validation error
- transcription error
- output validation error

Each class maps cleanly to one exit code. The CLI prints the error summary and exits. This keeps failures legible for humans and easy to consume for the future Go runner.

## Testing Strategy

The implementation should start with test-first development for the CLI contract and VTT rules.

### Required tests

1. `test_vtt_writes_header_and_cues`
   - verifies `WEBVTT`, timestamps, and cue text
2. `test_vtt_rejects_empty_segments`
   - verifies empty output is rejected
3. `test_cli_requires_input_and_output`
   - verifies missing args return exit code `2`
4. `test_cli_creates_parent_directory_for_output`
   - verifies nested output paths are created automatically
5. `test_cli_prints_json_on_success`
   - verifies stdout shape on success
6. `test_cli_returns_code_3_when_transcriber_fails`
   - verifies transcription failures map to exit code `3`

### Deferred tests

Real-audio integration tests are deferred to a later phase because they require model availability, longer runtime, and a less stable local environment. The initial phase focuses on deterministic unit and CLI contract tests.

## Assumptions

- The caller provides a readable local audio file
- The host environment can install Python and `uv` through `mise`
- The host environment can install and run the dependencies required by `faster-whisper`
- Model download and caching behavior can stay with the library defaults in the first version

## Risks

- The first local run may spend time downloading model assets
- Host-specific native dependency issues may affect setup
- Long audio files may increase local execution time significantly
- Whisper output quality depends on the chosen model and source audio quality

## Open Decisions Chosen For MVP

To remove ambiguity for implementation, the following choices are fixed for the first version:

- The package is a standalone top-level project at `whisper/`
- Dependency management uses `uv`
- Toolchain version management uses `mise`
- The command exposes exactly four user-facing flags
- Successful runs overwrite the target output file
- Success output uses one-line JSON on stdout
- Error output uses one-line plain text on stderr
- Empty transcription results count as failure

## Success Criteria

The subproject is complete when:

- `mise install` and `uv sync` prepare the environment inside `whisper/`
- `uv run whisper-cli --input ... --output ...` generates a valid `source.vtt`
- CLI success and failure paths return the expected exit codes
- The required tests pass locally
- The future Go runner can treat the CLI as a black-box command with a stable contract
