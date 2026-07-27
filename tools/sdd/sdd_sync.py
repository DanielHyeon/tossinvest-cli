#!/usr/bin/env python3
"""Refresh TossOS hard/advisory indexes with explicit failure reporting."""

from __future__ import annotations

import argparse
import shutil
import subprocess
from pathlib import Path

from check_index_freshness import record_index_state
from gbrain_project import GBRAIN_HOME, project_environment

ROOT = Path(__file__).resolve().parents[2]


def run(
    command: list[str],
    timeout: int = 600,
    env: dict[str, str] | None = None,
) -> bool:
    print("+", " ".join(command))
    try:
        proc = subprocess.run(
            command,
            cwd=ROOT,
            timeout=timeout,
            check=False,
            env=env,
        )
        return proc.returncode == 0
    except Exception as exc:
        print(f"  failed: {exc}")
        return False


def sync(full: bool = False, include_gbrain: bool = True) -> list[str]:
    failures: list[str] = []
    successes: set[str] = set()
    if shutil.which("codegraph"):
        if not (ROOT / ".codegraph").exists():
            command = ["codegraph", "init", "."]
        elif full:
            command = ["codegraph", "index", "--force", "."]
        else:
            command = ["codegraph", "sync", "."]
        if not run(command):
            failures.append("codegraph")
        else:
            successes.add("codegraph")
    else:
        failures.append("codegraph missing")

    if shutil.which("codegraphcontext"):
        stats = subprocess.run(
            ["codegraphcontext", "stats", "."],
            cwd=ROOT,
            capture_output=True,
            text=True,
            timeout=15,
            check=False,
        )
        missing = stats.returncode != 0 or "Repository not found" in (
            stats.stdout + stats.stderr
        )
        if missing:
            command = ["codegraphcontext", "index", "."]
        elif full:
            command = ["codegraphcontext", "index", ".", "--force"]
        else:
            command = ["codegraphcontext", "update", ".", "--quiet"]
        if not run(command, timeout=300):
            failures.append("codegraphcontext")
        else:
            successes.add("codegraphcontext")
    else:
        failures.append("codegraphcontext missing")

    if include_gbrain and shutil.which("gbrain"):
        source = (ROOT / ".gbrain-source").read_text(encoding="utf-8").strip()
        gbrain_env = project_environment()
        config_file = GBRAIN_HOME / ".gbrain" / "config.json"
        if not config_file.exists() and not run(
            ["gbrain", "init", "--pglite", "--no-embedding"],
            timeout=120,
            env=gbrain_env,
        ):
            failures.append("gbrain project initialization")
            record_index_state(successes)
            return failures
        try:
            listed = subprocess.run(
                ["gbrain", "sources", "list"],
                cwd=ROOT,
                capture_output=True,
                text=True,
                timeout=15,
                check=False,
                env=gbrain_env,
            )
        except subprocess.TimeoutExpired:
            failures.append("gbrain engine locked during source probe")
            record_index_state(successes)
            return failures
        if source not in listed.stdout:
            if not run(
                ["gbrain", "sources", "add", source, "--path", str(ROOT)],
                timeout=60,
                env=gbrain_env,
            ):
                failures.append("gbrain source registration")
        command = [
            "gbrain",
            "sync",
            "--source",
            source,
            "--strategy",
            "code",
            "--no-embed",
            "--yes",
        ]
        if full:
            command.append("--full")
        if not run(command, timeout=1800, env=gbrain_env):
            failures.append("gbrain project sync")
        else:
            successes.add("gbrain")
    elif include_gbrain:
        failures.append("gbrain missing")
    record_index_state(successes)
    return failures


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--full", action="store_true")
    parser.add_argument("--no-gbrain", action="store_true")
    args = parser.parse_args()
    failures = sync(args.full, not args.no_gbrain)
    if failures:
        print("[sdd-sync] incomplete:")
        for failure in failures:
            print(f"  - {failure}")
        return 1
    print("[sdd-sync] all indexes current")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
