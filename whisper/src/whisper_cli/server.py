from __future__ import annotations

import shutil
import threading
import os
from contextlib import asynccontextmanager
from dataclasses import dataclass
from datetime import UTC, datetime
from pathlib import Path
from typing import Callable
from uuid import uuid4

import anyio
from fastapi import FastAPI, File, Form, Request, UploadFile
from fastapi.responses import JSONResponse, Response

from whisper_cli.transcribe import TranscriptionResult, transcribe_audio
from whisper_cli.vtt import Segment, render_vtt


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
    audio_path: Path
    vtt_path: Path
    model: str
    compute_type: str | None
    requested_language: str
    job_id: str | None
    language: str | None
    duration_seconds: float | None
    segments: list[dict[str, object]] | None
    error_message: str | None
    created_at: datetime
    updated_at: datetime


class TranscriptionService:
    def __init__(
        self,
        *,
        work_dir: Path,
        transcribe: Callable[..., TranscriptionResult] = transcribe_audio,
    ) -> None:
        self.work_dir = work_dir
        self.tasks: dict[str, TranscriptionTask] = {}
        self.transcribe = transcribe
        self.queue: list[str] = []
        self.condition = threading.Condition()
        self.worker: threading.Thread | None = None
        self.stopping = False

    async def create(
        self,
        *,
        audio: UploadFile,
        model: str,
        language: str,
        compute_type: str | None,
        job_id: str | None,
    ) -> TranscriptionTask:
        task_id = f"tr_{uuid4().hex}"
        task_dir = self.work_dir / task_id
        task_dir.mkdir(parents=True, exist_ok=False)

        audio_path = task_dir / "audio.mp3"
        try:
            with audio_path.open("wb") as audio_file:
                while chunk := await audio.read(1024 * 1024):
                    await anyio.to_thread.run_sync(audio_file.write, chunk)
        except Exception:
            shutil.rmtree(task_dir, ignore_errors=True)
            raise

        now = datetime.now(UTC)
        task = TranscriptionTask(
            id=task_id,
            status="queued",
            progress=0,
            progress_text="等待转写",
            audio_path=audio_path,
            vtt_path=task_dir / "source.vtt",
            model=model,
            compute_type=compute_type,
            requested_language=language,
            job_id=job_id,
            language=None,
            duration_seconds=None,
            segments=None,
            error_message=None,
            created_at=now,
            updated_at=now,
        )
        with self.condition:
            self.tasks[task_id] = task
            self.queue.append(task_id)
            self.condition.notify()
        return task

    def get(self, task_id: str) -> TranscriptionTask:
        with self.condition:
            task = self.tasks.get(task_id)
            if task is None:
                raise APIError(404, "not_found", "transcription not found")
            return task

    def start(self) -> None:
        with self.condition:
            if self.worker is not None and self.worker.is_alive():
                return
            self.stopping = False
            self.worker = threading.Thread(target=self._worker_loop, daemon=True)
            self.worker.start()

    def stop(self) -> None:
        with self.condition:
            self.stopping = True
            self.condition.notify_all()
            worker = self.worker
        if worker is not None:
            worker.join()
            with self.condition:
                if self.worker is worker:
                    self.worker = None
        else:
            with self.condition:
                self.worker = None

    def run_next(self, timeout: float | None = None) -> bool:
        return self._run_next(timeout, stop_when_stopping=False)

    def _run_next(self, timeout: float | None, *, stop_when_stopping: bool) -> bool:
        with self.condition:
            if stop_when_stopping and self.stopping:
                return False
            if not self.queue:
                self.condition.wait(timeout=timeout)
            if stop_when_stopping and self.stopping:
                return False
            if not self.queue:
                return False
            task = self.tasks[self.queue.pop(0)]
            now = datetime.now(UTC)
            task.status = "running"
            task.progress = 10
            task.progress_text = "正在转写音频"
            task.updated_at = now

        try:
            result = self.transcribe(
                input_path=task.audio_path,
                model_name=task.model,
                language=task.requested_language,
                compute_type=task.compute_type or "default",
            )
            task.vtt_path.write_text(render_vtt(result.segments), encoding="utf-8")
            with self.condition:
                task.status = "completed"
                task.progress = 100
                task.progress_text = "转写完成"
                task.language = result.language
                task.duration_seconds = result.duration_seconds
                task.segments = [serialize_segment(segment) for segment in result.segments]
                task.error_message = None
                task.updated_at = datetime.now(UTC)
        except Exception as exc:
            with self.condition:
                task.status = "failed"
                task.progress = 100
                task.progress_text = "转写失败"
                task.error_message = str(exc)
                task.updated_at = datetime.now(UTC)
        return True

    def _worker_loop(self) -> None:
        while True:
            processed = self._run_next(0.2, stop_when_stopping=True)
            if not processed:
                with self.condition:
                    if self.stopping:
                        return

    def snapshot(self, task_id: str) -> dict[str, object]:
        with self.condition:
            return serialize_task(self.get(task_id))

    def completed_vtt_path(self, task_id: str) -> Path:
        with self.condition:
            task = self.get(task_id)
            if task.status != "completed":
                raise APIError(409, "not_ready", "transcription is not completed")
            return task.vtt_path

    def delete(self, task_id: str) -> None:
        with self.condition:
            task = self.tasks.pop(task_id, None)
            self.queue = [queued_id for queued_id in self.queue if queued_id != task_id]
        if task is not None:
            shutil.rmtree(task.audio_path.parent, ignore_errors=True)


def create_app(
    *, service: TranscriptionService | None = None, start_worker: bool = True
) -> FastAPI:
    @asynccontextmanager
    async def lifespan(app: FastAPI):
        if app.state.start_worker:
            app.state.transcription_service.start()
        yield
        if app.state.start_worker:
            app.state.transcription_service.stop()

    app = FastAPI(title="Lets Sub It Whisper Service", lifespan=lifespan)
    app.state.transcription_service = service or TranscriptionService(
        work_dir=Path(os.getenv("LSI_WHISPER_WORK_DIR", "/data/transcriptions"))
    )
    app.state.start_worker = start_worker

    @app.exception_handler(APIError)
    def api_error_handler(request: Request, exc: APIError) -> JSONResponse:
        return JSONResponse(
            status_code=exc.status_code,
            content={"error": {"code": exc.code, "message": exc.message}},
        )

    @app.get("/healthz")
    def healthz() -> dict[str, str]:
        return {"status": "ok"}

    @app.post("/transcriptions", status_code=202)
    async def create_transcription(
        audio: UploadFile | None = File(default=None),
        model: str | None = Form(default=None),
        language: str | None = Form(default=None),
        compute_type: str | None = Form(default=None, alias="computeType"),
        job_id: str | None = Form(default=None, alias="jobId"),
    ) -> dict[str, object]:
        if audio is None:
            raise APIError(400, "invalid_request", "audio is required")
        if not model:
            raise APIError(400, "invalid_request", "model is required")
        if not language:
            raise APIError(400, "invalid_request", "language is required")

        task = await app.state.transcription_service.create(
            audio=audio,
            model=model,
            language=language,
            compute_type=compute_type,
            job_id=job_id,
        )
        return {"transcription": app.state.transcription_service.snapshot(task.id)}

    @app.get("/transcriptions/{task_id}")
    def get_transcription(task_id: str) -> dict[str, object]:
        return {"transcription": app.state.transcription_service.snapshot(task_id)}

    @app.get("/transcriptions/{task_id}/vtt")
    def get_transcription_vtt(task_id: str) -> Response:
        vtt_path = app.state.transcription_service.completed_vtt_path(task_id)
        return Response(
            content=vtt_path.read_text(encoding="utf-8"),
            media_type="text/vtt; charset=utf-8",
        )

    @app.delete("/transcriptions/{task_id}", status_code=204)
    def delete_transcription(task_id: str) -> Response:
        app.state.transcription_service.delete(task_id)
        return Response(status_code=204)

    return app


def serialize_task(task: TranscriptionTask) -> dict[str, object]:
    return {
        "id": task.id,
        "status": task.status,
        "progress": task.progress,
        "progressText": task.progress_text,
        "language": task.language,
        "durationSeconds": task.duration_seconds,
        "segments": task.segments,
        "errorMessage": task.error_message,
        "createdAt": format_datetime(task.created_at),
        "updatedAt": format_datetime(task.updated_at),
    }


def format_datetime(value: datetime) -> str:
    return value.isoformat().replace("+00:00", "Z")


def serialize_segment(segment: Segment) -> dict[str, object]:
    return {"start": segment.start, "end": segment.end, "text": segment.text}


app = create_app()
