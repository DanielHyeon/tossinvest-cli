#!/usr/bin/env python3
"""a130 변이 하네스 — 사본에서만 변이하고, 무변이 대조군이 GREEN 이어야 판을 돎(a129 하네스와 같은 모양).

사용: python3 a130_mutate.py
- 저장소 루트는 이 파일 위치에서 유도함(절대경로 하드코딩 없음).
- 사본 = `git archive HEAD` + 워킹트리의 a130 파일 덮어쓰기. 사본 경로에 pid 를 붙임.
- 각 변이: 원문 조각이 사본 파일에 **정확히 한 번** 있어야 함(닿음 단언) → 바꿈 → 대상 시험 → 시험 실패면 CAUGHT → 원복.
  컴파일 실패는 시험이 잡은 것이 아니므로 BUILD 로 따로 셈.
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
    "internal/journal/trade_outcomes.go",
    "internal/journal/apply_hook.go",
    "internal/journal/fills.go",
    "internal/journal/trade_outcome_refusal.go",
    "internal/journal/a130_adopted_outcome_test.go",
    "internal/journal/a130_journal_copy_census_test.go",
    "internal/journal/adoption_static_test.go",
    "internal/app/engine/gateway.go",
    "internal/app/engine/engine.go",
    "internal/app/engine/trade_outcome_refusal_log.go",
    "internal/app/engine/a130_trade_outcome_refusal_log_test.go",
    "internal/console/templates_portfolio.go",
    "internal/console/a130_history_note_test.go",
]
RUNS = [
    ("./internal/journal/", "TestA130|Adopt|Outcome|Backfill|TestRealized|TestAnAnalytics|TestAdoptionID"),
    ("./internal/app/engine/", "TestA130|TestBinding|TestTheBoundExit|TestTheProjectionGuard"),
    ("./internal/console/", "TestA130|TestTheHistoryScreen"),
]
R = "internal/journal/trade_outcome_refusal.go"
T = "internal/journal/trade_outcomes.go"
MUTANTS = [
    ("M1 the post-adoption increase check is removed", R,
     "\t\t\tif createdAt.After(adoptedAt) {\n\t\t\t\tincreased = true",
     "\t\t\tif false && createdAt.After(adoptedAt) {\n\t\t\t\tincreased = true"),
    ("M2 the post-adoption disposal check is removed", R,
     "\t\t\tif !createdAt.Before(adoptedAt) {\n\t\t\t\tdisposed = true",
     "\t\t\tif false && !createdAt.Before(adoptedAt) {\n\t\t\t\tdisposed = true"),
    ("M3 the re-adoption check is removed", R,
     "\tif found {\n\t\treturn OutcomeRefusalReadopted", "\tif false && found {\n\t\treturn OutcomeRefusalReadopted"),
    ("M4 a same-second increase counts as post-adoption", R,
     "\t\t\tif createdAt.After(adoptedAt) {\n\t\t\t\tincreased = true",
     "\t\t\tif !createdAt.Before(adoptedAt) {\n\t\t\t\tincreased = true"),
    ("M5 a same-second decrease is not post-adoption", R,
     "\t\t\tif !createdAt.Before(adoptedAt) {\n\t\t\t\tdisposed = true",
     "\t\t\tif createdAt.After(adoptedAt) {\n\t\t\t\tdisposed = true"),
    ("M6 the adoption-epoch refusal is not consulted", T,
     "if reason := adoptionEpochRefusal(ctx, r, positionID); reason != \"\" {",
     "if reason := adoptionEpochRefusal(ctx, r, positionID); false && reason != \"\" {"),
    ("M7 the sell-quantity comparison is reverted to always-equal", T,
     "\tif sell.quantity.Cmp(adopted) != 0 {", "\tif sell.quantity.Cmp(sell.quantity) != 0 {"),
    ("M8 the denominator quantity is swapped for one share", T,
     "denominator := new(big.Rat).Mul(ratOf(risk), buy.quantity)",
     "denominator := new(big.Rat).Mul(ratOf(risk), new(big.Rat).SetInt64(1))"),
    ("M9 the refusal is not noted (silent again)", "internal/journal/apply_hook.go",
     "\t\t\ttx.noteOutcomeRefusal(ctx, where.PositionID, reason)", "\t\t\t_ = reason"),
    ("M10 the refusal is reported before the commit", "internal/journal/fills.go",
     "\tif err := tx.Commit(); err != nil {\n\t\treturn res, fmt.Errorf(\"journal: committing the fill snapshot of %s: %w\", orderID, err)\n\t}\n\t// a130: an outcome refusal is reported only for a close that committed.\n\tj.reportOutcomeRefusals(refusals)\n",
     "\tj.reportOutcomeRefusals(refusals)\n\tif err := tx.Commit(); err != nil {\n\t\treturn res, fmt.Errorf(\"journal: committing the fill snapshot of %s: %w\", orderID, err)\n\t}\n"),
    ("M11 the sink panic is not recovered", R,
     "\t\t\tdefer func() { _ = recover() }()\n\t\t\tsink(r)", "\t\t\tsink(r)"),
    ("M12 the engine sink writes on the caller's goroutine", "internal/app/engine/trade_outcome_refusal_log.go",
     "\t\tgo func() {", "\t\tfunc() {"),
    ("M13 NewContext binds without the sink", "internal/app/engine/engine.go",
     "bindApplyHooks(jrn, tradeOutcomeRefusalSink(opts.Logger))", "bindApplyHooks(jrn)"),
    ("M14 the handle never carries the sink", "internal/journal/apply_hook.go",
     ", outcomeRefused: hooks.OutcomeRefused}", "}"),
    ("M15 the engine line gains a position id field", "internal/app/engine/trade_outcome_refusal_log.go",
     "\t\t\t\t\"reason\", string(r.Reason))", "\t\t\t\t\"reason\", string(r.Reason), \"position_id\", r.Symbol)"),
    ("M16 sell_unpriced is reported as a mismatch", T,
     "\t\treturn buy, sell, OutcomeRefusalSellUnpriced", "\t\treturn buy, sell, OutcomeRefusalSellQuantityMismatch"),
]


def run(cmd: list[str], cwd: Path) -> tuple[int, str]:
    env = dict(os.environ, GOFLAGS="-trimpath")
    p = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, env=env)
    return p.returncode, p.stdout + p.stderr


def tests(copy: Path) -> tuple[str, str]:
    out, ok = "", True
    for pkg, pattern in RUNS:
        rc, o = run(["go", "test", "-count=1", pkg, "-run", pattern], copy)
        out += o
        ok = ok and rc == 0
    failed = [l.strip() for l in out.splitlines() if l.strip().startswith("--- FAIL") or "panic:" in l]
    if "[build failed]" in out or "[setup failed]" in out:
        return "BUILD", "; ".join([l for l in out.splitlines() if l.startswith("#") or ".go:" in l][:4])
    return ("GREEN" if ok else "RED"), "; ".join(failed[:6])


def main() -> int:
    head = subprocess.run(["git", "-C", str(ROOT), "rev-parse", "--short", "HEAD"],
                          capture_output=True, text=True, check=True).stdout.strip()
    copy = Path(tempfile.mkdtemp(prefix=f"a130-mutate-{os.getpid()}-"))
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
        print(f"# a130 mutation — HEAD {head} + working-tree a130 files · copy {repo}", flush=True)
        state, why = tests(repo)
        print(f"CONTROL {state} {why}", flush=True)
        if state != "GREEN":
            print("# control is not GREEN — no mutant is judged")
            return 2
        caught = 0
        for name, rel, old, new in MUTANTS:
            path = repo / rel
            text = path.read_text()
            n = text.count(old)
            if n != 1:
                print(f"UNREACHED {name} (fragment occurs {n} times)", flush=True)
                continue
            path.write_text(text.replace(old, new))
            state, why = tests(repo)
            path.write_text(text)
            verdict = {"RED": "CAUGHT", "GREEN": "SURVIVED", "BUILD": "BUILD-FAILED"}[state]
            caught += verdict == "CAUGHT"
            print(f"{verdict} {name} :: {why}", flush=True)
        print(f"# {caught}/{len(MUTANTS)} CAUGHT")
        return 0
    finally:
        shutil.rmtree(copy, ignore_errors=True)


if __name__ == "__main__":
    sys.exit(main())
