import asyncio
from pathlib import Path

from fastapi.testclient import TestClient

from whisper_cli.server import TranscriptionService, create_app


class ChunkedUpload:
    def __init__(self, data: bytes) -> None:
        self.data = data
        self.offset = 0

    async def read(self, size: int = -1) -> bytes:
        if size == -1:
            raise AssertionError("upload must be read in bounded chunks")
        chunk = self.data[self.offset : self.offset + size]
        self.offset += len(chunk)
        return chunk


def new_client(tmp_path: Path) -> tuple[TestClient, TranscriptionService]:
    service = TranscriptionService(work_dir=tmp_path)
    return TestClient(create_app(service=service, start_worker=False)), service


def test_healthz_returns_ok(tmp_path):
    client, _ = new_client(tmp_path)

    response = client.get("/healthz")

    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def test_create_transcription_requires_audio(tmp_path):
    client, _ = new_client(tmp_path)

    response = client.post(
        "/transcriptions",
        data={"model": "small", "language": "en"},
    )

    assert response.status_code == 400
    assert response.json() == {
        "error": {"code": "invalid_request", "message": "audio is required"}
    }


def test_create_transcription_requires_model(tmp_path):
    client, _ = new_client(tmp_path)

    response = client.post(
        "/transcriptions",
        data={"language": "en"},
        files={"audio": ("audio.mp3", b"audio", "audio/mpeg")},
    )

    assert response.status_code == 400
    assert response.json() == {
        "error": {"code": "invalid_request", "message": "model is required"}
    }


def test_create_transcription_requires_language(tmp_path):
    client, _ = new_client(tmp_path)

    response = client.post(
        "/transcriptions",
        data={"model": "small"},
        files={"audio": ("audio.mp3", b"audio", "audio/mpeg")},
    )

    assert response.status_code == 400
    assert response.json() == {
        "error": {"code": "invalid_request", "message": "language is required"}
    }


def test_create_transcription_stores_upload_and_returns_queued_task(tmp_path):
    client, service = new_client(tmp_path)

    response = client.post(
        "/transcriptions",
        data={"model": "small", "computeType": "int8", "language": "en", "jobId": "job_1"},
        files={"audio": ("audio.mp3", b"audio-data", "audio/mpeg")},
    )

    assert response.status_code == 202
    body = response.json()["transcription"]
    assert body["id"].startswith("tr_")
    assert body["status"] == "queued"
    assert body["progress"] == 0
    assert body["progressText"] == "等待转写"
    assert body["language"] is None
    assert body["durationSeconds"] is None
    assert body["segments"] is None
    assert body["errorMessage"] is None
    assert body["createdAt"].endswith("Z")
    assert body["updatedAt"].endswith("Z")

    task = service.get(body["id"])
    assert task.audio_path.read_bytes() == b"audio-data"
    assert task.model == "small"
    assert task.compute_type == "int8"
    assert task.requested_language == "en"
    assert task.job_id == "job_1"


def test_transcription_service_streams_upload_to_disk(tmp_path):
    service = TranscriptionService(work_dir=tmp_path)
    upload = ChunkedUpload(b"audio-data")

    task = asyncio.run(
        service.create(
            audio=upload,
            model="small",
            language="en",
            compute_type=None,
            job_id=None,
        )
    )

    assert task.audio_path.read_bytes() == b"audio-data"
