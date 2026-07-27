#!/usr/bin/env python3
"""Coalesce advisory index and graph refreshes behind one detached worker."""

from __future__ import annotations

import argparse
import json
import os
import shutil
import subprocess
import sys
import tempfile
import time
import uuid
from pathlib import Path

import context_graph_sync

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
STATE = ROOT / ".sdd" / "history"
QUEUE = STATE / "refresh-queue"
LOCK = STATE / "index-refresh.lock"
EVENTS = STATE / "events"


def enqueue(source: str) -> Path:
    QUEUE.mkdir(parents=True, exist_ok=True)
    marker = QUEUE / f"{time.time_ns()}-{os.getpid()}-{uuid.uuid4().hex}.json"
    temporary = marker.with_suffix(".tmp")
    temporary.write_text(json.dumps({"source": source}) + "\n", encoding="utf-8")
    temporary.replace(marker)
    return marker


def process_start_identity(pid: int) -> str | None:
    try:
        fields = (
            Path(f"/proc/{pid}/stat")
            .read_text(encoding="utf-8")
            .rsplit(")", 1)[1]
            .split()
        )
        return fields[19]
    except (OSError, IndexError):
        try:
            os.kill(pid, 0)
        except OSError:
            return None
        return "alive"


def lock_owner() -> dict:
    try:
        value = json.loads(LOCK.read_text(encoding="utf-8"))
        return value if isinstance(value, dict) else {}
    except (OSError, ValueError):
        return {}


def owner_alive(owner: dict) -> bool:
    try:
        pid = int(owner["pid"])
    except (KeyError, TypeError, ValueError):
        return False
    current = process_start_identity(pid)
    expected = owner.get("start")
    return current is not None and (not expected or expected == current)


def write_owner(token: str, pid: int) -> None:
    value = {
        "pid": pid,
        "start": process_start_identity(pid),
        "token": token,
        "heartbeat": time.time(),
    }
    if lock_owner().get("token") != token:
        raise RuntimeError("index refresh lock ownership changed")
    with tempfile.NamedTemporaryFile(
        "w",
        encoding="utf-8",
        dir=LOCK.parent,
        prefix=".index-refresh-owner-",
        delete=False,
    ) as handle:
        handle.write(json.dumps(value, sort_keys=True) + "\n")
        temporary = Path(handle.name)
    try:
        if lock_owner().get("token") != token:
            raise RuntimeError("index refresh lock ownership changed")
        temporary.replace(LOCK)
    finally:
        temporary.unlink(missing_ok=True)


def release(token: str) -> None:
    if lock_owner().get("token") == token:
        LOCK.unlink(missing_ok=True)


def claim() -> str | None:
    STATE.mkdir(parents=True, exist_ok=True)
    token = uuid.uuid4().hex
    try:
        descriptor = os.open(LOCK, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
    except FileExistsError:
        try:
            owner = lock_owner()
            if owner_alive(owner):
                return None
            if not owner and time.time() - LOCK.stat().st_mtime <= 5:
                return None
            LOCK.unlink()
            descriptor = os.open(LOCK, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
        except (FileExistsError, FileNotFoundError, OSError):
            return None
    with os.fdopen(descriptor, "w", encoding="utf-8") as handle:
        handle.write(
            json.dumps(
                {
                    "pid": os.getpid(),
                    "start": process_start_identity(os.getpid()),
                    "token": token,
                    "heartbeat": time.time(),
                },
                sort_keys=True,
            )
            + "\n"
        )
    return token


def schedule(source: str) -> None:
    enqueue(source)
    token = claim()
    if not token:
        return
    try:
        worker = subprocess.Popen(
            [sys.executable, str(Path(__file__).resolve()), "--run", "--token", token],
            cwd=ROOT,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            stdin=subprocess.DEVNULL,
            start_new_session=True,
        )
        write_owner(token, worker.pid)
    except Exception:
        release(token)


def run_command(command: list[str], timeout: int) -> None:
    if not shutil.which(command[0]):
        return
    try:
        subprocess.run(
            command,
            cwd=ROOT,
            timeout=timeout,
            check=False,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
    except Exception:
        pass
    finally:
        try:
            LOCK.touch()
        except OSError:
            pass


def refresh(sources: set[str]) -> None:
    run_command(["codegraph", "sync", ".", "--quiet"], 300)
    run_command(["codegraphcontext", "update", ".", "--quiet"], 300)
    event_sources = sources & {"agent-saves", "commits"}
    if os.environ.get("TOSSOS_CONTEXT_GRAPH_SYNC") == "1":
        for source in sorted(event_sources):
            context_graph_sync.sync(source, EVENTS / f"{source}.jsonl")
    if os.environ.get("TOSSOS_TYPEDB_SYNC") == "1" and event_sources:
        python = ROOT / ".sdd" / ".venv" / "bin" / "python"
        if python.exists():
            run_command([str(python), str(HERE / "ingest_typedb.py")], 300)


def pending() -> list[Path]:
    return sorted(QUEUE.glob("*.json")) if QUEUE.exists() else []


def run_worker(token: str | None = None) -> int:
    if not token:
        return 2
    owner = lock_owner()
    if owner.get("token") != token:
        return 0
    write_owner(token, os.getpid())
    owns_lock = True
    try:
        while True:
            batch = pending()
            if not batch:
                release(token)
                owns_lock = False
                if pending():
                    schedule("coalesced")
                return 0
            sources: set[str] = set()
            for marker in batch:
                try:
                    sources.add(json.loads(marker.read_text(encoding="utf-8"))["source"])
                except (OSError, ValueError, KeyError, TypeError):
                    pass
            refresh(sources)
            for marker in batch:
                marker.unlink(missing_ok=True)
    finally:
        if owns_lock:
            release(token)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--enqueue")
    parser.add_argument("--run", action="store_true")
    parser.add_argument("--token")
    args = parser.parse_args()
    if args.run:
        return run_worker(args.token)
    if args.enqueue:
        schedule(args.enqueue)
        return 0
    parser.error("--enqueue SOURCE or --run is required")
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
