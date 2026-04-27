from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Any

import yaml


@dataclass(slots=True)
class ProjectConfig:
    dataset_root: Path
    target_subdirs: list[str]
    ttn_keywords: list[str]
    non_ttn_keywords: list[str]
    supported_extensions: list[str]
    pdf_max_pages: int
    inventory_output: Path
    extraction_output: Path
    dataset_output_root: Path
    text_output_dir: Path

    @classmethod
    def from_file(cls, path: str | Path) -> "ProjectConfig":
        config_path = Path(path).resolve()
        raw = yaml.safe_load(config_path.read_text(encoding="utf-8"))
        return cls(
            dataset_root=(config_path.parent / raw["dataset_root"]).resolve(),
            target_subdirs=list(raw["target_subdirs"]),
            ttn_keywords=[item.casefold() for item in raw["ttn_keywords"]],
            non_ttn_keywords=[item.casefold() for item in raw["non_ttn_keywords"]],
            supported_extensions=[item.casefold() for item in raw["supported_extensions"]],
            pdf_max_pages=int(raw.get("pdf_max_pages", 3)),
            inventory_output=(config_path.parent / raw["inventory_output"]).resolve(),
            extraction_output=(config_path.parent / raw["extraction_output"]).resolve(),
            dataset_output_root=(config_path.parent / raw["dataset_output_root"]).resolve(),
            text_output_dir=(config_path.parent / raw["text_output_dir"]).resolve(),
        )

    def to_dict(self) -> dict[str, Any]:
        return {
            "dataset_root": str(self.dataset_root),
            "target_subdirs": self.target_subdirs,
            "ttn_keywords": self.ttn_keywords,
            "non_ttn_keywords": self.non_ttn_keywords,
            "supported_extensions": self.supported_extensions,
            "pdf_max_pages": self.pdf_max_pages,
            "inventory_output": str(self.inventory_output),
            "extraction_output": str(self.extraction_output),
            "dataset_output_root": str(self.dataset_output_root),
            "text_output_dir": str(self.text_output_dir),
        }
