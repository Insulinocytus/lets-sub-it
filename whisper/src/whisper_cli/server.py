from __future__ import annotations

from fastapi import FastAPI


def create_app(*, start_worker: bool = True) -> FastAPI:
    app = FastAPI(title="Lets Sub It Whisper Service")

    @app.get("/healthz")
    def healthz() -> dict[str, str]:
        return {"status": "ok"}

    return app


app = create_app()
