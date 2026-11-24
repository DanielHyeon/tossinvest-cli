#!/usr/bin/env python3
"""a199 1.4 — CodeGraph callers 원문 수집(저장소 루트에서 실행, worktree 색인 필요: `codegraph init .`).

출력: 같은 폴더의 codegraph-callers.txt — 심볼마다 머리(버전 · HEAD)와 CLI 원문.
"""

from __future__ import annotations

import subprocess
from pathlib import Path

HERE = Path(__file__).resolve().parent
SYMBOLS = [
    "verifyGate", "runInterlock", "refuseStartup", "logGateDecision", "NewContext", "runEngineRun",
    "engineRuntime", "ExitObserver", "ReconcileDriver", "judgeAttestation", "newAttestationWatch",
    "AllReasonCodes", "productionStrategyWorker",
]


def run(*cmd: str) -> str:
    p = subprocess.run(list(cmd), capture_output=True, text=True)
    return p.stdout + p.stderr + (f"[rc={p.returncode}]\n" if p.returncode else "")


def main() -> None:
    version = run("codegraph", "--version").strip()
    head = run("git", "rev-parse", "HEAD").strip()
    parts = []
    for s in SYMBOLS:
        parts.append(f"### codegraph {version} callers {s} (HEAD {head})\n")
        parts.append(run("codegraph", "callers", s))
        parts.append("\n")
    (HERE / "codegraph-callers.txt").write_text("".join(parts), encoding="utf-8")


if __name__ == "__main__":
    main()
