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

    return TranscriptionResult(
        language=info.language,
        duration_seconds=segments[-1].end,
        segments=segments,
    )
