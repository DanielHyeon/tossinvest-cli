#!/usr/bin/env python3
"""Best-effort local-file ingest into a TossOS-namespaced Create Context Graph run."""

from __future__ import annotations

import base64
import hashlib
import json
import os
import shutil
import subprocess
import sys
import tempfile
import time
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

import collect_events as collector

ROOT = Path(__file__).resolve().parents[2]
CHECKPOINTS = ROOT / ".sdd" / "history" / "checkpoints"
BATCH_SIZE = 100
RUNS = ROOT / "tmp" / "create-context-graph" / "runs"


def checkpoint_path(source: str, channel: str = "http") -> Path:
    if channel == "http":
        return CHECKPOINTS / f"context-graph-{source}.json"
    return CHECKPOINTS / f"context-graph-{channel}-{source}.json"


def read_new_events(path: Path, offset: int) -> tuple[list[dict], int]:
    if not path.exists():
        return [], 0
    size = path.stat().st_size
    if size < offset:
        offset = 0
    events: list[dict] = []
    consumed = offset
    with path.open("rb") as handle:
        handle.seek(offset)
        for raw in handle:
            if not raw.endswith(b"\n"):
                break
            consumed += len(raw)
            try:
                value = json.loads(raw.decode("utf-8"))
            except (UnicodeDecodeError, ValueError):
                continue
            masked = collector.mask_event(value)
            if masked is not None:
                events.append(masked)
    return events, consumed


def load_offset(source: str, channel: str = "http") -> int:
    try:
        return int(
            json.loads(checkpoint_path(source, channel).read_text(encoding="utf-8"))[
                "offset"
            ]
        )
    except (OSError, ValueError, KeyError, TypeError):
        return 0


def save_offset(source: str, offset: int, channel: str = "http") -> None:
    path = checkpoint_path(source, channel)
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_suffix(".tmp")
    temporary.write_text(json.dumps({"offset": offset}) + "\n", encoding="utf-8")
    temporary.replace(path)


def ingest_events_http(source: str, events: list[dict]) -> None:
    """Guarantee a project/source namespace even when CCG's extractor is unavailable."""
    username = os.environ["NEO4J_USERNAME"]
    password = os.environ["NEO4J_PASSWORD"]
    endpoint = os.environ.get("NEO4J_HTTP_URI", "http://localhost:7474").rstrip("/")
    statements = []
    source_name = f"tossos-{source}"
    for event in events:
        canonical = json.dumps(event, ensure_ascii=False, sort_keys=True)
        digest = hashlib.sha256(f"{source_name}\0{canonical}".encode()).hexdigest()
        statements.append(
            {
                "statement": (
                    "MERGE (e:SDDEvent {event_hash: $hash, source: $source}) "
                    "SET e.event_type = $type, "
                    "e.summary = $summary, e.occurred_at = $ts"
                ),
                "parameters": {
                    "hash": digest,
                    "source": source_name,
                    "type": event.get("type", ""),
                    "summary": event.get("summary", ""),
                    "ts": event.get("ts", ""),
                },
            }
        )
    if not statements:
        return
    token = base64.b64encode(f"{username}:{password}".encode()).decode()
    for start in range(0, len(statements), BATCH_SIZE):
        request = urllib.request.Request(
            f"{endpoint}/db/neo4j/tx/commit",
            data=json.dumps({"statements": statements[start : start + BATCH_SIZE]}).encode(),
            headers={"Content-Type": "application/json", "Authorization": f"Basic {token}"},
        )
        with urllib.request.urlopen(request, timeout=10) as response:
            result = json.load(response)
        if result.get("errors"):
            raise RuntimeError(result["errors"][0].get("message", "Neo4j ingest failed"))


def cleanup_stale_runs(runs: Path = RUNS, max_age_seconds: int = 3600) -> None:
    if not runs.exists():
        return
    cutoff = time.time() - max_age_seconds
    for path in runs.glob(".tossos-*"):
        try:
            if path.is_dir() and path.stat().st_mtime < cutoff:
                shutil.rmtree(path)
        except OSError:
            continue


def sync(source: str, path: Path) -> int:
    binary = shutil.which("create-context-graph")
    required = ("NEO4J_USERNAME", "NEO4J_PASSWORD")
    if not binary or not path.exists() or any(not os.environ.get(name) for name in required):
        return 0
    events, next_offset = read_new_events(path, load_offset(source, "http"))
    ccg_events, next_ccg_offset = read_new_events(path, load_offset(source, "ccg"))
    if not events and not ccg_events:
        return 0
    runs = ROOT / "tmp" / "create-context-graph" / "runs"
    runs.mkdir(parents=True, exist_ok=True, mode=0o700)
    runs.chmod(0o700)
    cleanup_stale_runs(runs)
    if ccg_events:
        stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
        ccg_ok = False
        with tempfile.TemporaryDirectory(
            prefix=f".tossos-{source}-{stamp}-",
            dir=runs,
        ) as temporary:
            workspace = Path(temporary)
            workspace.chmod(0o700)
            data_path = workspace / "events.jsonl"
            data_path.write_text(
                "".join(
                    json.dumps(event, ensure_ascii=False, sort_keys=True) + "\n"
                    for event in ccg_events
                ),
                encoding="utf-8",
            )
            data_path.chmod(0o600)
            output = workspace / "output"
            output.mkdir(mode=0o700)
            try:
                # CCG is a credential-free local-file adapter/validator.
                # Namespaced Neo4j persistence is performed below through HTTP.
                completed = subprocess.run(
                    [
                        binary,
                        f"tossos-{source}",
                        "--domain",
                        "software-engineering",
                        "--connector",
                        "local-file",
                        "--local-file-path",
                        str(data_path),
                        "--dry-run",
                        "--output-dir",
                        str(output),
                    ],
                    cwd=ROOT,
                    timeout=120,
                    check=False,
                    stdout=subprocess.DEVNULL,
                    stderr=subprocess.DEVNULL,
                )
                ccg_ok = completed.returncode == 0
            except Exception as exc:
                print(f"[context-graph] CCG adapter warning: {exc}", file=sys.stderr)
            if not ccg_ok:
                print(
                    "[context-graph] CCG adapter failed; its checkpoint was not advanced",
                    file=sys.stderr,
                )
        if ccg_ok:
            save_offset(source, next_ccg_offset, "ccg")
    if events:
        try:
            ingest_events_http(source, events)
        except Exception as exc:
            print(f"[context-graph] Neo4j ingest warning: {exc}", file=sys.stderr)
            return 0
        save_offset(source, next_offset, "http")
    return 0
