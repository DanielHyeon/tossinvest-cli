**Freeze is blocked by missing production operating-mode enforcement.** D7 preserves a durable `ENTRY_BLOCKED` row, but production never binds its projector or restores it at startup. Consequently, suppressing or acknowledging the alert latch can leave entry allowed without mode-relaxation approval.

Static review only. No files changed, tests executed, network calls, or engine execution.

References below use:

- `D`, `T`, `Δ`: this change’s `design.md`, `tasks.md`, `specs/engine-safety/spec.md`.
- `AD`: `internal/app/engine/alertdelivery.go`
- `N`: `internal/obs/notifier.go`
- `J`: `internal/journal/alert_claim.go`
- `O`: `internal/journal/outbox.go`
- `M`: `internal/journal/operating_mode.go`
- `G`: `internal/execgw/retry.go`

**Earlier findings**

“RESOLVED” means addressed adequately in the proposal, not implemented or tested.

| ID | Status | Reason and evidence |
|---|---|---|
| F1 | RESOLVED | D:37–44 specifies transaction-local attempts read before commit; existing applied return lacks it, J:330–334. |
| F2 | PARTIAL | Epoch fencing repairs latch ordering, but preserved escalation is unenforced in production; D:260–272, AC1 below. |
| F3 | RESOLVED | Per-row record/claim-error counter and list-error counter cover failures that cannot increment attempts; D:334–371; AD:168–171,222–225. |
| F4 | RESOLVED | Serial batch service included; 214 s and 316 s arithmetic correct under H; D:150–165; AD:122–161. |
| F5 | RESOLVED | Remaining starvation cases explicitly distinguished from under-limit prioritization; D:123–130; AD:178–192. |
| F6 | RESOLVED | Deployment escalation correctly conditional on selection and another failure; D:425–427; engine/gateway.go:153–167. |
| F7 | RESOLVED | OFF boundary correctly identifies startup refusal, not absent Notifier; cmd/tossctl/engine.go:220–221. |
| F8 | RESOLVED | Fixed details, sanitized new logs, escalation-failure fallback specified; D:390–406; obs/log.go:162–165. |
| F9 | RESOLVED | First detail retained; G:526–533; D:75. |
| F10 | RESOLVED | Error checked before zero-valued `SettleApplied`; D:46; J:108–109,322–324. |
| F11 | RESOLVED | Like-for-like single-row timings replace mixed failure comparison; D:83–94; N:428–433,525–528. |
| F12 | RESOLVED | Additional shared-connection cost becomes explicit measurement obligation; T:31–38; journal/journal.go:174. |
| F13 | RESOLVED | Missing-publisher latency difference acknowledged; D:103–104; N:429–431,570–573. |
| N1 | RESOLVED | Executor no longer takes `n.mu`; database and logging work remain outside gate lock; D:260–272,315–325. |
| N2 | RESOLVED | Error outcomes now conservatively retain judgments; approved exceptions are explicit; D:275–278; Δ:34–38. |
| N3 | RESOLVED | Published-but-unsettled errors immediately block/escalate and retain lease; D:61–73; N:465–491. |
| N4 | RESOLVED | Counter lifetime, rearm carryover, pruning and release exclusions specified; D:334–350; O:334–340. |
| N5 | RESOLVED | General upper-bound claim withdrawn; H and unbounded cases explicit; D:141–174. |
| N6 | RESOLVED | Failed-attempt `NotFound` blocks without escalation; N:519–523 → N:309–315; D:58. |
| N7 | PARTIAL | Failure policies covered, but production enforcement/startup integration missing from tasks; T:25–28,43–61; AC1. |
| N8 | RESOLVED | Deployment wording no longer promises unconditional first-cycle escalation; D:425–427. |
| R1 | RESOLVED | No executor-held Notifier mutex across journal waits; D:315–325; N:254–255,851–852 establish old contention. |
| R2 | RESOLVED | Account/raw-error exclusion applies to new logs; D:397–401; M:393–394,448–449 show necessity. |
| R3 | RESOLVED | Other rows and zero-write settlements cannot reset affected row’s counter; D:334–344; J:337–365. |
| R4 | RESOLVED | Global pending-count fence removed; clear epochs replace it; D:275–278,352–368. |
| R5 | RESOLVED | Design/spec/tasks consistently prohibit journal/log work inside gate lock; Δ:40–43; T:79–80. |
| R6 | RESOLVED | Duplicate suppression limited to live, unreplaced lease; D:70–73; J:162–165; O:455. |
| R7 | RESOLVED | Lease expiry alone does not invalidate token CAS; D:176–178; J:349–365. |
| R8 | RESOLVED | Per-row counting restores minimum two inter-cycle waits; D:373; AD:133,154–161. |
| Q1 | RESOLVED | Removing executor `n.mu` eliminates its nested `n.mu → g.mu` wait; D:315–325; strategy_entry_gate_authority.go:60–72. |
| Q2 | RESOLVED | Epoch increments at `Clear`, not acknowledgement start; D:237–242; N:851–875. |
| Q3 | RESOLVED | Failed/partial acknowledgements do not call `Clear`; N:864–875; T:51–56,69–72. |
| Q4 | RESOLVED | Old judgment survives delivery/rearm without clear, conservatively; D:292; O:334–340. |
| Q5 | RESOLVED | Complete-list pruning and truncated-list retention explicitly specified; D:339,346–350. |
| Q6 | RESOLVED | H, list/escalation costs and replacement a092 formula specified; D:150–182; T:97–99. |
| Q7 | RESOLVED | F2 disposition now identifies B′; D:444. |
| V1 | RESOLVED | Superseded by explicit conservative error policy; no unreliable acknowledgement-state suppression; D:275–278,295. |
| V2 | PARTIAL | Count–clear race correctly identified as existing, but reliance on surviving mode enforcement fails; N:870–875; D:296; AC1. |
| V3 | RESOLVED | Proposal/design/tasks align on post-settlement read, after release on failure path; D:265–267; T:79. |
| V4 | RESOLVED | Contract permits bounded gate operations rather than falsely forbidding all shared locks; Δ:40–43; G:526–543. |
| V5 | RESOLVED | Rearm carryover and list-counter reset semantics explicit; D:349–371; T:59–61. |
| V6 | RESOLVED | Final listing cost included; conditional formula and handoff corrected; D:156–182; T:98–99. |
| W1 | PARTIAL | Rejection holds for latch ordering, not claimed operational protection: durable mode lacks production enforcement; D:297; AC1. |
| W2 | RESOLVED | List counter uses explicit epoch transition rules without row lookup; D:352–371. |
| W3 | RESOLVED | Retry-count-based unconditional block removed; remaining fallback is specifically escalation-write failure; D:226–228,312–313. |
| W4 | RESOLVED | Lock-only judgment preserved through application; D:221; T:42,52; N:519–523. |
| W5 | RESOLVED | Current normative epoch contract aligned; Δ:124–135; T:69–80. |
| X1 | RESOLVED | Deferred-judgment map removed; D:312–313. |
| X2 | RESOLVED | Clear ordering handles delivered-then-acknowledged rows without requiring impossible ACK state; O:494–503; D:298. |
| X3 | RESOLVED | Deferred map removed; all counter pruning exceptions enumerated; D:334–344; Δ:45–52. |
| X4 | RESOLVED | Shared journal contention separated from mutex isolation; D:315–319; journal/journal.go:174. |
| X5 | RESOLVED | Current disposition and B8 map correctly describe missing-publisher baseline; AD:205–212. |
| Y1 | PARTIAL | Escalation is no longer discarded, but row persistence alone does not enforce entry blocking; D:220–224; AC1. |
| Y2 | RESOLVED | `acknowledged_at` comparison removed; timestamp precedes database execution, O:485,494; D:275–278. |
| Y3 | RESOLVED | Spec includes settled-claim, complete-list and list-success pruning; Δ:45–52. |
| Y4 | RESOLVED | Isolation assertions separated from transaction-contention measurements; T:31–38. |
| Y5 | RESOLVED | Failure path releases before epoch wait; retained successful-publish lease and expiry limitation recorded; D:266,280–282. |
| Y6 | RESOLVED | Proposal correctly describes B8 and count–clear chronology; AD:205–212; N:870–875. |
| Z1 | RESOLVED | Threshold decision precedes epoch reset; escalation retained; D:352–368; T:57–60. |
| Z2 | RESOLVED | Escalation failure unconditionally restores alert block; D:226–228; T:53. |
| Z3 | RESOLVED | Judgment/action table preserves lock-only `NotFound` and unconditional applicable escalation; D:218–224; Δ:25–29. |
| Z4 | RESOLVED | SELECT/commit failure atomicity required for all three settlement callers; D:37–41; T:39–40. |
| Z5 | RESOLVED | Self-expanding tolerance replaced by fixed baseline margin and failing control; T:34–36. |
| AA1 | RESOLVED | Fallback no longer depends on earlier conditional-block result; D:226–228; T:53. |
| AA2 | RESOLVED | Threshold-before-reset transition and three expected outcomes explicit; D:352–363; T:57–59. |
| AA3 | RESOLVED | Fixed 250 ms margin reused; T:35; a098_the_backlog_does_not_delay_protection_test.go:61–65. |
| AA4 | RESOLVED | Additive delivery-list method preserves existing callers; D:108–110; N:855; O:517–529. |
| AB1 | RESOLVED | Full-backlog query cost recognized; fixed-budget size measurements and stop condition required; D:112–119; T:37–38. |
| AB2 | RESOLVED | Conservative epoch-read-window exception explicit in normative text, scenarios and tests; D:302–310; Δ:36–38,86–100; T:44–45. |

**New findings**

| ID | Severity | Finding | Evidence | Suggested fix |
|---|---|---|---|---|
| AC1 | **P0** | **Durable escalation is not connected to production entry enforcement or startup restoration.** `ClearEpoch → acknowledgement/Clear → conditional block rejected → successful escalation` can finish with no alert latch and no mode latch. Restart with delivered/acknowledged backlog likewise restores neither, despite durable `ENTRY_BLOCKED`. | Repository-wide non-test search finds **zero callers** of `SetModeProjector` and `RestoreOperatingModeProjection`. Engine constructs its gate at `internal/app/engine/gateway.go:249`, restores only alert backlog at `:269`, and passes that gate to gateway at `:302`. Projection is optional at `M:475–476`; enforcement reads gate latches at `G:565–593`, reached through `internal/execgw/gateway.go:855–859`. Existing tests supply missing wiring themselves: `internal/obs/escalation_test.go:32`, `internal/execgw/modegate_test.go:66,213`. | Make production projector binding and startup restoration part of this change or an explicit landing prerequisite. Require production-builder tests: after acknowledgement, `CheckEntryFor` must still reject because of mode; repeat after restart with **zero PENDING rows**. Do not manually bind/restore only in the test fixture. |
| AC2 | **P2** | **The existing mode projector has additional concurrency gaps.** Binding it does not establish atomic or ordered projection: it removes the mode latch before reacquiring the lock to add it, and concurrent post-commit projections can arrive out of commit order. This is an existing dependency defect, not introduced by a124. | `internal/execgw/modegate.go:35–50`; `M:468–476`. Neither design nor T:43–61 exercises concurrent projection or an entry check between delete and replacement. | Record a separate prerequisite/follow-up for atomic, ordered projection. Add deterministic entry-check and stale-projection tests; preserve the prohibition on holding locks across remote calls. |

**D7, D8 and principle E**

- **Mutex borrowing is gone.** Executor acquires only `g.mu` for local gate operations. No new executor `n.mu → journal/g.mu` chain is specified. Exit still shares the journal connection; T:2.6 appropriately measures that separately. Existing synchronous remote-lock problems remain outside this change.
- **Epoch logic is conservative for latch ordering.** A clear observed after the sampled epoch necessarily follows settlement; a clear between settlement and sampling can cause extra blocking. That is the explicitly permitted approximation, not exact equivalence.
- **Acknowledgement timestamps are no longer used.** Removing them is correct: `O:485` captures time before `O:494` executes the update.
- **E is an extension of the acknowledgement rule, not a complete safety proof.** W1’s late clear would also remove an on-time alert latch. However, “the mode remains” protects trading only when that mode is actually enforced. AC1 invalidates that premise.
- **D8’s counter rules are coherent.** Other-row success, release success and zero-write settlements cannot erase failures. Successful failed-attempt recording resets the transient counter but advances durable attempts. Threshold-before-reset conservatively preserves escalation. Cancellation exclusion is specifically the executor’s lifetime context.
- **Tests 2.9/2.10/4.1 cover the named epoch/counter mutations, but can miss AC1.** Database mode plus alert-reason assertions can both pass while entry remains allowed. Task 2.3’s PENDING restart also masks missing mode restoration because backlog restoration independently blocks entry.

**D1 and D6**

D1’s additive read is implementable using the existing transaction; using `j.db` would wait for its own occupied connection. Failed-attempt `NotFound → latch only` matches `N:519–523 → N:309–315`; delivered-settlement failure correctly takes the separate block-and-escalate path.

D6’s examples are correct: `C=102 s`, first threshold `2×102+10=214 s`, plus `Q≤102 s` gives `316 s`. These remain conditional calculations, not operational guarantees or measurements.

**D2 and D3**

Both are adequately justified. Limit **3** matches the existing default retry budget; increasing it would weaken blocking. Counting a missing publisher matches existing intended behavior and a092’s explicit requirement. The approximately **0 s versus 4–6 s** missing-publisher latency difference is disclosed.

Production can finish the proposed acknowledgement/escalation sequence with durable `ENTRY_BLOCKED` but no enforced entry block; the proposal and integration tests must close that gap.

VERDICT: REJECT