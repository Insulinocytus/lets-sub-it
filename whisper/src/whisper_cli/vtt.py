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
