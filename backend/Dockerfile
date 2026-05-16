# ---- Stage 1: Go build ----
FROM golang:1.22-bookworm AS go-builder

WORKDIR /build

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=1 go build -o /build/server ./cmd/server

# ---- Stage 2: Runtime ----
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates ffmpeg yt-dlp && \
    rm -rf /var/lib/apt/lists/*

COPY --from=go-builder /build/server /usr/local/bin/server

ENV LSI_ADDR=0.0.0.0:8080 \
    LSI_DB_PATH=/data/backend.sqlite3 \
    LSI_WORK_DIR=/data/jobs

RUN mkdir -p /data/jobs

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/server"]
