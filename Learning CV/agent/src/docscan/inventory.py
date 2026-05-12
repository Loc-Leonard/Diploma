from __future__ import annotations

import json
from collections import Counter
from dataclasses import asdict, dataclass
from pathlib import Path

from .config import ProjectConfig


@dataclass(slots=True)
class DocumentRecord:
    path: str
    relative_path: str
    extension: str
    size_bytes: int
    predicted_type: str
    source_group: str


def classify_document(path: Path, config: ProjectConfig) -> str:
    name = path.name.casefold()
    if any(keyword in name for keyword in config.non_ttn_keywords):
        return "non_ttn_related"
    if any(keyword in name for keyword in config.ttn_keywords):
        return "ttn"
    return "unknown"


def iter_target_files(config: ProjectConfig) -> list[Path]:
    files: list[Path] = []
    for subdir in config.target_subdirs:
        root = config.dataset_root / subdir
        if not root.exists():
            continue
        for path in root.rglob("*"):
            if path.is_file() and path.suffix.casefold() in config.supported_extensions:
                files.append(path)
    return sorted(files)


def iter_all_files(config: ProjectConfig) -> list[Path]:
    return sorted(path for path in config.dataset_root.rglob("*") if path.is_file())


def build_inventory(config: ProjectConfig) -> dict:
    records: list[DocumentRecord] = []
    for path in iter_target_files(config):
        records.append(
            DocumentRecord(
                path=str(path),
                relative_path=str(path.relative_to(config.dataset_root)),
                extension=path.suffix.casefold(),
                size_bytes=path.stat().st_size,
                predicted_type=classify_document(path, config),
                source_group=path.parent.name,
            )
        )

    stats = {
        "total_files": len(records),
        "by_type": dict(Counter(record.predicted_type for record in records)),
        "by_extension": dict(Counter(record.extension for record in records)),
    }
    return {
        "config": config.to_dict(),
        "stats": stats,
        "records": [asdict(record) for record in records],
    }


def save_inventory(payload: dict, output_path: Path) -> Path:
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(
        json.dumps(payload, ensure_ascii=False, indent=2),
        encoding="utf-8",
    )
    return output_path
