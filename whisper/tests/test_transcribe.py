from types import SimpleNamespace

import pytest

from whisper_cli.transcribe import (
    InputValidationError,
    TranscriptionResult,
    transcribe_audio,
)


class FakeSegment:
    def __init__(self, start: float, end: float, text: str) -> None:
        self.start = start
        self.end = end
        self.text = text


class FakeModel:
    def __init__(self, model_name: str) -> None:
        self.model_name = model_name

    def transcribe(self, input_path: str, language: str):
        info = SimpleNamespace(language=language, duration=2.5)
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


def test_transcribe_audio_uses_sdk_reported_duration(monkeypatch, tmp_path):
    input_path = tmp_path / "audio.mp3"
    input_path.write_bytes(b"audio")

    class DurationModel:
        def __init__(self, model_name: str) -> None:
            self.model_name = model_name

        def transcribe(self, input_path: str, language: str):
            info = SimpleNamespace(language=language, duration=3.75)
            segments = [
                FakeSegment(0.0, 1.25, "hello"),
                FakeSegment(1.25, 2.5, "world"),
            ]
            return iter(segments), info

    monkeypatch.setattr("whisper_cli.transcribe.WhisperModel", DurationModel)

    result = transcribe_audio(
        input_path=input_path,
        model_name="small",
        language="ja",
    )

    assert result.duration_seconds == 3.75


def test_transcribe_audio_allows_empty_segment_output(monkeypatch, tmp_path):
    input_path = tmp_path / "audio.mp3"
    input_path.write_bytes(b"audio")

    class EmptyModel:
        def __init__(self, model_name: str) -> None:
            self.model_name = model_name

        def transcribe(self, input_path: str, language: str):
            return iter([]), SimpleNamespace(language=language, duration=0.0)

    monkeypatch.setattr("whisper_cli.transcribe.WhisperModel", EmptyModel)

    result = transcribe_audio(
        input_path=input_path,
        model_name="small",
        language="ja",
    )

    assert result.segments == []


def test_transcribe_audio_rejects_english_only_model_with_non_english_language(
    monkeypatch, tmp_path
):
    input_path = tmp_path / "audio.mp3"
    input_path.write_bytes(b"audio")

    class EnglishOnlyModel:
        def __init__(self, model_name: str) -> None:
            self.model_name = model_name
            self.model = SimpleNamespace(is_multilingual=False)

        def transcribe(self, input_path: str, language: str):
            info = SimpleNamespace(language="en", duration=2.5)
            segments = [FakeSegment(0.0, 1.0, "hello")]
            return iter(segments), info

    monkeypatch.setattr("whisper_cli.transcribe.WhisperModel", EnglishOnlyModel)

    with pytest.raises(InputValidationError, match="only supports language 'en'"):
        transcribe_audio(
            input_path=input_path,
            model_name="small.en",
            language="ja",
        )
