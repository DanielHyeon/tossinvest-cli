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
    "internal/journal/trade_outcome_backfill.go",
    "internal/official/a130_intake_candle.go",
    "internal/journal/a130_adopted_outcome_test.go",
    "internal/journal/a130_trade_outcome_backfill_test.go",
    "internal/journal/a130_journal_copy_census_test.go",
    "internal/journal/adoption_static_test.go",
    "internal/app/engine/gateway.go",
    "internal/app/engine/engine.go",
    "internal/app/engine/trade_outcome_refusal_log.go",
    "internal/app/engine/a130_trade_outcome_refusal_log_test.go",
    "internal/console/templates_portfolio.go",
    "internal/console/a130_history_note_test.go",
    "tools/a130-outcome-backfill/main.go",
    "tools/a130-outcome-backfill/main_test.go",
]
RUNS = [
    ("./internal/journal/", "TestA130|Adopt|Outcome|Backfill|TestRealized|TestAnAnalytics|TestAdoptionID|TestCampaignCore"),
    ("./internal/app/engine/", "TestA130|TestBinding|TestTheBoundExit|TestTheProjectionGuard"),
    ("./internal/console/", "TestA130|TestTheHistoryScreen"),
    ("./tools/a130-outcome-backfill/", "."),
    ("./internal/official/", "TestA130"),
]
R = "internal/journal/trade_outcome_refusal.go"
T = "internal/journal/trade_outcomes.go"
B = "internal/journal/trade_outcome_backfill.go"
TOOL = "tools/a130-outcome-backfill/main.go"
MUTANTS = [
    ("M1 a post-adoption increase is not a lot", R,
     "\t\t\tif createdAt.After(adoptedAt) {\n\t\t\t\tdelta :=", "\t\t\tif false && createdAt.After(adoptedAt) {\n\t\t\t\tdelta :="),
    ("M2 the bar label is the minute the adjustment started in", R,
     "return l.CreatedAt.UTC().Truncate(time.Minute).Add(time.Minute)", "return l.CreatedAt.UTC().Truncate(time.Minute)"),
    ("M3 a bar of another minute is accepted", R,
     "if !ok || !candle.Label.Equal(lot.BarLabel()) {", "if !ok {"),
    ("M4 nil prices ignore the lots", R,
     "\tif prices == nil {\n\t\treturn quantity, notional, nil, OutcomeRefusalNeedsIntakePrice", "\tif false && prices == nil {\n\t\treturn quantity, notional, nil, OutcomeRefusalNeedsIntakePrice"),
    ("M5 a same-second increase (the fold-in) becomes a lot", R,
     "\t\t\tif createdAt.After(adoptedAt) {\n\t\t\t\tdelta :=", "\t\t\tif !createdAt.Before(adoptedAt) {\n\t\t\t\tdelta :="),
    ("M6 the post-adoption disposal check is removed", R,
     "\t\t\tif !createdAt.Before(adoptedAt) {\n\t\t\t\tdisposed = true", "\t\t\tif false && !createdAt.Before(adoptedAt) {\n\t\t\t\tdisposed = true"),
    ("M7 a same-second decrease is not post-adoption", R,
     "\t\t\tif !createdAt.Before(adoptedAt) {\n\t\t\t\tdisposed = true", "\t\t\tif createdAt.After(adoptedAt) {\n\t\t\t\tdisposed = true"),
    ("M8 the re-adoption check is removed", R,
     "\tif found {\n\t\treturn nil, OutcomeRefusalReadopted", "\tif false && found {\n\t\treturn nil, OutcomeRefusalReadopted"),
    ("M9 the sell is compared with the adoption-time quantity again", T,
     "\tif sell.quantity.Cmp(buy.quantity) != 0 {", "\tif sell.quantity.Cmp(adopted) != 0 {"),
    ("M10 the lots' notional is dropped", T,
     "buy.notional = new(big.Rat).Add(new(big.Rat).Mul(adopted, observed), addedNotional)", "buy.notional = new(big.Rat).Mul(adopted, observed); _ = addedNotional"),
    ("M11 the denominator quantity is swapped for one share", T,
     "denominator := new(big.Rat).Mul(ratOf(risk), buy.quantity)", "denominator := new(big.Rat).Mul(ratOf(risk), new(big.Rat).SetInt64(1))"),
    ("M12 the foreign-instance check is removed", R,
     "\t\tcase isAttributed && settled.Before(opened):", "\t\tcase false && isAttributed && settled.Before(opened):"),
    ("M13 the unattributed-local-sell check is removed", R,
     "\t\tcase !isAttributed && settled.After(opened) && notAfterClose(settled):", "\t\tcase false && !isAttributed && settled.After(opened) && notAfterClose(settled):"),
    ("M14 the integrity check reads the fill commit time", R,
     "\t\tSELECT f.order_id, TRIM(f.trading_day), MIN(a.settled_at)", "\t\tSELECT f.order_id, TRIM(f.trading_day), MIN(f.committed_at)"),
    ("M15 the refusal is not noted (silent again)", "internal/journal/apply_hook.go",
     "\t\t\ttx.noteOutcomeRefusal(ctx, where.PositionID, reason)", "\t\t\t_ = reason"),
    ("M16 the refusal is reported before the commit", "internal/journal/fills.go",
     "\tif err := tx.Commit(); err != nil {\n\t\treturn res, fmt.Errorf(\"journal: committing the fill snapshot of %s: %w\", orderID, err)\n\t}\n\t// a130: an outcome refusal is reported only for a close that committed.\n\tj.reportOutcomeRefusals(refusals)\n",
     "\tj.reportOutcomeRefusals(refusals)\n\tif err := tx.Commit(); err != nil {\n\t\treturn res, fmt.Errorf(\"journal: committing the fill snapshot of %s: %w\", orderID, err)\n\t}\n"),
    ("M17 the sink panic is not recovered", R,
     "\t\t\tdefer func() { _ = recover() }()\n\t\t\tsink(r)", "\t\t\tsink(r)"),
    ("M18 the engine sink writes on the caller's goroutine", "internal/app/engine/trade_outcome_refusal_log.go",
     "\t\tgo func() {", "\t\tfunc() {"),
    ("M19 NewContext binds without the sink", "internal/app/engine/engine.go",
     "bindApplyHooks(jrn, tradeOutcomeRefusalSink(opts.Logger))", "bindApplyHooks(jrn)"),
    ("M20 the handle never carries the sink", "internal/journal/apply_hook.go",
     ", outcomeRefused: hooks.OutcomeRefused}", "}"),
    ("M21 the backfill population admits engine-entered positions", B,
     "\t\t   AND p.adoption_id IS NOT NULL\n", ""),
    ("M22 the backfill population admits adjustment-closed positions", B,
     "\t\t   AND EXISTS (SELECT 1 FROM exit_events e WHERE e.position_id = p.id AND e.action = ?)\n",
     "\t\t   AND (? IS NOT NULL)\n"),
    ("M23 a moved item is written anyway", B,
     "if reason != \"\" || !sameOutcome(outcome, item.Outcome) {", "if reason != \"\" {"),
    # M24 은 동등 변이로 기록함: 쓴 분봉의 종가는 결과 값(손익 · R)을 정하고 결과 값이 토큰에 들어가므로, 분봉을 토큰에서 빼도 분봉 변경은
    # 결과 값으로 잡힘(TestA130TheBackfillWritesOnlyWithTheConfirmedToken 의 「분봉 값 변경」). 이름(Label)은 묶음에서 결정적으로 유도됨.
    ("M24 the token ignores the bars used (equivalent — see comment)", B,
     "\t\tfor _, c := range it.IntakeCandles {", "\t\tfor _, c := range it.IntakeCandles[:0] {"),
    ("M25 the tool opens the writer without the schema guard", TOOL,
     "\t\tif err := schemaGuard(ro.SchemaVersion(), journal.SchemaVersion); err != nil {", "\t\tif err := schemaGuard(journal.SchemaVersion, journal.SchemaVersion); err != nil {"),
    ("M26 sell_unpriced is reported as a mismatch", T,
     "\t\treturn buy, sell, OutcomeRefusalSellUnpriced", "\t\treturn buy, sell, OutcomeRefusalSellQuantityMismatch"),
    ("M27 the engine line gains a position id field", "internal/app/engine/trade_outcome_refusal_log.go",
     "\t\t\t\t\"reason\", string(r.Reason))", "\t\t\t\t\"reason\", string(r.Reason), \"position_id\", r.Symbol)"),
    ("M28 the tool keeps a bar from another minute", "internal/official/a130_intake_candle.go",
     "\t\tif at.UTC().Equal(label) {", "\t\tif true || at.UTC().Equal(label) {"),
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
    # --tree <dir>: 이미 조립된 트리 사본(저장소 밖)을 대상으로 함 — 사본을 다시 복사해 그 복사본에서 변이함(원본 무접촉).
    tree = sys.argv[sys.argv.index("--tree") + 1] if "--tree" in sys.argv else ""
    head = subprocess.run(["git", "-C", str(ROOT), "rev-parse", "--short", "HEAD"],
                          capture_output=True, text=True, check=True).stdout.strip()
    copy = Path(tempfile.mkdtemp(prefix=f"a130-mutate-{os.getpid()}-"))
    try:
        repo = copy / "repo"
        if tree:
            shutil.copytree(tree, repo, symlinks=True)
            print(f"# a130 mutation — tree copy of {tree} · copy {repo}", flush=True)
        else:
            archive = copy / "head.tar"
            subprocess.run(["git", "-C", str(ROOT), "archive", "--format=tar", "-o", str(archive), "HEAD"], check=True)
            with tarfile.open(archive) as tf:
                tf.extractall(repo, filter="data")
            archive.unlink()
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
