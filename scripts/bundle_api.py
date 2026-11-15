#!/usr/bin/env python3
"""
Bundle decomposed OpenAPI/Swagger files into a single specification file.

Usage:
    python3 scripts/bundle_api.py
    python3 scripts/bundle_api.py --input api/src/openapi.json --output api/services.openapi.json
    python3 scripts/bundle_api.py --watch
"""

import argparse
import json
import os
import sys
import time
from pathlib import Path


def bundle(input_path: Path) -> dict:
    base_dir = input_path.parent
    with open(input_path, "r", encoding="utf-8") as f:
        root = json.load(f)

    def resolve_ref(ref_path: str):
        target = (base_dir / ref_path).resolve()
        if not target.exists():
            raise FileNotFoundError(f"Referenced file not found: {target} (from {ref_path})")
        with open(target, "r", encoding="utf-8") as rf:
            return json.load(rf)

    bundled = {
        "openapi": root.get("openapi", "3.0.3"),
        "info": root.get("info", {}),
        "servers": root.get("servers", []),
        "tags": root.get("tags", []),
        "paths": {},
        "components": {}
    }

    # Resolve paths
    for path, ref_obj in root.get("paths", {}).items():
        if isinstance(ref_obj, dict) and "$ref" in ref_obj:
            bundled["paths"][path] = resolve_ref(ref_obj["$ref"])
        else:
            bundled["paths"][path] = ref_obj

    # Resolve components
    for comp_type, comp_map in root.get("components", {}).items():
        bundled["components"][comp_type] = {}
        if isinstance(comp_map, dict):
            for name, ref_obj in comp_map.items():
                if isinstance(ref_obj, dict) and "$ref" in ref_obj:
                    bundled["components"][comp_type][name] = resolve_ref(ref_obj["$ref"])
                else:
                    bundled["components"][comp_type][name] = ref_obj
        else:
            bundled["components"][comp_type] = comp_map

    return bundled


def build(input_path: Path, output_path: Path) -> None:
    output_path.parent.mkdir(parents=True, exist_ok=True)
    spec = bundle(input_path)
    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(spec, f, ensure_ascii=False, indent=2)
        f.write("\n")
    print(f"✓ API successfully bundled: {output_path}")


def watch(input_path: Path, output_path: Path) -> None:
    base_dir = input_path.parent
    print(f"Watching for changes in {base_dir}... Press Ctrl+C to stop.")
    
    def get_mtimes():
        mtimes = {}
        for p in base_dir.rglob("*.json"):
            try:
                mtimes[p] = p.stat().st_mtime
            except OSError:
                pass
        return mtimes

    last_mtimes = get_mtimes()
    build(input_path, output_path)

    try:
        while True:
            time.sleep(1)
            current_mtimes = get_mtimes()
            if current_mtimes != last_mtimes:
                print("Change detected, rebuilding...")
                last_mtimes = current_mtimes
                try:
                    build(input_path, output_path)
                except Exception as e:
                    print(f"Error during bundling: {e}", file=sys.stderr)
    except KeyboardInterrupt:
        print("\nWatcher stopped.")


def main():
    parser = argparse.ArgumentParser(description="Bundle decomposed OpenAPI specification files.")
    parser.add_argument(
        "-i", "--input",
        type=Path,
        default=Path("api/src/openapi.json"),
        help="Path to root openapi.json (default: api/src/openapi.json)"
    )
    parser.add_argument(
        "-o", "--output",
        type=Path,
        default=Path("api/services.openapi.json"),
        help="Path to output bundled file (default: api/services.openapi.json)"
    )
    parser.add_argument(
        "-w", "--watch",
        action="store_true",
        help="Watch directory for changes and automatically rebuild"
    )

    args = parser.parse_args()

    if not args.input.exists():
        print(f"Error: Input file {args.input} not found.", file=sys.stderr)
        sys.exit(1)

    if args.watch:
        watch(args.input, args.output)
    else:
        build(args.input, args.output)


if __name__ == "__main__":
    main()
