# a156 / a087 read-only readiness — 2026-10-18

Workspace: /tmp/tossos-a148-impl-20261018. No production/test/spec edits or executions. Memory registry had no a156/a087 hit; ai-memory retrieval skill read, but memory_query is not exposed in this session. Current source is primary evidence; no claim of fresh CodeGraph/AST census.

## Recommendation: useful offline work exists now

Do not equate blocked a156 operational tasks 4.1/4.2 with blocked fixture work 2.2/2.3. Smallest useful addition is a test-only safety scenario using existing disposable journal/engine harnesses: partially filled campaign with unresolved entry order/reservation, restart, attempted next leg while explicit CampaignReconcile/position closing/pending protective sell or an existing entry-block latch remains active, then confirmed terminal reconciliation. Assert no duplicate order or early exposure release, and separately assert existing protective exit still operates when entry is disabled. Add only coverage missing after mapping existing tests; no generic rollout service or invented policy selector.

Reuse exact existing symbols:
- internal/positioncampaign/replay.go: Replay, RemainingFromCap, validateReplayLegQuantities.
- internal/positioncampaign/watermark.go: LegLedger.Observe, LinkOrder, MarkTerminal. Do not call MarkTerminal merely from 404 or absence.
- internal/positioncampaign/campaign.go: AdmitExposure (closing/reconcile/closed/risk-reducing pending refuse).
- internal/journal/position_campaign.go: PlanCampaignLeg, campaignExposureBlockedInTx; existing generation-scoped state/pending checks, CAS and idempotency rather than parallel state machine.
- internal/execgw/riskguardian_qfinal.go: PrecheckQFinalEntry, IssuePrecheckedQFinalEntry. Current reusable admission mechanism does not by itself prove a new later-leg policy/campaign binding.
- Existing candidate fixtures: TestExposureAdmissionReadsPositionAndPendingRiskReduction, TestCampaignFillRollbackAndRestartAreDeterministic, TestCampaignZeroAndPartialUnchangedTerminalObservationsCancelResidualIdempotently; a150 ExitSeedsAdmittedPolicyAfterConfigurationChange, ExistingExitRemainsManagedWhenLatchFails, RaisedFloorPersistsAcrossRestartAndPendingSell.

The new scenario should target an uncovered composed transition rather than duplicate these unit assertions. Later-leg quantity expected values need a Manager-defined residual risk/exposure formula and stop binding before tests can honestly assert a new qFinal. Existing planned leg weights are not that formula.

## Material design gaps before new production behavior

- Exact promotable new policy/version/digest absent: a153 is inconclusive, a154 hypothetical only, a155 offline diagnostics only. No source evidence supports assigning them operational approval. Existing internal/strategyrouter/production_family_activation.go activation document/verification/Desired/LeaseCeiling are reuse points, not permission to generate a new activation.
- Later-leg risk budget recomputation, pending-reservation accounting, target campaign generation and monotonic-stop binding need explicit input and expected quantity contracts. No new generic readiness checker without actual caller and schema.
- Intraday closing requires authoritative calendar/auction deadlines, cancel escalation and halt/unfilled policy. Observed source exit loop has no auction deadline policy. Cannot invent 15:30 timing or extend closing to weekly/adopted positions. Pure fixtures can preserve owner exclusion and unknown-not-terminal, but cannot prove unimplemented flattening.
- Tasks 4.1/4.2 and archival remain blocked by explicit bounded activation approval, actual observed evidence and authorized live proof. Offline test success must not check those tasks complete.

## a087 task 0.5 can proceed read-only now

Current direct source inspection:
- internal/app/engine/exitloop.go:395 ExitObserver.Run invokes ObserveOnce repeatedly, sleeping Interval; no regular-session check in this loop.
- ObserveOnce:457 uses current UTC observation timestamp, SLO deferral, workingSet, quote observation and freshness; time presence is not a market-session gate.
- judge:961 receives observed quote and cycleObservation, evaluates existing protection; record:1288 handles proposal persistence/order flow.
- sellIntent:1788 uses observed price or baseline and creates current LIMIT intent; has no calendar/session argument.
- internal/clock/market.go:146 InRegularSession exists; no match in current engine files. No auction/session selector found in inspected exitpolicy files.

This supports continued 0.5 upstream tracing but is not complete proof of reachability: trace construction/startup, quote usability/lease and any trading-service gate with fresh CodeGraph/AST before marking 0.5 complete. Task requires human D7 disposition using this census AND section 5.2 actual broker-session evidence. Do not infer broker MARKET support from source.

Phase constraints: a087 Phase 1 explicitly closed as nonimplementation (population zero); existing observed-price guard pin remains. Phase 2 sections 1–3 still require section 5 measurements/allowed disposition, 0.5 + D7, 0.9 and US human confirmation. This investigation authorizes neither MARKET gateway changes nor live probe execution. a156 must not bypass this by introducing EOD market behavior.

Clarification: an outstanding buy/reservation alone is not a universal later-leg blocker. Use an explicit source-defined admission refusal state. TransitionCampaign maps lineage/generation/watermark mismatches into RECONCILE; normal partial progress remains ACTIVE. Do not manufacture a new no-pending-buy policy.
