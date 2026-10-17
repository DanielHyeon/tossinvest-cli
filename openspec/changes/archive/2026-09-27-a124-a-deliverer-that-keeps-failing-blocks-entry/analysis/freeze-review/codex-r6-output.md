**REJECT.** Revision 5 removes the executor’s Notifier-mutex hazard, but its replacement can still discard a required entry-blocking judgement.

Read-only review completed. No files changed; no tests, network calls, or engine commands run. **RESOLVED below means resolved in the proposal, not implementation verified.** `design.md`, `tasks.md`, and delta `spec.md` refer to `openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/`.

**Earlier findings**

| ID | Status | Evidence / reason |
|---|---|---|
| F1 | RESOLVED | Transaction-local attempts read explicitly added before commit; existing CAS supports it. `design.md:34–39`; `internal/journal/alert_claim.go:322–334`. |
| F2 | PARTIAL | Ordinary acknowledgement races are fenced; unconditional fallback can still relatch after acknowledgement. `design.md:224–231`; W3 below. |
| F3 | PARTIAL | Per-row counter covers failed recording, but aggregate listing-failure fencing remains unspecified. `design.md:289–317`; W2. |
| F4 | RESOLVED | Formula now includes intervening batch work; 214/316 seconds calculate correctly under H. `design.md:137–152`; `internal/app/engine/alertdelivery.go:122–161`. |
| F5 | RESOLVED | Selection guarantee is scoped; held-row and exhausted-tier starvation explicitly retained as limitations. `design.md:103–117`; `internal/journal/outbox.go:517–522`. |
| F6 | RESOLVED | Deployment escalation correctly conditional on selection and another failure; startup restores only the latch. `design.md:368–370`; `internal/app/engine/gateway.go:153–167`. |
| F7 | RESOLVED | OFF boundary corrected to engine refusal, not absence of Notifier. `internal/app/engine/gateway.go:323`; `cmd/tossctl/engine.go:220–221`. |
| F8 | RESOLVED | New details/logs have an explicit allowlist and escalation errors are classified without raw error text. `design.md:338–350`; `internal/obs/log.go:162–165`. |
| F9 | RESOLVED | First detail wins; repeated Block does not update it. `design.md:70`; `internal/execgw/retry.go:526–533`. |
| F10 | RESOLVED | Error-first classification specified, avoiding zero-valued `SettleApplied`. `design.md:41`; `internal/journal/alert_claim.go:109,324`. |
| F11 | RESOLVED | Single-row, matching-failure comparison corrected to 4/34 seconds. `design.md:78–89`; `internal/obs/notifier.go:428–433,525–528`. |
| F12 | RESOLVED | Shared-connection cost recorded and protection-path measurement assigned; measurement remains pending. `tasks.md:31–33`; `internal/journal/journal.go:174`. |
| F13 | RESOLVED | Missing-publisher delay explicitly accepted, not described as timing parity. `design.md:98–99`; `internal/obs/notifier.go:429–431`. |
| N1 | RESOLVED | Executor no longer holds `n.mu` over settlement, escalation, or logging. `design.md:178–180,275–285`. |
| N2 | PARTIAL | Successful row lookup handles acknowledged rows; bounded fallback weakens the unconditional acknowledgement promise. `design.md:228–231`; W3. |
| N3 | RESOLVED | Successful publish followed by failed settlement now triggers immediate latch/escalation while retaining lease. `design.md:56–68`; `internal/obs/notifier.go:465–491`. |
| N4 | RESOLVED | Per-row lifecycle, complete-list pruning, rearm carryover, and cancellation exclusion specified. `design.md:295–313`; `internal/journal/outbox.go:337–340`. |
| N5 | RESOLVED | General finite bound withdrawn; conditional assumptions and unbounded waits named. `design.md:128–166`. |
| N6 | PARTIAL | D1 correctly specifies NotFound → latch only; shared D7 helper escalates unconditionally. `design.md:53,224,231`; W4. |
| N7 | PARTIAL | Coverage substantially expanded, but aggregate-failure fencing and contradictory acknowledgement requirements remain. `tasks.md:38–56`; delta `spec.md:21–38`; W2–W3. |
| N8 | RESOLVED | Existing high-attempt rows do not inevitably create a mode row. `design.md:368–370`; `internal/journal/operating_mode.go:410–421`. |
| R1 | RESOLVED | No executor-owned Notifier mutex while waiting for journal. `design.md:178–180,284–285`; existing Notifier lock is `internal/obs/notifier.go:254–255`. |
| R2 | RESOLVED | Account/raw-error prohibition now applies to new logs; actual error-leak source verified. `design.md:343–347`; `internal/journal/operating_mode.go:394,449,470`. |
| R3 | RESOLVED | Other rows and zero-write settlement outcomes no longer reset a row’s counter. `design.md:297–299`; `internal/journal/alert_claim.go:337–365`. |
| R4 | RESOLVED | Old acknowledged row is checked by its own ID, not by another row’s presence. `design.md:227–228`; `internal/journal/outbox.go:543–556`. |
| R5 | PARTIAL | Lock-placement contradiction removed; new fallback/acknowledgement contradiction remains. Delta `spec.md:21–32`; W3. |
| R6 | RESOLVED | Lease protection explicitly expires; expired-lease resend is covered in tasks. `design.md:65–68`; `tasks.md:49–51`; `internal/journal/alert_claim.go:162–165`. |
| R7 | RESOLVED | Unsupported 54-second mutex bound removed; settlement checks token/state, not expiry. `design.md:163–166`; `internal/journal/outbox.go:472`. |
| R8 | RESOLVED | One increment per row per cycle restores the nominal two-interval minimum. `design.md:319`; `internal/app/engine/alertdelivery.go:133,154–161`. |
| Q1 | RESOLVED | Executor no longer bridges `n.mu` to broker-held `g.mu`. `design.md:180`; `internal/execgw/strategy_entry_gate_authority.go:60–72`. |
| Q2 | RESOLVED | Epoch changes atomically with Clear, not at acknowledgement entry. `design.md:189–194`; `internal/obs/notifier.go:870–875`. |
| Q3 | PARTIAL | Failed/partial acknowledgements preserve epochs, but a stale successful Clear plus later delivery can still discard a required judgement. W1. |
| Q4 | RESOLVED | Same-ID delivery/rearm without Clear conservatively preserves the old judgement. `design.md:261`; `internal/journal/outbox.go:334–342`. |
| Q5 | RESOLVED | Pruning rules and potentially unbounded retained entries documented. `design.md:300–307`. |
| Q6 | RESOLVED | H, escalation/listing costs, and replacement of a092’s double-counting formula are explicit. `design.md:137–170`; `tasks.md:80–82`. |
| Q7 | RESOLVED | F2 disposition now identifies B′ and superseded C. `design.md:387`. |
| V1 | RESOLVED | Pre/post-claim epoch bracket rejects overlapping Clear before publishing. `design.md:214–217`; claim commits at `internal/journal/alert_claim.go:299–302`. |
| V2 | PARTIAL | Pending-row recheck fixes the stated case, but treating DELIVERED as proof of acknowledgement leaves W1. `design.md:228,244–247`. |
| V3 | RESOLVED | Proposal/design/spec agree on two reads; GREEN task still contains stale “after claim” wording. `proposal.md:88`; `tasks.md:63`; W5. |
| V4 | RESOLVED | Spec permits bounded gate-state sections and forbids external work inside them. Delta `spec.md:29–32`; `internal/flatten/flatten.go:643`. |
| V5 | RESOLVED | Rearm carryover and listing-counter reset-before-increment explicitly defined. `design.md:309–313`; `tasks.md:43–46`. |
| V6 | RESOLVED | Final listing cost included; a092 formula replacement assigned. `design.md:145`; `tasks.md:81–82`. |

**New findings**

| ID | Severity | Finding | Evidence | Suggested fix |
|---|---|---|---|---|
| W1 | **P0** | **D7 can permanently discard a required judgement without acknowledgement of the affected row.** `DELIVERED` proves transport recovery, not the required subsequent human confirmation. | `design.md:228,246–247` drops any non-PENDING row. `internal/obs/notifier.go:870–875` separates count from Clear; `internal/execgw/replay.go:551` independently inserts; `internal/journal/outbox.go:453–456` delivers without acknowledging. Canonical requirement: `openspec/specs/engine-safety/spec.md:183`. | Require evidence that the relevant judgement/episode was covered by acknowledgement. Preserve the judgement when delivery alone is known. Add the interleaving below to tasks and mutation coverage. |
| W2 | **P1** | **Listing-failure judgement has no executable epoch-mismatch procedure.** D8 routes its aggregate counter through `latchConfirmed(row,e)`, but listing failure supplies no row ID. | `design.md:309–317` versus `design.md:222–228`; current failure returns before row iteration at `internal/app/engine/alertdelivery.go:150–154`; `LookupAlert` requires an ID at `internal/journal/outbox.go:543`. | Define a separate aggregate-failure witness/recheck policy, including concurrent Clear, recovered listing, new rows, and continued read failure. Specify deterministic tests; do not invent a sentinel row. |
| W3 | **P1** | **Unconditional fallback conflicts with “acknowledgement wins.”** After the third recheck, a final acknowledgement can settle the row and Clear; the unconditional Block then resurrects the latch and escalates without any new failure. | `design.md:227–231,248–249`; delta `spec.md:21–26` both forbids resurrection and requires unconditional fallback; `spec.md:74–75` promises no latch/escalation. Acknowledgement updates state at `internal/journal/outbox.go:494–499`. | Resolve the contract explicitly. Either preserve acknowledgement ordering with a bounded deferred decision, or document an approved fail-closed exception consistently. “Rare” does not prove the existing SHALL. |
| W4 | **P2** | **D7 helper loses D1’s latch-only verdict.** Both helper exits escalate, including failure-settlement NotFound/unknown outcomes. | `design.md:53–54` versus `224,231`; `tasks.md:37` expects no escalation. Actual parity: `internal/obs/notifier.go:519–523,309–314`. | Carry `shouldEscalate` through normal, refreshed-epoch, and fallback branches; test each latch-only outcome through the helper. |
| W5 | **P3** | **Residual text contradicts the corrected protocol.** D8 still asserts epoch change proves no PENDING row; epoch spec repeats the obsolete implication; GREEN task mentions only post-claim reading. | `design.md:293` contradicts `244–247`; delta `spec.md:113–114` contradicts `24–25`; `tasks.md:63` contradicts `38–42`. Also `design.md:355` says no new protection wait despite `275–279`. | Replace obsolete claims with the actual, qualified contracts; align GREEN task with both epoch reads. |

W1’s concrete schedule:

1. `Acknowledge` counts zero, then pauses before `Clear`.
2. An independent writer creates B. The executor claims B with equal pre/post epochs and eventually commits B’s third failed attempt.
3. Executor releases B’s lease, then pauses before applying its judgement.
4. The old acknowledgement finally calls `Clear`, increments the epoch, and returns. It never acknowledged B.
5. The synchronous sender claims B and successfully delivers it.
6. Executor’s fence fails; lookup returns `DELIVERED`; line 228 drops the judgement. **Neither latch nor escalation occurs, and B will not return in PendingAlerts.**

This combines cases ㉩ and ㉤. The case table therefore is not exhaustive. It also disproves the claim at `design.md:268–272` that every mistake is conservative.

**D7, D8, D1, and timing assessment**

The **epoch bracket itself is sound** for an acquired claim. Incrementing on **every Clear request is necessary**, including an absent latch: otherwise case ㉣ leaves no evidence of completed human acknowledgement. Preserving `revision`’s existing mutation-only behavior is correct (`internal/execgw/retry.go:529–542`).

I found no other production path clearing `ReasonAlertUndelivered`: Notifier’s two calls are at `internal/obs/notifier.go:846,875`; direct projection deletions target other reasons (`internal/execgw/modegate.go:37`, `symbolgate.go:184–190`). This supports the caller restriction, **not** the stronger claim that every epoch change covers every outstanding judgement.

The new executor lock order removes the prior `n.mu → g.mu → broker` propagation. Its gate sections perform only map operations; no remote publish occurs under its locks. Protection paths can still contend briefly on `g.mu`, and journal operations still share one connection (`internal/journal/journal.go:174`). Thus “no newly propagated remote wait” is justified; “no additional waiting whatsoever” is not. Planned task 2.6 remains necessary.

D8’s **per-row accounting is substantially repaired**: another row’s success cannot erase failures; release is not a successful attempt record; truncated listings cannot prune omitted rows; Run-context cancellation is excluded while transport-child timeout counts. Rearm carryover is conservative and documented. Complete-list pruning and epoch resets are reasonable lifecycle rules, subject to the explicitly recorded stale-count exception. The unresolved issues are W1’s dropped mature judgement and W2’s aggregate path.

D1’s additive read is implementable: use **`tx.QueryRowContext` after successful CAS and before commit**, return Attempts only after commit succeeds, and roll back on read failure. Using the journal’s pooled DB handle inside that transaction would wait on its own sole connection. NotFound parity is accurately described in D1; W4 concerns propagation through the proposed helper.

D6’s numerical examples are correct under H, with costs set to zero:

| Configuration | From first cycle start | Including maximum Q |
|---|---:|---:|
| One row, immediate failure | 4 s | 6 s |
| One row, 10 s timeout | 34 s | 46 s |
| Ten rows, all 10 s timeout | 214 s | 316 s |

These are conditional calculations, not measured operational deadlines. The code runs a complete sequential batch before sleeping (`internal/app/engine/alertdelivery.go:122–161`); transport supplies the timeout (`internal/obs/ntfy.go:95–100`).

**D2 and D3**

Both choices are adequately justified.

- **Limit 3:** reuses `DefaultCriticalAttempts` and preserves the established attempt budget (`internal/obs/notifier.go:40–48`). Timing parity applies only to the stated single-row examples.
- **Count missing publisher:** required to avoid permanent exemption from failure detection; supported by `internal/app/engine/exitwiring.go:62–69` and a092 delta `spec.md:76`. The accepted approximately 4–6-second delay differs from synchronous immediate blocking.

VERDICT: REJECT — D7 can discard a required judgement without human acknowledgement, and aggregate-failure fencing plus fallback acknowledgement semantics remain incomplete or contradictory.