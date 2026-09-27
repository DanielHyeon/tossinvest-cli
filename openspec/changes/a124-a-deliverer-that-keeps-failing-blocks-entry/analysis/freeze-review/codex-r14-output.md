Revision 8 is freezeable within its stated scope. It does **not** establish durable operating-mode enforcement or resolve W1’s protection gap. No files changed; no network calls, tests, or engine commands ran.

References: `D`, `T`, `S`, `P` mean this change’s `design.md`, `tasks.md`, delta `specs/engine-safety/spec.md`, and `proposal.md`. Code paths below are relative to `internal/` unless stated otherwise. **RESOLVED means resolved at proposal level**, not implemented or tested.

**1. Earlier findings**

| ID | Status | Evidence and reason |
|---|---|---|
| F1 | RESOLVED | D:42–49 specifies same-transaction attempts read before commit; current API lacks it (`journal/alert_claim.go:330–334`). |
| F2 | RESOLVED | D:224–281 distinguishes settlement, conditional latch, unconditional escalation, and escalation-failure fallback; permitted conservative exceptions are explicit. `obs/notifier.go:870–875` clears only the alert latch. |
| F3 | RESOLVED | D:343–380 covers claim/write/list errors with explicit counters; current failure paths merely log (`app/engine/alertdelivery.go:169–171,222–226`). |
| F4 | RESOLVED | D:149–170 includes intervening batch service and queue delay; sequential execution confirms 214/304/316-second examples (`app/engine/alertdelivery.go:122–161`). |
| F5 | RESOLVED | D:126–135 limits the fairness claim and records unbounded exhausted-row starvation, including missing-latch exceptions. Current selection is oldest-ID limited (`journal/outbox.go:517–529`). |
| F6 | RESOLVED | D:514–516 describes conditional deployment behavior; startup restores only pending backlog (`app/engine/gateway.go:153–167`). |
| F7 | RESOLVED | Correct OFF boundary is engine refusal, not absent notifier (`cmd/tossctl/engine.go:220–221`; `app/engine/gateway.go:323`). |
| F8 | RESOLVED | D:401–415 requires sanitized new details/logs and failure handling; raw transition errors can contain account references (`journal/operating_mode.go:394,449`). |
| F9 | RESOLVED | D:80 correctly describes insert-if-absent, preserving initial detail (`execgw/retry.go:526–533`). |
| F10 | RESOLVED | D:51 and T:41–42 require error-first handling; `SettleApplied` is zero (`journal/alert_claim.go:105–109`). |
| F11 | RESOLVED | D:88–99 compares identical failure patterns using actual defaults (`obs/notifier.go:45–48`; `obs/alert_lease.go:17`). |
| F12 | RESOLVED | T:33–40 requires missing-publisher and shared-connection measurements; journal has one connection (`journal/journal.go:174`). |
| F13 | RESOLVED | D:108–109 explicitly accepts approximately 0 versus 4–6 seconds; synchronous nil-publisher path breaks immediately (`obs/notifier.go:429–431`). |
| N1 | RESOLVED | D:259,269–281 keeps journal/escalation/logging outside gate lock; executor never acquires notifier mutex. |
| N2 | RESOLVED | D:284–287 explicitly retains error-obscured judgments conservatively instead of claiming exact acknowledgement recovery (`obs/notifier.go:864–875`). |
| N3 | RESOLVED | D:66–78 adds immediate delivered-settlement-failure handling matching synchronous behavior (`obs/notifier.go:465–491`). |
| N4 | RESOLVED | D:343–380 defines per-row lifecycle, pruning and listing counter; release success is excluded (`journal/alert_claim.go:368`). |
| N5 | RESOLVED | D:146–179 withdraws unconditional upper bounds and names assumptions/unbounded waits; executor publishes sequentially (`app/engine/alertdelivery.go:154–161`). |
| N6 | RESOLVED | Failed-attempt `NotFound` means latch only: `obs/notifier.go:519–523` returns lost; `:309–315` suppresses escalation. |
| N7 | RESOLVED | S:23–67, T:41–82 cover settlement outcomes, counters, exceptions, sanitation and production enforcement boundary. |
| N8 | RESOLVED | D:514–516 correctly requires selection and another failure; equal/stricter modes produce no transition (`journal/operating_mode.go:410–421`). |
| R1 | RESOLVED | B′ removes executor acquisition of `n.mu`; existing notifier holds it across delivery (`obs/notifier.go:254–255,309`). |
| R2 | RESOLVED | D:406–410 prohibits account/raw-error fields in new logs; account-bearing errors exist (`journal/operating_mode.go:394,449`). |
| R3 | RESOLVED | D:343–353 resets only the relevant row’s applied settlement, not unrelated success or zero-row results (`journal/alert_claim.go:337–364`). |
| R4 | RESOLVED | Global backlog fencing is removed; D:284–287 uses reason-clear epochs rather than inferring episode identity from `UndeliveredCount`. |
| R5 | RESOLVED | D:269–281, S:46–49 and T:90–91 consistently place escalation outside gate lock. |
| R6 | RESOLVED | D:76–78 and T:66–69 qualify resend suppression by lease survival; synchronous implementation also permits expiry (`obs/notifier.go:486–490`). |
| R7 | RESOLVED | D:177–183 withdraws the mutex-time bound; settlement checks state/token, not expiry (`journal/outbox.go:455,472`; `journal/alert_claim.go:349–364`). |
| R8 | RESOLVED | Per-row, once-per-cycle counting restores minimum two intervals; cycle sleeps after processing (`app/engine/alertdelivery.go:133,161`). |
| Q1 | RESOLVED | Executor avoids notifier mutex while waiting on gate; existing dispatch holds gate through callback (`execgw/strategy_entry_gate_authority.go:60–72`). |
| Q2 | RESOLVED | D:246–251 increments at `Clear`, matching actual acknowledgement release point (`obs/notifier.go:875`). |
| Q3 | RESOLVED | Partial/failed acknowledgement does not call `Clear` (`obs/notifier.go:864–875`); D:299 preserves judgments. |
| Q4 | RESOLVED | D:301 records conservative late application after delivery/rearm; delivery itself does not clear gate (`journal/outbox.go:449–456`). |
| Q5 | RESOLVED | D:348–356 records complete-list pruning and residual map growth; capped selection is not absence evidence (`journal/outbox.go:520–529`). |
| Q6 | RESOLVED | D:155–187 supplies assumptions, listing/escalation costs and replacement formula; T:111–112 carries it downstream. |
| Q7 | RESOLVED | D:537 now identifies B′, not the abandoned notifier-lock solution. |
| V1 | RESOLVED | D:304 and S:43–44 explicitly allow conservative relatching when an error hides prior acknowledgement; no timestamp inference. |
| V2 | PARTIAL | D:305,465–468 honestly retains count–clear race; independent enqueue can follow count before clear (`obs/notifier.go:870–875`; `execgw/replay.go:551`). |
| V3 | RESOLVED | P:87–92 and D:274–276 place epoch read after settlement; failure path additionally releases first. |
| V4 | RESOLVED | S:46–49 permits bounded gate-state critical sections, rather than claiming no shared lock (`execgw/retry.go:526–543`). |
| V5 | RESOLVED | D:358–380 explicitly accepts same-ID carryover and defines listing-counter reset; rearm resets durable attempts (`journal/outbox.go:336–340`). |
| V6 | RESOLVED | D:163 includes final listing cost; T:111–112 replaces the obsolete downstream formula. |
| W1(a) | RESOLVED | For specified ordering, on-time latch is erased by the same later `Clear`; delayed suppression loses no additional latch (`obs/notifier.go:875`; D:305–313). |
| W1(b) | NOT RESOLVED | B-specific alert protection can disappear without B acknowledgement; mode row supplies no production enforcement (`journal/operating_mode.go:475–476`; D:464–475). |
| W2 | RESOLVED | D:379–380 applies explicit epoch/counter transitions to listing failures without requiring a row. |
| W3 | RESOLVED | Deferred retry/exhaustion relatching removed; remaining unconditional fallback is specifically escalation-write failure (D:234–237; S:34–35). |
| W4 | RESOLVED | D:229 retains latch-only judgments through application, matching `obs/notifier.go:519–523,309–315`. |
| W5 | RESOLVED | Current epoch contract and implementation tasks agree (S:130–141; T:71–74,90–91). |
| X1 | RESOLVED | Deferred-judgment map removed (D:321–322); no stale epoch refresh/replay mechanism remains. |
| X2 | RESOLVED | D:307 handles delivered-then-human-clear by ordering; delivered rows cannot become acknowledged through pending-only update (`journal/outbox.go:494–503`). |
| X3 | RESOLVED | D:343–380 and S:53–58 enumerate counter cleanup and exclude release success. |
| X4 | RESOLVED | D:324–328 separates mutex isolation from shared-journal contention (`journal/journal.go:174`). |
| X5 | RESOLVED | Current deliverOne map correctly identifies B8 as log/release only (`app/engine/alertdelivery.go:205–212`); D:573 uses current disposition. |
| Y1 | RESOLVED | Escalation survives late clear as a **ledger state**, with enforcement claim withdrawn (D:207–208,458–459; `obs/notifier.go:840–878`). |
| Y2 | RESOLVED | Timestamp fencing removed; acknowledgement timestamp precedes SQL (`journal/outbox.go:485,494`). |
| Y3 | RESOLVED | S:56–58 contains all specified pruning exceptions; T:61–63 tests them. |
| Y4 | RESOLVED | T:33–40 separates isolation assertions from transaction-contention measurement against fixed allowance (`journal/journal.go:174`). |
| Y5 | RESOLVED | D:275,289–291 releases before epoch read on failed-send path and records retained-lease expiry on delivered-error path (`app/engine/alertdelivery.go:225,229–234`). |
| Y6 | RESOLVED | P:32–34 correctly states missing publisher leaves attempts unchanged; P:137–139 qualifies count–clear ordering. |
| Z1 | RESOLVED | D:365–377 makes threshold judgment precede epoch reset, preserving escalation. |
| Z2 | RESOLVED | D:234–237 mandates latch fallback when escalation fails, even after conditional rejection; S:34–35 agrees. |
| Z3 | RESOLVED | D:55–73,228–232 and S:31–35 agree on latch-only versus latch-plus-escalation outcomes. |
| Z4 | RESOLVED | T:41–42 requires SELECT/commit failure atomicity for all settlement callers (`journal/alert_claim.go:322–344`). |
| Z5 | RESOLVED | T:36–38 replaces self-adjusting allowance with fixed bound and deliberately failing measurement control. |
| AA1 | RESOLVED | Escalation failure relatches regardless of prior conditional result (D:234,281), covering clear between latch and failed write. |
| AA2 | RESOLVED | D:363–377 and T:59–62 agree on threshold-first transitions and intervening-clear outcomes. |
| AA3 | RESOLVED | T:37 uses existing fixed 250 ms allowance (`app/engine/a098_the_backlog_does_not_delay_protection_test.go:65`). |
| AA4 | RESOLVED | D:113–115 introduces separate selection API; existing callers retain `PendingAlerts(ctx, limit)` (`obs/notifier.go:855`; `app/engine/alertops.go:117`). |
| AB1 | RESOLVED | D:117–124 records backlog-dependent query cost; T:39–40 requires measurement and escalation if fixed allowance fails. |
| AB2 | RESOLVED | D:311–319 and S:34–44,83–106 expressly document conservative exceptions. |
| AC1 | RESOLVED | As proposal defect: D:419–482 and S:13–17 withdraw fictitious enforcement. Actual wiring gap remains downstream (`app/engine/gateway.go:249–274`). |
| AC2 | RESOLVED | As recording obligation: D:488–494 correctly identifies both latent defects; code still deletes/unlocks/reblocks (`execgw/modegate.go:35–50`) and projects after commit (`journal/operating_mode.go:468–476`). |
| AD1 | PARTIAL | D:438–461 and T:78–82 are corrected; residual broad summaries remain at D:106,505—AE1 below. |
| AD2 | RESOLVED | S:13–17 now concerns only mode-row enforcement and preserves reblock exceptions; T:81–82 tests both. |
| AD3 | PARTIAL | Detailed closers are exact at D:464–470; disposition summary D:578 still uses overly broad “closes”—AE2. |
| AD4 | RESOLVED | D:423–426 acknowledges transactional readers and supplies base coordinates; actual reads are `journal/operating_mode.go:398,676–685`. |

W1(b) remains **honestly unresolved** in D:474–475. The detailed closers now correctly say:

- **(i)** Atomic count–clear removes arrivals **between count and clear**. It does not cover B arriving after acknowledgement’s ID snapshot, failing, then being delivered before atomic count–clear (`obs/notifier.go:854–875`; `journal/outbox.go:453–456,494–503`).
- **(ii)** Projector wiring plus startup restore requires AC2 repair, restore before admission, successful escalation, and no approved relaxation (D:469–470). It supplies mode protection, not B-specific acknowledgement coverage.

**2. New findings**

| ID | Severity | Finding | Evidence | Suggested fix |
|---|---|---|---|---|
| AE1 | P2 | Narrowing is not literally complete everywhere: D3 still says not counting missing publisher permanently leaves entry open; safety summary still calls alert latch the sole production blocking mechanism. Governing D10 correctly limits this, so no new executable contradiction. | D:106,505 versus D:438–439; other blockers are enforced by `execgw/retry.go:568–590` and `execgw/symbolgate.go:229–250`. | Say “adds no delivery-failure latch” and “the only blocking mechanism **added by a124**.” |
| AE2 | P2 | W1 disposition summary still says count–clear repair **or** projector wiring closes protection, without the narrowed trajectories/prerequisites. Detailed D10 is correct. | D:578 versus D:465–470; snapshot/count/clear are distinct operations (`obs/notifier.go:854–875`). | Replace summary with “reduces specified trajectories under D10 prerequisites”; retain unresolved protection status. |
| AE3 | P3 | Restore logic-map commentary still says durable escalation means “nothing to restore.” That can recreate AC1’s mistaken inference even though branch enumeration is correct. | `analysis/function-logic/internal-app-engine--restorealertentrylatch/function-logic-map.md:37`; actual restore checks only backlog (`app/engine/gateway.go:153–167`), separate mode restore exists (`journal/operating_mode.go:574–588`). | State “mode row persists; gate projection is not restored here.” |

D10’s other-check list is correct **as examples, not an exhaustive admission checklist**: all account latches, query freshness, reconciliation refresh and symbol blocks remain authoritative. `parkAlert` sets unresolved-order latch before enqueue (`execgw/replay.go:534–551`); alert acknowledgement does not clear it.

The replacement spec sentence is testable and consistent with late-application exceptions. Production has no `SetModeProjector` or `RestoreOperatingModeProjection` callers; projection is optional (`journal/operating_mode.go:475–476`), and tracer supplies `ModeNormal` (`app/engine/tracer.go:473`). No fixture-created mode enforcement is needed.

T:75–82 is implementable:

- Inspect both reason keys through `EntryGate.Blocks()` (`execgw/retry.go:597–604`).
- Inspect persisted mode separately through `CurrentOperatingMode` (`journal/operating_mode.go:553–561`).
- Assert gate admission only with successful reconciliation refresh, fresh required observations and no other blockers. Production thresholds remain intact (`app/engine/gateway.go:249,272–274`).
- Both reblock variants can use the production-built gate/notifier with deterministic settlement/epoch/failure seams. The failed-escalation variant must not assert that the failed write created a mode row.

Existing production-builder test infrastructure supports this separation (`app/engine/a098_restart_does_not_release_the_gate_test.go:35–66,109–115`).

No judgment-rule change was identified in D1, D7’s action table/order, or D8’s transitions. They retain committed-attempt judgment, error-first handling, post-settlement epoch read, conditional latch, independent escalation, unconditional escalation-failure fallback, and threshold-before-reset counters. This is semantic verification against the recorded prior rules; this export does not provide revision snapshots for an independent byte-level diff.

**3. D2 and D3**

Both remain adequately justified.

- **D2:** Three uses the existing default (`obs/notifier.go:45`). Single-row immediate failures take approximately 4 seconds; three 10-second timeouts take approximately 34 seconds, excluding journal/gate costs. D6 correctly separates batching and queue delay.
- **D3:** Missing publisher already causes synchronous latch/escalation (`obs/notifier.go:429–431,570–573,223–228`), and a092 explicitly requires counting it (`a092…/specs/engine-safety/spec.md:76,108–110`). The accepted 4–6-second executor delay is disclosed. AE1 concerns wording, not this decision.

VERDICT: PASS — No P0/P1 remains in the scoped proposal; W1 protection remains explicitly unresolved, and implementation, contention measurements, and downstream mode enforcement remain unproven.