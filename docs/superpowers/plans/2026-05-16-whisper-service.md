# Whisper Service Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the backend's `whisper-cli` process call with a separate HTTP Whisper transcription service that accepts uploaded audio, runs asynchronous transcription, and lets the backend poll for VTT output.

**Architecture:** Keep backend as the owner of YouTube downloads, SQLite job state, translation, packaging, and subtitle serving. Add a Python FastAPI Whisper service under `whisper/` that owns uploaded audio, transcription task state, and completed VTT output. Add a Go `Transcriber` boundary so `RealRunner` depends on an interface instead of command execution or raw HTTP details.

**Tech Stack:** Go 1.22, Python 3.12, FastAPI, Uvicorn, faster-whisper, pytest, Go `httptest`, Docker Compose, Taskfile.

---

## Scope Check

The approved spec covers one vertical migration: split Whisper transcription out of the backend while preserving the existing backend-facing job API. The work touches multiple modules, but the changes are tightly coupled and should ship together because backend, whisper service, Docker Compose, and developer commands must agree on one communication contract.

## File Structure

Create or modify these files:

| Path | Action | Responsibility |
|---|---|---|
| `whisper/pyproject.toml` | Modify | Add FastAPI service runtime dependencies and remove `whisper-cli` console script. |
| `whisper/src/whisper_cli/server.py` | Create | FastAPI app, error format, in-memory transcription queue, status endpoints, VTT endpoint, service factory for tests. |
| `whisper/src/whisper_cli/cli.py` | Delete | Remove deprecated CLI entrypoint. |
| `whisper/tests/test_server.py` | Create | Contract tests for Whisper HTTP API and task state transitions. |
| `whisper/tests/test_cli.py` | Delete | Remove CLI contract tests because CLI is no longer supported. |
| `backend/internal/app/config.go` | Modify | Add Whisper service URL, timeout, and poll interval config. |
| `backend/internal/app/config_test.go` | Modify | Cover new Whisper config defaults and custom values. |
| `backend/internal/app/app.go` | Modify | Wire HTTP transcriber and stop requiring `whisper-cli` on `PATH`. |
| `backend/internal/app/app_test.go` | Modify | Verify `checkTools()` no longer requires `whisper-cli`. |
| `backend/internal/runner/transcriber.go` | Replace | Define `Transcriber`, `TranscriptionRequest`, and shared output validation. |
| `backend/internal/runner/http_transcriber.go` | Create | HTTP client that uploads audio, polls status, downloads VTT, and writes `source.vtt`. |
| `backend/internal/runner/http_transcriber_test.go` | Create | Unit tests for upload, polling, VTT write, failed task, invalid response, and empty VTT. |
| `backend/internal/runner/real_runner.go` | Modify | Use `Transcriber` interface during `transcribing` stage. |
| `backend/internal/runner/real_runner_test.go` | Modify | Replace `whisper-cli` exec assertions with fake transcriber assertions. |
| `backend/Dockerfile` | Modify | Remove Python Whisper build stage from backend image. |
| `whisper/Dockerfile` | Create | Build and run the Whisper FastAPI service. |
| `docker-compose.yml` | Modify | Add `whisper` service, internal URL, healthcheck, and Whisper volumes. |
| `.env.example` | Modify | Add Whisper service URL, timeout, and poll interval variables. |
| `Taskfile.yml` | Modify | Add `dev:whisper`; remove backend dependency on `whisper-cli` PATH. |

## Task 1: Add Whisper Service Skeleton

**Files:**
- Modify: `whisper/pyproject.toml`
- Create: `whisper/src/whisper_cli/server.py`
- Create: `whisper/tests/test_server.py`

- [ ] **Step 1: Write the failing healthcheck test**

Create `whisper/tests/test_server.py` with this content:

```python
from fastapi.testclient import TestClient

from whisper_cli.server import create_app


def test_healthz_returns_ok():
    client = TestClient(create_app(start_worker=False))

    response = client.get("/healthz")

    assert response.status_code == 200
    assert response.json() == {"status": "ok"}
```

- [ ] **Step 2: Run the focused test and verify it fails**

Run from `whisper/`:

```bash
mise exec -- uv run pytest tests/test_server.py::test_healthz_returns_ok -q
```

Expected: FAIL because `whisper_cli.server` does not exist.

- [ ] **Step 3: Add FastAPI dependencies**

Run from `whisper/`:

```bash
mise exec -- uv add 'fastapi>=0.128.0' 'uvicorn[standard]>=0.34.0' 'python-multipart>=0.0.20'
mise exec -- uv add --dev 'httpx>=0.28.0'
```

Expected: `pyproject.toml` contains these dependency entries and `uv.lock` is updated.

- [ ] **Step 4: Implement the minimal service app**

Create `whisper/src/whisper_cli/server.py` with this content:

```python
from __future__ import annotations

from fastapi import FastAPI


def create_app(*, start_worker: bool = True) -> FastAPI:
    app = FastAPI(title="Lets Sub It Whisper Service")

    @app.get("/healthz")
    def healthz() -> dict[str, str]:
        return {"status": "ok"}

    return app


app = create_app()
```

- [ ] **Step 5: Run the focused test and verify it passes**

Run from `whisper/`:

```bash
mise exec -- uv run pytest tests/test_server.py::test_healthz_returns_ok -q
```

Expected: PASS.

- [ ] **Step 6: Commit**

Run from repository root:

```bash
git add whisper/pyproject.toml whisper/uv.lock whisper/src/whisper_cli/server.py whisper/tests/test_server.py
git commit -m "feat(whisper): add service health endpoint"
```

## Task 2: Add Transcription Creation and Validation

**Files:**
- Modify: `whisper/src/whisper_cli/server.py`
- Modify: `whisper/tests/test_server.py`

- [ ] **Step 1: Replace server tests with create-task contract tests**

Replace `whisper/tests/test_server.py` with this content:

```python
from pathlib import Path

from fastapi.testclient import TestClient

from whisper_cli.server import TranscriptionService, create_app


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
```

- [ ] **Step 2: Run the focused tests and verify they fail**

Run from `whisper/`:

```bash
mise exec -- uv run pytest tests/test_server.py -q
```

Expected: FAIL because `TranscriptionService` and `/transcriptions` do not exist.

- [ ] **Step 3: Implement task creation, upload storage, and uniform errors**

Replace `whisper/src/whisper_cli/server.py` with this content:

```python
from __future__ import annotations

import os
import uuid
from dataclasses import dataclass
from datetime import UTC, datetime
from pathlib import Path
from threading import Lock
from typing import Annotated

from fastapi import FastAPI, File, Form, Request, UploadFile
from fastapi.responses import JSONResponse


def utc_now() -> datetime:
    return datetime.now(UTC)


def format_time(value: datetime) -> str:
    return value.isoformat().replace("+00:00", "Z")


class APIError(Exception):
    def __init__(self, status_code: int, code: str, message: str) -> None:
        self.status_code = status_code
        self.code = code
        self.message = message


@dataclass
class TranscriptionTask:
    id: str
    status: str
    progress: int
    progress_text: str
    model: str
    compute_type: str
    requested_language: str
    job_id: str | None
    audio_path: Path
    vtt_path: Path
    language: str | None
    duration_seconds: float | None
    segments: int | None
    error_message: str | None
    created_at: datetime
    updated_at: datetime


class TranscriptionService:
    def __init__(self, *, work_dir: Path | str | None = None) -> None:
        self.work_dir = Path(work_dir or os.getenv("LSI_WHISPER_WORK_DIR", "/data/transcriptions"))
        self._lock = Lock()
        self._tasks: dict[str, TranscriptionTask] = {}

    def create(
        self,
        *,
        audio: UploadFile,
        model: str,
        compute_type: str,
        language: str,
        job_id: str | None,
    ) -> TranscriptionTask:
        task_id = "tr_" + uuid.uuid4().hex
        task_dir = self.work_dir / task_id
        task_dir.mkdir(parents=True, exist_ok=False)
        audio_path = task_dir / "audio.mp3"
        vtt_path = task_dir / "source.vtt"

        with audio_path.open("wb") as output:
            while chunk := audio.file.read(1024 * 1024):
                output.write(chunk)

        now = utc_now()
        task = TranscriptionTask(
            id=task_id,
            status="queued",
            progress=0,
            progress_text="等待转写",
            model=model,
            compute_type=compute_type,
            requested_language=language,
            job_id=job_id,
            audio_path=audio_path,
            vtt_path=vtt_path,
            language=None,
            duration_seconds=None,
            segments=None,
            error_message=None,
            created_at=now,
            updated_at=now,
        )
        with self._lock:
            self._tasks[task_id] = task
        return task

    def get(self, task_id: str) -> TranscriptionTask:
        with self._lock:
            task = self._tasks.get(task_id)
        if task is None:
            raise APIError(404, "not_found", "transcription not found")
        return task


def task_response(task: TranscriptionTask) -> dict[str, object]:
    return {
        "id": task.id,
        "status": task.status,
        "progress": task.progress,
        "progressText": task.progress_text,
        "language": task.language,
        "durationSeconds": task.duration_seconds,
        "segments": task.segments,
        "errorMessage": task.error_message,
        "createdAt": format_time(task.created_at),
        "updatedAt": format_time(task.updated_at),
    }


def create_app(
    *,
    service: TranscriptionService | None = None,
    start_worker: bool = True,
) -> FastAPI:
    transcription_service = service or TranscriptionService()
    app = FastAPI(title="Lets Sub It Whisper Service")
    app.state.transcription_service = transcription_service

    @app.exception_handler(APIError)
    def handle_api_error(request: Request, exc: APIError) -> JSONResponse:
        return JSONResponse(
            status_code=exc.status_code,
            content={"error": {"code": exc.code, "message": exc.message}},
        )

    @app.get("/healthz")
    def healthz() -> dict[str, str]:
        return {"status": "ok"}

    @app.post("/transcriptions", status_code=202)
    def create_transcription(
        audio: Annotated[UploadFile | None, File()] = None,
        model: Annotated[str | None, Form()] = None,
        language: Annotated[str | None, Form()] = None,
        computeType: Annotated[str, Form()] = "default",
        jobId: Annotated[str | None, Form()] = None,
    ) -> dict[str, dict[str, object]]:
        if audio is None:
            raise APIError(400, "invalid_request", "audio is required")
        if model is None or model.strip() == "":
            raise APIError(400, "invalid_request", "model is required")
        if language is None or language.strip() == "":
            raise APIError(400, "invalid_request", "language is required")

        task = transcription_service.create(
            audio=audio,
            model=model.strip(),
            compute_type=computeType.strip() or "default",
            language=language.strip(),
            job_id=jobId,
        )
        return {"transcription": task_response(task)}

    return app


app = create_app()
```

- [ ] **Step 4: Run the service tests and verify they pass**

Run from `whisper/`:

```bash
mise exec -- uv run pytest tests/test_server.py -q
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run from repository root:

```bash
git add whisper/src/whisper_cli/server.py whisper/tests/test_server.py
git commit -m "feat(whisper): accept transcription uploads"
```

## Task 3: Add Worker Execution, Status, and VTT Endpoints

**Files:**
- Modify: `whisper/src/whisper_cli/server.py`
- Modify: `whisper/tests/test_server.py`

- [ ] **Step 1: Extend API tests for status, VTT, completion, and failure**

Append these tests to `whisper/tests/test_server.py`:

```python
from whisper_cli.transcribe import TranscriptionResult
from whisper_cli.vtt import Segment


def complete_result() -> TranscriptionResult:
    return TranscriptionResult(
        language="en",
        duration_seconds=2.5,
        segments=[
            Segment(start=0.0, end=1.25, text="hello"),
            Segment(start=1.25, end=2.5, text="world"),
        ],
    )


def test_get_transcription_returns_queued_status(tmp_path):
    client, _ = new_client(tmp_path)
    created = client.post(
        "/transcriptions",
        data={"model": "small", "language": "en"},
        files={"audio": ("audio.mp3", b"audio", "audio/mpeg")},
    ).json()["transcription"]

    response = client.get(f"/transcriptions/{created['id']}")

    assert response.status_code == 200
    assert response.json()["transcription"]["status"] == "queued"


def test_get_transcription_not_found(tmp_path):
    client, _ = new_client(tmp_path)

    response = client.get("/transcriptions/tr_missing")

    assert response.status_code == 404
    assert response.json() == {
        "error": {"code": "not_found", "message": "transcription not found"}
    }


def test_vtt_before_completion_returns_conflict(tmp_path):
    client, _ = new_client(tmp_path)
    created = client.post(
        "/transcriptions",
        data={"model": "small", "language": "en"},
        files={"audio": ("audio.mp3", b"audio", "audio/mpeg")},
    ).json()["transcription"]

    response = client.get(f"/transcriptions/{created['id']}/vtt")

    assert response.status_code == 409
    assert response.json() == {
        "error": {"code": "not_ready", "message": "transcription is not completed"}
    }


def test_run_next_completes_task_and_vtt_endpoint_returns_webvtt(tmp_path):
    calls = []

    def fake_transcribe(**kwargs):
        calls.append(kwargs)
        return complete_result()

    service = TranscriptionService(work_dir=tmp_path, transcribe=fake_transcribe)
    client = TestClient(create_app(service=service, start_worker=False))
    created = client.post(
        "/transcriptions",
        data={"model": "small", "computeType": "int8", "language": "en"},
        files={"audio": ("audio.mp3", b"audio", "audio/mpeg")},
    ).json()["transcription"]

    assert service.run_next() is True

    status = client.get(f"/transcriptions/{created['id']}")
    assert status.status_code == 200
    body = status.json()["transcription"]
    assert body["status"] == "completed"
    assert body["progress"] == 100
    assert body["progressText"] == "转写完成"
    assert body["language"] == "en"
    assert body["durationSeconds"] == 2.5
    assert body["segments"] == 2

    vtt = client.get(f"/transcriptions/{created['id']}/vtt")
    assert vtt.status_code == 200
    assert vtt.headers["content-type"].startswith("text/vtt")
    assert vtt.text.startswith("WEBVTT")
    assert "hello" in vtt.text
    assert calls[0]["model_name"] == "small"
    assert calls[0]["compute_type"] == "int8"
    assert calls[0]["language"] == "en"


def test_run_next_marks_task_failed_when_transcriber_fails(tmp_path):
    def fake_transcribe(**kwargs):
        raise RuntimeError("model download error")

    service = TranscriptionService(work_dir=tmp_path, transcribe=fake_transcribe)
    client = TestClient(create_app(service=service, start_worker=False))
    created = client.post(
        "/transcriptions",
        data={"model": "small", "language": "en"},
        files={"audio": ("audio.mp3", b"audio", "audio/mpeg")},
    ).json()["transcription"]

    assert service.run_next() is True

    status = client.get(f"/transcriptions/{created['id']}")
    body = status.json()["transcription"]
    assert body["status"] == "failed"
    assert body["progress"] == 100
    assert body["progressText"] == "转写失败"
    assert body["errorMessage"] == "model download error"

    vtt = client.get(f"/transcriptions/{created['id']}/vtt")
    assert vtt.status_code == 409
```

- [ ] **Step 2: Run the focused tests and verify they fail**

Run from `whisper/`:

```bash
mise exec -- uv run pytest tests/test_server.py -q
```

Expected: FAIL because `TranscriptionService` has no `transcribe` injection, no queue execution, no status route, and no VTT route.

- [ ] **Step 3: Replace server implementation with queue and endpoint support**

Replace `whisper/src/whisper_cli/server.py` with this content:

```python
from __future__ import annotations

import os
import queue
import threading
import uuid
from collections.abc import Callable
from dataclasses import dataclass
from datetime import UTC, datetime
from pathlib import Path
from threading import Lock
from typing import Annotated

from fastapi import FastAPI, File, Form, Request, UploadFile
from fastapi.responses import JSONResponse, Response

from whisper_cli.transcribe import TranscriptionResult, transcribe_audio
from whisper_cli.vtt import render_vtt


TranscribeFunc = Callable[..., TranscriptionResult]


def utc_now() -> datetime:
    return datetime.now(UTC)


def format_time(value: datetime) -> str:
    return value.isoformat().replace("+00:00", "Z")


class APIError(Exception):
    def __init__(self, status_code: int, code: str, message: str) -> None:
        self.status_code = status_code
        self.code = code
        self.message = message


@dataclass
class TranscriptionTask:
    id: str
    status: str
    progress: int
    progress_text: str
    model: str
    compute_type: str
    requested_language: str
    job_id: str | None
    audio_path: Path
    vtt_path: Path
    language: str | None
    duration_seconds: float | None
    segments: int | None
    error_message: str | None
    created_at: datetime
    updated_at: datetime


class TranscriptionService:
    def __init__(
        self,
        *,
        work_dir: Path | str | None = None,
        transcribe: TranscribeFunc = transcribe_audio,
    ) -> None:
        self.work_dir = Path(work_dir or os.getenv("LSI_WHISPER_WORK_DIR", "/data/transcriptions"))
        self.transcribe = transcribe
        self._lock = Lock()
        self._tasks: dict[str, TranscriptionTask] = {}
        self._queue: queue.Queue[str] = queue.Queue()
        self._stop = threading.Event()
        self._worker: threading.Thread | None = None

    def start(self) -> None:
        if self._worker is not None:
            return
        self._worker = threading.Thread(target=self._worker_loop, name="whisper-worker", daemon=True)
        self._worker.start()

    def stop(self) -> None:
        self._stop.set()
        if self._worker is not None:
            self._worker.join(timeout=5)

    def _worker_loop(self) -> None:
        while not self._stop.is_set():
            try:
                self.run_next(timeout=0.2)
            except Exception:
                continue

    def create(
        self,
        *,
        audio: UploadFile,
        model: str,
        compute_type: str,
        language: str,
        job_id: str | None,
    ) -> TranscriptionTask:
        task_id = "tr_" + uuid.uuid4().hex
        task_dir = self.work_dir / task_id
        task_dir.mkdir(parents=True, exist_ok=False)
        audio_path = task_dir / "audio.mp3"
        vtt_path = task_dir / "source.vtt"

        with audio_path.open("wb") as output:
            while chunk := audio.file.read(1024 * 1024):
                output.write(chunk)

        now = utc_now()
        task = TranscriptionTask(
            id=task_id,
            status="queued",
            progress=0,
            progress_text="等待转写",
            model=model,
            compute_type=compute_type,
            requested_language=language,
            job_id=job_id,
            audio_path=audio_path,
            vtt_path=vtt_path,
            language=None,
            duration_seconds=None,
            segments=None,
            error_message=None,
            created_at=now,
            updated_at=now,
        )
        with self._lock:
            self._tasks[task_id] = task
        self._queue.put(task_id)
        return task

    def get(self, task_id: str) -> TranscriptionTask:
        with self._lock:
            task = self._tasks.get(task_id)
        if task is None:
            raise APIError(404, "not_found", "transcription not found")
        return task

    def run_next(self, timeout: float | None = None) -> bool:
        try:
            task_id = self._queue.get(timeout=timeout)
        except queue.Empty:
            return False
        try:
            task = self.get(task_id)
            self._mark_running(task)
            result = self.transcribe(
                input_path=task.audio_path,
                model_name=task.model,
                compute_type=task.compute_type,
                language=task.requested_language,
            )
            content = render_vtt(result.segments)
            task.vtt_path.write_text(content, encoding="utf-8")
            self._mark_completed(task, result)
            return True
        except Exception as exc:
            task = self.get(task_id)
            self._mark_failed(task, str(exc))
            return True
        finally:
            self._queue.task_done()

    def _mark_running(self, task: TranscriptionTask) -> None:
        with self._lock:
            task.status = "running"
            task.progress = 10
            task.progress_text = "正在转写音频"
            task.updated_at = utc_now()

    def _mark_completed(self, task: TranscriptionTask, result: TranscriptionResult) -> None:
        with self._lock:
            task.status = "completed"
            task.progress = 100
            task.progress_text = "转写完成"
            task.language = result.language
            task.duration_seconds = result.duration_seconds
            task.segments = len(result.segments)
            task.error_message = None
            task.updated_at = utc_now()

    def _mark_failed(self, task: TranscriptionTask, message: str) -> None:
        with self._lock:
            task.status = "failed"
            task.progress = 100
            task.progress_text = "转写失败"
            task.error_message = message
            task.updated_at = utc_now()


def task_response(task: TranscriptionTask) -> dict[str, object]:
    return {
        "id": task.id,
        "status": task.status,
        "progress": task.progress,
        "progressText": task.progress_text,
        "language": task.language,
        "durationSeconds": task.duration_seconds,
        "segments": task.segments,
        "errorMessage": task.error_message,
        "createdAt": format_time(task.created_at),
        "updatedAt": format_time(task.updated_at),
    }


def create_app(
    *,
    service: TranscriptionService | None = None,
    start_worker: bool = True,
) -> FastAPI:
    transcription_service = service or TranscriptionService()
    app = FastAPI(title="Lets Sub It Whisper Service")
    app.state.transcription_service = transcription_service

    if start_worker:
        transcription_service.start()

    @app.exception_handler(APIError)
    def handle_api_error(request: Request, exc: APIError) -> JSONResponse:
        return JSONResponse(
            status_code=exc.status_code,
            content={"error": {"code": exc.code, "message": exc.message}},
        )

    @app.get("/healthz")
    def healthz() -> dict[str, str]:
        return {"status": "ok"}

    @app.post("/transcriptions", status_code=202)
    def create_transcription(
        audio: Annotated[UploadFile | None, File()] = None,
        model: Annotated[str | None, Form()] = None,
        language: Annotated[str | None, Form()] = None,
        computeType: Annotated[str, Form()] = "default",
        jobId: Annotated[str | None, Form()] = None,
    ) -> dict[str, dict[str, object]]:
        if audio is None:
            raise APIError(400, "invalid_request", "audio is required")
        if model is None or model.strip() == "":
            raise APIError(400, "invalid_request", "model is required")
        if language is None or language.strip() == "":
            raise APIError(400, "invalid_request", "language is required")

        task = transcription_service.create(
            audio=audio,
            model=model.strip(),
            compute_type=computeType.strip() or "default",
            language=language.strip(),
            job_id=jobId,
        )
        return {"transcription": task_response(task)}

    @app.get("/transcriptions/{task_id}")
    def get_transcription(task_id: str) -> dict[str, dict[str, object]]:
        return {"transcription": task_response(transcription_service.get(task_id))}

    @app.get("/transcriptions/{task_id}/vtt")
    def get_transcription_vtt(task_id: str) -> Response:
        task = transcription_service.get(task_id)
        if task.status != "completed":
            raise APIError(409, "not_ready", "transcription is not completed")
        return Response(task.vtt_path.read_text(encoding="utf-8"), media_type="text/vtt; charset=utf-8")

    return app


app = create_app()
```

- [ ] **Step 4: Run the service tests and verify they pass**

Run from `whisper/`:

```bash
mise exec -- uv run pytest tests/test_server.py -q
```

Expected: PASS.

- [ ] **Step 5: Run all Whisper tests**

Run from repository root:

```bash
task test:whisper
```

Expected: all Whisper tests pass.

- [ ] **Step 6: Commit**

Run from repository root:

```bash
git add whisper/src/whisper_cli/server.py whisper/tests/test_server.py
git commit -m "feat(whisper): add transcription task endpoints"
```

## Task 4: Remove Deprecated Whisper CLI

**Files:**
- Modify: `whisper/pyproject.toml`
- Delete: `whisper/src/whisper_cli/cli.py`
- Delete: `whisper/tests/test_cli.py`

- [ ] **Step 1: Remove the CLI script entrypoint from pyproject**

Apply this patch:

```diff
*** Begin Patch
*** Update File: whisper/pyproject.toml
@@
-[project.scripts]
-whisper-cli = "whisper_cli.cli:main_entry"
-
 [tool.pytest.ini_options]
 pythonpath = ["src"]
 addopts = "-q"
*** End Patch
```

- [ ] **Step 2: Delete CLI files**

Use `apply_patch` with this patch:

```diff
*** Begin Patch
*** Delete File: whisper/src/whisper_cli/cli.py
*** Delete File: whisper/tests/test_cli.py
*** End Patch
```

- [ ] **Step 3: Verify the package no longer exposes `whisper-cli` tests or script**

Run from repository root:

```bash
task test:whisper
```

Expected: all remaining Whisper tests pass.

- [ ] **Step 4: Commit**

Run from repository root:

```bash
git add whisper/pyproject.toml whisper/src/whisper_cli/cli.py whisper/tests/test_cli.py
git commit -m "refactor(whisper): remove deprecated cli entrypoint"
```

## Task 5: Add Backend Whisper Service Configuration and Tool Checks

**Files:**
- Modify: `backend/internal/app/config.go`
- Modify: `backend/internal/app/config_test.go`
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/app/app_test.go`

- [ ] **Step 1: Add failing config tests**

Append these tests to `backend/internal/app/config_test.go`:

```go
func TestLoadConfigWhisperServiceDefaults(t *testing.T) {
	t.Setenv("LSI_WHISPER_BASE_URL", "")
	t.Setenv("LSI_WHISPER_TIMEOUT", "")
	t.Setenv("LSI_WHISPER_POLL_INTERVAL", "")

	config := LoadConfig()

	if config.WhisperBaseURL != "http://127.0.0.1:8081" {
		t.Fatalf("WhisperBaseURL = %q", config.WhisperBaseURL)
	}
	if config.WhisperTimeout != 30*time.Minute {
		t.Fatalf("WhisperTimeout = %v, want %v", config.WhisperTimeout, 30*time.Minute)
	}
	if config.WhisperPollInterval != 2*time.Second {
		t.Fatalf("WhisperPollInterval = %v, want %v", config.WhisperPollInterval, 2*time.Second)
	}
}

func TestLoadConfigWhisperServiceCustomValues(t *testing.T) {
	t.Setenv("LSI_WHISPER_BASE_URL", "http://whisper:8081")
	t.Setenv("LSI_WHISPER_TIMEOUT", "45m")
	t.Setenv("LSI_WHISPER_POLL_INTERVAL", "500ms")

	config := LoadConfig()

	if config.WhisperBaseURL != "http://whisper:8081" {
		t.Fatalf("WhisperBaseURL = %q", config.WhisperBaseURL)
	}
	if config.WhisperTimeout != 45*time.Minute {
		t.Fatalf("WhisperTimeout = %v", config.WhisperTimeout)
	}
	if config.WhisperPollInterval != 500*time.Millisecond {
		t.Fatalf("WhisperPollInterval = %v", config.WhisperPollInterval)
	}
}
```

- [ ] **Step 2: Replace the tool-check test**

Replace `TestCheckToolsRequiresWhisperCLI` in `backend/internal/app/app_test.go` with this test:

```go
func TestCheckToolsDoesNotRequireWhisperCLI(t *testing.T) {
	origLookPath := lookPath
	t.Cleanup(func() { lookPath = origLookPath })

	var checked []string
	lookPath = func(tool string) (string, error) {
		checked = append(checked, tool)
		return "/usr/bin/" + tool, nil
	}

	if err := checkTools(); err != nil {
		t.Fatalf("checkTools() error = %v", err)
	}
	for _, tool := range checked {
		if tool == "whisper-cli" {
			t.Fatalf("checkTools() checked whisper-cli; checked = %#v", checked)
		}
	}
}
```

- [ ] **Step 3: Run app tests and verify they fail**

Run from `backend/`:

```bash
mise exec -- go test ./internal/app -run 'TestLoadConfigWhisperService|TestCheckToolsDoesNotRequireWhisperCLI'
```

Expected: FAIL because the config fields do not exist and `checkTools()` still checks `whisper-cli`.

- [ ] **Step 4: Add config fields and remove `whisper-cli` tool check**

Apply this patch:

```diff
*** Begin Patch
*** Update File: backend/internal/app/config.go
@@
 	WhisperModel       string
 	WhisperComputeType string
+	WhisperBaseURL     string
+	WhisperTimeout     time.Duration
+	WhisperPollInterval time.Duration
 	LLMBaseURL         string
@@
 		WhisperModel:       envOrDefault("LSI_WHISPER_MODEL", "small"),
 		WhisperComputeType: envOrDefault("LSI_WHISPER_COMPUTE_TYPE", "default"),
+		WhisperBaseURL:     envOrDefault("LSI_WHISPER_BASE_URL", "http://127.0.0.1:8081"),
+		WhisperTimeout:     envDurationOrDefault("LSI_WHISPER_TIMEOUT", 30*time.Minute),
+		WhisperPollInterval: envDurationOrDefault("LSI_WHISPER_POLL_INTERVAL", 2*time.Second),
 		LLMBaseURL:         envOrDefault("LSI_LLM_BASE_URL", "https://api.openai.com"),
*** End Patch
```

Then run `gofmt` after the patch because the aligned struct fields will need formatting:

```bash
gofmt -w backend/internal/app/config.go
```

Apply this patch:

```diff
*** Begin Patch
*** Update File: backend/internal/app/app.go
@@
-	for _, tool := range []string{"yt-dlp", "ffmpeg", "whisper-cli"} {
+	for _, tool := range []string{"yt-dlp", "ffmpeg"} {
*** End Patch
```

- [ ] **Step 5: Run app tests and verify they pass**

Run from `backend/`:

```bash
mise exec -- go test ./internal/app
```

Expected: PASS.

- [ ] **Step 6: Commit**

Run from repository root:

```bash
git add backend/internal/app/config.go backend/internal/app/config_test.go backend/internal/app/app.go backend/internal/app/app_test.go
git commit -m "feat(backend): configure whisper service client"
```

## Task 6: Implement Backend HTTP Transcriber Client

**Files:**
- Replace: `backend/internal/runner/transcriber.go`
- Create: `backend/internal/runner/http_transcriber.go`
- Create: `backend/internal/runner/http_transcriber_test.go`

- [ ] **Step 1: Add failing HTTP transcriber tests**

Create `backend/internal/runner/http_transcriber_test.go` with this content:

```go
package runner

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHTTPTranscriberUploadsAudioPollsAndWritesVTT(t *testing.T) {
	var sawUpload bool
	var pollCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/transcriptions":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s", r.Method)
			}
			if err := r.ParseMultipartForm(8 << 20); err != nil {
				t.Fatalf("ParseMultipartForm() error = %v", err)
			}
			if got := r.FormValue("model"); got != "tiny" {
				t.Fatalf("model = %q", got)
			}
			if got := r.FormValue("computeType"); got != "int8" {
				t.Fatalf("computeType = %q", got)
			}
			if got := r.FormValue("language"); got != "ja" {
				t.Fatalf("language = %q", got)
			}
			if got := r.FormValue("jobId"); got != "job_1" {
				t.Fatalf("jobId = %q", got)
			}
			file, _, err := r.FormFile("audio")
			if err != nil {
				t.Fatalf("FormFile(audio) error = %v", err)
			}
			defer file.Close()
			data, err := io.ReadAll(file)
			if err != nil {
				t.Fatalf("ReadAll(audio) error = %v", err)
			}
			if string(data) != "fake-audio" {
				t.Fatalf("audio = %q", string(data))
			}
			sawUpload = true
			writeWhisperJSON(w, http.StatusAccepted, map[string]any{"transcription": map[string]any{"id": "tr_1", "status": "queued", "progressText": "等待转写"}})
		case "/transcriptions/tr_1":
			pollCount++
			writeWhisperJSON(w, http.StatusOK, map[string]any{"transcription": map[string]any{"id": "tr_1", "status": "completed", "progressText": "转写完成"}})
		case "/transcriptions/tr_1/vtt":
			w.Header().Set("Content-Type", "text/vtt; charset=utf-8")
			_, _ = w.Write([]byte("WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nhello\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	workDir := t.TempDir()
	audioPath := filepath.Join(workDir, "audio.mp3")
	if err := os.WriteFile(audioPath, []byte("fake-audio"), 0o644); err != nil {
		t.Fatalf("WriteFile(audio) error = %v", err)
	}
	sourcePath := filepath.Join(workDir, "source.vtt")
	var progress []string
	transcriber := NewHTTPTranscriber(server.URL, time.Second, time.Millisecond, server.Client())

	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{
		JobID:       "job_1",
		AudioPath:   audioPath,
		SourcePath:  sourcePath,
		Model:       "tiny",
		ComputeType: "int8",
		Language:    "ja",
		OnProgress: func(text string) error {
			progress = append(progress, text)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}
	if !sawUpload {
		t.Fatal("upload was not received")
	}
	if pollCount != 1 {
		t.Fatalf("pollCount = %d, want 1", pollCount)
	}
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("ReadFile(source.vtt) error = %v", err)
	}
	if !strings.Contains(string(content), "hello") {
		t.Fatalf("source.vtt = %q", string(content))
	}
	if len(progress) == 0 || progress[len(progress)-1] != "转写完成" {
		t.Fatalf("progress = %#v", progress)
	}
}

func TestHTTPTranscriberFailsWhenWhisperReportsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/transcriptions":
			writeWhisperJSON(w, http.StatusAccepted, map[string]any{"transcription": map[string]any{"id": "tr_1", "status": "queued", "progressText": "等待转写"}})
		case "/transcriptions/tr_1":
			writeWhisperJSON(w, http.StatusOK, map[string]any{"transcription": map[string]any{"id": "tr_1", "status": "failed", "progressText": "转写失败", "errorMessage": "model download error"}})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	audioPath := filepath.Join(t.TempDir(), "audio.mp3")
	if err := os.WriteFile(audioPath, []byte("fake-audio"), 0o644); err != nil {
		t.Fatalf("WriteFile(audio) error = %v", err)
	}
	transcriber := NewHTTPTranscriber(server.URL, time.Second, time.Millisecond, server.Client())

	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{JobID: "job_1", AudioPath: audioPath, SourcePath: filepath.Join(t.TempDir(), "source.vtt"), Model: "tiny", ComputeType: "int8", Language: "ja"})
	if err == nil || !strings.Contains(err.Error(), "model download error") {
		t.Fatalf("Transcribe() error = %v, want model download error", err)
	}
}

func TestHTTPTranscriberRejectsEmptyVTT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/transcriptions":
			writeWhisperJSON(w, http.StatusAccepted, map[string]any{"transcription": map[string]any{"id": "tr_1", "status": "queued", "progressText": "等待转写"}})
		case "/transcriptions/tr_1":
			writeWhisperJSON(w, http.StatusOK, map[string]any{"transcription": map[string]any{"id": "tr_1", "status": "completed", "progressText": "转写完成"}})
		case "/transcriptions/tr_1/vtt":
			w.Header().Set("Content-Type", "text/vtt; charset=utf-8")
			_, _ = w.Write([]byte(""))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	audioPath := filepath.Join(t.TempDir(), "audio.mp3")
	if err := os.WriteFile(audioPath, []byte("fake-audio"), 0o644); err != nil {
		t.Fatalf("WriteFile(audio) error = %v", err)
	}
	transcriber := NewHTTPTranscriber(server.URL, time.Second, time.Millisecond, server.Client())

	err := transcriber.Transcribe(context.Background(), TranscriptionRequest{JobID: "job_1", AudioPath: audioPath, SourcePath: filepath.Join(t.TempDir(), "source.vtt"), Model: "tiny", ComputeType: "int8", Language: "ja"})
	if err == nil || !strings.Contains(err.Error(), "empty source.vtt") {
		t.Fatalf("Transcribe() error = %v, want empty source.vtt", err)
	}
}

func writeWhisperJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run from `backend/`:

```bash
mise exec -- go test ./internal/runner -run TestHTTPTranscriber
```

Expected: FAIL because `NewHTTPTranscriber` and `TranscriptionRequest` do not exist.

- [ ] **Step 3: Replace the old CLI transcriber contract**

Replace `backend/internal/runner/transcriber.go` with this content:

```go
package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Transcriber interface {
	Transcribe(ctx context.Context, request TranscriptionRequest) error
}

type TranscriptionRequest struct {
	JobID       string
	AudioPath   string
	SourcePath  string
	Model       string
	ComputeType string
	Language    string
	OnProgress  func(text string) error
}

func ensureValidSourceVTT(sourcePath string) error {
	info, statErr := os.Stat(sourcePath)
	if statErr != nil {
		return fmt.Errorf("whisper service did not create source.vtt: %w", statErr)
	}
	if info.Size() == 0 {
		return fmt.Errorf("whisper service created empty source.vtt")
	}
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read source.vtt: %w", err)
	}
	if !strings.HasPrefix(string(content), "WEBVTT") {
		return fmt.Errorf("whisper service returned invalid source.vtt")
	}
	return nil
}

func ensureSourceDir(sourcePath string) error {
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		return fmt.Errorf("create transcript directory: %w", err)
	}
	return nil
}
```

- [ ] **Step 4: Implement the HTTP transcriber**

Create `backend/internal/runner/http_transcriber.go` with this content:

```go
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type HTTPTranscriber struct {
	baseURL      string
	timeout      time.Duration
	pollInterval time.Duration
	client       *http.Client
}

type whisperTranscriptionResponse struct {
	Transcription whisperTranscription `json:"transcription"`
}

type whisperTranscription struct {
	ID            string  `json:"id"`
	Status        string  `json:"status"`
	ProgressText  string  `json:"progressText"`
	ErrorMessage  *string `json:"errorMessage"`
}

type whisperErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewHTTPTranscriber(baseURL string, timeout time.Duration, pollInterval time.Duration, client *http.Client) *HTTPTranscriber {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPTranscriber{
		baseURL:      strings.TrimRight(baseURL, "/"),
		timeout:      timeout,
		pollInterval: pollInterval,
		client:       client,
	}
}

func (t *HTTPTranscriber) Transcribe(ctx context.Context, request TranscriptionRequest) error {
	if err := ensureSourceDir(request.SourcePath); err != nil {
		return err
	}
	requestCtx := ctx
	var cancel context.CancelFunc
	if t.timeout > 0 {
		requestCtx, cancel = context.WithTimeout(ctx, t.timeout)
		defer cancel()
	}

	transcriptionID, err := t.createTranscription(requestCtx, request)
	if err != nil {
		return err
	}

	for {
		status, err := t.getTranscription(requestCtx, transcriptionID)
		if err != nil {
			return err
		}
		if request.OnProgress != nil && strings.TrimSpace(status.ProgressText) != "" {
			if err := request.OnProgress(status.ProgressText); err != nil {
				return fmt.Errorf("update transcription progress: %w", err)
			}
		}

		switch status.Status {
		case "queued", "running":
			if err := sleepContext(requestCtx, t.pollInterval); err != nil {
				return fmt.Errorf("wait before polling whisper service: %w", err)
			}
		case "completed":
			if err := t.downloadVTT(requestCtx, transcriptionID, request.SourcePath); err != nil {
				return err
			}
			return ensureValidSourceVTT(request.SourcePath)
		case "failed":
			if status.ErrorMessage != nil && strings.TrimSpace(*status.ErrorMessage) != "" {
				return fmt.Errorf("whisper transcription failed: %s", *status.ErrorMessage)
			}
			return fmt.Errorf("whisper transcription failed")
		default:
			return fmt.Errorf("whisper transcription returned invalid status %q", status.Status)
		}
	}
}

func (t *HTTPTranscriber) createTranscription(ctx context.Context, request TranscriptionRequest) (string, error) {
	reader, writer := io.Pipe()
	multipartWriter := multipart.NewWriter(writer)

	go func() {
		defer writer.Close()
		defer multipartWriter.Close()

		if err := multipartWriter.WriteField("model", request.Model); err != nil {
			_ = writer.CloseWithError(err)
			return
		}
		if err := multipartWriter.WriteField("computeType", request.ComputeType); err != nil {
			_ = writer.CloseWithError(err)
			return
		}
		if err := multipartWriter.WriteField("language", request.Language); err != nil {
			_ = writer.CloseWithError(err)
			return
		}
		if request.JobID != "" {
			if err := multipartWriter.WriteField("jobId", request.JobID); err != nil {
				_ = writer.CloseWithError(err)
				return
			}
		}

		file, err := os.Open(request.AudioPath)
		if err != nil {
			_ = writer.CloseWithError(err)
			return
		}
		defer file.Close()

		part, err := multipartWriter.CreateFormFile("audio", filepath.Base(request.AudioPath))
		if err != nil {
			_ = writer.CloseWithError(err)
			return
		}
		if _, err := io.Copy(part, file); err != nil {
			_ = writer.CloseWithError(err)
			return
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/transcriptions", reader)
	if err != nil {
		return "", fmt.Errorf("create whisper transcription request: %w", err)
	}
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send whisper transcription request: %w", err)
	}
	defer resp.Body.Close()

	var decoded whisperTranscriptionResponse
	if err := decodeWhisperResponse(resp, http.StatusAccepted, &decoded); err != nil {
		return "", err
	}
	if decoded.Transcription.ID == "" {
		return "", fmt.Errorf("whisper transcription response missing id")
	}
	slog.Debug("whisper transcription created", "job_id", request.JobID, "transcription_id", decoded.Transcription.ID)
	return decoded.Transcription.ID, nil
}

func (t *HTTPTranscriber) getTranscription(ctx context.Context, transcriptionID string) (whisperTranscription, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.baseURL+"/transcriptions/"+transcriptionID, nil)
	if err != nil {
		return whisperTranscription{}, fmt.Errorf("create whisper status request: %w", err)
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return whisperTranscription{}, fmt.Errorf("send whisper status request: %w", err)
	}
	defer resp.Body.Close()

	var decoded whisperTranscriptionResponse
	if err := decodeWhisperResponse(resp, http.StatusOK, &decoded); err != nil {
		return whisperTranscription{}, err
	}
	if decoded.Transcription.ID == "" || decoded.Transcription.Status == "" {
		return whisperTranscription{}, fmt.Errorf("whisper status response missing required fields")
	}
	return decoded.Transcription, nil
}

func (t *HTTPTranscriber) downloadVTT(ctx context.Context, transcriptionID string, sourcePath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.baseURL+"/transcriptions/"+transcriptionID+"/vtt", nil)
	if err != nil {
		return fmt.Errorf("create whisper vtt request: %w", err)
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("send whisper vtt request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return decodeWhisperResponse(resp, http.StatusOK, &whisperTranscriptionResponse{})
	}

	file, err := os.Create(sourcePath)
	if err != nil {
		return fmt.Errorf("create source.vtt: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("write source.vtt: %w", err)
	}
	return nil
}

func decodeWhisperResponse(resp *http.Response, wantStatus int, target any) error {
	if resp.StatusCode != wantStatus {
		var apiErr whisperErrorResponse
		if err := json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&apiErr); err == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("whisper service request failed with status %d: %s", resp.StatusCode, apiErr.Error.Message)
		}
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("whisper service request failed with status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode whisper service response: %w", err)
	}
	return nil
}
```

- [ ] **Step 5: Run HTTP transcriber tests**

Run from `backend/`:

```bash
mise exec -- go test ./internal/runner -run TestHTTPTranscriber
```

Expected: PASS.

- [ ] **Step 6: Commit**

Run from repository root:

```bash
git add backend/internal/runner/transcriber.go backend/internal/runner/http_transcriber.go backend/internal/runner/http_transcriber_test.go
git commit -m "feat(backend): add whisper http transcriber"
```

## Task 7: Integrate Transcriber Interface into RealRunner

**Files:**
- Modify: `backend/internal/runner/real_runner.go`
- Modify: `backend/internal/runner/real_runner_test.go`

- [ ] **Step 1: Update RealRunner constructor and transcribing stage**

Apply this patch:

```diff
*** Begin Patch
*** Update File: backend/internal/runner/real_runner.go
@@
 	whisperModel       string
 	whisperComputeType string
+	transcriber        Transcriber
 	translator         Translator
 }
 
-func NewRealRunner(store Store, downloadTimeout time.Duration, whisperModel string, whisperComputeType string, translator Translator) *RealRunner {
-	return &RealRunner{store: store, downloadTimeout: downloadTimeout, whisperModel: whisperModel, whisperComputeType: whisperComputeType, translator: translator}
+func NewRealRunner(store Store, downloadTimeout time.Duration, whisperModel string, whisperComputeType string, transcriber Transcriber, translator Translator) *RealRunner {
+	return &RealRunner{store: store, downloadTimeout: downloadTimeout, whisperModel: whisperModel, whisperComputeType: whisperComputeType, transcriber: transcriber, translator: translator}
 }
@@
-	if err := r.set(job.ID, store.StatusTranscribing, "调用 whisper-cli 生成 source.vtt", ""); err != nil {
+	if err := r.set(job.ID, store.StatusTranscribing, "正在提交音频到 Whisper 服务", ""); err != nil {
 		return r.fail(job, store.StatusTranscribing, err, jobStartedAt)
 	}
@@
-	if err := transcribeAudio(ctx, job.ID, audioPath, sourcePath, r.whisperModel, r.whisperComputeType, job.SourceLanguage); err != nil {
+	if err := r.transcriber.Transcribe(ctx, TranscriptionRequest{
+		JobID:       job.ID,
+		AudioPath:   audioPath,
+		SourcePath:  sourcePath,
+		Model:       r.whisperModel,
+		ComputeType: r.whisperComputeType,
+		Language:    job.SourceLanguage,
+		OnProgress: func(text string) error {
+			return r.set(job.ID, store.StatusTranscribing, text, "")
+		},
+	}); err != nil {
 		return r.fail(job, store.StatusTranscribing, err, jobStartedAt)
 	}
*** End Patch
```

Run:

```bash
gofmt -w backend/internal/runner/real_runner.go
```

- [ ] **Step 2: Update runner tests to use fake transcriber**

In `backend/internal/runner/real_runner_test.go`, add this helper near `fakeTranslator`:

```go
type fakeTranscriber struct {
	vtt      string
	err      error
	calls    []TranscriptionRequest
	progress []string
}

func (t *fakeTranscriber) Transcribe(ctx context.Context, request TranscriptionRequest) error {
	t.calls = append(t.calls, request)
	if t.err != nil {
		return t.err
	}
	for _, text := range t.progress {
		if request.OnProgress != nil {
			if err := request.OnProgress(text); err != nil {
				return err
			}
		}
	}
	content := t.vtt
	if content == "" {
		content = "WEBVTT\n\n00:00:00.000 --> 00:00:01.000\nreal transcript\n"
	}
	return os.WriteFile(request.SourcePath, []byte(content), 0o644)
}
```

- [ ] **Step 3: Replace `NewRealRunner` calls in tests**

Update every `NewRealRunner(...)` call in `backend/internal/runner/real_runner_test.go` to pass a fake transcriber before the translator. For success cases use:

```go
transcriber := &fakeTranscriber{}
err := NewRealRunner(testStore, 10*time.Minute, "small", "default", transcriber, fakeTranslator{translations: []string{"translated"}}).Start(context.Background(), job)
```

For the main completion test, use the existing `translator := fakeTranslator{translations: []string{"translated one"}}` and create:

```go
transcriber := &fakeTranscriber{}
if err := NewRealRunner(testStore, 10*time.Minute, "tiny", "int8", transcriber, translator).Start(context.Background(), job); err != nil {
	t.Fatalf("Start() error = %v", err)
}
```

Replace the old `whisperCall := findExecCall(calls, "whisper-cli")` assertion block with:

```go
if len(transcriber.calls) != 1 {
	t.Fatalf("transcriber calls = %#v, want 1 call", transcriber.calls)
}
transcribeCall := transcriber.calls[0]
sourcePath := filepath.Join(jobDir, "source.vtt")
if transcribeCall.AudioPath != audioPath {
	t.Fatalf("AudioPath = %q, want %q", transcribeCall.AudioPath, audioPath)
}
if transcribeCall.SourcePath != sourcePath {
	t.Fatalf("SourcePath = %q, want %q", transcribeCall.SourcePath, sourcePath)
}
if transcribeCall.Model != "tiny" {
	t.Fatalf("Model = %q, want tiny", transcribeCall.Model)
}
if transcribeCall.ComputeType != "int8" {
	t.Fatalf("ComputeType = %q, want int8", transcribeCall.ComputeType)
}
if transcribeCall.Language != "zh" {
	t.Fatalf("Language = %q, want zh", transcribeCall.Language)
}
```

For transcription failure test, use:

```go
transcriber := &fakeTranscriber{err: errors.New("model download error")}
err := NewRealRunner(testStore, 10*time.Minute, "small", "default", transcriber, fakeTranslator{}).Start(context.Background(), job)
```

- [ ] **Step 4: Run runner tests and fix remaining constructor call sites**

Run from `backend/`:

```bash
mise exec -- go test ./internal/runner
```

Expected before fixing all call sites: compile failures identifying remaining old constructor calls. Update each call using the same fake transcriber pattern, then rerun until PASS.

- [ ] **Step 5: Continue to app wiring before committing**

Do not commit yet. At this point `backend/internal/app/app.go` still calls the old `NewRealRunner` signature, so the whole backend package set will not compile. Continue directly to Task 8, then commit the runner and app wiring together after `go test ./...` passes.

## Task 8: Wire HTTP Transcriber into Backend App

**Files:**
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/app/app_test.go`

- [ ] **Step 1: Wire `NewHTTPTranscriber` in `NewHTTPHandler`**

Apply this patch:

```diff
*** Begin Patch
*** Update File: backend/internal/app/app.go
@@
 	}
 	translator := runner.NewChatTranslator(config.LLMBaseURL, config.LLMAPIKey, config.LLMModel, config.LLMTimeout, http.DefaultClient)
-	jobRunner := runner.NewRealRunner(database, config.DownloadTimeout, config.WhisperModel, config.WhisperComputeType, translator)
+	transcriber := runner.NewHTTPTranscriber(config.WhisperBaseURL, config.WhisperTimeout, config.WhisperPollInterval, http.DefaultClient)
+	jobRunner := runner.NewRealRunner(database, config.DownloadTimeout, config.WhisperModel, config.WhisperComputeType, transcriber, translator)
*** End Patch
```

Run:

```bash
gofmt -w backend/internal/app/app.go
```

- [ ] **Step 2: Update `NewHTTPHandler` config literals in tests**

In `backend/internal/app/app_test.go`, update the `Config{...}` literal in `TestNewHTTPHandlerRequiresToolsByDefault` to include:

```go
		WhisperBaseURL:      "http://127.0.0.1:8081",
		WhisperTimeout:      time.Minute,
		WhisperPollInterval: time.Second,
```

Add `time` to the import list:

```go
import (
	"errors"
	"strings"
	"testing"
	"time"
)
```

- [ ] **Step 3: Run backend tests**

Run from `backend/`:

```bash
mise exec -- go test ./...
```

Expected: PASS.

- [ ] **Step 4: Commit**

Run from repository root:

```bash
git add backend/internal/runner/real_runner.go backend/internal/runner/real_runner_test.go backend/internal/app/app.go backend/internal/app/app_test.go
git commit -m "feat(backend): use whisper service transcriber"
```

## Task 9: Update Docker, Compose, Env, and Taskfile

**Files:**
- Modify: `backend/Dockerfile`
- Create: `whisper/Dockerfile`
- Modify: `docker-compose.yml`
- Modify: `.env.example`
- Modify: `Taskfile.yml`

- [ ] **Step 1: Update backend Dockerfile**

Replace `backend/Dockerfile` with this content:

```dockerfile
# ---- Stage 1: Go build ----
FROM golang:1.22-bookworm AS go-builder

WORKDIR /build

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=1 go build -o /build/server ./cmd/server

# ---- Stage 2: Runtime ----
FROM python:3.12-slim-bookworm

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates ffmpeg && \
    rm -rf /var/lib/apt/lists/* && \
    pip install --no-cache-dir yt-dlp

COPY --from=go-builder /build/server /usr/local/bin/server

ENV LSI_ADDR=0.0.0.0:8080 \
    LSI_DB_PATH=/data/backend.sqlite3 \
    LSI_WORK_DIR=/data/jobs

RUN mkdir -p /data/jobs

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/server"]
```

- [ ] **Step 2: Add Whisper Dockerfile**

Create `whisper/Dockerfile` with this content:

```dockerfile
FROM python:3.12-slim-bookworm

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates ffmpeg && \
    rm -rf /var/lib/apt/lists/* && \
    pip install --no-cache-dir uv

COPY whisper/ ./
RUN uv sync --no-dev --no-editable

ENV PATH="/app/.venv/bin:$PATH" \
    HF_HOME=/huggingface \
    LSI_WHISPER_WORK_DIR=/data/transcriptions

RUN mkdir -p /data/transcriptions /huggingface

EXPOSE 8081

CMD ["uvicorn", "whisper_cli.server:app", "--host", "0.0.0.0", "--port", "8081"]
```

- [ ] **Step 3: Update Docker Compose**

Replace `docker-compose.yml` with this content:

```yaml
services:
  backend:
    build:
      context: .
      dockerfile: backend/Dockerfile
    env_file:
      - .env
    environment:
      LSI_WHISPER_BASE_URL: http://whisper:8081
    ports:
      - "${LSI_DOCKER_BIND_HOST:?set LSI_DOCKER_BIND_HOST in .env}:8080:8080"
    volumes:
      - lsi-data:/data
    depends_on:
      whisper:
        condition: service_healthy
    restart: unless-stopped

  whisper:
    build:
      context: .
      dockerfile: whisper/Dockerfile
    env_file:
      - .env
    volumes:
      - lsi-hf-cache:/huggingface
      - lsi-whisper-data:/data
    healthcheck:
      test: ["CMD", "python", "-c", "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8081/healthz', timeout=2)"]
      interval: 10s
      timeout: 3s
      retries: 6
    restart: unless-stopped

volumes:
  lsi-data:
  lsi-hf-cache:
  lsi-whisper-data:
```

- [ ] **Step 4: Update `.env.example`**

Apply this patch:

```diff
*** Begin Patch
*** Update File: .env.example
@@
 # Whisper 模型（首次处理视频时自动下载到独立持久卷）
 LSI_WHISPER_MODEL=small
 LSI_WHISPER_COMPUTE_TYPE=default
+LSI_WHISPER_BASE_URL=http://whisper:8081
+LSI_WHISPER_TIMEOUT=30m
+LSI_WHISPER_POLL_INTERVAL=2s
 
 # 下载超时
 LSI_DOWNLOAD_TIMEOUT=10m
*** End Patch
```

- [ ] **Step 5: Update Taskfile**

Apply this patch:

```diff
*** Begin Patch
*** Update File: Taskfile.yml
@@
   dev:backend:
     desc: Run backend API with the local real toolchain
     dir: backend
     cmds:
-      - task deps:whisper
       - >-
-        PATH="$PWD/../whisper/.venv/bin:$PATH"
         LSI_DOWNLOAD_TIMEOUT=${LSI_DOWNLOAD_TIMEOUT:-10m}
         LSI_WHISPER_MODEL=${LSI_WHISPER_MODEL:-small}
+        LSI_WHISPER_BASE_URL=${LSI_WHISPER_BASE_URL:-http://127.0.0.1:8081}
+        LSI_WHISPER_TIMEOUT=${LSI_WHISPER_TIMEOUT:-30m}
+        LSI_WHISPER_POLL_INTERVAL=${LSI_WHISPER_POLL_INTERVAL:-2s}
         LSI_LLM_BASE_URL=${LSI_LLM_BASE_URL:-https://api.openai.com}
         LSI_LLM_TIMEOUT=${LSI_LLM_TIMEOUT:-2m}
         LSI_ADDR=${LSI_ADDR:-127.0.0.1:8080}
         mise exec -- go run ./cmd/server
 
+  dev:whisper:
+    desc: Run the Whisper transcription service
+    dir: whisper
+    cmds:
+      - mise exec -- uv sync --dev
+      - >-
+        LSI_WHISPER_WORK_DIR=${LSI_WHISPER_WORK_DIR:-./data/transcriptions}
+        HF_HOME=${HF_HOME:-./data/huggingface}
+        mise exec -- uv run uvicorn whisper_cli.server:app --host 127.0.0.1 --port 8081
+
   dev:extension:
*** End Patch
```

- [ ] **Step 6: Validate module tests after config changes**

Run from repository root:

```bash
task test:backend
task test:whisper
```

Expected: both pass.

- [ ] **Step 7: Validate Docker build**

Run from repository root:

```bash
task docker:build
```

Expected: Compose builds and starts both `backend` and `whisper` services. If external model downloads make this too slow in the current environment, record the exact failure output and verify at least `docker compose config` succeeds.

- [ ] **Step 8: Commit**

Run from repository root:

```bash
git add backend/Dockerfile whisper/Dockerfile docker-compose.yml .env.example Taskfile.yml
git commit -m "build: split whisper service container"
```

## Task 10: Final Verification and Cleanup

**Files:**
- Inspect only: repository root and changed files

- [ ] **Step 1: Search for deprecated CLI references**

Run from repository root:

```bash
rg "whisper-cli|exec whisper|PATH=\"\$PWD/../whisper/.venv/bin" backend whisper Taskfile.yml docker-compose.yml .env.example docs/superpowers/specs docs/superpowers/plans
```

Expected: only historical design documents may mention `whisper-cli`. Runtime code, Dockerfiles, Taskfile commands, and tests must not require it.

- [ ] **Step 2: Run backend tests**

Run from repository root:

```bash
task test:backend
```

Expected: PASS.

- [ ] **Step 3: Run Whisper tests**

Run from repository root:

```bash
task test:whisper
```

Expected: PASS.

- [ ] **Step 4: Run full project check if time allows**

Run from repository root:

```bash
task check
```

Expected: PASS. If extension dependency setup is missing in the environment, record the missing dependency error and confirm backend and whisper checks passed.

- [ ] **Step 5: Inspect git status**

Run from repository root:

```bash
git status --short
```

Expected: clean working tree. If generated lockfile changes remain, include them in the final commit below.

- [ ] **Step 6: Commit verification cleanup if needed**

If Step 5 shows remaining intentional changes, inspect them first:

```bash
git diff --stat
git diff -- whisper/uv.lock
```

If the only remaining change is an intentional `whisper/uv.lock` update from dependency resolution, commit it:

```bash
git add whisper/uv.lock
git commit -m "chore(whisper): update service dependency lockfile"
```

If Step 5 is clean, do not create an empty commit. If Step 5 shows any other file, stop and review why it was not included in the earlier task commits before staging it.

## Plan Self-Review

- Spec coverage: The plan implements HTTP upload, asynchronous Whisper tasks, status polling, VTT retrieval, CLI removal, backend `Transcriber` boundary, Docker Compose split, env config, Taskfile updates, and tests for backend and whisper modules.
- Placeholder scan: The plan contains concrete file paths, commands, code snippets, API fields, expected statuses, and commit messages. It does not rely on unspecified handlers or unnamed tests.
- Type consistency: Python response fields use camelCase JSON matching the spec. Go request fields use `TranscriptionRequest` consistently. The backend config field names match their planned use in `app.go`.
