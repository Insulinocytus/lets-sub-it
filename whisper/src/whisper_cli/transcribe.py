from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

from faster_whisper import WhisperModel

from whisper_cli.vtt import Segment


class InputValidationError(Exception):
    pass


def _raise_input_validation_error(exc: ValueError) -> None:
    message = str(exc)
    if "Invalid model size" in message or "valid language code" in message:
        raise InputValidationError(message) from exc


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
    try:
        model = WhisperModel(model_name)
    except ValueError as exc:
        _raise_input_validation_error(exc)
        raise

    if getattr(getattr(model, "model", None), "is_multilingual", True) is False:
        if language != "en":
            raise InputValidationError(
                f"model '{model_name}' only supports language 'en'"
            )

    try:
        raw_segments, info = model.transcribe(str(input_path), language=language)
    except ValueError as exc:
        _raise_input_validation_error(exc)
        raise
    segments = [
        Segment(start=segment.start, end=segment.end, text=segment.text)
        for segment in raw_segments
    ]
    if not segments:
        raise RuntimeError("transcription produced no segments")

    return TranscriptionResult(
        language=language,
        duration_seconds=info.duration,
        segments=segments,
    )
