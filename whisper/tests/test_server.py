from fastapi.testclient import TestClient

from whisper_cli.server import create_app


def test_healthz_returns_ok():
    client = TestClient(create_app(start_worker=False))

    response = client.get("/healthz")

    assert response.status_code == 200
    assert response.json() == {"status": "ok"}
