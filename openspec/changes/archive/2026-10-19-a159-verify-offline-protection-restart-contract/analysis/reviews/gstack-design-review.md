# a159 gstack design review

2026-10-19. **PASS for the approved test-only design scope.** Codex-host gstack design/readiness review; no implementation, test execution, full gate, live readiness or performance approval claimed. No production/test/source evidence edits performed by this reviewer. a156/a157 remain outside scope.

Reviewed proposal/design/spec/tasks, existing RecordFill composite fixture, working-BUY cancellation/alert fixture, campaign admission/reconcile paths, research report and nine fresh AST files. All nine AST source hashes match current corresponding source files. Existing component coverage supports feasibility investigation but does not prove the composed test is already implemented or that all private helpers are accessible from the chosen test package.

## Contract assessment

- Same temporary DB and confirmed order/attempt/campaign lineage is maintained across actual RecordFill and production apply hooks, explicit protection reconcile, close/reopen and duplicate replay. Manual position SQL cannot substitute for the claimed partial fill. The 10 ordered / 3 filled / 7 residual oracle is independent and source-consistent.
- New-leg rejection uses fresh command key, valid sequence and current version with exact ErrExposureBlocked. Existing command replay may return its historical result before checking a latch; the design correctly excludes that false positive. State/event/version/quantity assertions and unblocked/different-scope controls prevent blanket-denial success.
- Phase 5 now fixes cancellation failure/uncertainty, matching TestAnUncancellableEntryWithholdsTheLiquidationAndAlertsPastTheBound. That existing fixture allows submission once cancellation succeeds; the new design explicitly does not change this into a universal terminal-fill prerequisite. Phase 6 terminal evidence is a sufficient fixture setup, not a new global BUY cancellation contract.
- Protection evaluation, delayed alert and submit count are distinct assertions. Confirmed attempt versus working order is distinguished; the latter must remain visible to the fake unsettled-order reader. Pending SELL expects preservation and zero new sends; non-breach expects zero. Floor reconciliation may require another genuine fake-clock observation before an authorized capped sell. No invented hidden residual, 404 terminal proof, stop widening or order-policy change.
- Two independent disposable mutations must fail behavioral assertions, not compilation or altered test expectations. Original source must pass. Existing framework/helper reuse is preferred, and production limitations/defects return to Manager rather than being bypassed with exported test APIs or a new framework.

No actionable P1/P2 design blocker found. Implementation review must still verify the complete same-lineage composition, ownership/guards on the real fill path, actual scoped controls, no reservation release on ambiguous cancellation, and the two meaningful mutation receipts. Source/test counts or several unrelated passing unit tests cannot replace that proof.

## Scope and completion

KR offline regression only. No new qFinal, policy activation, deadlines/auction handling, MARKET support, collection, operating setting, dependency or schema change. a156 live acceptance and all a157 work remain untouched. Only this separate change can complete after independent/gstack implementation review, Manager inspection and actual aggregate gate/lifecycle receipts. Grill-me scope choice is resolved by user approval; remaining fixture details are source questions. Impeccable is N/A because no UI changes. Research-first appropriately uses existing code/harnesses; no external research needed for these internal contracts.

## Reviewed document hashes

- `openspec/changes/a159-verify-offline-protection-restart-contract/proposal.md`: `6cdab1b56127ac3368c151b02bdee3d16f981be8adf22141c475de6e81b0e3b9`
- `openspec/changes/a159-verify-offline-protection-restart-contract/design.md`: `cfc765391931206c16f0b6fb1cab9e55954fb10d571464db532387c5637f1b30`
- `openspec/changes/a159-verify-offline-protection-restart-contract/tasks.md`: `e2f2ede5a101be4d5f07f7a33aa326f69c6a2194a6e2cb17fc8a77502cc21569`
- `openspec/changes/a159-verify-offline-protection-restart-contract/review.md`: `2d8b8b3367868b5cfcf58b953c759568f89020966d440f9ba996bde4ad0bc6cd`
- `openspec/changes/a159-verify-offline-protection-restart-contract/specs/strategy-runtime/spec.md`: `b01803e471e0830b57fd4217d28d166c868f26da21a1b5cfc858af7f8c762416`

Research: `/tmp/tossos-offline-regression-research-20261019.md`; AST directory `/tmp/tossos-offline-regression-research-20261019-evidence/`.
