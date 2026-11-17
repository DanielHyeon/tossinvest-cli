# 코드 리뷰 2회차(최종) 원문 — Codex(gpt-6-astra), codex-cli 0.165.0, `codex exec -s read-only`, 2026-11-17, 대상 diff `code-review-2.diff`(1회차 뒤 interdiff `code-review-1-to-2.interdiff`)

## 프롬프트
```
You are an adversarial senior engineer doing the SECOND and FINAL code review of a HIGH-RISK change to a live-money automated trading engine (Go). You did not write it. Read-only: do not modify files, no network, no make/docker/tossctl. If the sandbox allows, you may run `go test -count=1 -run TestA192 ./internal/reconcile/ ./internal/journal/ ./internal/app/engine/`; otherwise review statically.

Repository: current directory (worktree; base 6277ec34e; uncommitted). Full diff: openspec/changes/a192-a-missing-order-block-is-released-by-the-orders-terminal-proof/analysis/review/code-review-2.diff ; changes since round 1: code-review-1-to-2.interdiff (same directory). Your round-1 findings: code-review-1-codex.md (same directory).
Mutation ledgers: ../mutation/mutation-run-2.txt (29/29 CAUGHT after the compile-failure classification fix) and mutation-run-3.txt (adds m30-m32; may still be running).

Round-1 dispositions by the author: #1 fixed — C12 now counts any fill event after the pre-insert watermark whose symbol is the bound symbol OR blank (unattributable legacy), regardless of account (conservative superset of design's account+symbol; legacy CTE not used because a blank symbol already holds); #2 mutation harness now classifies build failures as NOT_REACHED, invalid mutants fixed, run 2 = 29/29 behavioural CAUGHT, plus a fail_closed-without-fill interleaving test; #3 hold reasons now recorded per binding (first verdict and tx re-verdict as "tx_<reason>") and logged once per change as reconcile.block_release_held; #4 broader amendment veto left as a Manager decision.

Scope: (a) verify those fixes are correct and introduce no new defect (data races: Tracker.OrderTerminalHoldReasons takes t.mu; ReconcileDriver.terminalHoldLogged is touched only from RunOnce); (b) any remaining real defect against the invariants in design.md D8 and the spec delta (release only by cause; Terminal==nil identical to base; no deadlock; logs without account/order/quantity/price). List new guard/policy suggestions separately — they will not be implemented.

Output: numbered findings with severity (BLOCKER/MAJOR/MINOR/NIT), claim, file:line, proposed change; "real defects" vs "suggestions" separated; one-paragraph verdict. Concise.
```

## 응답
**Real defects**

1. **MINOR — Adjustment release leaves stale terminal bindings and hold diagnostics.**  
   [mismatch.go:802](/home/daniel/tossos-work/a192-missing-order-block-release/internal/reconcile/mismatch.go:802) removes released blocks but not their bindings. Binding deletion occurs only in the terminal-release path or a later candidate verdict. After adjustment release, that block no longer reaches candidate evaluation, so [OrderTerminalHoldReasons:429](/home/daniel/tossos-work/a192-missing-order-block-release/internal/reconcile/order_terminal_release.go:429) can retain `"armed"` indefinitely. The existing `TestA192ACreditWinsOverTheTerminalProof` supplies this sequence but never checks diagnostics.  
   **Change:** delete the corresponding binding for every committed tracker release; extend that test to assert no remaining hold reason. This violates D2’s cleanup contract but does not establish unsafe release.

**New guard / policy suggestions**

2. **MINOR — Record the conservative scope expansions explicitly.**  
   [reconcile_order_terminal_release.go:193](/home/daniel/tossos-work/a192-missing-order-block-release/internal/journal/reconcile_order_terminal_release.go:193) vetoes foreign-account same-symbol fills and all blank-symbol fills; [line 207](/home/daniel/tossos-work/a192-missing-order-block-release/internal/journal/reconcile_order_terminal_release.go:207) vetoes any amendment attempt regardless of scope/outcome. Both can retain otherwise eligible blocks. **Proposed disposition:** document Manager acceptance; no additional guard requested.

**Verdict:** No new BLOCKER/MAJOR implementation defect established. C12 now closes the blank-symbol escape; transaction hold reasons and the fill-free `fail_closed` interleaving test are present. Static inspection found no new unsafe release, lock inversion, log disclosure, or D8/`Terminal == nil` behavioral regression. Acceptance evidence remains incomplete: run 2 records 29/29 caught, but current-source run 3 finished **30 caught, 2 NOT_REACHED**—m11/m12 patterns did not match; the supplemental run-4 file was empty when inspected. The authorized test command failed before execution because `/tmp/go-build…` creation was denied. Thus this is a static review, not a runtime pass. No files or network were changed/accessed. `impeccable: N/A — backend-only`.
