**REJECT.** Revision 5 removes the executor’s Notifier-mutex dependency, but its epoch proof still permits stale relatching and suppression of required blocking.

Read-only review; no tests, network calls, engine execution, or file changes. “RESOLVED” below means resolved **in the proposed design**, not implementation verified. References to `design.md`, `proposal.md`, `tasks.md`, and the delta `spec.md` mean files under `openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/`.

**1. Earlier findings**

| ID | Status | Why |
|---|---|---|
| F1 | RESOLVED | D1 supplies the missing field through a same-transaction read before commit; current applied return lacks it (`design.md:32–37`; `internal/journal/alert_claim.go:330–334`). |
| F2 | PARTIAL | Conditional blocking fixes clear-after-snapshot races, but claim→snapshot and count→clear gaps remain; see V1–V2 (`design.md:211–219`). |
| F3 | PARTIAL | D8 counts otherwise invisible failures, but its epoch-based reset inherits the invalid proof; see V2 (`design.md:267–287`). |
| F4 | RESOLVED | Other batch members now contribute between retries; 214/316-second examples are correct under zero-cost assumptions (`design.md:141–149`; `internal/app/engine/alertdelivery.go:154–162`). |
| F5 | PARTIAL | Exhausted-row starvation is addressed; held-row obstruction and exhausted-tier starvation remain explicitly accepted limitations (`design.md:101–115`; `internal/app/engine/alertdelivery.go:178–192`). |
| F6 | RESOLVED | Deployment escalation is correctly conditional on selection and another failed attempt; startup restores only the latch (`design.md:338–340`; `internal/app/engine/gateway.go:153–167`). |
| F7 | RESOLVED | OFF boundary corrected to engine refusal; Notifier creation is unconditional (`internal/app/engine/gateway.go:323`; `cmd/tossctl/engine.go:220–221`). |
| F8 | RESOLVED | D9 prohibits account/raw-error leakage in new output and preserves the latch on escalation failure (`design.md:308–320`; account-bearing errors: `internal/journal/operating_mode.go:394,449`). |
| F9 | RESOLVED | First detail is retained; repeated `Block` does not update it (`design.md:68`; `internal/execgw/retry.go:529–532`). |
| F10 | RESOLVED | Error inspection explicitly precedes zero-valued outcome inspection (`design.md:39`; `internal/journal/alert_claim.go:109,318`). |
| F11 | RESOLVED | Like-for-like single-row timing replaces the mismatched comparison (`design.md:76–86`; `internal/obs/notifier.go:425–433,525–528`). |
| F12 | RESOLVED | Added journal contention is recorded and a missing-publisher protection test is required; latency remains unverified (`design.md:341–342`; `tasks.md:31–33`; `internal/journal/journal.go:174`). |
| F13 | RESOLVED | Immediate synchronous nil-publisher blocking versus deferred executor blocking is disclosed (`design.md:96–97`; `internal/obs/notifier.go:429–431,571`). |
| N1 | RESOLVED | Executor no longer holds `n.mu` across settlement, escalation, or logging (`design.md:211–219,259–260`; existing protection acquisition: `internal/obs/notifier.go:254`). |
| N2 | PARTIAL | Errors after a captured epoch are fenced, but acknowledgement before that capture is missed; V1 (`design.md:212`; `internal/obs/notifier.go:870–875`). |
| N3 | PARTIAL | Success-plus-settlement-failure now blocks immediately and retains the lease, but V1 defeats its acknowledgement guarantee (`design.md:54–66`). |
| N4 | PARTIAL | Per-row lifecycle and pruning are substantially specified; episode carryover and listing-counter epoch reset still need clarification; V5 (`design.md:270–284`). |
| N5 | RESOLVED | General finite-bound claim withdrawn; explicit assumptions and unbounded cases replace it (`design.md:126–158`). |
| N6 | RESOLVED | Failed-attempt `NotFound` blocks only: synchronous `lost=true` suppresses escalation (`internal/obs/notifier.go:519–523,309–315`). |
| N7 | PARTIAL | Coverage expanded considerably, but the lock requirement contradicts D7 and the decisive missing interleavings are absent; V1–V4 (`spec.md:21–28`; `tasks.md:38–41`). |
| N8 | RESOLVED | Existing exhausted rows need not escalate on the first cycle or at all after successful delivery (`design.md:338–340`; `internal/app/engine/alertdelivery.go:229–241`). |
| R1 | RESOLVED | Executor-owned Notifier critical section is removed; no journal wait is nested beneath it (`design.md:175–177,211–219`). |
| R2 | RESOLVED | New logs use an allowlist and classify escalation errors without forwarding their text (`design.md:313–317`; `internal/journal/operating_mode.go:394,449`). |
| R3 | RESOLVED | Another row’s success and zero-row settlements no longer reset this row’s counter (`design.md:272–274`; `internal/journal/alert_claim.go:337–365`). |
| R4 | PARTIAL | Global-count recheck is gone, but post-clear epoch adoption and the count→clear gap still invalidate episode attribution; V1–V2. |
| R5 | PARTIAL | Escalation placement is consistent; the new blanket shared-lock prohibition conflicts with the required gate mutex; V4 (`spec.md:26–27,95`). |
| R6 | RESOLVED | Suppression is explicitly limited to an unexpired, unreplaced lease; expiry replay is included in tests (`design.md:64–66`; `tasks.md:48–50`). |
| R7 | RESOLVED | Unsupported 54-second mutex bound removed; token replacement, not expiry alone, determines settlement loss (`design.md:160–163`; `internal/journal/outbox.go:472`). |
| R8 | RESOLVED | Per-row counting restores at least two inter-cycle waits before three failures (`design.md:289`; `internal/app/engine/alertdelivery.go:133,154–162`). |
| Q1 | RESOLVED | Executor no longer holds `n.mu` while awaiting `g.mu`; the remote-wait chain is removed, subject to V4’s wording correction (`design.md:175–177,251–254`). |
| Q2 | RESOLVED | Epoch increment moves to the gate-clear critical section, closing the earlier acknowledgement-start window (`design.md:186–196,235`). |
| Q3 | PARTIAL | Ordinary partial/failed acknowledgements preserve counters, but a stale zero-count observation can still suppress a genuinely new failure; V2. |
| Q4 | RESOLVED | Late blocking after delivery/rearm is explicitly conservative and consistent with manual release (`design.md:239`; `openspec/specs/engine-safety/spec.md:181–183`). |
| Q5 | RESOLVED | Complete-list pruning and potentially unbounded retention during continuously full batches are now disclosed (`design.md:276–282`). |
| Q6 | PARTIAL | H is stronger and double-counting is recognized, but terminal listing cost is missing and task 5.2 retains the old instruction; V6 (`design.md:141–167`; `tasks.md:80`). |
| Q7 | RESOLVED | F2 disposition now names B′ and escalation outside the lock (`design.md:357`). |

**2. New findings**

| ID | Severity | Finding | Evidence | Suggested fix |
|---|---|---|---|---|
| V1 | **P1** | **Claim→epoch read is not atomic. An acknowledged old operation can adopt the new epoch and relatch.** | `design.md:211–228`; claim commits before returning (`internal/journal/alert_claim.go:299–302`); acknowledgement ignores ownership and clears the claim (`internal/journal/outbox.go:494–499`); gate clear follows (`internal/obs/notifier.go:875`). | Define a protocol that detects clears overlapping claim acquisition, including claim errors. Bracket acquisition with epoch observations and safely release/retry or validate the claimed episode before publishing. Add the exact interleaving below. Do not hold a protection mutex across journal work. |
| V2 | **P1** | **Proof (P) assumes zero pending at `Clear`, but the code establishes zero only at an earlier query. A separate recorder can insert between them, causing required blocking and counters to be discarded.** | `design.md:224–248,267–268`; separate count/clear (`internal/obs/notifier.go:870–875`); independent recorder (`internal/execgw/replay.go:551–558` → `internal/journal/outbox.go:131–153`); a092 explicitly identifies this requirement (`a092…/spec.md:72`). | Establish the count→clear boundary against **every** producer, or replace the global-zero inference with episode-specific acknowledgement evidence. Caller-count testing alone cannot prove this. Add an independent-enqueue race test. |
| V3 | **P1** | **Proposal mandates the snapshot timing that design/spec explicitly reject.** | `proposal.md:88`: before claim; `design.md:212,230–231` and `spec.md:22–23`: after claim; `tasks.md:70` treats before-claim placement as a mutation that must fail. | Align all artifacts to the corrected V1 protocol before freeze. |
| V4 | **P1** | **Spec forbids any mutex protection waits on, while D7 requires acquiring precisely such a mutex.** Removing `n.mu` eliminates remote-wait propagation, not all waiting. | `spec.md:26,95`; `design.md:251–254`; flatten calls `Gate.Block` (`internal/flatten/flatten.go:643`), which acquires `g.mu` (`internal/execgw/retry.go:527`). | Explicitly permit bounded gate-state critical sections; prohibit holding them across external work or another blocking dependency. Replace “nothing new waits” with the actual bounded-contention claim. |
| V5 | **P2** | **D8 needs explicit episode-carryover and listing-counter reset rules.** Same-ID delivery/rearm without a clear can retain old failures; listing-counter epoch reset is not stated as clearly as row-entry reset. | `design.md:272–284`; rearm resets persisted attempts (`internal/journal/outbox.go:334–342`), while D8 keys memory by ID and clear epoch. | Record whether early conservative blocking from retained old-episode failures is accepted; add its expected test result. State that an epoch change resets the listing counter before counting the current error, without waiting for threshold comparison. |
| V6 | **P2** | **D6 omits the final cycle’s listing cost and task 5.2 still says to redefine the period.** | Listing precedes attempts (`internal/app/engine/alertdelivery.go:150–161`); `design.md:141–143,165–167`; `tasks.md:80`. | Use `Q + (L−1)C + I_list + T + S + M` under the stated cycle-start convention. Call it a conditional bound; equality also needs zero initial attempts. Change task 5.2 to replace the a092 formula. |

The two missing schedules are decisive:

- **V1 — old work adopts a fresh epoch:** executor acquires A → operator acknowledges A and clears → executor reads the **new** epoch → publish succeeds → delivery settlement returns a journal error → conditional block succeeds and escalation follows. D1’s immediate-error policy applies. “Immediately after claim” does not prevent this scheduling gap; proof (P)’s assertion that A was PENDING when the epoch was read is false.
- **V2 — fresh work is mistaken for acknowledged work:** acknowledgement counts zero and pauses before `Clear` → independent `EnqueueAlert` creates B → executor claims B and reads the old epoch → publish succeeds but settlement errors → acknowledgement executes its stale clear → executor drops B’s mandatory immediate-block decision. A subsequent successful delivery can remove B without that decision ever being applied. The replay path’s separate unresolved-order latch does not fulfill the required alert latch and durable escalation.

**The author’s deviation from M1 is correct and necessary:** incrementing on every clear request closes case ㉣, where no latch exists yet. Incrementing only on actual deletion loses the acknowledgement. Preserve `revision`’s existing behavior (`internal/execgw/retry.go:540–542`). This correction is **necessary but insufficient** because V1–V2 concern which episode the epoch represents.

No additional clearing route for this reason was found: the two calls are `Notifier.Acknowledge` at `internal/obs/notifier.go:846,875`. Direct projection deletions affect different reasons (`internal/execgw/modegate.go:37`; `internal/execgw/symbolgate.go:184–190`). However, identifying the clearing caller does not prove its zero-count observation remains current.

D7 introduces no evident executor lock-order cycle: its journal work precedes or follows gate locking, and automatic escalation projects only after commit (`internal/journal/operating_mode.go:468–476`). The existing synchronous `n.mu → g.mu → broker-wait` exposure remains a separately recorded issue. D8’s per-row counters resist other-row successes and ordinary partial acknowledgements; parent-context cancellation exclusion correctly distinguishes shutdown from Ntfy’s child timeout (`internal/obs/ntfy.go:95–100`).

**3. D2 and D3**

- **D2, limit 3: adequately justified.** It preserves the default synchronous retry count. Single-row 4/34-second calculations are correct with journal costs excluded; batch timing must use D6. This is not a universal latency-equivalence claim.
- **D3, count missing publisher: adequately justified.** Current wiring explicitly specifies blocking/escalation for absent transport (`internal/app/engine/exitwiring.go:62–70`), and a092 requires counting it (`a092…/spec.md:108–110`). The additional delay and journal contention are disclosed.
- **D1’s additive read and NotFound parity are sound.** Reading through the transaction after the successful CAS prevents rearm from changing the returned count before commit; rollback-on-read-error is conservative. Failed-attempt `NotFound` correctly blocks without escalation.

VERDICT: REJECT — V1–V2 invalidate the epoch safety proof, and V3–V4 leave contradictory implementation contracts that must be fixed before freeze.