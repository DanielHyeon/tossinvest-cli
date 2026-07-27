#!/usr/bin/env python3
"""Persist the immutable git base used by Function Logic Map completion checks."""

from __future__ import annotations

import argparse
import re
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]


def capture(change: str, root: Path = ROOT) -> Path:
    if not re.fullmatch(r"[a-z0-9][a-z0-9-]*", change):
        raise ValueError(f"invalid OpenSpec change id: {change!r}")
    change_dir = root / "openspec" / "changes" / change
    if not (change_dir / "proposal.md").is_file():
        raise ValueError(f"OpenSpec change does not exist: {change}")
    target = change_dir / "base-commit.txt"
    if target.exists():
        raise ValueError(f"change base already captured: {target}")
    process = subprocess.run(
        ["git", "rev-parse", "HEAD"],
        cwd=root,
        capture_output=True,
        text=True,
        check=True,
    )
    commit = process.stdout.strip()
    if not re.fullmatch(r"[0-9a-f]{40}", commit):
        raise ValueError("git did not return a full commit id")
    target.write_text(commit + "\n", encoding="utf-8")
    return target


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--change", required=True)
    parser.add_argument("--root", default=str(ROOT))
    args = parser.parse_args()
    try:
        path = capture(args.change, Path(args.root))
    except (OSError, ValueError, subprocess.SubprocessError) as exc:
        print(f"[change-base] {exc}")
        return 1
    print(f"[change-base] captured {path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
