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
    if total_milliseconds < 0:
        raise ValueError("segment timestamps must be non-negative")
    total_seconds, milliseconds = divmod(total_milliseconds, 1000)
    minutes_total, whole_seconds = divmod(total_seconds, 60)
    hours, minutes = divmod(minutes_total, 60)
    return f"{hours:02d}:{minutes:02d}:{whole_seconds:02d}.{milliseconds:03d}"


def format_timestamp_from_milliseconds(total_milliseconds: int) -> str:
    if total_milliseconds < 0:
        raise ValueError("segment timestamps must be non-negative")
    total_seconds, milliseconds = divmod(total_milliseconds, 1000)
    minutes_total, whole_seconds = divmod(total_seconds, 60)
    hours, minutes = divmod(minutes_total, 60)
    return f"{hours:02d}:{minutes:02d}:{whole_seconds:02d}.{milliseconds:03d}"


def render_vtt(segments: Sequence[Segment]) -> str:
    if not segments:
        raise ValueError("segments must not be empty")

    lines: list[str] = ["WEBVTT", ""]
    previous_start_ms = 0

    for index, segment in enumerate(segments):
        text = segment.text.strip()
        if not text:
            raise ValueError("segment text must not be empty")
        if "-->" in text:
            raise ValueError("segment text must not contain -->")
        if any(not line.strip() for line in text.splitlines()):
            raise ValueError("segment text must not contain blank lines")
        if segment.start < 0 or segment.end < 0:
            raise ValueError("segment timestamps must be non-negative")
        start_ms = round(segment.start * 1000)
        end_ms = round(segment.end * 1000)
        if start_ms < 0 or end_ms < 0:
            raise ValueError("segment timestamps must be non-negative")
        if end_ms <= start_ms:
            raise ValueError("segment end must be greater than segment start")
        if index > 0 and start_ms < previous_start_ms:
            raise ValueError("segment start must be monotonic")

        lines.append(
            f"{format_timestamp_from_milliseconds(start_ms)} --> "
            f"{format_timestamp_from_milliseconds(end_ms)}"
        )
        lines.append(text)
        lines.append("")
        previous_start_ms = start_ms

    return "\n".join(lines).rstrip() + "\n"
