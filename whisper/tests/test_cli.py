import json

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
