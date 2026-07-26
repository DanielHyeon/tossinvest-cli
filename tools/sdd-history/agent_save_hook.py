#!/usr/bin/env python3
"""Capture agent edits and refresh advisory indexes without blocking the editor."""

from __future__ import annotations

import json
import sys
from pathlib import Path

import collect_events as events
import refresh_indexes

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]


def read_payload() -> dict:
    if sys.stdin.isatty():
        return {}
    try:
        value = json.load(sys.stdin)
        return value if isinstance(value, dict) else {}
    except Exception:
        return {}


def changed_paths(payload: dict) -> list[str]:
    tool_input = payload.get("tool_input") or {}
    candidates = [
        tool_input.get("file_path"),
        tool_input.get("path"),
        payload.get("file_path"),
    ]
    return sorted(
        {
            str(Path(value).resolve().relative_to(ROOT))
            for value in candidates
            if isinstance(value, str) and value and Path(value).resolve().is_relative_to(ROOT)
        }
    )


def run(payload: dict | None = None, events_dir: Path = events.EVENTS_DIR) -> int:
    try:
        payload = read_payload() if payload is None else payload
        paths = changed_paths(payload)
        written = events.collect_event(
            {
                "type": "agent-save",
                "actor": str(payload.get("agent") or "agent"),
                "summary": "agent file save",
                "paths": paths,
                "refs": [f"file:{path}" for path in paths],
            },
            events_dir=events_dir,
            filename="agent-saves.jsonl",
        )
        if not written:
            print("[sdd-hook] warning: local event ledger write failed", file=sys.stderr)
        refresh_indexes.schedule("agent-saves")
    except Exception as exc:
        print(f"[sdd-hook] warning: {exc}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(run())
