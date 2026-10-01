Access rule: do not read, list or write anything under ~/.codex, and do not access paths outside the working directory. If you did access either at any point, your FIRST line must be `ACCESS-VIOLATION: <what>` before the verdict line.

You are a cross-model adversarial reviewer (proposal-freeze, round 2) for OpenSpec change `a091-a-stop-that-sold-nothing-is-critical`
in TossOS, a live-money automated trading engine written in Go. The working directory is a read-only `git archive` of the commit under
review. Do not modify files. Report only what you verify in this tree, with file:line evidence.

Safety invariants that override everything: never weaken or delay stop-loss immediacy (§0.3); do not change submitted quantity, price or
thresholds (§0.9); account for every broker request (§0.4); toggle-OFF must equal upstream behaviour; never put raw account numbers in
logs or alerts.

Read first: `openspec/changes/a091-a-stop-that-sold-nothing-is-critical/{proposal.md,design.md,tasks.md,issues.md,specs/engine-safety/spec.md}`,
then `review.md` (round 1 rejected freeze on 2026-08-06; the "2차 판" section records the rebase and FLM regeneration). The canonical
specs are `openspec/specs/engine-safety/spec.md` (「등급화된 알림」, and 「무관리 보유 보고의 등급은 사실이 정한다」 from a095) and
`openspec/specs/exit-policy/spec.md`.

What the change will do: add `obs.EventExitStopSoldNothing = "exit.stop_sold_nothing"` to `criticalEvents` (`internal/obs/event.go`).
When `ExitObserver.applyFloor` (`internal/app/engine/exitloop.go` ~1617-1661) returns zero for a PROTECTIVE exit (both the floor-error path
B2 and the tail zero-floor path), report it with that kind (critical, which means a durable outbox row through `obs.RecordOnly`,
`internal/obs/record_only.go`). Partial caps and take-profit zeros stay on `EventExitProposalCapped` (normal). B2's `logErr` kind changes
for protective exits. The zero-case title stops saying "일부만 나갔다". `submit` passes `isProtective(proposal)`. Return values of
`applyFloor` do not change.

Verify or refute, citing code:
1. Every code claim and number in design.md D1–D6 and the range table (coordinates, 19 critical kinds, retry/timeout numbers, the M1 chain).
2. §0.3: the loop cost a091 adds (design D5 「0주 기록」). Include the record-failure path (gate latch plus escalate, and whether escalate
   runs in the exit goroutine) and contention on `n.mu`.
3. Entry-gate and operating-mode consequences of a new critical kind when delivery fails or no transport is configured. Compare with a095's
   canonical decision that an alerts-off default engine must not block entries (`internal/app/engine/adoption.go` ~440-452,
   `NotificationsEnabled`). Is a091 consistent with that decision, or must it state or handle the alerts-off case?
4. Episode key `exit.stop_sold_nothing|<position id>`, shared by B2 and the tail. PENDING rows keep the first content (a097). Is the spec
   SHALL "원인은 세부 정보로 구분해 담아야 한다" satisfiable?
5. The protective/take-profit split (`isProtective`) for every orderable action, including recovery snapshots.
6. Whether every SHALL and Scenario in the delta has a RED task that fails on a wrong implementation, and whether any task premise is
   false (replay 5.1/5.2, the harness's alert sink).

Output: first line `VERDICT: PASS` or `VERDICT: FAIL`. Then numbered findings, each with severity (P0 blocks freeze, P1 fix before freeze,
P2 fix or name, P3 nit), the exact sentence or gap, file:line evidence, and the minimal fix.
