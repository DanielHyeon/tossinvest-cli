Freeze remains blocked. Revision 6 removes the executor’s Notifier-mutex hazard, but principle E does **not** preserve operating-mode state, and `acknowledged_at` cannot establish the ordering its error path requires.

Read-only review; no files changed, network calls, engine execution, or Go tests. “RESOLVED” below means resolved **in the proposal**, not implementation verified.

References: **D** = [design.md](/mnt/D/Axipient/workspace/TossOS/openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/design.md), **T** = [tasks.md](/mnt/D/Axipient/workspace/TossOS/openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/tasks.md), **S** = [delta spec.md](/mnt/D/Axipient/workspace/TossOS/openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/specs/engine-safety/spec.md). Code paths below are relative to `internal/`.

Earlier findings:

| Finding | Status | Reason and evidence |
|---|---|---|
| F1 | RESOLVED | D:30–36 specifies `tx.QueryRowContext` before commit; implementable at `journal/alert_claim.go:330`, under the existing CAS. |
| F2 | PARTIAL | Clear-epoch fencing addresses delayed latches, but dropping escalation violates claimed mode equivalence; Y1 below. D:181–183,283. |
| F3 | PARTIAL | D8 supplies per-row failure counting, but its final decision inherits the unsafe timestamp filter; Y2. D:314–316. |
| F4 | RESOLVED | D:140–149 includes intervening batch service; 214/316 seconds are correct under H. `app/engine/alertdelivery.go:122–136,154–162`. |
| F5 | PARTIAL | Exhausted-row displacement fixed; held-row and exhausted-tier starvation remain explicitly accepted limitations. D:100–114; `journal/outbox.go:518–522`. |
| F6 | RESOLVED | Deployment escalation now correctly conditional on selection and another failure. D:372–374; restoration only latches, `app/engine/gateway.go:153–167`. |
| F7 | RESOLVED | OFF explanation corrected: startup refuses automation OFF; Notifier construction is unconditional. `cmd/tossctl/engine.go:220–221`; `app/engine/gateway.go:323`. |
| F8 | RESOLVED | D:341–353 defines sanitized new details/logs and escalation failure handling; necessary because `obs/log.go:162–165` prints raw errors. |
| F9 | RESOLVED | First detail retained; no update on repeated Block. D:67; `execgw/retry.go:526–533`. |
| F10 | RESOLVED | Error checked before zero-valued outcome. D:38; `journal/alert_claim.go:109,323–324`. |
| F11 | RESOLVED | Like-for-like single-row timing corrected to 4/34 seconds. D:75–85; `obs/notifier.go:45–48,525–528`. |
| F12 | RESOLVED | Added journal contention recognized and measurement requested. D:375–376; T:31–33; shared connection at `journal/journal.go:174`. |
| F13 | RESOLVED | Missing-publisher delay difference explicitly accepted. D:95–96; synchronous immediate break at `obs/notifier.go:429–431`. |
| N1 | RESOLVED | Executor no longer holds `n.mu` around settlement, escalation, or logging. D:277–287; existing Notify mutex at `obs/notifier.go:254`. |
| N2 | PARTIAL | Error-hidden acknowledgment handled, but timestamp is not commit-order evidence; Y2. `journal/outbox.go:485–499`. |
| N3 | PARTIAL | Delivery-settlement failure policy and lease retention supplied, but erroneous acknowledgment suppression remains. D:53–65,237–240; Y2. |
| N4 | PARTIAL | Counter lifecycle specified in D:296–312, but claimed spec coverage of pruning exceptions is absent; Y3. |
| N5 | RESOLVED | General bound withdrawn; H and unbounded cases explicit. D:125–158. |
| N6 | RESOLVED | Failed-attempt NotFound means latch only: `deliver` returns `lost=true`, suppressing escalation. `obs/notifier.go:519–523,309–315`. |
| N7 | PARTIAL | Most policies covered; mode-equivalence contradiction, missing pruning exceptions, and test inconsistencies remain. S:21–42; Y1/Y3/Y4. |
| N8 | RESOLVED | Deployment statement no longer promises unconditional first-cycle escalation. D:372–374; `journal/operating_mode.go:410–421`. |
| R1 | RESOLVED | Shared Notifier exclusion removed entirely. D:119,277–287; exit/flatten Notify callers remain unchanged. |
| R2 | RESOLVED | New logs prohibit account references and raw errors. D:346–350; T:50–51. |
| R3 | RESOLVED | Per-row counters survive other-row success, non-applied settlements, and release success. D:298–301; T:44–45. |
| R4 | PARTIAL | Global backlog-count filter removed; replacement still permits incorrect suppression through pre-write timestamps. D:239,314–316; Y2. |
| R5 | RESOLVED | Escalation consistently outside gate lock. D:283–287; S:33–36; T:66–67. |
| R6 | RESOLVED | Suppression limited to a live, unreplaced lease; expiry explicitly tested. D:63–65; T:52–54; `journal/alert_claim.go:162–165`. |
| R7 | RESOLVED | Unsupported 54-second mutex bound withdrawn; expiry distinguished from token replacement. D:156–162; `journal/alert_claim.go:349–365`. |
| R8 | RESOLVED | Per-row counting restores at least two intervening cycle waits for three errors. D:322; `app/engine/alertdelivery.go:133,161`. |
| Q1 | RESOLVED | Executor does not nest `n.mu → g.mu`; existing synchronous hazard correctly remains a follow-up. D:277–287,380–381. |
| Q2 | RESOLVED | Epoch advances inside Clear, not at acknowledgment start. D:203–212; `execgw/retry.go:537–543`. |
| Q3 | RESOLVED | Failed/partial acknowledgment without Clear cannot reset the epoch. `obs/notifier.go:864–875`; D:298,318. |
| Q4 | RESOLVED | Post-settlement delivery/rearm without Clear retains the judgment; pre-settlement rearm loses the old token. D:259–261; `journal/outbox.go:334–340`. |
| Q5 | RESOLVED | Complete-list pruning and potentially unbounded retention under truncated lists documented. D:301,308–309. |
| Q6 | RESOLVED | H, listing/escalation costs, and replacement of a092’s double-counting formula specified. D:134–166; T:83–85. |
| Q7 | RESOLVED | F2 disposition now identifies B′ and escalation outside locks. D:391. |
| V1 | PARTIAL | Normal earlier acknowledgment covered; same-second conservatism explicit, but timestamp ordering remains unsound. D:239–249; Y2. |
| V2 | PARTIAL | Existing count/Clear race correctly identified; claimed full-state equivalence does not follow. `obs/notifier.go:870–875`; Y1. |
| V3 | RESOLVED | Main proposal/design now agree on post-settlement epoch capture. D:233–234; T:66. |
| V4 | RESOLVED | Contract permits gate-only critical sections and forbids external work inside them. S:33–36; `execgw/retry.go:526–543`. |
| V5 | RESOLVED | Same-ID carryover explicitly conservative; epoch reset precedes counter increment. D:298,311–319; T:46–49. |
| V6 | RESOLVED | Final listing cost included; a092 formula replacement task updated. D:142,164–166; T:84. |
| W1 | PARTIAL | Rejection holds for the alert-latch-only counterfactual, not operating mode or proof of actual human acknowledgment of B; Y1. |
| W2 | RESOLVED | Listing failure uses its own counter and epoch, without inventing a row lookup. D:318–320. |
| W3 | RESOLVED | Unconditional fallback block and retry loop removed. D:243,274–275. |
| W4 | RESOLVED | Latch-only decisions retain their no-escalation flag through application. D:50–51,242; T:37. |
| W5 | RESOLVED | Current fencing language uses E instead of the former pending-state retry rule. S:21–31; T:66–67. |
| X1 | RESOLVED | Deferred-judgment map removed; its fresh-epoch bypass no longer exists. D:274–275. |
| X2 | PARTIAL | Epoch detects delivery followed by manual Clear, but suppressing the accompanying durable escalation breaks equivalence. D:265; Y1. |
| X3 | PARTIAL | Deferred map removed and release exclusion specified; pruning exceptions still missing from spec. D:299–306; S:38–43; Y3. |
| X4 | PARTIAL | Design withdraws “exit dwell unchanged,” but task 2.6 still requires it under artificial settlement/escalation delay. D:279–281; T:31–33. |
| X5 | RESOLVED | V2 disposition and deliverOne FLM B8 corrected. D:427; FLM B8 now correctly records no attempt increment. Proposal still has separate stale claims; Y6. |

New findings:

| ID | Severity | Finding | Evidence | Suggested fix |
|---|---|---|---|---|
| Y1 | **P0** | **E’s gate-and-mode equivalence is impossible as written.** An on-time judgment escalates durably; a later acknowledgment clears only the alert latch. Revision 6 instead discards both latch and escalation. | D:181–183,242–243,283; S:21–31 versus S:59–62. `obs/notifier.go:870–875`; `journal/operating_mode.go:423–433,444–476`; `execgw/modegate.go:35–50`. | Separate alert-latch cancellation from durable mode obligation. Preserve escalation unless an appropriately ordered, explicit mode-relaxation decision supersedes it. Correct the equivalence oracle. |
| Y2 | **P1** | **`acknowledged_at < t_j` can suppress a required judgment without any Clear.** The timestamp precedes connection acquisition and acknowledgment completion. Sharing a clock does not repair this. Wall-clock rollback creates another inversion. | D:239–249,314–316; `journal/outbox.go:485,494–499`; single connection `journal/journal.go:174`; nonmonotonic wall clock `clock/clock.go:77–83`. | Use ordering evidence tied to the acknowledged episode and completed mutation, not wall-clock timestamp comparison. Until order is proven, do not suppress the judgment. |
| Y3 | **P1** | **D8 claims spec exceptions that do not exist.** Design deletes counters on ClaimSettled and absence from a complete listing; spec says only applied settlement or epoch change breaks continuity. | D:299–306; S:38–43. ClaimSettled source: `journal/alert_claim.go:263–265`; listing: `journal/outbox.go:517–529`. | Add both exceptions explicitly, including complete-versus-truncated listing rules and listing-counter success reset. Add a direct ClaimSettled-pruning test. |
| Y4 | **P1** | **Protection acceptance criteria contradict the design.** Task 2.6 requires unchanged exit dwell while executor settlement/escalation is artificially delayed. Holding the shared DB connection necessarily blocks an exit’s journal operation. A delay outside the transaction would miss this contention. | T:31–33 versus D:279–281; `journal/journal.go:174`; `obs/notifier.go:262`; `app/engine/exitloop.go:1710`. | Separate mutex/remote-wait isolation from journal-contention measurement. Specify actual transaction-held injection and an explicit baseline/acceptance criterion; retain “stop and report” on regression. |
| Y5 | **P2** | **Epoch-capture gap is not necessarily a tiny, lease-free window.** `ClearEpoch` can wait indefinitely for `g.mu`; failed-attempt settlement retains its lease until the later release. | D:162,234–235,269; `journal/outbox.go:462–473`; `execgw/strategy_entry_gate_authority.go:60–72`; `execgw/gateway.go:708`. | Document potentially unbounded capture latency and retained-lease expiry during that wait. Test Clear-before-capture and token replacement during the wait. |
| Y6 | **P3** | **Proposal retains two false statements:** current missing-publisher branch increments attempts; evidence fixed *after* Clear should be discarded. | `proposal.md:32–33,120–121`; actual branch `app/engine/alertdelivery.go:205–212`; contrary intended rule D:266. | Correct B8’s baseline behavior and reverse the mistaken temporal wording. |

The decisive counterexamples and test consequences:

- **Y1 — mode loss:** Start NORMAL. Third failure commits → capture epoch → another sender delivers → operator acknowledges empty backlog and clears → conditional block rejects → no escalation. Revision 6 ends NORMAL. The required on-time reference executes Block **and Escalate** before delivery/acknowledgment and ends ENTRY_BLOCKED. Acknowledge never relaxes that mode. Task **2.9** currently describes a reference containing only `Block`; implemented literally, it can falsely pass. Task **4.1**’s generic “remove escalation” mutation does not establish this delayed-path property. Test both projected gate reasons and persisted mode with no mode-relaxation approval.

- **Y2 — false earlier acknowledgment:** Acknowledge(A) captures timestamp at second 10, then stalls before its UPDATE completes. Executor’s delivery-settlement error returns at second 12. Acknowledgment commits afterward; B remains pending, so no Clear occurs. Lookup sees A acknowledged at second 10 and discards the judgment. An on-time judgment at second 12 would remain latched because partial acknowledgment does not Clear. Tasks **2.9/2.10** cover completed earlier acknowledgments and equal seconds, not this start-versus-completion inversion. Add this interleaving for immediate delivery errors and the third D8 error, plus clock rollback.

- **Epoch mutation coverage:** Task **4.1** claims case ㉬ kills moving epoch capture before settlement. The stated sequence “Clear → new B → settlement” need not kill that mutant: both captures can still occur after Clear. The test must force **mutant capture → Clear → settlement → correct capture**.

**W1’s rejected argument is only partly sound.** For `ReasonAlertUndelivered` alone, a stale Clear after evidence would erase an on-time latch too. That establishes equivalence to an existing race. It does **not** establish that the operator saw B: a092 `spec.md:66` ties release to acknowledged knowledge, while `:72` explicitly forbids the count/Clear gap. Nor does stale Clear erase an on-time operating-mode escalation. The race can remain an explicitly owned a092 defect, but it cannot justify dropping the durable consequence or claiming full-state equivalence.

D7’s current lock structure is otherwise sound: no executor `n.mu`, no remote publish under `g.mu`, no transaction held while acquiring the conditional-block lock. Mode projection occurs after commit (`journal/operating_mode.go:468–476`). Existing synchronous `n.mu → g.mu → broker-wait` exposure remains a separate, correctly recorded defect. Additional interleavings include capture contention, partial acknowledgment, acknowledgment timestamp capture before completion, and Clear between conditional Block and escalation; the old two-case argument is not exhaustive.

D8’s counter mechanics are substantially sound: unrelated successes cannot erase failures; non-applied settlements and successful release cannot reset them; complete-list pruning is reasonable; rearm carryover errs conservatively. Run-context cancellation exclusion is appropriate because shutdown restoration reads pending rows. Intermittent successful failed-attempt writes eventually advance D1’s persisted attempts (`journal/outbox.go:471`). The remaining safety defect is the final timestamp-based suppression, not the counter arithmetic.

D1’s additive `Attempts` read is implementable within the existing transaction. Read failure must roll back; commit failure must return error before any outcome interpretation. Failed-attempt NotFound correctly latches **without** escalation, while delivered-settlement NotFound escalates (`obs/notifier.go:452–491,519–523,309–315`).

D6’s conditional arithmetic checks out: with zero ancillary costs, `C=B·T+2`, giving **4/6**, **34/46**, and **214/316 seconds** for first-cycle/maximum-Q timing; the tenth row finishes its third attempt at **304 seconds**. These are conditional calculations, not measured operating bounds.

D2 and D3 are adequately justified. Limit **3** matches the wired synchronous default (`obs/notifier.go:45`, `app/engine/exitwiring.go:73–80`). Counting missing publisher closes a real permanent-open case and matches the existing intended direction (`exitwiring.go:62–70`). The accepted approximately 4–6-second missing-publisher delay must remain explicitly distinct from synchronous immediate blocking.

VERDICT: REJECT — mode-equivalence and acknowledgment-ordering defects remain, with spec and acceptance-test inconsistencies that must be fixed before freeze.