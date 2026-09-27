**REJECT: R1 remains partially unresolved. V5 can still pass without delivery or rearm.** No new production P0/P1 identified within this narrowed review.

Read-only inspection only. No tests, network calls, engine execution, or file changes. Mutation results are supplied evidence, not independently reproduced.

References below:

- `J`: `internal/app/engine/a124_the_deliverer_judges_internal_test.go`
- `L`: `internal/app/engine/a124_judging_does_not_delay_protection_test.go`
- `S`: `internal/app/engine/a124_the_executor_holds_no_stop_lock_testseams_test.go`
- `W`: `internal/app/engine/a124_the_worst_latch_time_internal_test.go`
- Document paths are relative to `openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/`.

| Review item | Assessment | Evidence |
|---|---|---|
| R1: ㉪ | **Resolved.** Hook execution, acquired claim, applied delivery, DELIVERED state before clear, and both final outcomes are checked. Removing the distinguishing delivery now fails. | `J:975–1000` |
| R1: failed acknowledgement | **Resolved for the requested schedule.** SQL trigger produces an actual acknowledgement failure; the first row’s ACKNOWLEDGED state establishes partial progress. Hook execution, unchanged epoch, latch and durable mode are asserted. | `J:1008–1045` |
| R1: failed acknowledgement preserves run | **Resolved.** Held O cannot supply a competing judgement; O’s acknowledgement and the returned failure are observed. Count is checked before/after acknowledgement; the next failure must latch and escalate. | `J:1053–1075` |
| R1: V5 | **Still deficient.** Both final outcomes are now checked, but successful claim acquisition does not distinguish rearm from expired-PENDING lease acquisition. | `J:1118–1132`; finding T1 below |
| R1: lease expiry during gate wait | **Lease-expiry schedule resolved; republication remains untested.** Hook arrival and `claim.Stole` establish expiry/reacquisition while the gate is held; gate and durable mode are checked after release. No second publish occurs. | `S:82–109` |
| R2 | **Resolved.** Clear invalidates the captured epoch; escalation attempt, NORMAL durable mode and fallback latch are independently asserted. | `J:1260–1278` |
| F4 | **Resolved for conditional-latch logging.** Logging follows actual insertion, rather than matching epoch or an existing latch. Existing-latch and fresh-latch cases expect zero/one lines. | `internal/app/engine/alertdelivery.go:411–423,462–472`; `J:1284–1307` |
| F7 | **Resolved.** Each successful measurement cancels and joins its executor before returning. Failure to stop marks the test failed. | `L:239–247,298–299` |

**Mutation evidence supports R2 and F4.** M30 adds the cancellation-specific escalation escape; M31 substitutes `applied` for `inserted`. Their transformations are at `analysis/harness/mutate_a124.py:164–170`; the corresponding named tests fail in `analysis/harness/mutation-ledger.tsv:38–39`. Neither establishes V5’s missing event.

**The API change preserves the stale-generation scenario.** `internal/execgw/retry.go:574–575` returns `(false, false)` before any latch/revision mutation. Matching epoch returns `(true, true)` only for insertion, otherwise `(true, false)` (`:577–582`). Thus the spec’s “returns false” remains true for `applied` (`specs/engine-safety/spec.md:151–153`). Later escalation/fallback belongs to the enclosing judgement, not this conditional operation.

Remaining findings:

| ID | Severity | Finding | Evidence | Suggested fix |
|---|---|---|---|---|
| T1 / R1 | **P1** | **V5 can pass with delivery/rearm absent.** Remove the delivery arrangement—or return `SettleLeaseLost, nil`—and the row remains PENDING. Advancing two hours expires its lease; `ClaimAlertForDelivery` still returns the same ID and `ClaimAcquired`. The third record failure then satisfies both final assertions. | `J:1118–1132`; PENDING explicitly means owed **without rearm** at `internal/journal/outbox.go:379–381`; acquisition follows at `:247–259`. | Require `SettleApplied` and persisted DELIVERED before advancing. After rearm, assert PENDING with reset settlement fields and `!again.Stole`; validate release outcome. Add a targeted mutant suppressing delivery. |
| T2 / R3 | **P2** | Overlap validation still covers only the selected median sample. Other samples influence median selection without proving equivalent overlap. Scaling checks an outbox write but not escalation; the negative control checks excessive duration without checking overlap. | `L:323–342,350–354,397–406` | Validate required workload/overlap for every sample before aggregation, including control. Require escalation overlap where claiming judgement-write coverage. |
| T3 / R4 | **P2** | Operational transaction verification remains unsupported by the prescribed benchmark. It measures average **claim + successful delivery** cost on a fresh temporary journal, not individual failed-attempt, release and escalation transactions or an operational journal copy. | `tasks.md:125–131`; `internal/app/engine/a098_cycle_cost_bench_test.go:37–43,120–143` | Retain the TMPDIR correction; provide separate timings for the specified transactions on a representative isolated copy, including tail durations. Do not describe this aggregate benchmark as directly verifying the ≈16 ms premise. |
| T4 | **P2** | The new D6 test does not satisfy task 4.4’s explicit “including queue wait” requirement. Disclosure is honest, but task closure remains unsupported. | `W:8–10,49–70`; `tasks.md:108` | Add controlled Q cases and expected totals, or obtain an explicit contract revision narrowing task 4.4. |
| T5 / R1 | **P2** | Lease theft proves another sender *can acquire*, not that republication is permitted and recorded as task 2.12 requires. | `S:95–109`; `tasks.md:71–74` | Publish through a recording publisher using the stolen claim; assert the second send and its settlement outcome. |

R3’s other fixes are valid: calibration now rejects values outside **½–2× target** (`L:213–218`), and per-sample executor shutdown removes the previous contamination. That tolerance validates approximate synthetic load, not actual transaction duration or a maximum-delay guarantee.

The new timing test honestly measures **simulated elapsed time from first cycle start to first latch**: publisher-driven fake-clock advances plus manually inserted inter-cycle waits. It checks 4/34/214 seconds (`W:39–41`). It excludes Q, real database/fsync cost, gate contention, logging/escalation time and scheduler delay; it calls `cycle` directly and manually models `Run`’s sleep (`W:56–62`). This is useful formula evidence, not a measured operational worst-case bound.

`tasks.md:113–114` correctly requires archive disclosure that §6 deployment/remeasurement remains unexecuted. That disclosure does not close T3 or task 4.4.

VERDICT: REJECT — R1’s P1 evidence gap remains because V5 still passes when its distinguishing delivery/rearm event is absent.