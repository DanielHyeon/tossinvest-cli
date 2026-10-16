## 1. Pre-edit evidence

- [x] 1.1 Capture current AST/logic/branch maps for `M0ExactPrerequisites`, CLI validation, and direct Runner admission; record the immutable base and show existing TTL skip/fail/refuse/missing behavior.
- [x] 1.2 Add failing focused tests proving only TTL is exempt. With no `pending-create` checkpoint, every other latest non-PASS verdict, Outstanding artifact, or non-pending M0Unsettled state rejects before broker factory, confirmation, planned mutation, cleanup, or request; retain the existing `pending-create` read-only recovery then HOLD exception.

## 2. Narrow prerequisite change

- [x] 2.1 Change only the M0 prerequisite predicate so `StepIdempotencyTTLEdge` and `StepConditionalTrigger` are excluded; retain iteration and latest-PASS requirements for every other Step.
- [x] 2.2 Keep CLI and direct Runner strict trigger-only validation unchanged: no TTL opt-in, one trigger redo, confirm-each, resume, receipt, outstanding and checkpoint holds, raw-read source, and execution exclusion.
- [x] 2.3 Preserve existing TTL verdicts; verify missing/nonterminal TTL uses ordinary no-opt-in skipped recording only and never executes the TTL step body or sends its first/replay/cancel order; ordinary skipped append and `Plan.Excluded` metadata remain allowed.

## 3. Verification

- [x] 3.1 Run focused CLI and `internal/verifylive` tests with fake-broker request logs for TTL PASS/skipped/failed/refused/missing/nonterminal and assert zero TTL broker mutation, TTL step-body dispatch, and TTL planned-mutation line in M0 mode, while allowing ordinary skipped append and `Plan.Excluded` metadata.
- [ ] 3.2 Run focused negative tests for each non-TTL prerequisite and Outstanding/non-pending M0Unsettled HOLD and pending-create read-only recovery HOLD, then run required repository checks.
- [x] 3.3 Update final AST/logic/branch maps and review evidence; record that this change does not run a live verification or alter operating settings.
