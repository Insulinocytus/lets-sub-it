FROM python:3.12-slim-bookworm

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg && \
    rm -rf /var/lib/apt/lists/* && \
    pip install --no-cache-dir uv

COPY whisper/ ./
RUN uv sync --no-dev --no-editable

ENV HF_HOME=/huggingface \
    LSI_WHISPER_WORK_DIR=/data/transcriptions

RUN mkdir -p /huggingface /data/transcriptions

EXPOSE 8081

CMD ["uv", "run", "uvicorn", "whisper_cli.server:app", "--host", "0.0.0.0", "--port", "8081"]
