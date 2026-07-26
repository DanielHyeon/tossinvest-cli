#!/usr/bin/env python3
"""Best-effort TypeDB ingestion for the TossOS SDD control graph."""

from __future__ import annotations

import argparse
import ctypes
import hashlib
import importlib.util
import json
import os
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
EVENTS = ROOT / ".sdd" / "history" / "events"
SCHEMA = ROOT / ".sdd" / "control-graph" / "typedb" / "schema.tql"
CHECKPOINT = ROOT / ".sdd" / "history" / "checkpoints" / "typedb.json"
ENDPOINT = os.environ.get("TYPEDB_ENDPOINT", "localhost:1729")
DATABASE = os.environ.get("TYPEDB_DATABASE", "tossos_sdd")
BATCH_SIZE = 100

spec = importlib.util.spec_from_file_location("collect_events", HERE / "collect_events.py")
collector = importlib.util.module_from_spec(spec)
sys.modules[spec.name] = collector
spec.loader.exec_module(collector)


def preload_python_runtime() -> None:
    """uv standalone Python keeps libpython outside the dynamic loader's default path."""
    version = f"{sys.version_info.major}.{sys.version_info.minor}"
    candidates = (
        Path(sys.base_prefix) / "lib" / f"libpython{version}.so.1.0",
        Path(sys.base_prefix) / "lib" / f"libpython{version}.so",
    )
    for candidate in candidates:
        if candidate.exists():
            ctypes.CDLL(str(candidate), mode=ctypes.RTLD_GLOBAL)
            return


def load_events(path: Path) -> list[dict]:
    files = sorted(path.glob("*.jsonl")) if path.is_dir() else [path]
    events = []
    for file in files:
        if not file.exists():
            continue
        for line in file.read_text(encoding="utf-8").splitlines():
            try:
                value = json.loads(line)
            except ValueError:
                continue
            masked = collector.mask_event(value)
            if masked is not None:
                events.append(masked)
    return events


def load_offsets() -> dict[str, int]:
    try:
        value = json.loads(CHECKPOINT.read_text(encoding="utf-8"))
        return {str(key): int(offset) for key, offset in value.items()}
    except (OSError, ValueError, TypeError):
        return {}


def save_offsets(offsets: dict[str, int]) -> None:
    CHECKPOINT.parent.mkdir(parents=True, exist_ok=True)
    temporary = CHECKPOINT.with_suffix(".tmp")
    temporary.write_text(json.dumps(offsets, sort_keys=True) + "\n", encoding="utf-8")
    temporary.replace(CHECKPOINT)


def pending_events(path: Path, offsets: dict[str, int]) -> tuple[list[dict], dict[str, int]]:
    files = sorted(path.glob("*.jsonl")) if path.is_dir() else [path]
    events: list[dict] = []
    updated = dict(offsets)
    for file in files:
        if not file.exists():
            continue
        key = str(file.resolve())
        offset = offsets.get(key, 0)
        if file.stat().st_size < offset:
            offset = 0
        consumed = offset
        with file.open("rb") as handle:
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
        updated[key] = consumed
    return events, updated


def stable_hash(event: dict) -> str:
    payload = json.dumps(event, ensure_ascii=False, sort_keys=True).encode()
    return hashlib.sha256(payload).hexdigest()


def quoted(value: object) -> str:
    text = str(value).replace("\\", "\\\\").replace('"', '\\"').replace("\n", " ")
    return f'"{text}"'


def connect(endpoint: str):
    preload_python_runtime()
    from typedb.driver import Credentials, DriverOptions, DriverTlsConfig, TypeDB

    username = os.environ.get("TYPEDB_USER")
    password = os.environ.get("TYPEDB_PASSWORD")
    if not username or not password:
        raise RuntimeError("TYPEDB_USER and TYPEDB_PASSWORD are required")
    options = DriverOptions(DriverTlsConfig.disabled(), request_timeout_millis=3000)
    return TypeDB.driver(
        endpoint,
        Credentials(username, password),
        options,
    )


def ingest(path: Path = EVENTS, endpoint: str = ENDPOINT, database: str = DATABASE) -> int:
    offsets = load_offsets()
    events, updated_offsets = pending_events(path, offsets)
    if not events:
        print("[typedb] best-effort inserted: 0")
        return 0
    try:
        preload_python_runtime()
        from typedb.driver import TransactionType

        driver = connect(endpoint)
    except Exception as exc:
        print(f"[typedb] unavailable, local ledger retained: {exc}", file=sys.stderr)
        return 0
    inserted = 0
    try:
        if not driver.databases.contains(database):
            driver.databases.create(database)
            if SCHEMA.exists():
                with driver.transaction(database, TransactionType.SCHEMA) as tx:
                    tx.query(SCHEMA.read_text(encoding="utf-8")).resolve()
                    tx.commit()
        for start in range(0, len(events), BATCH_SIZE):
            with driver.transaction(database, TransactionType.WRITE) as tx:
                for event in events[start : start + BATCH_SIZE]:
                    digest = stable_hash(event)
                    event_type = quoted(event.get("type", ""))
                    summary = quoted(event.get("summary", ""))
                    query = (
                        f"match not {{ $existing isa event, has event-hash {quoted(digest)}; }}; "
                        f"insert $event isa event, has event-hash {quoted(digest)}, "
                        f"has event-type {event_type}, has summary {summary};"
                    )
                    # Fail the transaction as a unit. Offsets are persisted only
                    # after every batch commits, so a failed event is retried.
                    tx.query(query).resolve()
                    inserted += 1
                tx.commit()
        save_offsets(updated_offsets)
    except Exception as exc:
        print(f"[typedb] ingest warning: {exc}", file=sys.stderr)
    finally:
        try:
            driver.close()
        except Exception:
            pass
    print(f"[typedb] best-effort inserted: {inserted}")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--events", default=str(EVENTS))
    parser.add_argument("--endpoint", default=ENDPOINT)
    parser.add_argument("--database", default=DATABASE)
    args = parser.parse_args()
    return ingest(Path(args.events), args.endpoint, args.database)


if __name__ == "__main__":
    raise SystemExit(main())
