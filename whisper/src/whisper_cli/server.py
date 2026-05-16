from __future__ import annotations

from dataclasses import dataclass
from datetime import UTC, datetime
from pathlib import Path
from uuid import uuid4

from fastapi import FastAPI, File, Form, Request, UploadFile
from fastapi.responses import JSONResponse


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
    def __init__(self, *, work_dir: Path) -> None:
        self.work_dir = work_dir
        self.tasks: dict[str, TranscriptionTask] = {}

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
        with audio_path.open("wb") as audio_file:
            while chunk := await audio.read(1024 * 1024):
                audio_file.write(chunk)

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
        self.tasks[task_id] = task
        return task

    def get(self, task_id: str) -> TranscriptionTask:
        task = self.tasks.get(task_id)
        if task is None:
            raise APIError(404, "not_found", "transcription not found")
        return task


def create_app(
    *, service: TranscriptionService | None = None, start_worker: bool = True
) -> FastAPI:
    app = FastAPI(title="Lets Sub It Whisper Service")
    app.state.transcription_service = service or TranscriptionService(work_dir=Path("work"))
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
        return {"transcription": serialize_task(task)}

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


app = create_app()
