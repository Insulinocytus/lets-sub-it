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


def test_cli_creates_parent_directory_for_output(tmp_path, monkeypatch, capsys):
    input_path = tmp_path / "audio.mp3"
    input_path.write_text("audio")
    output_path = tmp_path / "nested" / "result.vtt"

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

    assert exit_code == 0
    assert output_path.parent.is_dir()
    assert output_path.is_file()
    payload = json.loads(captured.out)
    assert payload["output"] == str(output_path.resolve())


def test_cli_prints_json_on_success(tmp_path, monkeypatch, capsys):
    input_path = tmp_path / "audio.mp3"
    input_path.write_text("audio")
    output_path = tmp_path / "result.vtt"

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

    assert exit_code == 0
    assert captured.err == ""
    assert json.loads(captured.out) == {
        "output": str(output_path.resolve()),
        "language": "ja",
        "duration_seconds": 2.5,
        "segments": 2,
    }


def test_cli_returns_code_3_when_transcriber_fails(tmp_path, monkeypatch, capsys):
    input_path = tmp_path / "audio.mp3"
    input_path.write_text("audio")
    output_path = tmp_path / "result.vtt"

    def raise_error(**_):
        raise RuntimeError("model download error")

    monkeypatch.setattr("whisper_cli.cli.transcribe_audio", raise_error)

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


def test_cli_writes_output_with_explicit_utf8_encoding(
    tmp_path, monkeypatch, capsys
):
    input_path = tmp_path / "audio.mp3"
    input_path.write_text("audio")
    output_path = tmp_path / "result.vtt"
    original_write_text = Path.write_text

    def fake_transcription_result() -> TranscriptionResult:
        return TranscriptionResult(
            language="ja",
            duration_seconds=1.0,
            segments=[Segment(start=0.0, end=1.0, text="こんにちは")],
        )

    def write_text_with_encoding_check(
        self, data, encoding=None, errors=None, newline=None
    ):
        if encoding is None:
            raise UnicodeEncodeError("ascii", data, 0, 1, "ordinal not in range")
        assert encoding == "utf-8"
        return original_write_text(
            self, data, encoding=encoding, errors=errors, newline=newline
        )

    monkeypatch.setattr(
        "whisper_cli.cli.transcribe_audio", lambda **_: fake_transcription_result()
    )
    monkeypatch.setattr(Path, "write_text", write_text_with_encoding_check)

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

    assert exit_code == 0
    assert captured.err == ""
    assert "こんにちは" in output_path.read_text(encoding="utf-8")


def test_cli_returns_code_4_when_output_path_is_directory(
    tmp_path, monkeypatch, capsys
):
    input_path = tmp_path / "audio.mp3"
    input_path.write_text("audio")
    output_path = tmp_path / "result.vtt"
    output_path.mkdir()

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

    assert exit_code == 4
    assert "output validation failed:" in captured.err
