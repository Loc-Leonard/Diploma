from __future__ import annotations

import json
import tempfile
from pathlib import Path

from fastapi import FastAPI, File, HTTPException, UploadFile

from .config import ProjectConfig
from .extraction import build_document_payload

CONFIG_PATH = Path(__file__).resolve().parents[2] / "config.yaml"
PROJECT_CONFIG = ProjectConfig.from_file(CONFIG_PATH)

app = FastAPI(title="docscan-cv", version="0.1.0")


@app.get("/health")
def healthcheck() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/process-file")
async def process_file(file: UploadFile = File(...)) -> dict:
    suffix = Path(file.filename or "upload.bin").suffix or ".bin"
    temp_path: Path | None = None
    try:
        with tempfile.NamedTemporaryFile(delete=False, suffix=suffix) as temp_file:
            temp_path = Path(temp_file.name)
            while True:
                chunk = await file.read(1024 * 1024)
                if not chunk:
                    break
                temp_file.write(chunk)

        payload = {
            "source_path": str(temp_path),
            **build_document_payload(temp_path, "ttn", PROJECT_CONFIG),
        }
        return json.loads(json.dumps(payload, ensure_ascii=False))
    except Exception as exc:
        raise HTTPException(status_code=500, detail=str(exc)) from exc
    finally:
        await file.close()
        if temp_path and temp_path.exists():
            temp_path.unlink()
