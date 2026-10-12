#!/usr/bin/env python3
"""a129 변이 하네스 — 사본에서만 변이하고, 무변이 대조군이 GREEN 이어야 판을 돎.

사용: python3 a129_mutate.py [--keep]
- 저장소 루트는 이 파일 위치에서 유도함(절대경로 하드코딩 없음).
- 사본 = `git archive HEAD` + 워킹트리의 a129 편집 파일 덮어쓰기. 사본 경로에 pid 를 붙임(두 판이 한 사본을 쓰지 않게).
- 각 변이: 원문 조각이 사본 파일에 **정확히 한 번** 있어야 함(닿음 단언) → 바꿈 → 대상 시험 → 실패면 CAUGHT → 원복.
- 결과는 표준출력 한 줄씩(판정 · 이름 · 실패 시험). 원장으로 쓰려면 리다이렉트.
"""

from __future__ import annotations

import os
import shutil
import subprocess
import sys
import tarfile
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[5]
FILES = [
    "internal/app/engine/exitloop.go",
    "internal/app/engine/exit_order_log.go",
    "internal/app/engine/a129_exit_order_log_test.go",
    "internal/app/engine/a129_export_test.go",
    "cmd/tossctl/engine.go",
    "cmd/tossctl/a129_engine_runtime_exit_order_log_test.go",
]
ENGINE_RUN = "TestA129"
CMD_RUN = "TestA129|TestA090EngineRuntime"

MUTANTS = [
    ("M1 whitelist gains the account field", "internal/app/engine/exit_order_log.go",
     "\tlogger.Event(t, line.args()...)\n",
     "\tlogger.Event(t, append(line.args(), obs.FieldAccount, o.opts.AccountRef)...)\n"),
    ("M2 engineRuntime stops wiring ExitOrderLog", "cmd/tossctl/engine.go",
     "\t\tExitOrderLog: logger,\n", ""),
    ("M3 the nil guard on queueing and following is removed", "internal/app/engine/exit_order_log.go",
     "\tif o.opts.ExitOrderLog == nil {\n\t\treturn\n\t}\n\tif refusal {\n", "\tif refusal {\n"),
    ("M4 the write panic is not recovered", "internal/app/engine/exit_order_log.go",
     "\tdefer func() { _ = recover() }()\n", ""),
    ("M5 an unchanged cumulative fill writes a line", "internal/app/engine/exit_order_log.go",
     "case cerr == nil && grew > 0:", "case cerr == nil && grew >= 0:"),
    ("M6 the refusal latch is removed", "internal/app/engine/exit_order_log.go",
     "\t\tif o.exitOrderRefused[line.positionID] == signature {\n\t\t\treturn\n\t\t}\n", ""),
    ("M7 the token filter passes everything", "internal/app/engine/exit_order_log.go",
     "\tif s == \"\" || exitOrderTokenPattern.MatchString(s) {\n", "\tif true {\n"),
    ("M8 the end-of-cycle hook runs before the judgement", "internal/app/engine/exitloop.go",
     "\tdefer o.endExitOrderCycle(ctx)\n", "\to.endExitOrderCycle(ctx)\n"),
    ("M9 a terminal fill does not end the follow", "internal/app/engine/exit_order_log.go",
     "\t\tif rec.Terminal {\n\t\t\tdelete(o.exitFills, key)\n\t\t}\n", ""),
    ("M10 the fill lookup is unscoped", "internal/app/engine/exit_order_log.go",
     "rec, err := o.opts.Journal.LookupFillScoped(ctx, w.scope)", "rec, err := o.opts.Journal.LookupFill(ctx, id)"),
    ("M11 an unfilled terminal order writes nothing", "internal/app/engine/exit_order_log.go",
     "\t\t\to.writeExitOrder(eventExitOrderEnded, true, w.line.filled(rec))\n", ""),
    ("M12 the confirmed branch queues no line", "internal/app/engine/exitloop.go",
     "\t\to.noteExitOrder(obs.EventOrderConfirmed, false, false, true, line)\n", ""),
    ("M1b whitelist gains the account field on the warn path", "internal/app/engine/exit_order_log.go",
     "\t\tlogger.Warn(t, line.args()...)\n",
     "\t\tlogger.Warn(t, append(line.args(), obs.FieldAccount, o.opts.AccountRef)...)\n"),
    ("M13 the latch outlives its episode", "internal/app/engine/exit_order_log.go",
     "\t\tif !o.exitOrderRefusedNow[id] {\n", "\t\tif false && !o.exitOrderRefusedNow[id] {\n"),
    ("M14 an endless follow never expires", "internal/app/engine/exit_order_log.go",
     "if now.Sub(w.since) >= exitFillFollowLimit {", "if false && now.Sub(w.since) >= exitFillFollowLimit {"),
    ("M15 B2 swallows a failed release", "internal/app/engine/exitloop.go",
     "line.because(exitOrderFloorZero))\n\t\treturn releaseErr\n", "line.because(exitOrderFloorZero))\n\t\t_ = releaseErr\n\t\treturn nil\n"),
    ("M16 B3 swallows a failed release", "internal/app/engine/exitloop.go",
     "line.because(exitOrderGuardianRefused))\n\t\treturn releaseErr\n", "line.because(exitOrderGuardianRefused))\n\t\t_ = releaseErr\n\t\treturn nil\n"),
    ("M17 B10 swallows a failed release", "internal/app/engine/exitloop.go",
     "\t\to.noteExitOrder(obs.EventOrderNotDispatched, true, true, false, line)\n\t\treturn releaseErr\n",
     "\t\to.noteExitOrder(obs.EventOrderNotDispatched, true, true, false, line)\n\t\t_ = releaseErr\n\t\treturn nil\n"),
    ("M18 B13 swallows a failed release", "internal/app/engine/exitloop.go",
     "\t\to.noteExitOrder(refusal, true, true, false, line)\n\t\treturn releaseErr\n",
     "\t\to.noteExitOrder(refusal, true, true, false, line)\n\t\t_ = releaseErr\n\t\treturn nil\n"),
]


def run(cmd: list[str], cwd: Path) -> tuple[int, str]:
    env = dict(os.environ, GOFLAGS="-trimpath")
    p = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, env=env)
    return p.returncode, p.stdout + p.stderr


def tests(copy: Path) -> tuple[str, str]:
    """GREEN · RED · BUILD 중 하나. 컴파일 실패는 시험이 잡은 것이 아니므로 CAUGHT 로 세지 않음."""
    rc1, out1 = run(["go", "test", "-count=1", "./internal/app/engine/", "-run", ENGINE_RUN], copy)
    rc2, out2 = run(["go", "test", "-count=1", "./cmd/tossctl/", "-run", CMD_RUN], copy)
    out = out1 + out2
    failed = [l.strip() for l in out.splitlines() if l.strip().startswith("--- FAIL") or "panic:" in l
              or "build failed" in l or l.startswith("#")]
    if "[build failed]" in out or "[setup failed]" in out:
        return "BUILD", "; ".join(failed[:6])
    return ("GREEN" if rc1 == 0 and rc2 == 0 else "RED"), "; ".join(failed[:6])


def main() -> int:
    keep = "--keep" in sys.argv
    head = subprocess.run(["git", "-C", str(ROOT), "rev-parse", "--short", "HEAD"],
                          capture_output=True, text=True, check=True).stdout.strip()
    copy = Path(tempfile.mkdtemp(prefix=f"a129-mutate-{os.getpid()}-"))
    try:
        archive = copy / "head.tar"
        subprocess.run(["git", "-C", str(ROOT), "archive", "--format=tar", "-o", str(archive), "HEAD"], check=True)
        with tarfile.open(archive) as tf:
            tf.extractall(copy / "repo", filter="data")
        archive.unlink()
        repo = copy / "repo"
        for rel in FILES:
            src = ROOT / rel
            if src.exists():
                (repo / rel).parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(src, repo / rel)
        print(f"# a129 mutation — HEAD {head} + working-tree a129 files · copy {repo}")
        state, why = tests(repo)
        print(f"CONTROL {state} {why}")
        if state != "GREEN":
            print("# control is not GREEN — no mutant is judged")
            return 2
        caught = 0
        for name, rel, old, new in MUTANTS:
            path = repo / rel
            text = path.read_text()
            if text.count(old) != 1:
                print(f"UNREACHED {name} — fragment count {text.count(old)} in {rel}")
                continue
            path.write_text(text.replace(old, new, 1))
            try:
                state, why = tests(repo)
            finally:
                path.write_text(text)
            verdict = {"GREEN": "SURVIVED", "RED": "CAUGHT", "BUILD": "BUILD-FAILED"}[state]
            caught += 1 if state == "RED" else 0
            print(f"{verdict} {name} — {why}")
        print(f"# {caught}/{len(MUTANTS)} caught")
        return 0 if caught == len(MUTANTS) else 1
    finally:
        if not keep:
            shutil.rmtree(copy, ignore_errors=True)


if __name__ == "__main__":
    raise SystemExit(main())
