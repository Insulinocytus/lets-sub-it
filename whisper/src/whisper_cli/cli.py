from __future__ import annotations

import argparse
import json
import sys
from collections.abc import Sequence
from pathlib import Path

from whisper_cli.transcribe import InputValidationError, transcribe_audio
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
        with input_path.open("rb"):
            pass
    except OSError as exc:
        print(f"input validation failed: {exc}", file=sys.stderr)
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
    except InputValidationError as exc:
        print(f"input validation failed: {exc}", file=sys.stderr)
        return 2
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
    except UnicodeError as exc:
        print(f"output validation failed: {exc}", file=sys.stderr)
        return 4

    print(
        json.dumps(
            {
                "output": str(output_path.resolve()),
                "language": result.language,
                "duration_seconds": result.duration_seconds,
                "segments": len(result.segments),
            }
        )
    )
    return 0


def main_entry() -> None:
    raise SystemExit(main())
