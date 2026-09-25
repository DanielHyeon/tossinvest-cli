**Freeze remains blocked:** D8 can erase a threshold crossing after it happened, losing the required escalation. The post-clear escalation-failure fallback is also incomplete, and contradictory instructions remain.

Read-only review; no files changed, tests run, network calls, or engine execution. “RESOLVED” below means resolved **in the proposal**, not implementation verified.

References: `D`, `T`, `S`, `P` denote this change’s `design.md`, `tasks.md`, `specs/engine-safety/spec.md`, and `proposal.md`. Source basenames refer to the supplied repository paths; ambiguous gateway references include their package.

Earlier findings:

| ID | Status | Reason / evidence |
|---|---|---|
| F1 | RESOLVED | D:30–36 specifies the missing additive result and same-transaction read before commit; current API indeed lacks attempts (`alert_claim.go:140–145,330–334`). |
| F2 | PARTIAL | Conditional latch application is sound with documented conservative exceptions; durable equivalence still fails in Z1/Z2 (`D:237–243,303,356`). |
| F3 | PARTIAL | Row/list counters cover otherwise unrecorded failures, but their epoch reset can erase a completed threshold crossing: Z1 (`D:303,322`). |
| F4 | RESOLVED | D:140–152 includes intervening batch service and final listing; consistent with sequential execution (`alertdelivery.go:122–136,154–162`). |
| F5 | RESOLVED | D:106–118 explicitly limits fairness and records held-row and exhausted-tier starvation; no universal fairness claim is justified. |
| F6 | RESOLVED | Deployment escalation is conditional on selection and another failure (`D:376–378`); startup restoration only latches (`engine/gateway.go:153–167`). |
| F7 | RESOLVED | Notifier construction is unconditional; OFF refuses runtime startup (`engine/gateway.go:323`, `cmd/tossctl/engine.go:220–221`). |
| F8 | PARTIAL | Sanitization is specified, but “escalation failure leaves the latch” is false after rejected conditional blocking: Z2 (`D:345–356`). |
| F9 | RESOLVED | First detail survives repeated blocking, correctly reflected in D (`retry.go:526–533`). |
| F10 | RESOLVED | Error-first handling explicitly prevents zero-value `SettleApplied` misuse (`D:38`, `alert_claim.go:109,318–334`; T:37–39). |
| F11 | RESOLVED | D:78–88 compares matching failure patterns: 4 s and 34 s, excluding ledger costs. |
| F12 | RESOLVED | Shared-connection overhead is acknowledged and measurement required (`journal.go:174`, T:31–35); acceptance-budget caveat remains Z5. |
| F13 | RESOLVED | Missing-publisher delay difference is recorded (`D:90–98`, `notifier.go:429–431,570–573`). |
| N1 | RESOLVED | Executor no longer borrows `n.mu`; ledger/escalation/logging stay outside gate critical sections (`D:234–243,283–291`). |
| N2 | PARTIAL | Error outcomes now deliberately block conservatively, but proposal still prescribes the rejected timestamp suppression: Z3 (`P:89–90`, `D:246–249`). |
| N3 | PARTIAL | Delivered-but-unrecorded outcomes are covered; D:62 still conditions both actions on the fence, contradicting D:243: Z3. |
| N4 | PARTIAL | Row identity, rearming, pruning and release exclusions are specified; threshold/reset ordering remains broken: Z1 (`D:303–323`). |
| N5 | RESOLVED | D:127–164 explicitly rejects a general bound and names required assumptions and unbounded terms. |
| N6 | RESOLVED | Failed-attempt `NotFound` means latch only: `notifier.go:519–523` returns `lost`, and :309–315 suppresses escalation. |
| N7 | PARTIAL | Coverage expanded substantially, but Z1/Z2 failure sequences and Z3 normative contradictions remain (`T:40–59,78–83`). |
| N8 | RESOLVED | D:376–378 no longer promises unconditional first-cycle escalation. |
| R1 | RESOLVED | Executor-held `n.mu` is removed; shared ledger contention remains separately acknowledged (`D:283–291`, `journal.go:174`). |
| R2 | RESOLVED | New logs prohibit account/raw errors and require sentinels (`D:349–354`, T:56–57); transition errors really can contain accounts (`operating_mode.go:394,449`). |
| R3 | RESOLVED | Other-row successes, `LeaseLost`, `AlreadySettled`, and release success cannot reset a row counter (`D:303–306`, T:49–50). |
| R4 | RESOLVED | Global backlog reread is removed; error outcomes use explicit conservative policy rather than treating another row as evidence (`D:246–249`). |
| R5 | PARTIAL | Lock placement is aligned, but escalation conditions remain contradictory: Z3 (`D:62,243`, T:72–73). |
| R6 | RESOLVED | Suppression lasts only while the lease survives without replacement; expiry/republication tests are required (`D:66–69,251–253`, T:58–61). |
| R7 | RESOLVED | Unsupported 54 s mutex bound is removed; settlement checks token/state, not expiry (`outbox.go:455,472`, `alert_claim.go:349–365`). |
| R8 | RESOLVED | Per-row counting prevents three different rows exhausting one counter in one batch (`D:303,326`, `alertdelivery.go:154–162`). |
| Q1 | RESOLVED | Removing executor ownership of `n.mu` breaks the new transitive wait; existing synchronous defect remains (`notifier.go:254,571`, `strategy_entry_gate_authority.go:60–72`). |
| Q2 | RESOLVED | Epoch increment belongs inside `Clear`’s gate lock, not acknowledgement entry (`D:207–220`, `retry.go:537–543`). |
| Q3 | PARTIAL | Failed/partial acknowledgements do not advance epochs; a real clear can still erase an already-earned escalation through D8: Z1. |
| Q4 | RESOLVED | Post-settlement delivery/rearm does not invalidate a confirmed historical failure; pre-settlement rearm invalidates the token (`outbox.go:334–341`, `alert_claim.go:357–365`). |
| Q5 | RESOLVED | Complete-list pruning and potentially unbounded retained entries under perpetual truncation are explicitly recorded (`D:305–314`). |
| Q6 | RESOLVED | H and final-listing cost are included; T:90–91 requires replacing, not reinterpreting, a092’s formula. |
| Q7 | RESOLVED | F2 disposition now identifies B′ and lock-free notifier interaction (`D:395`). |
| V1 | RESOLVED | Approval-hidden settlement errors intentionally produce conservative blocking; no unsafe state-based suppression remains in D7 (`D:246–249,267`). |
| V2 | PARTIAL | Count/clear race is correctly identified as existing; retained escalation makes the proposed disposition defensible only when escalation succeeds: Z2 (`notifier.go:870–875`). |
| V3 | PARTIAL | Proposal remains stale about approval timestamps and “immediate” epoch capture versus release-first order (`P:88–90`, `D:238–239`). |
| V4 | RESOLVED | Gate-map critical sections are allowed; waiting on external work inside them is forbidden (`S:36–39`, `retry.go:526–543`). |
| V5 | PARTIAL | Same-id conservative carryover is explicit; reset-before-increment loses a threshold in Z1 (`D:303,316–323`). |
| V6 | RESOLVED | Final `I_list` and conditional-bound language are present (`D:141–152`, T:90–91). |
| W1 | PARTIAL | Rejection argument holds for the existing count/clear interleaving with successful escalation; it neither fixes that race nor establishes unconditional durable preservation: Z2. |
| W2 | PARTIAL | Listing counter no longer requires a nonexistent row, but its reset has Z1’s same lost-threshold window (`D:322–324`). |
| W3 | RESOLVED | Bounded retries followed by unconditional blocking are removed (`D:278–280`). |
| W4 | RESOLVED | D1 distinguishes latch-only outcomes and T:39 explicitly preserves no-escalation behavior; spec wording needs Z3 qualification. |
| W5 | PARTIAL | Old machinery is removed, but conflicting policy sentences survive (`P:89–90`, `D:62,164,337`). |
| X1 | RESOLVED | Deferred-judgement map and its incorrect fresh-epoch replay are removed (`D:278–280`). |
| X2 | RESOLVED | Delivered-then-human-clear is handled through epochs, not unreachable `ACKNOWLEDGED` state (`outbox.go:494–499`, `D:270`). |
| X3 | RESOLVED | Deferred map removed; row-counter pruning and release exclusions are enumerated (`D:303–314`, `S:43–48`). |
| X4 | RESOLVED | D:283–286 distinguishes mutex isolation from shared-connection cost; no unchanged-exit-duration claim remains. |
| X5 | RESOLVED | Base B8 correctly means logging plus release, without attempts increment (`alertdelivery.go:205–212`; deliverOne FLM B8). |
| Y1 | PARTIAL | D7 preserves escalation after fence rejection, but D8 can avoid creating that judgement and D9 can lose its fallback: Z1/Z2. |
| Y2 | PARTIAL | D7/spec reject timestamps correctly; proposal still requires them (`D:246–249`, `S:29`, `P:89–90`). |
| Y3 | RESOLVED | All stated pruning exceptions now appear in S:43–48 and T:49–54. |
| Y4 | PARTIAL | Isolation and contention tests are separated, but a whole exit cycle’s “one transaction” allowance lacks justification: Z5 (`T:31–35`). |
| Y5 | RESOLVED | Failed-send release precedes epoch acquisition; retained-lease success-path expiry is recorded (`D:238,251–253`). |
| Y6 | RESOLVED | B8 baseline and follow-up temporal direction are corrected (`P:34–35,118–121`, `alertdelivery.go:205–212`). |

New findings:

| ID | Severity | Finding | Evidence | Suggested fix |
|---|---|---|---|---|
| Z1 | **P1** | **Epoch reset can erase a threshold already reached.** Counter `(2,e0)` → third error returns at `t_j` → human clear advances to `e1` → delayed epoch read sees `e1` → reset to 0, increment to 1. Neither blocking nor escalation is attempted. On-time application would have escalated before that clear, and acknowledgement cannot remove that mode. A subsequent successful listing/settled row can erase the remaining counter permanently. Applies to both row and listing counters. | `D:199–200,303,319–324`; acknowledgement only clears the latch (`notifier.go:870–875`); mode relaxation requires approval (`operating_mode.go:423–429`). | Separate confirmed threshold obligations from future counter resets. An ambiguously timed clear must not discard an already-earned escalation. Specify ordering without timestamps or holding locks across ledger work. |
| Z2 | **P1** | **D9’s failure fallback disappears when conditional blocking returns false.** Threshold → capture epoch → deliver/acknowledge and clear → fence rejects → escalation write fails. No alert latch remains; no durable mode was written. The settled row will not be selected again, and startup sees zero pending. “Memory latch remains; restart covers it” is false. | `D:241–243,356`; transition can fail before durability (`operating_mode.go:391–400,444–470`); listing selects only PENDING (`outbox.go:518`); startup does nothing at zero (`engine/gateway.go:158–159`). | Define preservation/retry and fail-closed handling for a confirmed escalation whose latch has been cleared. Explicitly cover failure after fence rejection and settled-row disappearance; do not claim restart recovery without durable evidence. |
| Z3 | **P1** | **Freeze documents still prescribe incompatible safety behavior.** Proposal requires rejected timestamp suppression; D1 conditions delivered-error escalation on fence success; D7 requires escalation irrespective of it. Spec E unconditionally says both actions, whereas failed-attempt `NotFound` must remain latch-only. | `P:89–90`; `D:49–50,62,243`; `S:20–26`; `T:39,72–73`; actual latch-only behavior: `notifier.go:519–523,309–315`. | Align proposal, D1, D7 and spec around an explicit outcome/action table. Apply E separately to whichever actions that outcome actually requires. |
| Z4 | P2 | Additive transaction read lacks an explicit failure-atomicity test obligation. A read failure after UPDATE must roll back attempts/state/lease; commit failure must not expose usable attempts. “Error result is zero” alone does not prove rollback. | `D:30–36`; `alert_claim.go:320–334`; `T:37–38,71`. | Add transaction-level injected SELECT/commit failure tests, asserting unchanged stored state and unusable result; retain all-three-caller coverage. |
| Z5 | P2 | A single connection guarantees serialization, not that a whole exit cycle encounters only one competing transaction. The proposed acceptance line may be an arbitrary budget rather than a derived bound. | `T:33–35`; `journal.go:174`; delivery performs claim/settlement/release (`alertdelivery.go:168,222,225`), and adds escalation. | Identify the exact measured operation and contention opportunities; test repeated interleaving across the cycle and use an explicitly justified protection budget. |

For **Z1**, task 2.10’s “third error then clear” expectation would catch the defect **if the clear is forced before the first post-error epoch read**. Current wording does not require that placement. A test clearing after the read passes. Add both row/list schedules:

`count=2 → third error returns → Clear → epoch read/reset → judgement`

Tasks 2.9 and 4.1 primarily exercise an already-created judgement or moving the read before settlement; neither substitutes for this schedule. Add a mutation that resets the counter before preserving the threshold obligation.

For **Z2**, tasks 2.9’s successful durable-state comparisons and 2.11’s sanitized-error checks do not require escalation failure after fence rejection. Add that exact schedule, then successful subsequent cycles and restart.

D7’s current lock design is materially improved. Executor publication holds neither `n.mu` nor `g.mu`; gate operations contain only map work. The old notifier-mutex borrowing argument no longer applies. Exit/flatten can still contend on the shared journal connection, and on short gate critical sections. Existing synchronous `n.mu → g.mu → broker-wait` exposure remains an activation prerequisite, not fixed here (`notifier.go:254,571`; `strategy_entry_gate_authority.go:60–72`; `execgw/gateway.go:708`; `flatten.go:643,694`).

**E is a qualified extension of “acknowledgement wins,” not exact equivalence in every execution.** For a fixed D1 judgement, a clear after epoch capture safely suppresses its alert latch. A clear between settlement and capture can cause conservative relatching, explicitly allowed by S:31–34. Error outcomes can also relatch an acknowledged row. Timestamp comparison cannot establish completion order and is correctly removed from D7. The two simple before/after cases therefore are not exhaustive; D8’s reset and escalation failure introduce non-conservative losses.

**W1’s rejection is defensible only within that boundary.** The old count/clear race would erase an on-time alert latch too; a092:72 already requires closing it. Successful escalation preserves the operating-mode block because acknowledgement does not relax modes. Calling this unconditional preservation is unsupported while Z2 remains.

**D1, D2, D3 and D6:** the same-transaction attempts read is implementable without a second connection; error-first handling and `NotFound` latch-only parity are correct. Limit **3** is adequately justified by existing defaults and matching single-row timings. Counting missing publisher is justified by `exitwiring.go:62–70` and a092:76; its slower detection is explicitly disclosed. D6’s conditional figures check out: **4/6 s**, **34/46 s**, **214/316 s**, with the tenth row at **304 s**, under its zero-overhead assumptions. These are calculated examples, not measured operational bounds.

The counter-reset race, missing post-clear escalation-failure policy, and contradictory safety instructions must be resolved before freeze.

VERDICT: REJECT