from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

from faster_whisper import WhisperModel

from whisper_cli.vtt import Segment


class InputValidationError(Exception):
    pass


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
    try:
        raw_segments, info = model.transcribe(str(input_path), language=language)
    except ValueError as exc:
        if "valid language code" in str(exc):
            raise InputValidationError(str(exc)) from exc
        raise
    segments = [
        Segment(start=segment.start, end=segment.end, text=segment.text)
        for segment in raw_segments
    ]
    if not segments:
        raise RuntimeError("transcription produced no segments")

    return TranscriptionResult(
        language=info.language,
        duration_seconds=info.duration,
        segments=segments,
    )
