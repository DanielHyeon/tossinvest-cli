#!/usr/bin/env python3
"""Capture the current commit synchronously, then queue non-blocking refreshes."""

from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path

import collect_events as events
import refresh_indexes

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
SCOPE = re.compile(r"^[A-Za-z]+(?:\(([^)]+)\))?!?:")


def head() -> tuple[str, str, str, str] | None:
    try:
        proc = subprocess.run(
            ["git", "log", "-1", "--format=%H%x1f%an%x1f%s%x1f%b"],
            cwd=ROOT,
            capture_output=True,
            text=True,
            timeout=10,
            check=False,
        )
        if proc.returncode:
            return None
        parts = (proc.stdout.rstrip("\n").split("\x1f", 3) + ["", "", "", ""])[:4]
        return tuple(parts)
    except Exception:
        return None


def paths(sha: str) -> list[str]:
    try:
        proc = subprocess.run(
            ["git", "diff-tree", "--no-commit-id", "--name-only", "-r", sha],
            cwd=ROOT,
            capture_output=True,
            text=True,
            timeout=10,
            check=False,
        )
        return [line for line in proc.stdout.splitlines() if line]
    except Exception:
        return []


def change_id(subject: str) -> str | None:
    match = SCOPE.match(subject.strip())
    if not match or not match.group(1):
        return None
    scope = match.group(1)
    matches = sorted((ROOT / "openspec" / "changes").glob(f"*{scope}*"))
    return matches[0].name if len(matches) == 1 else scope


def run(
    value: tuple[str, str, str, str] | None = None,
    events_dir: Path = events.EVENTS_DIR,
) -> int:
    try:
        value = head() if value is None else value
        if value is None:
            print("[commit-event] warning: cannot read committed HEAD metadata", file=sys.stderr)
            return 1
        sha, author, subject, body = value
        actor = "codex" if "codex" in body.casefold() else "claude" if "claude" in body.casefold() else author
        written = events.collect_event(
            {
                "type": "commit",
                "actor": actor,
                "summary": subject,
                "change_id": change_id(subject),
                "refs": [f"commit:{sha[:12]}"],
                "paths": paths(sha),
            },
            events_dir=events_dir,
            filename="commits.jsonl",
        )
        if not written:
            print("[commit-event] warning: local event ledger write failed", file=sys.stderr)
        refresh_indexes.schedule("commits")
    except Exception as exc:
        print(f"[commit-event] warning: {exc}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(run())
