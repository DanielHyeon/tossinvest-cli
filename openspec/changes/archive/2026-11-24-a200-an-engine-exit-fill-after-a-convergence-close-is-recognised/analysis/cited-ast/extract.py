#!/usr/bin/env python3
"""a200 1.3 — 인용 AST 재추출과 사전 증거 대조.

저장소 루트에서 돈다(`python3 <이 파일> [사전증거폴더]`). 목록의 함수마다
`go run ./tools/logic-map --file F --func Fn` 을 돌려 이 폴더에 `<이름>.ast.json` 으로 쓰고,
사전 증거 폴더가 주어지면 같은 이름 파일과 JSON 전체가 같은지(source_sha256 · start · end ·
분기 수)를 표로 찍는다. 사전 증거에 없는 함수는 「신규」로 찍는다.
"""
import json
import pathlib
import subprocess
import sys

HERE = pathlib.Path(__file__).resolve().parent
ROOT = pathlib.Path(subprocess.check_output(["git", "-C", str(HERE), "rev-parse", "--show-toplevel"], text=True).strip())

# (산출물 이름, 파일, logic-map --func 값)
TARGETS = [
    ("ApplyExitFill", "internal/journal/apply_hook.go", "ApplyExitFill"),
    ("ApplyPositionAdjustment", "internal/journal/position_adjustments.go", "Journal.ApplyPositionAdjustment"),
    ("Compare", "internal/reconcile/compare.go", "Comparer.Compare"),
    ("ConvergeQuantities", "internal/reconcile/converge.go", "Converger.ConvergeQuantities"),
    ("ExitObserver.submit", "internal/app/engine/exitloop.go", "ExitObserver.submit"),
    ("HeldBySymbol", "internal/reconcile/compare.go", "HeldBySymbol"),
    ("ProjectPosition", "internal/journal/position_projection.go", "ProjectPosition"),
    ("RunOnce", "internal/app/engine/reconcileloop.go", "ReconcileDriver.RunOnce"),
    ("Tracker.Observe", "internal/reconcile/mismatch.go", "Tracker.Observe"),
    ("Tracker.Refresh", "internal/reconcile/mismatch.go", "Tracker.Refresh"),
    ("Tracker.Resolve", "internal/reconcile/mismatch.go", "Tracker.Resolve"),
    ("Tracker.Restore", "internal/reconcile/mismatch.go", "Tracker.Restore"),
    ("Tracker.enterReconcileWithBinding", "internal/reconcile/order_terminal_release.go", "Tracker.enterReconcileWithBinding"),
    ("Tracker.orderTerminalCandidate", "internal/reconcile/order_terminal_release.go", "Tracker.orderTerminalCandidate"),
    ("Tracker.persist", "internal/reconcile/mismatch.go", "Tracker.persist"),
    ("Tracker.releaseOnOrderTerminal", "internal/reconcile/order_terminal_release.go", "Tracker.releaseOnOrderTerminal"),
    ("adoptionEpoch", "internal/journal/trade_outcome_refusal.go", "adoptionEpoch"),
    ("blockFromReconcileState", "internal/reconcile/mismatch.go", "blockFromReconcileState"),
    ("blocksFor", "internal/reconcile/mismatch.go", "blocksFor"),
    ("classifyDisposition", "internal/position/position.go", "classifyDisposition"),
    ("classifyMovement", "internal/position/position.go", "classifyMovement"),
    ("closeExitStateOnAdjustmentTx", "internal/journal/position_adjustments.go", "closeExitStateOnAdjustmentTx"),
    ("enterReconcileInTx", "internal/journal/position_projection.go", "enterReconcileInTx"),
    ("enterReconcileScopeInTx", "internal/journal/position_projection.go", "enterReconcileScopeInTx"),
    ("judgeHoldings", "internal/app/engine/adoption.go", "ReconcileDriver.judgeHoldings"),
    ("orderTerminalReleaseVerdict", "internal/reconcile/order_terminal_release.go", "orderTerminalReleaseVerdict"),
    ("position.Apply", "internal/position/position.go", "Apply"),
    ("reconcileCauseFor", "internal/journal/position_projection.go", "reconcileCauseFor"),
    ("reportBlockedHoldings", "internal/app/engine/reconcile_block_report.go", "ReconcileDriver.reportBlockedHoldings"),
    ("resolveFillOrigin", "internal/journal/position_projection.go", "resolveFillOrigin"),
    ("runApplyHooks", "internal/journal/apply_hook.go", "Journal.runApplyHooks"),
    ("soleLiveMarket", "internal/reconcile/converge.go", "soleLiveMarket"),
    ("symbolsInDispute", "internal/reconcile/mismatch.go", "symbolsInDispute"),
    ("writeProjection", "internal/journal/position_projection.go", "writeProjection"),
    # 1단계가 새로 인용하는 함수(사전 증거에 없음)
    ("currentInstance", "internal/journal/position_projection.go", "currentInstance"),
    ("exitContextForFill", "internal/journal/exit_state.go", "exitContextForFill"),
    ("computeTradeOutcome", "internal/journal/trade_outcomes.go", "computeTradeOutcome"),
    ("roundTripLegs", "internal/journal/trade_outcomes.go", "roundTripLegs"),
    ("adoptedRoundTripLegs", "internal/journal/trade_outcomes.go", "adoptedRoundTripLegs"),
    ("markExitCompletedTx", "internal/journal/exit_state.go", "markExitCompletedTx"),
    ("Journal.AttachExitIntent", "internal/journal/apply_hook.go", "Journal.AttachExitIntent"),
    ("Journal.ResolveExitProposal", "internal/journal/apply_hook.go", "Journal.ResolveExitProposal"),
    ("clearExitProposalTx", "internal/journal/apply_hook.go", "clearExitProposalTx"),
    ("ApplyTx.ResolvePending", "internal/journal/apply_hook.go", "ApplyTx.ResolvePending"),
]


def main() -> int:
    prior = pathlib.Path(sys.argv[1]) if len(sys.argv) > 1 else None
    failed = 0
    print("| 함수 | 파일 | 줄(현재) | 분기 | source_sha256 앞 12 | 사전 증거 대조 |")
    print("|---|---|---|---|---|---|")
    for name, file, func in TARGETS:
        proc = subprocess.run(["go", "run", "./tools/logic-map", "--file", file, "--func", func],
                              cwd=ROOT, capture_output=True, text=True)
        if proc.returncode != 0:
            # receiver 접두사가 맞지 않으면 접두사 없이 한 번 더 — 실패는 그대로 보고
            print(f"| {name} | {file} | — | — | — | 추출 실패: {proc.stderr.strip()[:120]} |")
            failed += 1
            continue
        out = HERE / f"{name}.ast.json"
        out.write_text(proc.stdout)
        cur = json.loads(proc.stdout)
        verdict = "신규"
        if prior is not None and (prior / f"{name}.ast.json").exists():
            old = json.loads((prior / f"{name}.ast.json").read_text())
            if old == cur:
                verdict = "동일(JSON 전체)"
            else:
                diffs = [k for k in sorted(set(old) | set(cur)) if old.get(k) != cur.get(k)]
                verdict = "다름: " + ",".join(diffs)
        print(f"| {name} | {file} | {cur['start']}-{cur['end']} | {len(cur.get('branches') or [])} | "
              f"{cur['source_sha256'][:12]} | {verdict} |")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
