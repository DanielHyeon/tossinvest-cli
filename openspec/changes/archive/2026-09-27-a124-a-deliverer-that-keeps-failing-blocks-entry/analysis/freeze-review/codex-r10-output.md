**REJECT.** Three freeze blockers remain: escalation-failure recovery trusts a stale `blocked` result; D8’s threshold/reset rules conflict with required tests; the protection-latency acceptance criterion grows with the regression it measures.

Read-only review of the supplied revision and source. No tests, network calls, engine execution, or file changes.

References below: `D`, `T`, and `S` mean this change’s `design.md`, `tasks.md`, and `specs/engine-safety/spec.md`. Source filenames retain their repository paths.

**1. Earlier findings**

“RESOLVED” means resolved at proposal level, not implementation verified.

| ID | Status | Reason / evidence |
|---|---|---|
| F1 | RESOLVED | Transaction-local `Attempts` read is explicitly specified before commit; current API lacks it (`D:34–41`; `internal/journal/alert_claim.go:330–334`). |
| F2 | PARTIAL | Epoch fencing removes the main race, but escalation-failure recovery misses a later clear; see AA1 (`D:253–260`). |
| F3 | RESOLVED | Row-local and listing-failure counters cover failures that cannot increment durable attempts (`D:313–343`). |
| F4 | RESOLVED | Batch service between retries is included; 214/316 seconds are arithmetically correct (`D:145–154`; `internal/app/engine/alertdelivery.go:122–162`). |
| F5 | RESOLVED | Remaining starvation cases are explicitly recorded; guarantee is limited to prioritizing below-limit rows (`D:110–119`). |
| F6 | RESOLVED | Deployment behavior is conditional on selection and subsequent failure; startup itself only restores a latch (`D:397–399`; `internal/app/engine/gateway.go:153–167`). |
| F7 | RESOLVED | OFF boundary corrected to engine startup refusal (`D:386–388`; `cmd/tossctl/engine.go:220–221`). |
| F8 | PARTIAL | New detail/log sanitization is specified; escalation-failure retention remains incomplete, AA1 (`D:364–377`). |
| F9 | RESOLVED | Correctly describes insert-only latch behavior (`D:72`; `internal/execgw/retry.go:526–533`). |
| F10 | RESOLVED | Error-first handling and RED coverage address zero-valued `SettleApplied` (`D:43`; `T:37–38`; `internal/journal/alert_claim.go:109`). |
| F11 | RESOLVED | Comparable one-row failure patterns now give 4/34 seconds (`D:80–90`). |
| F12 | PARTIAL | Publisher-absent contention is included, but its acceptance threshold is inadequate; AA3 (`T:31–36`). |
| F13 | RESOLVED | Publisher-absent delay difference is explicitly accepted (`D:100–101`; `internal/obs/notifier.go:429–431`). |
| N1 | RESOLVED | Executor no longer borrows `n.mu`; journal and logging remain outside gate locking (`D:299–309`). |
| N2 | PARTIAL | Error outcomes intentionally permit conservative reblocking; this supersedes, rather than fully preserves, the earlier acknowledgement guarantee (`D:263–270`; `S:35–37`). |
| N3 | RESOLVED | Delivery-settlement errors immediately latch/escalate and retain the lease (`D:58–70`; `internal/obs/notifier.go:475–491`). |
| N4 | PARTIAL | Lifecycle/pruning is specified, but threshold-versus-reset precedence conflicts with tasks; AA2 (`D:320–342`; `T:53–55`). |
| N5 | RESOLVED | General upper-bound claim withdrawn; conditional assumptions and unbounded cases are explicit (`D:130–163`). |
| N6 | RESOLVED | Failed-attempt `NotFound` latches without escalation: `lost=true` suppresses `owed` (`internal/obs/notifier.go:519–523,309–315`). |
| N7 | PARTIAL | Coverage substantially expanded; AA1–AA3 remain uncovered or contradictory. |
| N8 | RESOLVED | Deployment escalation now correctly depends on next selected failure (`D:397–399`). |
| R1 | RESOLVED | No executor-held `n.mu` across journal waits (`D:124,299–309`). |
| R2 | RESOLVED | New logs prohibit account and raw error strings; sentinel test required (`D:369–374`; `T:58–59`). |
| R3 | RESOLVED | Other-row success, `LeaseLost`, and release success cannot reset the row’s counter (`D:320–323`). |
| R4 | PARTIAL | Global-count fence removed; replacement reset semantics still conflict at threshold, AA2 (`D:336–342`). |
| R5 | RESOLVED | Lock scope is consistently gate-memory-only; escalation stays outside (`S:39–42`; `T:75–76`). |
| R6 | RESOLVED | Duplicate suppression explicitly ends with lease expiry/replacement (`D:68–70`; `T:60–63`). |
| R7 | RESOLVED | Unsupported 54-second bound withdrawn; token replacement, not expiry alone, invalidates settlement (`D:165–167`; `internal/journal/outbox.go:472`). |
| R8 | RESOLVED | Row counters advance at most once per cycle; approximately four seconds minimum with defaults (`D:345`; `internal/app/engine/alertdelivery.go:133,154–161`). |
| Q1 | RESOLVED | Dropping `n.mu` removes the executor’s propagation of broker-held `g.mu` waits (`D:299–309`; `internal/execgw/strategy_entry_gate_authority.go:60–72`). |
| Q2 | RESOLVED | Epoch increments at `Clear`, not acknowledgement start (`D:225–230`; `internal/obs/notifier.go:870–875`). |
| Q3 | RESOLVED | Partial/failed acknowledgements do not call `Clear`, so cannot invalidate evidence (`internal/obs/notifier.go:854–875`; `T:48,53`). |
| Q4 | RESOLVED | Post-settlement delivery/rearm without clear deliberately preserves the old judgement (`D:280`; `internal/journal/outbox.go:334–342`). |
| Q5 | RESOLVED | Complete-list pruning and potentially growing retention under truncated lists are recorded (`D:323,330–334`). |
| Q6 | RESOLVED | H assumptions, listing/escalation costs, and replacement of a092’s formula are specified (`D:139–171`; `T:93–95`). |
| Q7 | RESOLVED | F2 disposition now identifies B′, not discarded C (`D:416`). |
| V1 | PARTIAL | Accepted error ambiguity now deliberately reblocks after acknowledgement; it is a conservative exception, not exact acknowledgement parity (`D:283,290–292`). |
| V2 | PARTIAL | Existing count-to-clear race is correctly identified and assigned to a092; retained escalation protects the nominal case, but AA1 remains (`D:284–285`). |
| V3 | RESOLVED | Proposal and current design agree on post-settlement/post-release epoch capture (`proposal.md:88–91`; `D:253–255`). |
| V4 | RESOLVED | Spec permits bounded gate-memory sections rather than falsely claiming no shared lock (`S:39–42`). |
| V5 | PARTIAL | Episode carryover is recorded; listing-counter reset ordering still has contradictory instructions, AA2 (`D:333–342`). |
| V6 | RESOLVED | Final listing cost included; a092 formula replacement specified (`D:147,169–171`). |
| W1 | PARTIAL | Rejection argument holds for successful escalation: old clear removes the latch, not durable mode. It is incomplete under AA1’s failure interleaving (`D:285`; `internal/obs/notifier.go:875`). |
| W2 | PARTIAL | Listing counter needs no row lookup, but threshold/reset precedence remains ambiguous (`D:341–342`; `T:54–57`). |
| W3 | RESOLVED | Arbitrary retry-count fallback removed; new fallback is specifically escalation-write failure, though incomplete under AA1 (`D:215–216`). |
| W4 | RESOLVED | Action table preserves latch-only `NotFound` through application (`D:210`; `T:40,49`). |
| W5 | RESOLVED | Obsolete fence machinery removed from current contract (`D:296–297`; `T:75–76`). |
| X1 | RESOLVED | Deferred-judgement map removed (`D:296`). |
| X2 | RESOLVED | Delivered rows need not become ACKNOWLEDGED; gate-clear epoch captures manual release (`D:183–189,286`; `internal/journal/outbox.go:498`). |
| X3 | RESOLVED | Deferred-map obligations disappear; counter pruning and release exclusion are explicit (`D:320–331`; `S:49–51`). |
| X4 | PARTIAL | Shared journal contention acknowledged, but proposed acceptance cannot enforce the protection bound, AA3 (`D:301–303`; `T:34–36`). |
| X5 | RESOLVED | Baseline B8 correctly says log/release with unchanged attempts; V2 disposition updated (`internal/app/engine/alertdelivery.go:205–212`; `D:452`). |
| Y1 | RESOLVED | Clear no longer suppresses escalation in the nominal path (`D:259`; `S:25–27`). |
| Y2 | RESOLVED | `acknowledged_at` comparison removed; timestamp is indeed captured before the database write (`D:263–265`; `internal/journal/outbox.go:485,494`). |
| Y3 | RESOLVED | All specified pruning exceptions appear in delta spec (`S:49–51`). |
| Y4 | PARTIAL | Separation of isolation and database contention is good; acceptance remains self-adjusting, AA3 (`T:31–36`). |
| Y5 | RESOLVED | Failed-attempt release precedes epoch read; delivery-error lease expiry risk explicitly retained (`D:254,268–270`). |
| Y6 | RESOLVED | Publisher-absent baseline and count-to-clear chronology corrected (`proposal.md:32–34,118–121`). |
| Z1 | PARTIAL | The original post-third-error reset hole is closed, but replacement contradicts pre-third-error test expectations, AA2 (`D:336–342`; `T:54–55`). |
| Z2 | PARTIAL | Handles `blocked=false`; misses successful block subsequently cleared before escalation failure, AA1 (`D:260`). |
| Z3 | PARTIAL | D1 action table is aligned; D8 narrative/spec/tasks still disagree, AA2. |
| Z4 | RESOLVED | SELECT/commit failure atomicity coverage required for all three callers (`D:34–38`; `T:37–38`). |
| Z5 | PARTIAL | A formula is supplied, but measured regression determines its own allowed budget, AA3 (`T:34–36`). |

**2. New findings**

| ID | Severity | Finding | Evidence | Suggested fix |
|---|---|---|---|---|
| AA1 | **P1** | **Successful conditional block is not proof that a latch remains when escalation fails.** Interleaving: threshold → epoch read → conditional block returns `true` → operator acknowledges/clears → escalation fails. `!blocked` is false, so fallback does nothing. No alert latch, no durable escalation, no pending row for startup restoration. | `D:253–260,375–377`; clear actually deletes the latch: `internal/obs/notifier.go:864–875`, `internal/execgw/retry.go:537–543`; escalation can fail: `internal/journal/operating_mode.go:391–400,468–470`; startup skips empty backlog: `internal/app/engine/gateway.go:154–159`. | Apply the agreed conservative fallback on escalation error without trusting historical `blocked`. Add deterministic block-success → clear → escalation-error tests, including restart. `T:50` only requires block-rejected → escalation-error. |
| AA2 | **P1** | **D8 cannot satisfy its design and required tests simultaneously.** With `(count=2,e0) → Clear(e1) → third error`, D8 mandates no reset, uses `e0`, rejects the latch, and escalates. Task 2.10 says a clear **before** the threshold error yields latch + escalation; its preceding sentence says clear breaks the streak. Listing instructions also say reset-before-increment and then prohibit threshold reset. | `D:320,336–342`; `S:29,34–35,47–48`; `T:53–55`. Epoch mismatch really must reject: `S:125–126`. | Write one transition table with explicit precedence. Distinguish “third error before clear” from “clear between second and third error.” Align tests/spec with the selected conservative-overescalation policy. Add threshold/reset-order mutations to 4.1. |
| AA3 | **P1** | **Protection-latency acceptance is self-adjusting.** `baseline + operation_count × measured executor transaction maximum` increases when executor transactions become slower. Arbitrarily harmful finite delay can therefore pass. The retained transport-stall test does not exercise these transactions while its publisher is blocked. | `T:31–36`; shared connection: `internal/journal/journal.go:171–175`; existing test explicitly rejects this reasoning and uses fixed margins: `internal/app/engine/a098_the_backlog_does_not_delay_protection_test.go:61–72,182–190`. | Keep contention measurements, but require an independently fixed protection budget and matched base-versus-change workloads. Test publisher absence and delayed settlement/escalation against that budget; excessive delay must fail regardless of measured transaction length. |
| AA4 | **P2** | Parameterizing `PendingAlerts` conflicts with “no obs edits” unless API compatibility is deliberately preserved. Current callers in `Notifier.Flush` and `Acknowledge` use the two-argument signature. | `D:105–108`; `proposal.md:72,105`; `internal/journal/outbox.go:517`; `internal/obs/notifier.go:737,855`. | Specify a compatible API or explicitly allow mechanical caller updates; retain one canonical threshold source. |

D7’s **current** lock design is sound at the executor boundary: it does not borrow `n.mu`, and its gate critical sections contain only memory operations. The rejected C design’s two-interleaving argument is no longer relevant. Existing synchronous delivery still holds `n.mu` across publish (`internal/obs/notifier.go:254–255,406–433`); a124 does not introduce that behavior or fix it. Exit/flatten can additionally wait on the shared journal connection, so mutex isolation alone cannot establish unchanged protection latency.

Principle E is **a conservative approximation**, not literal equivalence to “acknowledgement wins.” A clear between settlement and epoch capture can be missed and cause reblocking; release before capture widens that window (`D:253–255,290–294`). Error outcomes can also reblock an acknowledged episode. These exceptions are recorded in `S:35–37`; they should qualify the unconditional scenarios at `S:76–97`. Removing `acknowledged_at` comparison was correct.

The non-conservative loss found here is **AA1**, not post-settlement epoch capture itself. Tests 2.9/4.1 do not currently demand that interleaving. Task 2.10 exposes AA2 only by requiring mutually inconsistent outcomes.

D1’s transaction-local SELECT is implementable: the existing UPDATE, SELECT, and commit can share `tx`; using `j.db` would contend with its own sole connection (`internal/journal/alert_claim.go:316–334`; `internal/journal/journal.go:174`). The two `NotFound` policies correctly differ: failed-attempt settlement is latch-only; successful publish with failed settlement also escalates (`internal/obs/notifier.go:452–491,519–523,309–315`).

D6’s illustrative calculations check out: `C=102s`, first-row third failure `2×102+10=214s`, last-row `304s`, and maximum stated queue allowance gives `316s`. These remain conditional calculations, not measured operational guarantees. `S/M/E/I_list` must represent appropriate bounded costs when used as an upper bound.

**3. D2 and D3**

Both are adequately justified.

- **Limit 3:** reuses the existing default and preserves retry count. One-row immediate failure/timeout comparisons are correctly 4/34 seconds, excluding local costs (`internal/obs/notifier.go:45–48,420–428,525–527`).
- **Missing publisher counts:** matches existing intended safety direction and a092’s explicit requirement (`internal/app/engine/exitwiring.go:60–70`; a092 delta `spec.md:76,108–110`). The executor’s additional delay is openly recorded.

VERDICT: REJECT — AA1 can lose both protection outcomes, AA2 leaves contradictory implementation requirements, and AA3 cannot enforce the protection-latency acceptance boundary.