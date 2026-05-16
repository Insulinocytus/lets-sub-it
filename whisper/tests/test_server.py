import asyncio
from pathlib import Path

import pytest
from fastapi.testclient import TestClient

from whisper_cli.server import TranscriptionService, create_app
from whisper_cli.transcribe import TranscriptionResult
from whisper_cli.vtt import Segment


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


class FailingUpload:
    def __init__(self) -> None:
        self.read_count = 0

    async def read(self, size: int = -1) -> bytes:
        if size == -1:
            raise AssertionError("upload must be read in bounded chunks")
        self.read_count += 1
        if self.read_count == 1:
            return b"partial"
        raise OSError("read failed")


def fake_transcribe(**kwargs) -> TranscriptionResult:
    return TranscriptionResult(
        language="en",
        duration_seconds=1.25,
        segments=[Segment(start=0.0, end=1.25, text="Hello world")],
    )


def failing_transcribe(**kwargs) -> TranscriptionResult:
    raise RuntimeError("transcriber failed")


def new_client(
    tmp_path: Path, *, transcribe=fake_transcribe
) -> tuple[TestClient, TranscriptionService]:
    service = TranscriptionService(work_dir=tmp_path, transcribe=transcribe)
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


def test_transcription_service_removes_task_dir_when_upload_fails(tmp_path):
    service = TranscriptionService(work_dir=tmp_path)

    with pytest.raises(OSError, match="read failed"):
        asyncio.run(
            service.create(
                audio=FailingUpload(),
                model="small",
                language="en",
                compute_type=None,
                job_id=None,
            )
        )

    assert list(tmp_path.glob("tr_*")) == []


def test_get_transcription_returns_queued_status(tmp_path):
    client, _ = new_client(tmp_path)
    create_response = client.post(
        "/transcriptions",
        data={"model": "small", "language": "en"},
        files={"audio": ("audio.mp3", b"audio", "audio/mpeg")},
    )
    task_id = create_response.json()["transcription"]["id"]

    response = client.get(f"/transcriptions/{task_id}")

    assert response.status_code == 200
    body = response.json()["transcription"]
    assert body["id"] == task_id
    assert body["status"] == "queued"
    assert body["progress"] == 0
    assert body["progressText"] == "等待转写"


def test_get_transcription_not_found(tmp_path):
    client, _ = new_client(tmp_path)

    response = client.get("/transcriptions/missing")

    assert response.status_code == 404
    assert response.json() == {
        "error": {"code": "not_found", "message": "transcription not found"}
    }


def test_vtt_before_completion_returns_conflict(tmp_path):
    client, _ = new_client(tmp_path)
    create_response = client.post(
        "/transcriptions",
        data={"model": "small", "language": "en"},
        files={"audio": ("audio.mp3", b"audio", "audio/mpeg")},
    )
    task_id = create_response.json()["transcription"]["id"]

    response = client.get(f"/transcriptions/{task_id}/vtt")

    assert response.status_code == 409
    assert response.json() == {
        "error": {
            "code": "not_ready",
            "message": "transcription is not completed",
        }
    }


def test_run_next_completes_task_and_vtt_endpoint_returns_webvtt(tmp_path):
    client, service = new_client(tmp_path)
    create_response = client.post(
        "/transcriptions",
        data={"model": "small", "language": "en"},
        files={"audio": ("audio.mp3", b"audio", "audio/mpeg")},
    )
    task_id = create_response.json()["transcription"]["id"]

    assert service.run_next() is True

    status_response = client.get(f"/transcriptions/{task_id}")
    assert status_response.status_code == 200
    body = status_response.json()["transcription"]
    assert body["status"] == "completed"
    assert body["progress"] == 100
    assert body["progressText"] == "转写完成"
    assert body["language"] == "en"
    assert body["durationSeconds"] == 1.25
    assert body["segments"] == [{"start": 0.0, "end": 1.25, "text": "Hello world"}]
    assert body["errorMessage"] is None

    vtt_response = client.get(f"/transcriptions/{task_id}/vtt")
    assert vtt_response.status_code == 200
    assert vtt_response.headers["content-type"] == "text/vtt; charset=utf-8"
    assert vtt_response.text == "WEBVTT\n\n00:00:00.000 --> 00:00:01.250\nHello world\n"


def test_run_next_marks_task_failed_when_transcriber_fails(tmp_path):
    client, service = new_client(tmp_path, transcribe=failing_transcribe)
    create_response = client.post(
        "/transcriptions",
        data={"model": "small", "language": "en"},
        files={"audio": ("audio.mp3", b"audio", "audio/mpeg")},
    )
    task_id = create_response.json()["transcription"]["id"]

    assert service.run_next() is True

    response = client.get(f"/transcriptions/{task_id}")
    assert response.status_code == 200
    body = response.json()["transcription"]
    assert body["status"] == "failed"
    assert body["progress"] == 100
    assert body["progressText"] == "转写失败"
    assert body["errorMessage"] == "transcriber failed"
