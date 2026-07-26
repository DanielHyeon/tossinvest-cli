#!/usr/bin/env python3
"""Masked local event ledger used by non-blocking SDD hooks."""

from __future__ import annotations

import json
import re
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
EVENTS_DIR = ROOT / ".sdd" / "history" / "events"
ALLOWED = {"type", "actor", "summary", "change_id", "refs", "paths", "ts"}
SENSITIVE_KEY = re.compile(
    r"(api[-_]?key|token|secret|password|authorization|cookie|session|private[-_]?key)",
    re.IGNORECASE,
)
EMAIL = re.compile(r"(?<![\w.+-])[\w.+-]+@[\w.-]+\.[A-Za-z]{2,}(?![\w.-])")
BEARER = re.compile(r"(?i)\bbearer\s+[A-Za-z0-9._~+/=-]+")
LONG_TOKEN = re.compile(r"\b[A-Za-z0-9_-]{32,}\b")
PEM = re.compile(r"-----BEGIN [A-Z ]*PRIVATE KEY-----")
SECRET_ASSIGNMENT = re.compile(
    r"(?i)\b(api[-_]?key|token|secret|password|authorization|cookie|session|private[-_]?key)"
    r"\s*[:=]\s*(?:\"[^\"]*\"|'[^']*'|[^\s,;]+)"
)


def mask_string(value: str) -> str:
    if PEM.search(value):
        return "[REDACTED_PRIVATE_KEY]"
    value = SECRET_ASSIGNMENT.sub(lambda match: f"{match.group(1)}=[REDACTED]", value)
    value = BEARER.sub("Bearer [REDACTED]", value)
    value = EMAIL.sub("[REDACTED_EMAIL]", value)
    value = LONG_TOKEN.sub("[REDACTED_TOKEN]", value)
    return value[:1000]


def mask_value(key: str, value):
    if SENSITIVE_KEY.search(key):
        return "[REDACTED]"
    if isinstance(value, str):
        return mask_string(value)
    if isinstance(value, list):
        return [mask_value(key, item) for item in value[:100]]
    if isinstance(value, dict):
        return {name: mask_value(str(name), item) for name, item in value.items()}
    if isinstance(value, (int, float, bool)) or value is None:
        return value
    return mask_string(str(value))


def mask_event(event: dict) -> dict | None:
    if not isinstance(event, dict) or not event.get("type"):
        return None
    masked = {key: mask_value(key, event[key]) for key in ALLOWED if key in event}
    masked.setdefault("ts", datetime.now(timezone.utc).isoformat())
    return masked


def collect_event(event: dict, events_dir: Path = EVENTS_DIR, filename: str = "events.jsonl") -> bool:
    masked = mask_event(event)
    if masked is None:
        return False
    try:
        events_dir.mkdir(parents=True, exist_ok=True)
        with (events_dir / filename).open("a", encoding="utf-8") as handle:
            handle.write(json.dumps(masked, ensure_ascii=False, sort_keys=True) + "\n")
        return True
    except OSError:
        return False
