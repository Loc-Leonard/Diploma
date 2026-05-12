from __future__ import annotations

import json
import re
from dataclasses import asdict, dataclass
from functools import lru_cache
from pathlib import Path
from statistics import mean
from typing import Any

import fitz
import numpy as np
from PIL import Image
from rapidocr_onnxruntime import RapidOCR

from .config import ProjectConfig
from .inventory import build_inventory, classify_document, iter_all_files

DIMENSION_PATTERN = re.compile(r"\b\d{2,4}\s*[xх]\s*\d{2,4}\b", re.IGNORECASE)
QUANTITY_PATTERN = re.compile(r"(?P<quantity>\d{2,6})(?:\s*(?P<unit>шт|м3|м2|т))?\b", re.IGNORECASE)
DOC_NUMBER_PATTERN = re.compile(r"\b(?P<number>\d{8})\b")
DATE_PATTERN = re.compile(r"\b\d{2}[./-]\d{2}[./-]\d{2,4}\b")
MATERIAL_PATTERN = re.compile(r"(борт\w*\s+кам\w+|бордюр\w*|поребрик\w*)", re.IGNORECASE)
CONTEXT_KEYWORDS = ("ттн", "наклад", "товар", "материал", "колич", "груз", "кирпич", "камень", "бордюр")


@dataclass(slots=True)
class ExtractionResult:
    document_path: str
    document_type: str
    material_name: str | None
    quantity: int | None
    unit: str | None
    document_number: str | None
    confidence: float
    extraction_method: str


@dataclass(slots=True)
class OCRPageResult:
    page_index: int
    text: str
    source: str
    confidence: float | None


@dataclass(slots=True)
class OCRDocumentResult:
    text: str
    average_confidence: float | None
    pages: list[OCRPageResult]
    method: str


@dataclass(slots=True)
class DatasetProcessResult:
    source_path: str
    relative_path: str
    output_json_path: str
    supported: bool
    predicted_type: str
    processing_status: str
    extraction: dict | None


@lru_cache(maxsize=1)
def get_ocr_engine() -> RapidOCR:
    return RapidOCR()


def pil_to_numpy(path: Path) -> np.ndarray:
    with Image.open(path) as image:
        rgb = image.convert("RGB")
        return np.array(rgb)


def render_pdf_page(page: fitz.Page, zoom: float = 2.0) -> np.ndarray:
    pix = page.get_pixmap(matrix=fitz.Matrix(zoom, zoom), alpha=False)
    return np.frombuffer(pix.samples, dtype=np.uint8).reshape(pix.height, pix.width, pix.n)


def run_ocr_on_image(image: np.ndarray) -> tuple[str, float | None]:
    result, _elapsed = get_ocr_engine()(image)
    if not result:
        return "", None

    lines = [item[1].strip() for item in result if len(item) >= 2 and str(item[1]).strip()]
    confidences = [float(item[2]) for item in result if len(item) >= 3]
    text = "\n".join(lines)
    confidence = round(mean(confidences), 4) if confidences else None
    return text, confidence


def extract_text_from_pdf(path: Path, max_pages: int) -> OCRDocumentResult:
    pages: list[OCRPageResult] = []
    methods: list[str] = []
    with fitz.open(path) as document:
        for page_index, page in enumerate(document):
            if page_index >= max_pages:
                break
            embedded_text = page.get_text("text").strip()
            if len(embedded_text) >= 20:
                pages.append(
                    OCRPageResult(
                        page_index=page_index,
                        text=embedded_text,
                        source="embedded_text",
                        confidence=1.0,
                    )
                )
                methods.append("embedded_text")
                continue

            image = render_pdf_page(page)
            ocr_text, confidence = run_ocr_on_image(image)
            pages.append(
                OCRPageResult(
                    page_index=page_index,
                    text=ocr_text,
                    source="ocr_rendered_page",
                    confidence=confidence,
                )
            )
            methods.append("ocr_rendered_page")

    text = "\n\n".join(page.text for page in pages if page.text)
    confidences = [page.confidence for page in pages if page.confidence is not None]
    average_confidence = round(mean(confidences), 4) if confidences else None
    method = "+".join(sorted(set(methods))) if methods else "none"
    return OCRDocumentResult(text=text, average_confidence=average_confidence, pages=pages, method=method)


def extract_text_from_image(path: Path) -> OCRDocumentResult:
    image = pil_to_numpy(path)
    text, confidence = run_ocr_on_image(image)
    return OCRDocumentResult(
        text=text,
        average_confidence=confidence,
        pages=[OCRPageResult(page_index=0, text=text, source="ocr_image", confidence=confidence)],
        method="ocr_image",
    )


def extract_document_text(path: Path, config: ProjectConfig) -> OCRDocumentResult:
    suffix = path.suffix.casefold()
    if suffix == ".pdf":
        return extract_text_from_pdf(path, config.pdf_max_pages)
    if suffix in {".jpg", ".jpeg", ".png"}:
        return extract_text_from_image(path)
    return OCRDocumentResult(text="", average_confidence=None, pages=[], method="unsupported")


def extract_from_filename(path: Path) -> ExtractionResult:
    stem = path.stem
    stem_casefold = stem.casefold()
    number_match = DOC_NUMBER_PATTERN.search(stem)
    search_zone = stem
    if number_match:
        search_zone = stem[number_match.end() :].strip()

    dimension_match = DIMENSION_PATTERN.search(stem)
    quantity_matches = list(QUANTITY_PATTERN.finditer(search_zone))
    has_context = bool(dimension_match or any(keyword in stem_casefold for keyword in CONTEXT_KEYWORDS))

    material_name = None
    if dimension_match:
        dimension = dimension_match.group(0).replace("х", "x").replace(" ", "")
        material_name = f"Бортовой камень {dimension}"

    quantity = None
    unit = None
    if has_context:
        quantity = int(quantity_matches[-1].group("quantity")) if quantity_matches else None
        unit = quantity_matches[-1].group("unit") if quantity_matches and quantity_matches[-1].group("unit") else None
    if unit is None and quantity is not None:
        unit = "шт"

    confidence = 0.15
    if material_name:
        confidence += 0.35
    if quantity is not None:
        confidence += 0.35
    if "ттн" in stem_casefold:
        confidence += 0.15

    return ExtractionResult(
        document_path=str(path),
        document_type="filename_heuristic",
        material_name=material_name,
        quantity=quantity,
        unit=unit,
        document_number=number_match.group("number") if number_match and has_context else None,
        confidence=round(min(confidence, 0.95), 2),
        extraction_method="filename_rules",
    )


def extract_from_text(text: str, path: Path, ocr_confidence: float | None, predicted_type: str) -> ExtractionResult:
    normalized = text.replace(",", ".")
    normalized_casefold = normalized.casefold()

    number_match = DOC_NUMBER_PATTERN.search(normalized)
    dimension_match = DIMENSION_PATTERN.search(normalized)
    quantity_matches = list(QUANTITY_PATTERN.finditer(normalized_casefold))
    material_match = MATERIAL_PATTERN.search(normalized)
    has_context = bool(material_match or dimension_match or any(keyword in normalized_casefold for keyword in CONTEXT_KEYWORDS))

    material_name = None
    if material_match and dimension_match:
        dimension = dimension_match.group(0).replace("х", "x").replace(" ", "")
        material_name = f"Бортовой камень {dimension}"
    elif material_match:
        material_name = material_match.group(0).strip()
    elif dimension_match:
        dimension = dimension_match.group(0).replace("х", "x").replace(" ", "")
        material_name = f"Бортовой камень {dimension}"

    quantity = None
    unit = None
    if has_context or predicted_type == "ttn":
        for match in reversed(quantity_matches):
            value = int(match.group("quantity"))
            if 1 <= value <= 100000:
                quantity = value
                unit = match.group("unit")
                break

    confidence = 0.2
    if material_name:
        confidence += 0.25
    if quantity is not None:
        confidence += 0.25
    if number_match:
        confidence += 0.15
    if ocr_confidence is not None:
        confidence += min(max(ocr_confidence, 0.0), 1.0) * 0.15

    return ExtractionResult(
        document_path=str(path),
        document_type="ocr_text",
        material_name=material_name,
        quantity=quantity,
        unit=unit,
        document_number=number_match.group("number") if number_match and (has_context or predicted_type == "ttn") else None,
        confidence=round(min(confidence, 0.95), 2),
        extraction_method="ocr_text_rules",
    )


def merge_extractions(ocr_result: ExtractionResult, filename_result: ExtractionResult) -> ExtractionResult:
    return ExtractionResult(
        document_path=ocr_result.document_path,
        document_type="merged",
        material_name=filename_result.material_name or ocr_result.material_name,
        quantity=filename_result.quantity if filename_result.quantity is not None else ocr_result.quantity,
        unit=filename_result.unit or ocr_result.unit,
        document_number=filename_result.document_number or ocr_result.document_number,
        confidence=max(ocr_result.confidence, filename_result.confidence),
        extraction_method="ocr_then_filename_fallback",
    )


def build_document_payload(path: Path, predicted_type: str, config: ProjectConfig) -> dict[str, Any]:
    filename_extraction = extract_from_filename(path)
    ocr_document = extract_document_text(path, config)
    text_extraction = extract_from_text(ocr_document.text, path, ocr_document.average_confidence, predicted_type)
    merged_extraction = merge_extractions(text_extraction, filename_extraction)

    return {
        "predicted_type": predicted_type,
        "ocr": {
            "method": ocr_document.method,
            "average_confidence": ocr_document.average_confidence,
            "recognized_text": ocr_document.text,
            "pages": [asdict(page) for page in ocr_document.pages],
        },
        "filename_extraction": asdict(filename_extraction),
        "ocr_extraction": asdict(text_extraction),
        "extraction": asdict(merged_extraction),
        "document_date_candidates": DATE_PATTERN.findall(ocr_document.text),
    }


def run_baseline_extraction(config: ProjectConfig, limit: int | None = None) -> list[ExtractionResult]:
    inventory = build_inventory(config)
    results: list[ExtractionResult] = []
    for record in inventory["records"]:
        if record["predicted_type"] == "non_ttn_related":
            continue
        payload = build_document_payload(Path(record["path"]), record["predicted_type"], config)
        extraction = payload["extraction"]
        results.append(ExtractionResult(**extraction))
        if limit is not None and len(results) >= limit:
            break
    return results


def save_extractions(results: list[ExtractionResult], output_path: Path) -> Path:
    output_path.parent.mkdir(parents=True, exist_ok=True)
    lines = [json.dumps(asdict(result), ensure_ascii=False) for result in results]
    output_path.write_text("\n".join(lines) + ("\n" if lines else ""), encoding="utf-8")
    return output_path


def save_processed_texts(results: list[ExtractionResult], output_dir: Path) -> Path:
    output_dir.mkdir(parents=True, exist_ok=True)
    for result in results:
        source_path = Path(result.document_path)
        text_path = output_dir / f"{source_path.stem}.txt"
        text = "\n".join(
            [
                f"source_file: {source_path.name}",
                f"document_path: {result.document_path}",
                f"document_type: {result.document_type}",
                f"material_name: {result.material_name or ''}",
                f"quantity: {result.quantity if result.quantity is not None else ''}",
                f"unit: {result.unit or ''}",
                f"document_number: {result.document_number or ''}",
                f"confidence: {result.confidence}",
                f"extraction_method: {result.extraction_method}",
            ]
        )
        text_path.write_text(text + "\n", encoding="utf-8")
    return output_dir


def mirror_output_dir(source_dir: Path, dataset_root: Path, output_root: Path) -> Path:
    relative_dir = source_dir.relative_to(dataset_root)
    mirrored_dir = output_root
    for part in relative_dir.parts:
        mirrored_dir = mirrored_dir / f"{part}_out"
    mirrored_dir.mkdir(parents=True, exist_ok=True)
    return mirrored_dir


def process_dataset_to_json(config: ProjectConfig) -> dict:
    output_root = config.dataset_output_root
    output_root.mkdir(parents=True, exist_ok=True)

    for directory in sorted(path for path in config.dataset_root.rglob("*") if path.is_dir()):
        mirror_output_dir(directory, config.dataset_root, output_root)

    processed: list[DatasetProcessResult] = []
    supported_count = 0
    unsupported_count = 0
    processed_count = 0

    for path in iter_all_files(config):
        mirrored_dir = mirror_output_dir(path.parent, config.dataset_root, output_root)
        output_json_path = mirrored_dir / f"{path.name}.json"
        predicted_type = classify_document(path, config)
        supported = path.suffix.casefold() in config.supported_extensions

        payload: dict[str, Any] = {
            "source_path": str(path),
            "relative_path": str(path.relative_to(config.dataset_root)),
            "predicted_type": predicted_type,
            "supported": supported,
        }

        extraction_payload = None
        processing_status = "unsupported"
        if supported:
            document_payload = build_document_payload(path, predicted_type, config)
            payload.update(document_payload)
            extraction_payload = payload["extraction"]
            processing_status = "processed"
            supported_count += 1
            processed_count += 1
        else:
            unsupported_count += 1
            payload["ocr"] = None
            payload["filename_extraction"] = None
            payload["ocr_extraction"] = None
            payload["extraction"] = None

        payload["processing_status"] = processing_status
        output_json_path.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")

        processed.append(
            DatasetProcessResult(
                source_path=str(path),
                relative_path=str(path.relative_to(config.dataset_root)),
                output_json_path=str(output_json_path),
                supported=supported,
                predicted_type=predicted_type,
                processing_status=processing_status,
                extraction=extraction_payload,
            )
        )

    return {
        "dataset_root": str(config.dataset_root),
        "output_root": str(output_root),
        "total_files": len(processed),
        "supported_files": supported_count,
        "unsupported_files": unsupported_count,
        "processed_files": processed_count,
        "results": [asdict(item) for item in processed],
    }
