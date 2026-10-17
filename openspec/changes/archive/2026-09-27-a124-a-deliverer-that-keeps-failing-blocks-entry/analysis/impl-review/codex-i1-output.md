**Three P1 issues remain: the cancellation exception violates D1, the latency evidence misses the escalation path, and several promised adversarial schedules are untested.** No P0 demonstrated by static inspection.

Read-only review; no tests, engine execution, network calls, or modifications performed. Recorded PASS results are evidence supplied by the implementation, not independently reproduced.

References below use filenames; production files resolve under the directories specified in your request. `judges_test` means `internal/app/engine/a124_the_deliverer_judges_internal_test.go`.

**1. Contract conformance**

D1 failure-record table:

| Row | Status | Evidence / reason |
|---|---|---|
| `err != nil` | CONFORMS | Error checked before zero-valued outcome; routes to D8. `alertdelivery.go:342–346`. |
| Applied, attempts below limit | CONFORMS | Clears record-failure run, returns without judgement. `alertdelivery.go:349–354`. |
| Applied, attempts at/above limit | CONFORMS | Conditional latch plus escalation. `alertdelivery.go:355`. |
| AlreadySettled / LeaseLost | CONFORMS | Neither judgement nor counter reset. `alertdelivery.go:356–357`. |
| NotFound | CONFORMS | Latch-only judgement. `alertdelivery.go:358–361`. |
| Unknown outcome | CONFORMS | Same conservative latch-only default. `alertdelivery.go:358–361`. |

D1 delivery-settlement table:

| Row | Status | Evidence / reason |
|---|---|---|
| Applied | CONFORMS | Clears counter; no judgement. `alertdelivery.go:381–383`. |
| AlreadySettled / LeaseLost | CONFORMS | No judgement or explicit release. `alertdelivery.go:389–391`. |
| NotFound / unknown | CONFORMS | Immediate latch-plus-escalation judgement; no release. `alertdelivery.go:392–394`. |
| Error | **DEVIATES** | Cancelled Run context suppresses judgement entirely. Otherwise conforms. `alertdelivery.go:370–379`. |

The committed-attempts input **CONFORMS**: same-transaction SELECT precedes Commit, populated result returned only after successful Commit; other outcomes/errors carry no attempts. `alert_claim.go:324–354,360–363`. Commit-failure test evidence has the qualification in I5 below.

D7 judgement→action table and ordering:

| Row / requirement | Status | Evidence / reason |
|---|---|---|
| Failed attempt reaches limit | CONFORMS | Current epoch, latch plus escalation. `alertdelivery.go:355`. |
| Failure-record NotFound / unknown | CONFORMS | Current epoch, latch only. `alertdelivery.go:361`. |
| Delivery error / NotFound / unknown | **DEVIATES** | Correct actions except cancellation escape. `alertdelivery.go:374–394`. |
| Record-failure counter reaches limit | CONFORMS | Uses previous increment’s epoch; deletes run; escalates. `alertdelivery.go:493–500,532–533`. |
| Listing-failure counter reaches limit | CONFORMS | Same transition helper; resets counter; escalates. `alertdelivery.go:508–515`. |
| Escalation despite rejected conditional latch | CONFORMS | Escalation independent of block result. `alertdelivery.go:417–424`. |
| Escalation failure unconditionally relatches | CONFORMS | Plain `Block` after error, regardless of earlier clear. `alertdelivery.go:424–425`. |
| Failed-record ordering | CONFORMS | Settlement → release → epoch → conditional block → escalation. `alertdelivery.go:336–355,416–425`. |
| Delivery ordering / lease retention | CONFORMS except cancellation | Settlement precedes epoch; no release. `alertdelivery.go:368–394`. |
| Lock isolation | CONFORMS by inspection | Epoch comparison/insertion use one gate lock and map operations; ledger and logging happen after release. `retry.go:559–580`; `alertdelivery.go:416–450`. No Notifier reference in executor. |

D8 AA2 transitions and pruning:

| Row / rule | Status | Evidence / reason |
|---|---|---|
| AA2 #1: threshold before epoch reset | CONFORMS | Threshold branch precedes epoch mismatch; returns previous epoch. `alertdelivery.go:532–536`. |
| AA2 #2: below threshold, epoch changed | CONFORMS | Restarts at `(1,current epoch)`. `alertdelivery.go:535–536`. |
| AA2 #3: otherwise increment | CONFORMS | Increments and stores current epoch. `alertdelivery.go:538`. |
| Claim / failed-record errors count per row | CONFORMS | Both reach `countRecordFailure`. `alertdelivery.go:260–265,342–345`. |
| Applied settlement clears own run | CONFORMS | Both settlement paths clear. `alertdelivery.go:351,382`. |
| LeaseLost / AlreadySettled / release Applied preserve run | CONFORMS | No reset in these paths. `alertdelivery.go:356–357,389–391,582–587`. |
| Claim already settled clears run | CONFORMS | Clears on settled disposition. `alertdelivery.go:287–292`. |
| Complete listing prunes absent IDs | CONFORMS | Only `len < batch`; absent IDs deleted. `alertdelivery.go:238–241,548–559`. |
| Truncated listing preserves absent IDs | CONFORMS | Pruning skipped. `alertdelivery.go:238`. |
| Run cancellation excluded from counts | CONFORMS | Both counters check Run context. `alertdelivery.go:485–508`. |
| Listing success resets listing run | CONFORMS | Explicit reset before row processing. `alertdelivery.go:237`. |
| Threshold judgement deletes/resets counter | CONFORMS | Row deletion / listing reset before judgement. `alertdelivery.go:499,514`. |
| Same-ID rearm carryover / restart | CONFORMS | Counters keyed only by ID; process-local fields. `alertdelivery.go:170–174,493–500`. |

Other requested decisions:

| Decision | Status | Evidence / reason |
|---|---|---|
| D2 limit 3 | CONFORMS | References `obs.DefaultCriticalAttempts`. `alertdelivery.go:86`. |
| D3 missing publisher counts | CONFORMS | Fixed cause enters failed-attempt path. `alertdelivery.go:303–325`. |
| D4 selection | CONFORMS | Below-limit first, ID order within tiers, no discard; legacy query unchanged. `outbox.go:522–551`. |
| D9 sanitization | CONFORMS by inspection | Fixed gate details; new logs omit raw error/account/token/content; escalation error uses fixed classification. `alertdelivery.go:97–103,440–478`. |
| D9 repeat suppression | CONFORMS | Escalation logs only on change; latch reporting once per observed epoch. `alertdelivery.go:454–478`. |
| D10 enforcement boundary | CONFORMS | Production projector/restoration callers remain absent; startup restores only PENDING backlog. `operating_mode.go:475–476`; `gateway.go:153–167`. |

**Recorded deviations**

- **Cancelled-context delivery error: not conservative relative to the frozen contract.** It removes both consequences, and checks `ctx.Err()` rather than whether cancellation caused the settlement error. “Row remains PENDING” is also not guaranteed: acknowledgement or another sender can settle it before the failed call. Shutdown may reduce immediate exposure, but restart restoration does not establish the claimed equivalence. I1 applies.
- **Empty AccountRef / nil Gate: not conservative as general behavior, but excluded by normal production assembly.** Nil gate is rejected by `Context.AlertDeliverer` (`auxiliary.go:157`). Account resolution rejects an empty account (`interlock.go:684–687`). Thus no current production failure demonstrated. Empty-account escalation nevertheless returns success without writing (`alertdelivery.go:435–436`), bypassing fallback relatching if this precondition is violated. Record these as assembly preconditions; preferably validate AccountRef in the factory too.

**2. New findings**

| ID | Severity | Finding | Evidence | Suggested fix |
|---|---|---|---|---|
| I1 | **P1** | Cancellation drops a mandatory delivery-error judgement, including an unrelated ledger error followed by cancellation. Neither latch nor escalation is attempted. The asserted PENDING restoration premise is unproven. | `alertdelivery.go:370–378`; frozen spec `specs/engine-safety/spec.md:60–63`; acknowledgement can remove PENDING at `outbox.go:494–499`. | Remove this exemption and retain D7 fallback behavior, or obtain an explicit contract revision with shutdown/interleaving evidence. Add a regression where cancellation follows the settlement error, including an already-settled row. |
| I2 | **P1** | Task 2.6’s “judging executor” measurement does not establish escalation contention. Seeds default to zero attempts; timing begins after the first publish, normally before any threshold judgement. The mode trigger has no firing assertion. The backlog selection also completes before the measured exit starts. | `a124_judging_does_not_delay_protection_test.go:122–123,146–148,186–205`; `outbox.go:58`; threshold `alertdelivery.go:352–355`. | Seed threshold-minus-one attempts; assert escalation/trigger execution; synchronize exit work with settlement, escalation, and selection while each holds the connection. Add missing-publisher variant. Retain fixed budget and oversized-delay negative control. |
| I3 | **P1** | Required adversarial schedules are missing despite broad coverage claims: delivered→empty acknowledgement, ㉩/㉪, failed acknowledgement, listing threshold followed by clear, same-ID carryover, and expiry while gate-read is blocked. Required escalation-under-gate-lock mutation is absent. | `judges_test:424–434` acknowledges a PENDING row, not a delivered one; `:618–646` tests row errors only; `:664–685` listing tests contain no clear; `:744–765` tests expiry between completed cycles. Mutation list ends at M26 in `analysis/harness/mutate_a124.py:147–157`. | Add the exact promised schedules with hook-fired assertions and both gate/mode assertions. Add the lock-scope mutant and demonstrate its behavioral kill. |
| I4 | **P2** | Atomicity stress assertion cannot detect the target stale relatch. A broken compare-unlock-clear-lock-insert implementation can leave `latched=true, epoch>e`; the test accepts it. `-race` need not detect this logical race. | `a124_clear_epoch_internal_test.go:139–146` rejects only `!latched && now == e`. | Add a deterministic comparison/insertion barrier and require a clear between them to prevent insertion; kill a split-lock mutant. |
| I5 | **P2** | “Commit failure” test accepts any error. Failure while creating/inserting the deferred-FK fixture would satisfy it without reaching Commit. | `a124_settle_reads_the_committed_attempts_test.go:204–210,235–241`. | Assert successful fault setup and attempts read, then require the commit-specific error. Retain rollback/state/token checks. |

No additional production fail-open path was found in the normal, noncancelled D1/D8 outcome handling. Epoch implementation does not introduce an observed clear-order violation outside the frozen conservative exceptions.

**Mutation evidence:** all 30 ledger entries contain named test failures rather than build-failure-only output. Runner rejects build/setup failures, checks unique replacement anchors, and requires green controls (`mutate_a124.py:155–166,186–205`). M12 has a relevant behavioral kill through `judges_test:881–889`, not merely its timeout failure. This supports genuine mutant detection, but does **not** substitute for the missing scenarios/mutant above.

Structural scans use the correct repository roots: engine tests ascend three directories, execgw tests two. The production-projector scan also has a positive control (`a124_the_enforcement_boundary_internal_test.go:89–98`).

**3. Tasks 2.1–2.14 coverage**

| Task | Assessment |
|---|---|
| 2.1 | **Partial:** immediate transport failure and nil publisher covered; nonresponding publisher/timeout variant absent. `judges_test:231–253`. |
| 2.2 | **Covered:** acknowledgement clears latch, durable mode remains. `judges_test:259–275`. |
| 2.3 | **Partial:** fresh gate restoration tested against the same open journal; no actual close/reopen durability check. `judges_test:278–290`. |
| 2.4–2.5 | **Covered:** fresh-row priority and exhausted-row preservation. `judges_test:308–332`; journal selection test `:270–310`. |
| 2.6 | **Incomplete:** fixed margin and oversized-delay control are meaningful, but I2 applies. Lock test covers successful stop Notify while executor waits for gate; no remote-publish stall/emergency-Notify variant. `a124_the_executor_holds_no_stop_lock_testseams_test.go:19–71`. |
| 2.7 | **Mostly covered:** stale listing, rearm reset, error-before-outcome, three callers × two faults. Commit-phase qualification: I5. |
| 2.8 | **Partial:** outcome tables and actual mid-send acknowledgement covered; LeaseLost is injected rather than produced by rearm interleaving. `judges_test:347–389`. |
| 2.9 | **Incomplete:** basic before/after/window and escalation-failure cases covered; I3’s schedules and several required mode assertions absent. |
| 2.10 | **Incomplete:** core AA2 behavior covered; listing-clear ordering, positive complete-list pruning, same-ID carryover, and failed-ack preservation lack direct tests. `judges_test:545–706,831–875`. |
| 2.11 | **Partial:** content/account/raw-error sentinels and nonempty-log control present; no lease-token sentinel injection. `judges_test:710–738`. |
| 2.12 | **Partial:** outcomes, immediate error latch, live-lease suppression and later expiry covered; blocked epoch-read/batch-expiry schedules absent. `judges_test:744–791`. |
| 2.13 | **Mostly covered:** epoch/revision semantics and caller scan present; atomicity assertion weak, I4. |
| 2.14 | **Covered substantially:** production-builder boundary, mode-row assertions, controlled admission, and both relatch variants present. `a124_the_enforcement_boundary_internal_test.go:77–214`. |

VERDICT: REJECT — resolve the cancellation contract violation and complete the missing safety-path measurements and adversarial tests before landing.