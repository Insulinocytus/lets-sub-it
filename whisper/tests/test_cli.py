from whisper_cli.cli import main


def test_cli_requires_all_required_arguments(capsys):
    exit_code = main([])

    captured = capsys.readouterr()

    assert exit_code == 2
    assert "required" in captured.err
