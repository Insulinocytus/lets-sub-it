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


def test_vtt_rejects_zero_duration_after_millisecond_rounding():
    with pytest.raises(ValueError, match="segment end must be greater than segment start"):
        render_vtt([Segment(start=1.0004, end=1.0005, text="x")])


def test_vtt_rejects_negative_timestamps():
    with pytest.raises(ValueError, match="segment timestamps must be non-negative"):
        render_vtt([Segment(start=-0.1, end=0.5, text="x")])
