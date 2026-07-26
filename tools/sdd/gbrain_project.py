#!/usr/bin/env python3
"""Run the shared GBrain CLI against TossOS' project-local data home."""

from __future__ import annotations

import os
import shutil
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
GBRAIN_HOME = ROOT / ".sdd" / "gbrain-home"


def project_environment() -> dict[str, str]:
    env = os.environ.copy()
    env["GBRAIN_HOME"] = str(GBRAIN_HOME)
    return env


def main() -> int:
    executable = shutil.which("gbrain")
    if executable is None:
        print("gbrain is not installed; run make sdd-doctor", file=sys.stderr)
        return 127
    if len(sys.argv) == 1:
        print("usage: gbrain_project.py <gbrain args...>", file=sys.stderr)
        return 2
    os.execve(executable, [executable, *sys.argv[1:]], project_environment())
    return 127


if __name__ == "__main__":
    raise SystemExit(main())
