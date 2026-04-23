# whisper-cli

## Setup

`mise install && cd whisper && uv sync --dev`

## Run

`uv run whisper-cli --input /path/to/audio.mp3 --output /tmp/source.vtt --model small --language ja`

## Test

`uv run pytest`

## Output Contract

- `0`: success
- `2`: input validation
- `3`: transcription
- `4`: output validation
