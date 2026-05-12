from __future__ import annotations

import argparse
import json
from dataclasses import asdict
from pathlib import Path

from .config import ProjectConfig
from .extraction import (
    process_dataset_to_json,
    run_baseline_extraction,
    save_extractions,
    save_processed_texts,
)
from .inventory import build_inventory, save_inventory


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Baseline CV tooling for TTN document scanning.")
    parser.add_argument(
        "--config",
        default="config.yaml",
        help="Path to project config YAML. Defaults to agent/config.yaml.",
    )
    subparsers = parser.add_subparsers(dest="command", required=True)

    subparsers.add_parser("inventory", help="Scan dataset and build inventory JSON.")

    extract_parser = subparsers.add_parser(
        "extract-baseline",
        help="Run heuristic extraction baseline using document filenames.",
    )
    extract_parser.add_argument("--limit", type=int, default=None, help="Optional max document count.")
    file_parser = subparsers.add_parser(
        "process-file",
        help="Process a single document and print extracted JSON payload to stdout.",
    )
    file_parser.add_argument("--input", required=True, help="Path to a source PDF/JPG/JPEG/PNG file.")
    file_parser.add_argument(
        "--predicted-type",
        default="ttn",
        help="Optional document type hint. Defaults to ttn.",
    )
    subparsers.add_parser(
        "process-dataset",
        help="Mirror the full dataset directory tree into output/*_out and write one processed JSON per source file.",
    )

    return parser


def main() -> None:
    parser = build_parser()
    args = parser.parse_args()

    config = ProjectConfig.from_file(Path(args.config))

    if args.command == "inventory":
        payload = build_inventory(config)
        path = save_inventory(payload, config.inventory_output)
        print(json.dumps({"saved_to": str(path), "stats": payload["stats"]}, ensure_ascii=False, indent=2))
        return

    if args.command == "extract-baseline":
        results = run_baseline_extraction(config, limit=args.limit)
        path = save_extractions(results, config.extraction_output)
        text_dir = save_processed_texts(results, config.text_output_dir)
        preview = [asdict(result) for result in results[:3]]
        print(
            json.dumps(
                {
                    "saved_to": str(path),
                    "text_output_dir": str(text_dir),
                    "documents": len(results),
                    "preview": preview,
                },
                ensure_ascii=False,
                indent=2,
            )
        )
        return

    if args.command == "process-dataset":
        payload = process_dataset_to_json(config)
        summary = {
            "dataset_root": payload["dataset_root"],
            "output_root": payload["output_root"],
            "total_files": payload["total_files"],
            "supported_files": payload["supported_files"],
            "unsupported_files": payload["unsupported_files"],
            "processed_files": payload["processed_files"],
        }
        print(json.dumps(summary, ensure_ascii=False, indent=2))
        return

    if args.command == "process-file":
        source_path = Path(args.input).resolve()
        payload = {
            "source_path": str(source_path),
            **build_document_payload(source_path, args.predicted_type, config),
        }
        print(json.dumps(payload, ensure_ascii=False, indent=2))
        return

    parser.error(f"Unsupported command: {args.command}")


if __name__ == "__main__":
    main()
