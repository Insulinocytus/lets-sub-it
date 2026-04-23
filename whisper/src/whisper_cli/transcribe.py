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
