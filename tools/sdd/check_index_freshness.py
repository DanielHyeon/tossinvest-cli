#!/usr/bin/env python3
"""Bind hard-evidence index freshness to the exact repository worktree."""

from __future__ import annotations

import argparse
import hashlib
import json
import shutil
import subprocess
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
STATE = ROOT / ".sdd" / "index-state.json"
HARD_INDEX = "codegraph"
ADVISORY_INDEXES = ("codegraphcontext", "gbrain")


def repository_fingerprint(root: Path = ROOT) -> str:
    process = subprocess.run(
        ["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"],
        cwd=root,
        capture_output=True,
        check=True,
    )
    paths = sorted(
        part.decode("utf-8", errors="surrogateescape")
        for part in process.stdout.split(b"\0")
        if part
    )
    digest = hashlib.sha256()
    for relative in paths:
        path = root / relative
        if not path.is_file():
            continue
        digest.update(relative.encode("utf-8", errors="surrogateescape"))
        digest.update(b"\0")
        digest.update(path.read_bytes())
        digest.update(b"\0")
    return digest.hexdigest()


def load_state(path: Path = STATE) -> dict:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
        return value if isinstance(value, dict) else {}
    except (OSError, ValueError):
        return {}


def record_index_state(
    indexes: set[str],
    root: Path = ROOT,
    state_path: Path | None = None,
) -> None:
    if not indexes:
        return
    path = state_path or root / ".sdd" / "index-state.json"
    state = load_state(path)
    fingerprint = repository_fingerprint(root)
    head = subprocess.run(
        ["git", "rev-parse", "HEAD"],
        cwd=root,
        capture_output=True,
        text=True,
        check=True,
    ).stdout.strip()
    records = state.setdefault("indexes", {})
    for name in sorted(indexes):
        records[name] = {
            "fingerprint": fingerprint,
            "head": head,
            "synced_at": datetime.now(timezone.utc).isoformat(),
        }
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_suffix(".tmp")
    temporary.write_text(json.dumps(state, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    temporary.replace(path)


def codegraph_index_status(root: Path = ROOT) -> tuple[bool, str]:
    database = root / ".codegraph" / "codegraph.db"
    if not database.is_file() or database.stat().st_size == 0:
        return False, ".codegraph/codegraph.db is missing or empty"
    binary = shutil.which("codegraph")
    if not binary:
        return False, "codegraph executable is missing"
    try:
        process = subprocess.run(
            [binary, "status", "."],
            cwd=root,
            capture_output=True,
            text=True,
            timeout=15,
            check=False,
        )
    except Exception as exc:
        return False, f"CodeGraph status probe failed: {exc}"
    output = process.stdout + process.stderr
    if process.returncode or "Index is up to date" not in output:
        return False, "CodeGraph status did not confirm an up-to-date index"
    return True, "CodeGraph database and status probe are healthy"


def check(root: Path = ROOT, state_path: Path | None = None) -> tuple[list[str], list[str]]:
    path = state_path or root / ".sdd" / "index-state.json"
    state = load_state(path).get("indexes", {})
    current = repository_fingerprint(root)
    errors: list[str] = []
    warnings: list[str] = []
    hard = state.get(HARD_INDEX, {})
    if hard.get("fingerprint") != current:
        errors.append("CodeGraph hard-evidence index is missing or stale; run `make sdd-sync`")
    index_ok, detail = codegraph_index_status(root)
    if not index_ok:
        errors.append(f"CodeGraph hard-evidence index is unusable: {detail}")
    for name in ADVISORY_INDEXES:
        if state.get(name, {}).get("fingerprint") != current:
            warnings.append(f"{name} advisory index is missing or stale")
    return errors, warnings


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", default=str(ROOT))
    args = parser.parse_args()
    errors, warnings = check(Path(args.root))
    for warning in warnings:
        print(f"[index-freshness] WARN: {warning}")
    for error in errors:
        print(f"[index-freshness] FAIL: {error}")
    if not errors:
        print("[index-freshness] CodeGraph hard-evidence index matches the worktree")
    return 1 if errors else 0


if __name__ == "__main__":
    raise SystemExit(main())
