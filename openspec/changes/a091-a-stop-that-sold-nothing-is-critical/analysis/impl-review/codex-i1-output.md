VERDICT: FAIL

1. **P1 — Timing test omits reporting work covered by the budget.**  
   `internal/app/engine/a091_replay_test.go:171–173` times only the inner `Notify`. It excludes the B2 error log (`exit_stop_sold_nothing.go:113`) and recording-failure log (`:191`). D5 explicitly includes the latter in the combined budget (`design.md:136`). Those synchronous operations can exceed 750 ms while the test passes.  
   **Minimal fix:** time the entire `reportZeroFloor` call; exercise both B2 and B7, including recording failure.

2. **P2 — “Failure escalation” measurement never measures a successful escalation transaction.**  
   `internal/app/engine/a091_replay_test.go:254–262` closes the notifier’s journal. Both alert recording and `EscalateOperatingMode` therefore encounter a closed database (`internal/obs/notifier.go:429`). The reported sub-millisecond result does not cover the durable escalation transaction budgeted in `design.md:136`.  
   **Minimal fix:** inject failure specifically into alert recording while leaving operating-mode writes functional; assert durable ENTRY_BLOCKED and measure that path separately.

3. **P2 — Acknowledgement contention is not established before measurement.**  
   `internal/app/engine/a091_replay_test.go:207–222` starts a goroutine that inserts 100 rows before acquiring the acknowledgement lock. Measurements begin without a readiness barrier (`:264–270`); insert and acknowledgement errors are ignored. The test can pass without any report encountering the claimed 100-row acknowledgement lock hold. The numerical limits at `:280` and `:285` are correct, but the workload is unproven.  
   **Minimal fix:** verify backlog creation and synchronize a measured report with actual acknowledgement lock ownership.

4. **P2 — Take-profit reports falsely describe a failed stop-loss.**  
   `internal/app/engine/exit_stop_sold_nothing.go:144` unconditionally says “손절이 나가지 않았다,” including nonprotective take-profit zeros. The take-profit test checks event counts, not body wording (`a091_stop_sold_nothing_test.go:502–513`).  
   **Minimal fix:** use neutral liquidation wording for nonprotective proposals and assert the take-profit body.

5. **P3 — Replay assertions do not establish the published table.**  
   `issues.md:85–88` records exact attempt/log counts. Tests accept merely ≥3 sends for failure (`internal/app/engine/a091_replay_test.go:108`) and any positive undelivered-log count without checking “no publisher” content (`:116`). Thus materially different results still pass.  
   **Minimal fix:** synchronize delivery cycles and assert the documented counts/content, or label the table as one observed run with weaker tested guarantees.

Verification limitation: source inspection only. No tests or scratch mutations executed under the read-only, tree-only restrictions; no surviving mutation or reproduced timing result is claimed.