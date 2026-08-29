# Review — a072-wire-multi-market-strategy-runtime

- Date: 2026-08-04
- Stage: isolated runtime/lease core complete; production assembly integration pending
- Voices: Manager safety review, independent operations/security review, isolated-core authority review

## Findings and disposition

- KR and US evaluation workers start in the same runtime/release with independent calendar, activation,
  evidence cursor, budget and failure envelope; one market never waits for the other's stability.
- Dispatch uses durable owner epoch/fencing and irreversible
  `ISSUED→CLAIMED→SUBMITTING→SUBMITTED|AMBIGUOUS|REFUSED`; authority A→B→A cannot revive a lease.
- Round 2 froze reservation disposition atomically with exact outcome: pre-transport or definitive broker
  rejection/no-accept/no-fill is `REFUSED+RELEASED`; acceptance is `SUBMITTED+TRANSFERRED`; durable
  transport uncertainty alone is `AMBIGUOUS+HELD`. Reconciliation changes disposition, never lease state.
- Market-worker faults latch/restart only that market's entry worker. Central integrity faults block all
  entry and require fenced safety-only fallback within 60 seconds while broker protection is preserved.
- Paired KR/US worker state is sealed and defaults both markets OFF/UNOBSERVED in one release. Each market has
  independent calendar, activation, evidence cursor, budget, refusal/latch and bounded restart state.
- Complete lineage and every authority generation/digest, exact scope, protection serial, owner epoch/token
  and expiry are sealed into package-private lease construction.
- Every pure transition declares expected/next revision for eventual durable CAS. Pre-transport drift, expiry,
  scope or fence failure is one atomic `REFUSED+RELEASED` result with zero broker requests.
- Terminal replay preserves original state/disposition and may release only a separately sealed exact retry
  HELD reservation. Missing SUBMITTING outcome evidence cannot synthesize a terminal result.
- Recovery resubmit requires the same operation identity, current protection generation/serial/digest, complete
  broker identity/query/cancel/dedup/idempotency capability and bounded attempts. It is authorized only from
  exact `AMBIGUOUS+HELD` and revalidates the lease-bound current authority plus owner fence; `SUBMITTING`,
  released/transferred disposition, drift and stale ownership yield zero authorization.
- Out-of-order nonterminal submit/classification calls atomically consume the lease as `REFUSED+RELEASED` with
  zero broker requests. The separate `CLAIMED` crash classifier remains the sole pre-transport crash path.
- Definitive outcome proof seals exact operation, lookup and response digests, authoritative/complete status,
  acceptance, fill count and pending/terminal presence. Contradictions and observations before lease issue are
  rejected without synthesizing a terminal result.
- Worker latch/effective state is exact. The first valid typed cycle refusal is immutable, abnormal return uses
  its fixed code, invalid refusal fails closed, and restart attempt/deadline arithmetic saturates without overflow.
- All journal/adapter identities use one bounded canonical UTF-8 contract (non-empty, at most 256 bytes, no
  whitespace or control characters); recomputing a seal cannot legitimize a noncanonical identity.
- Central fallback requires a newer owner fence and frozen RTO at most 60 seconds; it has no entry or lease
  issuance. Failure remains critical while broker protection is preserved.
- The new pure `strategyflow` composition binds KR and US continuation, reversal and weekly evaluators in
  one paired `OFF/UNOBSERVED` registry. Approved-candidate scope, router evidence/config, selected lane,
  existing-owner campaign and accepted campaign/leg/risk lineage are exact and sealed; substitution,
  unsupported binding, native lane refusal and incomplete lineage fail closed before Guardian.
- `strategyflow` production dependencies are statically denied access to journal, Gateway, trading,
  configuration, engine and operating writer packages. Its public production path fixes the real router and
  real lane registry; evaluator substitution exists only as a package-private test seam.
- Independent v25 review initially BLOCKed caller-created authority, shadow reservation disposition,
  caller-string broker outcome and restart discovery gaps. The checkpoint was reduced to an unissuable
  durable schema: all public mutation methods return `ErrStrategyDispatchDormant`, future raw rows are
  constrained to an actual q_final decision, aggregate HELD plus exactly five monetary HELD reservations,
  exact active owner/scope/current fence, and broker order identity is account-wide unique. Cold restart
  discovery is read-only. Re-review is CLEAN for this dormant scope only.
- Independent supervisor review initially BLOCKed bool-only recovery, shutdown/enqueue loss and unbounded
  mutation callbacks. Recovery/mutation callbacks were removed; shutdown and enqueue now share one mutex
  barrier and drain both queues. Only evaluation-only callbacks can cross a watchdog, at most one per market;
  the market becomes irreversibly disabled and late results have no action path. Re-review is CLEAN, while
  durable fault persistence and production Runtime wiring remain explicitly pending.
- The dormant production assembly adds the single supervisor loop after reconcile/exit/filldetect without
  changing their construction or Runtime supervision. Its helper has zero inputs, fixes KR/US OFF with nil
  cycles and has no production Trigger caller. Independent review is CLEAN for this inert scope.
- The strategyflow→q_final bridge audit found no authority-complete construction path. In particular, lane
  lineage lacks executable entry/stop/target prices, no production loader can seal all five risk-bucket
  policies/snapshots or prospective owner mapping, the legacy exposure collector is not a production
  multi-market authority, and official FX validity is not preserved into a frozen cap. A bridge was therefore
  not created; the exact blockers and implementation order are recorded in the pre-edit analysis.
- Wave B adversarial FX review initially BLOCKed four authority gaps: configured official clients could mint
  from arbitrary endpoints/transports, opaque evidence was discarded into caller-forgeable FX DTOs, KR
  identity freshness was caller-selected, and haircut values/digests were caller-minted. The remediation
  seals configuration before `New` returns and removes authority eligibility after any constructor-time
  endpoint or HTTP-client override. officialfx.Evidence now remains opaque through q_final precheck and
  final issuance; riskbucket.BindFXAuthority alone derives the arithmetic DTO at the Guardian clock and
  caller-provided FX fields are overwritten.
- Re-review then BLOCKed a retained-Option TOCTOU between origin validation and the HTTP read. The final
  boundary makes post-construction endpoint, transport and account option replay a mutex-protected no-op,
  and performs official-origin validation plus token/data GET under one read lock. Configured, custom and
  non-comparable transports cannot mint authority or trigger an interface-comparison panic. Independent
  re-review is CLEAN.
- Same-currency identity now requires a private sealed, 256-byte-bounded snapshot capability with a maximum
  five-minute window. Cross-currency evidence requires a separate private sealed haircut-policy capability;
  its multiplier is exact-decimal canonicalized before the policy digest and full evidence seal are made.
  No production policy/snapshot loader exists, so both public zero values fail closed and neither KR nor US
  FX authority can currently be minted by production callers.
- Exact execution terms for all six KR/US lanes now seal entry, stop and target price provenance, currency,
  minor scale, unit version and policy lineage. Weekly value additionally seals its complete RR preimage;
  reversal stop evidence is private and freshness-bounded.
- Independent execution-term review found two successive saved-stop weaknesses: caller-forgeable provenance,
  followed by a public scalar that could suppress the sealed saved-stop branch and retreat to the candidate
  stop. Both continuation and weekly value now choose the saved stop directly from private sealed authority;
  public scalar deletion, lowering or substitution cannot weaken it, and scalar-without-authority fails
  closed. Final independent re-review is CLEAN.
- The paired KR/US first-leg checkpoint adds one journal-owned v26 transaction and immutable companion for
  q_final, aggregate plus five monetary holds, strategy attempt, campaign/claim, risk owner and leg 1. The
  journal mints the prospective token; exact replay is verify-only and partial authority is never repaired.
  Independent review is CLEAN after adding live manifest/client-order joins, all-column v25 preservation,
  migration rollback coverage and fail-closed row-count errors.
- The first 3.4a review BLOCKed transport-free claim on three P1 gaps: release disposition lacked an exact
  final-state proof, operation/router lineage was not durable, and three attempt/strategy/client joins were
  not repeated at claim time. Remediation binds the fixed production RouterID/RouterRelease through
  strategyflow, first-leg request/binding and future lease, requires `operation_id == client_order_id`,
  repeats every live join, and proves one aggregate plus five distinct monetary holds are RELEASED/zero.
  Already-released rows normalize idempotently; missing/cross-scope rows roll the whole transaction back.
- A follow-up provenance review found that a fresh caller could name a different well-formed router pair.
  `RecordQFinalCampaignFirstLeg` now accepts only the exact sealed production RouterID/RouterRelease before
  opening a transaction, and paired KR/US fresh-forgery tests prove zero writes. Final re-review is CLEAN.
  Task 3.4a is complete; 3.4b (`CLAIMED→SUBMITTING` immediate pre-Gateway fence) remains pending.
- Engine Task 3.7 initially BLOCKed on a duration-based restart handoff race. The supervisor now publishes
  one absolute `RestartNotBefore` and recomputes the remaining interval at sleep time. KR/US × error/panic
  race regressions pass; failed entry stays OFF-latched while peer evaluation and all five safety loops
  continue. Independent re-review is CLEAN.
- Currency architecture review rejects per-market Guardians because they could each consume the full
  account-wide cap. Production integration will retain one account-base-currency Guardian and bind the same
  request-scoped frozen official FX evidence through sizing, aggregate/five-bucket reservation and Gateway.
  This decision grants no current authority; the production FX/policy/account snapshot loaders remain absent.

## Verification

- Strict OpenSpec validation: PASS.
- `go test ./internal/strategyruntime -count=1`: PASS.
- `go test -race ./internal/strategyruntime -count=1`: PASS.
- `go vet ./internal/strategyruntime`: PASS.
- `FuzzTerminalLeaseNeverHasOutgoingTransition` and `FuzzAuthorityGenerationNeverRevivesLease` (3s each): PASS.
- Statement coverage: 89.1%.
- Static/external-package tests prove no broker/live transport, engine/journal/gateway/scheduler writer,
  toggle/approval path or public authority constructor exists.
- `go test ./internal/strategyflow ./internal/strategyrouter ./internal/continuationlane ./internal/reversallane ./internal/weeklyvaluelane ./internal/strategydispatch -count=1`: PASS.
- Normal and `tossos_testseams` tagged tests across `strategyflow`, `strategyrouter`, `continuationlane`, `reversallane` and `weeklyvaluelane`: PASS.
- `go test -tags tossos_testseams ./internal/strategyflow -run TestProductionEvaluateUsesRealRouterAndAllSixConcreteEvaluators -count=1`: PASS; production `Evaluate` used the real sealed router and all six concrete KR/US continuation, reversal and weekly evaluators, preserving distinct candidate/router/lane evidence plus exact router/lane releases.
- `go test -race -tags tossos_testseams ./internal/strategyflow ./internal/strategyrouter ./internal/continuationlane ./internal/reversallane ./internal/weeklyvaluelane -count=1`: PASS.
- `go vet -tags tossos_testseams ./internal/strategyflow ./internal/strategyrouter ./internal/continuationlane ./internal/reversallane ./internal/weeklyvaluelane`: PASS.
- Production assembly, official Gateway CAS and full safety-loop integration remain pending by design.
- `go test ./internal/journal -count=1`: PASS on the integrated stable tree (180.938s).
- v25 focused/race/vet/diff checks and independent dormant-scope re-review: PASS/CLEAN.
- `go test ./internal/app/engine -count=1`: PASS; full package race PASS (396.594s); vet PASS.
- KR/US supervisor independent re-review after three Critical fixes: CLEAN.
- Dormant `cmd/tossctl` production assembly focused/race/vet and independent review: PASS/CLEAN.
- Wave B FX focused tests, affected-package race tests and vet for official, officialfx, riskbucket and
  execgw: PASS. No journal, broker, activation or LIVE mutation was added.
- Six-lane execution-term normal/tagged/race/tagged-race tests and vet: PASS. Saved-stop adversarial
  regressions and final independent review: PASS/CLEAN.
- Final integrated `go test ./internal/journal -count=1`: PASS (221.186s).
- Final focused journal tests for first-leg, v26 migration and claim: PASS; paired claim/first-leg race tests:
  PASS (65.554s); `go vet ./internal/journal`: PASS.
- `go test -race ./internal/strategyrouter ./internal/strategyflow -count=1`: PASS; tagged production
  strategyflow test: PASS; affected-package vet: PASS.
- Task 3.7 full engine tests/vet and independent re-review: PASS/CLEAN; deterministic paired handoff-race
  suites passed at repeated race counts.
- Paired composite settlement focused tests: PASS. KR and US cover byte-exact ACK identity, caller-field
  non-authority, cross-attempt refusal, ACK/fill ordering in both directions, late-link rollback, blank ACK,
  unknown outcome class and duplicate zero-resend. Focused race and `go vet ./internal/journal`: PASS.
- Strategy Gateway paired integration: normal and `tossos_testseams` race tests PASS. Stale or cross-market
  account-base FX reaches the exact pre-transport refusal handoff once, reaches neither SUBMITTING nor the
  broker, and the journal refusal matrix proves aggregate/five-bucket terminal normalization.
- A subsequent independent review found four P1 gaps: nil official verifier confirmation, terminal
  ACK-window residual holds, split prepared-refusal commits and missing production Campaign hook. Paired KR/US
  RED regressions now pass for all four. Core+claim+six-hold refusal rollback preserves the exact preimage;
  terminal zero/partial fill releases only the remainder while retaining fill evidence. Full engine, normal
  Gateway and `tossos_testseams` Gateway package suites pass. The independent re-review is CLEAN with
  P0/P1/P2 = 0/0/0; focused normal/race/vet checks passed and the reviewer changed no files.
- Claimed-crash recovery is now a no-send terminalizer rather than a dormant stub. A newer current owner can
  recover only exact `CLAIMED + RESERVED` rows with no durable transport-start marker. KR and US tests prove
  stale-owner fencing, `SUBMITTING` refusal, exact replay non-revival, optional prepared-core closure and
  injected late-write rollback of core, lease and all six holds. Focused tests passed (29.748s), the paired
  recovery race suite passed (24.176s), and journal vet plus diff checks passed. No retry/resubmit or official
  outcome authority was added.
- Full untagged repository verification with an explicit 30-minute package timeout passed. The earlier default
  run reached Go's 10-minute package timeout while journal was in SQLite migration fsync, with no assertion
  failure; journal then passed alone in 404.183s and the integrated run passed with journal 386.779s,
  engine 120.065s and execgw 108.331s.

## Final transport-boundary red-team review

- Verdict: CLEAN; CRITICAL/P1/high findings: 0.
- The final journal transport proof is performed inside the entry-gate critical section immediately before
  the official send and repeats the exact dispatch owner, lease, revision, state and expiry.
- `TestStrategyGatewayFinalGateWaitCannotUseExpiredReplacedOwnerKRUS` proves both markets fail closed when an
  owner expires or is replaced while waiting for the gate. Normal and race executions pass.
- Market-local final scheduler revalidation, monotonic entry-gate revision binding, SUBMITTING owner takeover
  grace and immutable lease-update guards were re-reviewed together; no remaining high-severity gap exists.

## Dormant deployment startup-interlock correction

- The first dormant image replacement exposed a production regression: the strategy-only official
  `GET /api/v1/exchange-rate` read had been added to the global engine startup attestation, so a missing
  supervised FX probe stopped every existing engine loop even though the KR/US strategy workers remained OFF.
- The exchange-rate read is now excluded from the global startup dependency set. US strategy entry still
  fails closed inside its market-local FX authority and dispatch path; this correction grants no fallback FX,
  lane activation, toggle, approval or broker mutation authority.
- RED regressions cover the global endpoint set and the existing KR-preservation/US-FX-failure behavior.
  Targeted normal/race/vet checks, the full Go suite, diff/logic-map checks and an independent review all pass.
  The independent verdict is CLEAN with no high-severity finding.

## Verdict

The paired first-leg transaction, transport-free claim CAS, claimed no-send crash recovery, strategy-only
atomic Gateway settlement and market-isolated engine recovery are approved for the next production-integration
checkpoint. Production activation/calendar/bucket/account/exposure assembly, SUBMITTING exact-outcome recovery
and a071 protection assembly remain pending.
KR and US remain independently OFF until human activation and every current lease authority are valid. The
completed work grants no lane, automation, toggle or LIVE approval authority.
