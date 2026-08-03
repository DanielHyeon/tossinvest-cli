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

## Verdict

Isolated core approved for production-integration review. KR and US remain independently OFF until human
activation and every current lease authority are valid. This core has no transport dependency and grants no
lane, automation, toggle or LIVE approval authority.
