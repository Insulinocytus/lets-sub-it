# Whisper SDK CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `whisper/` 下构建一个可被后续 Go runner 通过 `exec` 调用的 `whisper-cli`，直接使用 `faster-whisper` Python SDK 生成并校验 `source.vtt`。

**Architecture:** 根目录 `mise.toml` 统一管理 Python 和 `uv`；`whisper/` 作为独立 Python 子项目，`cli.py` 负责参数解析、stdout/stderr 和退出码映射，`vtt.py` 负责 WebVTT 序列化与校验，`transcribe.py` 负责 `faster-whisper` SDK 适配。实现顺序采用 TDD，先锁 CLI 契约与 VTT 规则，再接入真实 SDK。

**Tech Stack:** Python 3.12, uv, mise, faster-whisper, pytest

---

## File Structure

- Create: `mise.toml`
  - monorepo 根目录统一工具链版本
- Create: `whisper/pyproject.toml`
  - Python package、依赖、console script、pytest 配置
- Create: `whisper/README.md`
  - 本地安装、运行和测试说明
- Create: `whisper/src/whisper_cli/__init__.py`
  - package 边界
- Create: `whisper/src/whisper_cli/cli.py`
  - CLI 参数解析、退出码、文件输出、stdout/stderr
- Create: `whisper/src/whisper_cli/vtt.py`
  - `Segment` 数据结构、时间戳格式化、WebVTT 渲染和基础校验
- Create: `whisper/src/whisper_cli/transcribe.py`
  - `TranscriptionResult` 数据结构和 `faster-whisper` SDK 适配
- Create: `whisper/tests/test_cli.py`
  - CLI 契约测试
- Create: `whisper/tests/test_vtt.py`
  - WebVTT 规则测试
- Create: `whisper/tests/test_transcribe.py`
  - SDK 适配测试，全部使用 fake model，不依赖真实模型下载
- Create: `whisper/uv.lock`
  - 通过 `uv lock` 生成的锁文件

### Task 1: Scaffold Project And Lock CLI Required Args

**Files:**
- Create: `mise.toml`
- Create: `whisper/pyproject.toml`
- Create: `whisper/README.md`
- Create: `whisper/src/whisper_cli/__init__.py`
- Create: `whisper/src/whisper_cli/cli.py`
- Test: `whisper/tests/test_cli.py`

- [ ] **Step 1: Create toolchain config from the monorepo root**

Run:

```bash
mise use python@3.12 uv@latest
```

Expected `mise.toml`:

```toml
# mise.toml
[tools]
python = "3.12"
uv = "latest"
```

- [ ] **Step 2: Create package scaffold**

```toml
# whisper/pyproject.toml
[build-system]
requires = ["hatchling>=1.27.0"]
build-backend = "hatchling.build"

[project]
name = "whisper-cli"
version = "0.1.0"
description = "Local faster-whisper CLI that writes validated VTT output"
readme = "README.md"
requires-python = ">=3.12"
dependencies = [
  "faster-whisper>=1.1.1",
]

[dependency-groups]
dev = [
  "pytest>=8.3.5",
]

[project.scripts]
whisper-cli = "whisper_cli.cli:main_entry"

[tool.pytest.ini_options]
pythonpath = ["src"]
addopts = "-q"
```

```python
# whisper/src/whisper_cli/__init__.py
__all__: list[str] = []
```

````markdown
# whisper/README.md

## Setup

```bash
mise install
cd whisper
uv sync --dev
```

## Run

```bash
uv run whisper-cli --input /path/to/audio.mp3 --output /tmp/source.vtt --model small --language ja
```

## Test

```bash
uv run pytest
```
````

- [ ] **Step 3: Install dependencies and generate the lockfile**

Run:

```bash
mise install
cd whisper
uv sync --dev
uv lock
```

Expected:

```text
Python and uv are installed, the virtual environment is created, and whisper/uv.lock exists.
```

- [ ] **Step 4: Write the failing test for required CLI arguments**

```python
# whisper/tests/test_cli.py
from whisper_cli.cli import main


def test_cli_requires_all_required_arguments(capsys):
    exit_code = main([])

    captured = capsys.readouterr()

    assert exit_code == 2
    assert "required" in captured.err
```

- [ ] **Step 5: Run test to verify it fails**

Run:

```bash
cd whisper
uv run pytest tests/test_cli.py::test_cli_requires_all_required_arguments -v
```

Expected:

```text
FAIL because whisper_cli.cli does not exist yet.
```

- [ ] **Step 6: Write the minimal CLI implementation**

```python
# whisper/src/whisper_cli/cli.py
from __future__ import annotations

import argparse
from collections.abc import Sequence


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="whisper-cli")
    parser.add_argument("--input", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--model", required=True)
    parser.add_argument("--language", required=True)
    return parser


def main(argv: Sequence[str] | None = None) -> int:
    parser = build_parser()
    try:
        parser.parse_args(argv)
    except SystemExit as exc:
        return int(exc.code)
    return 0


def main_entry() -> None:
    raise SystemExit(main())
```

- [ ] **Step 7: Run test to verify it passes**

Run:

```bash
cd whisper
uv run pytest tests/test_cli.py::test_cli_requires_all_required_arguments -v
```

Expected:

```text
PASS
```

- [ ] **Step 8: Commit**

```bash
git add mise.toml whisper/pyproject.toml whisper/README.md whisper/src/whisper_cli/__init__.py whisper/src/whisper_cli/cli.py whisper/tests/test_cli.py whisper/uv.lock
git commit -m "build(whisper): scaffold sdk cli project"
```

### Task 2: Implement WebVTT Rendering And Validation

**Files:**
- Create: `whisper/src/whisper_cli/vtt.py`
- Test: `whisper/tests/test_vtt.py`

- [ ] **Step 1: Write the failing tests for WebVTT output**

```python
# whisper/tests/test_vtt.py
import pytest

from whisper_cli.vtt import Segment, render_vtt


def test_vtt_writes_header_and_cues():
    content = render_vtt(
        [
            Segment(start=0.0, end=1.25, text="hello"),
            Segment(start=1.25, end=2.5, text="world"),
        ]
    )

    assert content.startswith("WEBVTT\n\n")
    assert "00:00:00.000 --> 00:00:01.250" in content
    assert "00:00:01.250 --> 00:00:02.500" in content
    assert "hello" in content
    assert "world" in content


def test_vtt_rejects_empty_segments():
    with pytest.raises(ValueError, match="segments must not be empty"):
        render_vtt([])
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
cd whisper
uv run pytest tests/test_vtt.py -v
```

Expected:

```text
FAIL because whisper_cli.vtt does not exist yet.
```

- [ ] **Step 3: Write the minimal VTT implementation**

```python
# whisper/src/whisper_cli/vtt.py
from __future__ import annotations

from dataclasses import dataclass
from typing import Sequence


@dataclass(frozen=True)
class Segment:
    start: float
    end: float
    text: str


def format_timestamp(seconds: float) -> str:
    total_milliseconds = round(seconds * 1000)
    total_seconds, milliseconds = divmod(total_milliseconds, 1000)
    minutes_total, whole_seconds = divmod(total_seconds, 60)
    hours, minutes = divmod(minutes_total, 60)
    return f"{hours:02d}:{minutes:02d}:{whole_seconds:02d}.{milliseconds:03d}"


def render_vtt(segments: Sequence[Segment]) -> str:
    if not segments:
        raise ValueError("segments must not be empty")

    lines: list[str] = ["WEBVTT", ""]
    previous_start = 0.0

    for index, segment in enumerate(segments):
        text = segment.text.strip()
        if not text:
            raise ValueError("segment text must not be empty")
        if segment.end <= segment.start:
            raise ValueError("segment end must be greater than segment start")
        if index > 0 and segment.start < previous_start:
            raise ValueError("segment start must be monotonic")

        lines.append(
            f"{format_timestamp(segment.start)} --> {format_timestamp(segment.end)}"
        )
        lines.append(text)
        lines.append("")
        previous_start = segment.start

    return "\n".join(lines).rstrip() + "\n"
```

- [ ] **Step 4: Run tests to verify they pass**

Run:

```bash
cd whisper
uv run pytest tests/test_vtt.py -v
```

Expected:

```text
2 passed
```

- [ ] **Step 5: Commit**

```bash
git add whisper/src/whisper_cli/vtt.py whisper/tests/test_vtt.py
git commit -m "feat(whisper): add vtt rendering"
```

### Task 3: Implement CLI Output Contract And Error Mapping

**Files:**
- Modify: `whisper/src/whisper_cli/cli.py`
- Create: `whisper/src/whisper_cli/transcribe.py`
- Modify: `whisper/tests/test_cli.py`

- [ ] **Step 1: Expand CLI contract tests**

```python
# whisper/tests/test_cli.py
import json
from pathlib import Path

from whisper_cli.cli import main
from whisper_cli.transcribe import TranscriptionResult
from whisper_cli.vtt import Segment


def fake_result() -> TranscriptionResult:
    return TranscriptionResult(
        language="ja",
        duration_seconds=2.5,
        segments=[
            Segment(start=0.0, end=1.25, text="hello"),
            Segment(start=1.25, end=2.5, text="world"),
        ],
    )


def test_cli_requires_all_required_arguments(capsys):
    exit_code = main([])

    captured = capsys.readouterr()

    assert exit_code == 2
    assert "required" in captured.err


def test_cli_creates_parent_directory_for_output(monkeypatch, tmp_path):
    input_path = tmp_path / "audio.mp3"
    output_path = tmp_path / "nested" / "source.vtt"
    input_path.write_bytes(b"audio")

    monkeypatch.setattr("whisper_cli.cli.transcribe_audio", lambda **_: fake_result())

    exit_code = main(
        [
            "--input",
            str(input_path),
            "--output",
            str(output_path),
            "--model",
            "small",
            "--language",
            "ja",
        ]
    )

    assert exit_code == 0
    assert output_path.exists()


def test_cli_prints_json_on_success(monkeypatch, tmp_path, capsys):
    input_path = tmp_path / "audio.mp3"
    output_path = tmp_path / "source.vtt"
    input_path.write_bytes(b"audio")

    monkeypatch.setattr("whisper_cli.cli.transcribe_audio", lambda **_: fake_result())

    exit_code = main(
        [
            "--input",
            str(input_path),
            "--output",
            str(output_path),
            "--model",
            "small",
            "--language",
            "ja",
        ]
    )

    captured = capsys.readouterr()
    payload = json.loads(captured.out)

    assert exit_code == 0
    assert payload == {
        "output": str(output_path.resolve()),
        "language": "ja",
        "duration_seconds": 2.5,
        "segments": 2,
    }


def test_cli_returns_code_3_when_transcriber_fails(monkeypatch, tmp_path, capsys):
    input_path = tmp_path / "audio.mp3"
    output_path = tmp_path / "source.vtt"
    input_path.write_bytes(b"audio")

    def raise_failure(**_):
        raise RuntimeError("model download error")

    monkeypatch.setattr("whisper_cli.cli.transcribe_audio", raise_failure)

    exit_code = main(
        [
            "--input",
            str(input_path),
            "--output",
            str(output_path),
            "--model",
            "small",
            "--language",
            "ja",
        ]
    )

    captured = capsys.readouterr()

    assert exit_code == 3
    assert "transcription failed: model download error" in captured.err
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
cd whisper
uv run pytest tests/test_cli.py -v
```

Expected:

```text
FAIL because whisper_cli.transcribe does not exist and cli.py does not implement the success path.
```

- [ ] **Step 3: Write minimal transcribe result types and full CLI contract**

```python
# whisper/src/whisper_cli/transcribe.py
from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

from whisper_cli.vtt import Segment


@dataclass(frozen=True)
class TranscriptionResult:
    language: str
    duration_seconds: float
    segments: list[Segment]


def transcribe_audio(
    *,
    input_path: Path,
    model_name: str,
    language: str,
) -> TranscriptionResult:
    raise RuntimeError("SDK integration not implemented yet")
```

```python
# whisper/src/whisper_cli/cli.py
from __future__ import annotations

import argparse
import json
import sys
from collections.abc import Sequence
from pathlib import Path

from whisper_cli.transcribe import transcribe_audio
from whisper_cli.vtt import render_vtt


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="whisper-cli")
    parser.add_argument("--input", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--model", required=True)
    parser.add_argument("--language", required=True)
    return parser


def main(argv: Sequence[str] | None = None) -> int:
    parser = build_parser()
    try:
        args = parser.parse_args(argv)
    except SystemExit as exc:
        return int(exc.code)

    input_path = Path(args.input)
    output_path = Path(args.output)

    if not input_path.is_file():
        print(f"input validation failed: file not found: {input_path}", file=sys.stderr)
        return 2

    try:
        output_path.parent.mkdir(parents=True, exist_ok=True)
    except OSError as exc:
        print(f"input validation failed: {exc}", file=sys.stderr)
        return 2

    try:
        result = transcribe_audio(
            input_path=input_path,
            model_name=args.model,
            language=args.language,
        )
    except Exception as exc:
        print(f"transcription failed: {exc}", file=sys.stderr)
        return 3

    try:
        content = render_vtt(result.segments)
    except ValueError as exc:
        print(f"output validation failed: {exc}", file=sys.stderr)
        return 4
    try:
        output_path.write_text(content, encoding="utf-8")
    except OSError as exc:
        print(f"input validation failed: {exc}", file=sys.stderr)
        return 2

    payload = {
        "output": str(output_path.resolve()),
        "language": result.language,
        "duration_seconds": result.duration_seconds,
        "segments": len(result.segments),
    }
    print(json.dumps(payload, ensure_ascii=True))
    return 0


def main_entry() -> None:
    raise SystemExit(main())
```

- [ ] **Step 4: Run tests to verify they pass**

Run:

```bash
cd whisper
uv run pytest tests/test_cli.py tests/test_vtt.py -v
```

Expected:

```text
6 passed
```

- [ ] **Step 5: Commit**

```bash
git add whisper/src/whisper_cli/cli.py whisper/src/whisper_cli/transcribe.py whisper/tests/test_cli.py whisper/tests/test_vtt.py
git commit -m "feat(whisper): add cli contract and exit codes"
```

### Task 4: Integrate The Faster-Whisper SDK

**Files:**
- Modify: `whisper/src/whisper_cli/transcribe.py`
- Create: `whisper/tests/test_transcribe.py`
- Modify: `whisper/README.md`

- [ ] **Step 1: Write the failing SDK adapter tests**

```python
# whisper/tests/test_transcribe.py
from types import SimpleNamespace

import pytest

from whisper_cli.transcribe import TranscriptionResult, transcribe_audio


class FakeSegment:
    def __init__(self, start: float, end: float, text: str) -> None:
        self.start = start
        self.end = end
        self.text = text


class FakeModel:
    def __init__(self, model_name: str) -> None:
        self.model_name = model_name

    def transcribe(self, input_path: str, language: str):
        info = SimpleNamespace(language=language)
        segments = [
            FakeSegment(0.0, 1.25, "hello"),
            FakeSegment(1.25, 2.5, "world"),
        ]
        return iter(segments), info


def test_transcribe_audio_uses_sdk_and_builds_result(monkeypatch, tmp_path):
    input_path = tmp_path / "audio.mp3"
    input_path.write_bytes(b"audio")

    monkeypatch.setattr("whisper_cli.transcribe.WhisperModel", FakeModel)

    result = transcribe_audio(
        input_path=input_path,
        model_name="small",
        language="ja",
    )

    assert isinstance(result, TranscriptionResult)
    assert result.language == "ja"
    assert result.duration_seconds == 2.5
    assert [segment.text for segment in result.segments] == ["hello", "world"]


def test_transcribe_audio_rejects_empty_segment_output(monkeypatch, tmp_path):
    input_path = tmp_path / "audio.mp3"
    input_path.write_bytes(b"audio")

    class EmptyModel:
        def __init__(self, model_name: str) -> None:
            self.model_name = model_name

        def transcribe(self, input_path: str, language: str):
            return iter([]), SimpleNamespace(language=language)

    monkeypatch.setattr("whisper_cli.transcribe.WhisperModel", EmptyModel)

    with pytest.raises(RuntimeError, match="transcription produced no segments"):
        transcribe_audio(
            input_path=input_path,
            model_name="small",
            language="ja",
        )
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
cd whisper
uv run pytest tests/test_transcribe.py -v
```

Expected:

```text
FAIL because whisper_cli.transcribe has not imported WhisperModel yet and transcribe_audio is still a placeholder.
```

- [ ] **Step 3: Implement the faster-whisper adapter**

```python
# whisper/src/whisper_cli/transcribe.py
from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

from faster_whisper import WhisperModel

from whisper_cli.vtt import Segment


@dataclass(frozen=True)
class TranscriptionResult:
    language: str
    duration_seconds: float
    segments: list[Segment]


def transcribe_audio(
    *,
    input_path: Path,
    model_name: str,
    language: str,
) -> TranscriptionResult:
    model = WhisperModel(model_name)
    raw_segments, info = model.transcribe(str(input_path), language=language)

    segments = [
        Segment(start=segment.start, end=segment.end, text=segment.text)
        for segment in raw_segments
    ]
    if not segments:
        raise RuntimeError("transcription produced no segments")

    duration_seconds = segments[-1].end
    return TranscriptionResult(
        language=info.language,
        duration_seconds=duration_seconds,
        segments=segments,
    )
```

- [ ] **Step 4: Update README with the final developer workflow**

````markdown
# whisper/README.md

## Setup

```bash
mise install
cd whisper
uv sync --dev
```

## Run

```bash
uv run whisper-cli \
  --input /path/to/audio.mp3 \
  --output /tmp/source.vtt \
  --model small \
  --language ja
```

## Test

```bash
uv run pytest
```

## Output Contract

- success exit code: `0`
- input validation exit code: `2`
- transcription exit code: `3`
- output validation exit code: `4`
````

- [ ] **Step 5: Run tests to verify they pass**

Run:

```bash
cd whisper
uv run pytest tests/test_transcribe.py tests/test_cli.py tests/test_vtt.py -v
```

Expected:

```text
8 passed
```

- [ ] **Step 6: Run the full local verification**

Run:

```bash
cd whisper
uv run pytest -v
```

Expected:

```text
All whisper tests pass with no failures.
```

- [ ] **Step 7: Commit**

```bash
git add whisper/src/whisper_cli/transcribe.py whisper/tests/test_transcribe.py whisper/tests/test_cli.py whisper/tests/test_vtt.py whisper/README.md
git commit -m "feat(whisper): integrate faster-whisper sdk"
```

## Self-Review

### Spec Coverage

- 根目录 `mise.toml` 统一工具链版本: Task 1
- `whisper/` 作为独立 Python 子项目: Task 1
- `uv` 依赖和锁文件: Task 1
- `whisper-cli` 直接调用 `faster-whisper` Python SDK: Task 4
- `--input`、`--output`、`--model`、`--language` 四个必填参数: Task 1, Task 3
- 生成并校验 `source.vtt`: Task 2, Task 3
- stdout JSON 契约: Task 3
- 退出码 `0/2/3/4`: Task 3
- TDD 顺序和首批失败测试: Task 1, Task 2, Task 3, Task 4
- README 开发者说明: Task 1, Task 4

### Placeholder Scan

- 已检查，没有 `TBD`、`TODO`、`implement later` 这类占位词
- 每个代码步骤都给出具体文件内容
- 每个运行步骤都给出具体命令和预期结果

### Type Consistency

- `Segment` 统一定义在 `whisper/src/whisper_cli/vtt.py`
- `TranscriptionResult` 统一定义在 `whisper/src/whisper_cli/transcribe.py`
- `main(argv)` 统一返回 `int`
- `main_entry()` 统一作为 console script 入口
