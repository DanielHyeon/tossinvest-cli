Access rule: do not read, list or write anything under ~/.codex, and do not access paths outside the review tree. Do not use shell process
substitution. If you did access anything outside the tree, your FIRST line must be `ACCESS-VIOLATION: <what>`.

You are a cross-model adversarial reviewer of the IMPLEMENTATION of OpenSpec change `a091-a-stop-that-sold-nothing-is-critical` in TossOS,
a live-money automated trading engine written in Go. The read-only review tree is `/tmp/claude-1000/a091-i1-tree` (a git archive of
97a6f717). Read ONLY there. You may run `go test -run <name> ./internal/...` inside the tree to prove a claim; set
`GOCACHE=/tmp/claude-1000/a091-codex-gocache`.

The frozen contract (authority): `openspec/changes/a091-a-stop-that-sold-nothing-is-critical/{design.md,specs/engine-safety/spec.md,specs/exit-policy/spec.md,tasks.md}`.
`review.md` records the freeze rounds, the Manager rulings and the implementation lot. The implementation lives in
`internal/app/engine/exit_stop_sold_nothing.go` (new), `internal/app/engine/exitloop.go` (`applyFloor`, `submit`, `ExitObserverOptions`),
`internal/app/engine/exitwiring.go` (`Context.ExitObserver`), `internal/obs/event.go` and `internal/obs/notifier.go` (`escalate`). Tests:
`internal/app/engine/a091_*_test.go`, `internal/obs/a091_*_test.go`, `internal/riskcalc/a091_*_test.go`. Evidence:
`analysis/implementation/{mutation-ledger.md,pre-edit.md}` and `issues.md` 「8/2 재생 결과」.

Safety invariants that override everything: never weaken or delay stop-loss immediacy (§0.3); submitted quantity, price and thresholds are
unchanged (§0.9); no new broker requests (§0.4); toggle-OFF equals upstream behaviour; no raw account numbers in logs or alerts.

Attack these specifically:
1. Bypassing the alerts-off gate at the config-load boundary.
   - Is `NotificationsEnabled` always the loaded `notifications.enabled`?
   - What about a refused notifications block, `enabled=true` with no transport, a config reload or hot settings change while the engine
     runs (`internal/console/settings_notifications.go`, `Notifications.Disable`), and other `ExitObserver` constructors or wiring paths?
   - Can an alerts-off engine still produce a critical `exit.stop_sold_nothing` row by any path?
2. The `context.WithoutCancel` record and the shutdown window.
   - Can the uncancellable record outlive the journal (shutdown ordering in `internal/app/engine/runtime.go`, `cmd/tossctl/engine.go`)?
   - Can it deadlock against `n.mu` holders during shutdown, or extend shutdown without bound?
   - Is the cancellation-only predicate (`cancellationOnly`) correct for every error shape the floor path returns?
3. Mutations at the boundaries of the four cause categories in `classifyZero` / `zeroIsCritical` / `reportZeroFloor`.
   - Which mutation would survive the tests? Name it and prove it by applying it in a scratch copy, NOT in the review tree.
   - Pay special attention to the protective/take-profit split, the Holdings bound, B2 versus B7, ctx done versus not, and the log-only path.
4. Honesty of the 750ms / measure-and-record criterion text (design D5 7판, tasks 5.3, review 「수락 기준 정정」).
   - Does the test actually enforce what the text claims (owned cells ≤ 750ms; Acknowledge cell < 5s)?
   - Are the measured numbers reproducible?
   - Is anything stated as measured that was not?
5. Anything else: return values unchanged (§0.3/§0.9), H2 one-kind, the account canary scope, and the 4-arm replay's assertions versus the
   table in `issues.md`.

Output: first line `VERDICT: PASS` or `VERDICT: FAIL`. Then numbered findings, each with severity (P0/P1/P2/P3), file:line evidence, and
the minimal fix. Only report what you verified.
