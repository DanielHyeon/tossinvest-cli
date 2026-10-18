# Independent a159 design review

Verdict: PASS — design freeze only. 2026-10-19. Reviewed baseline `1dbff7ced3653e3e9c1aaa892d18d76b1a50d511`; no implementation, test execution, mutation result or gate success claimed.

## Scope and source findings

All four canonical artifacts agree on a test-only KR composed regression using the same persisted lineage. No production seam, schema, policy, live transport, a156 completion or a157 modification is authorized. Existing synthetic fixture results cannot establish live acceptance.

The corrected phase 5 explicitly fixes unsuccessful/uncertain cancellation as the reason for zero sells. Phase 6 terminal BUY evidence is a sufficient fixture state, not a new mandatory BUY policy. This resolves the earlier research overstatement: `ExitObserver.clearTheSymbol` (`internal/app/engine/exitloop.go:1611`) separately reads working orders and gateway unsettled mutations; confirmed BUY cancellation can clear, whereas confirmed SELL cancellation still waits for terminal evidence. `Gateway.UnsettledOnSymbol` (`internal/execgw/a094_shared_judgements.go:21`) delegates to journal PendingAttempts and must not be replaced by a misleading empty fake.

`Journal.PlanCampaignLeg` (`internal/journal/position_campaign.go:522`) checks command replay before CAS, then exposure/entry blocking before sequence validation. Design correctly requires a fresh command, latest version and valid next sequence. The positive control must independently establish its own existing first leg, read its own current version, and use its own fresh command/plan key and positive quantity. Merely obtaining ErrExposureBlocked does not establish sequence validity because blocking precedes that validation. Matched valid unblocked admission closes this false-green route.

`Journal.ReconcileStrategyProtection` (`internal/journal/strategy_protection_reconcile.go:14`) binds account/market/symbol/decision/generation and stored position identity, then uses a deterministic per-generation command to avoid duplicate events. Reopen and duplicate assertions must compare after the first committed reconcile; they must not expect the initial reconcile to leave version unchanged. Wrong-scope requests and unaffected valid sibling admission provide distinct scope evidence.

The genuine v4 admission harness `pairedStrategyDispatchCycleFixture` (`internal/app/engine/strategy_dispatch_cycle_test.go:244`) is reusable, but its existing spy only returns a CONFIRMED outcome. Public Prepare/DispatchStrategyVerified/RecordFill plus production hooks must create the order/partial fill; existing a150 manual position insertion cannot substitute. This is feasible source-level composition, still subject to implementation verification.

## Acceptance details to retain in implementation review

- Assert exact 10/3/7 values before terminal processing; do not expect residual seven after authoritative terminal cancellation releases the remainder.
- Record admitted policy identity, generation and effective stop before close and after reopen; keep existing first-floor suppression separate from cancellation-delay observations.
- Pending SELL preservation means the same outstanding protective proposal is not duplicated by replay. It must not introduce a universal prohibition on existing CancelPendingFirst replacement behavior for a different liquidation proposal.
- Assert real persisted no-write/refusal and duplicate state, not only fake counters. New leg rejection is PlanCampaignLeg coverage, not proof of broker or later-leg qFinal submission.
- Mutation runs must fail relevant behavioral assertions in disposable variants, with unchanged-source GREEN and actual tagged test selection.

No actionable design blocker remains. Implementation must demonstrate these assertions; PASS here does not pre-approve the resulting fixture or evidence.

## Reviewed artifact hashes

- `proposal.md`: `6cdab1b56127ac3368c151b02bdee3d16f981be8adf22141c475de6e81b0e3b9`
- `design.md`: `cfc765391931206c16f0b6fb1cab9e55954fb10d571464db532387c5637f1b30`
- `tasks.md`: `e2f2ede5a101be4d5f07f7a33aa326f69c6a2194a6e2cb17fc8a77502cc21569`
- `specs/strategy-runtime/spec.md`: `b01803e471e0830b57fd4217d28d166c868f26da21a1b5cfc858af7f8c762416`

## Source evidence hashes

- `internal/journal/position_campaign.go`: `0168dc2a026488a43161803e4441b8d40c561a244c0a7043006ff29e8d1e078b`
- `internal/journal/strategy_protection_reconcile.go`: `57a172e322e569c691ac6354910311af19250e646f4d74f0bae9c30911c36be4`
- `internal/app/engine/exitloop.go`: `ba30515476584fead62907157bb5ba985271057b2da882c6fda79e2b29951598`
- `internal/execgw/a094_shared_judgements.go`: `ccc2c9efa6c13ad98de62f04faaff23796c2913bed46226a8d215a15373a9901`
- `internal/app/engine/strategy_dispatch_cycle_test.go`: `52e4ecf3e6fc26e0152f604f3fb78e6752347eef5ac696a66c0bf5f00c063996`
