# ---- Stage 1: Go build ----
FROM golang:1.22-bookworm AS go-builder

WORKDIR /build

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=1 go build -o /build/server ./cmd/server

# ---- Stage 2: Python whisper-cli ----
FROM python:3.12-slim-bookworm AS py-builder

WORKDIR /build

RUN pip install uv

COPY whisper/pyproject.toml whisper/uv.lock ./
RUN uv venv /opt/whisper-venv && \
    uv pip install --python /opt/whisper-venv/bin/python -r uv.lock --no-dev

COPY whisper/ ./
RUN uv pip install --python /opt/whisper-venv/bin/python --no-dev .

# ---- Stage 3: Runtime ----
FROM python:3.12-slim-bookworm

RUN apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg && \
    rm -rf /var/lib/apt/lists/* && \
    pip install --no-cache-dir yt-dlp

COPY --from=go-builder /build/server /usr/local/bin/server
COPY --from=py-builder /opt/whisper-venv /opt/whisper-venv

ENV PATH="/opt/whisper-venv/bin:$PATH" \
    LSI_ADDR=0.0.0.0:8080 \
    LSI_DB_PATH=/data/backend.sqlite3 \
    LSI_WORK_DIR=/data/jobs \
    LSI_RUNNER_MODE=real

RUN mkdir -p /data

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/server"]