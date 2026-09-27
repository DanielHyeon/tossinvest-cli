Freeze remains blocked by two P1 defects in D7. This is a proposal review, not implementation acceptance; no files changed, tests run, network calls made, or engine commands executed.

References below: `D` = change `design.md`, `T` = `tasks.md`, `Δ` = change `specs/engine-safety/spec.md`. Code references use the filenames under the paths specified in the request. **RESOLVED means resolved in the proposal.**

| Earlier finding | Status | Evidence / reason |
|---|---|---|
| F1 | RESOLVED | D1 specifies a transaction-local read before commit; compatible with `alert_claim.go:322–334`, unlike the current result at `:140–145`. |
| F2 | PARTIAL | Normal acknowledgement races are fenced, but delivered-row acknowledgement and deferred replay still permit re-latching; X1–X2 below. |
| F3 | RESOLVED | D8 supplies per-row recording/claim-error counters and listing-error handling (`D:309–335`). |
| F4 | RESOLVED | Batch service is included; 214 s / 316 s arithmetic matches sequential cycles (`alertdelivery.go:122–161`; `D:146–155`). |
| F5 | PARTIAL | Priority starvation is addressed; held-row and exhausted-tier starvation remain explicitly accepted (`D:106–117`; `alert_claim.go:162–165`). |
| F6 | RESOLVED | Deployment escalation is correctly conditional on subsequent selection and failure (`D:384–387`; `gateway.go:153–167`). |
| F7 | RESOLVED | OFF boundary corrected to engine refusal; Notifier remains unconditionally constructed (`cmd/tossctl/engine.go:220–221`; `gateway.go:323`). |
| F8 | RESOLVED | D9 requires fixed details and allowlisted new logs; necessary because `Logger.Error` appends raw errors (`log.go:162–165`). |
| F9 | RESOLVED | First detail retained; repeated Block does not update it (`retry.go:526–533`). |
| F10 | RESOLVED | Error-first handling specified; `SettleApplied` remains zero (`alert_claim.go:109,324`; `D:43`). |
| F11 | RESOLVED | Like-for-like single-row comparison corrected to 4 s / 34 s (`notifier.go:428–526`; `D:87–96`). |
| F12 | RESOLVED | Shared-connection cost acknowledged and testing planned; not yet verified (`journal.go:174`; `T:31–33`). See X4 for overstatement. |
| F13 | RESOLVED | Missing-publisher delay explicitly accepted; synchronous path breaks immediately (`notifier.go:429–431`; `D:104`). |
| N1 | RESOLVED | Executor no longer acquires `n.mu`; settlement, escalation and logs are outside gate locking (`D:181–205,294–298`). |
| N2 | PARTIAL | Ordinary post-clear recording errors are fenced; deferred replay and delivered-row acknowledgement remain defective, X1–X2. |
| N3 | PARTIAL | Success/settlement-failure outcomes and lease retention match code; acknowledgement handling remains incomplete (`notifier.go:435–491`; X1–X2). |
| N4 | RESOLVED | Per-row lifecycle, epoch reset, full-list pruning and carryover are specified (`D:309–327`); wording exceptions need X3. |
| N5 | RESOLVED | Universal upper-bound claim withdrawn; H and unbounded dependencies stated (`D:131–164`). |
| N6 | RESOLVED | Failure-record NotFound blocks only: `deliver` returns `lost=true`, suppressing escalation (`notifier.go:519–523,309–315`). |
| N7 | PARTIAL | Coverage expanded substantially, but deferred replay and counter exceptions remain inconsistent or underspecified; X1–X3. |
| N8 | RESOLVED | First-cycle escalation no longer unconditional (`D:384–387`; `operating_mode.go:410–421`). |
| R1 | RESOLVED | Executor does not hold the protection path’s Notifier mutex across journal work (`D:181–183,215–235`). |
| R2 | RESOLVED | New logs exclude account/raw-error fields; mode errors can contain account references (`operating_mode.go:394,449`; `D:359–365`). |
| R3 | RESOLVED | Other-row success and zero-write settlement outcomes no longer reset a row’s counter (`D:311–315`). |
| R4 | PARTIAL | Row lookup replaces global-count inference, but its acknowledgement model misses delivered-then-manually-cleared rows; X2. |
| R5 | PARTIAL | Locking contract aligned; counter lifecycle and deferred execution still lack one consistent contract, X1/X3. |
| R6 | RESOLVED | Lease suppression explicitly ends at expiry/replacement (`D:72–75`; `alert_claim.go:162–165`). |
| R7 | RESOLVED | Expiry alone does not invalidate settlement; state/token determine outcome (`alert_claim.go:349–365`; `outbox.go:472`). |
| R8 | RESOLVED | Per-row counting restores three observations separated by two waits (`D:337`; `alertdelivery.go:133,161`). |
| Q1 | RESOLVED | No new `n.mu → g.mu` chain; existing synchronous chain remains a documented prerequisite (`notifier.go:254,484`; `strategy_entry_gate_authority.go:60–72`). |
| Q2 | RESOLVED | Epoch changes atomically with Clear, not acknowledgement entry (`D:192–202`; `notifier.go:870–875`). |
| Q3 | RESOLVED | Failed/partial acknowledgements do not call Clear, preserving counters and judgments (`notifier.go:854–875`). |
| Q4 | RESOLVED | No-clear delivery/re-arm carryover is explicitly conservative (`D:271`; `outbox.go:334–342`). |
| Q5 | RESOLVED | Complete-list pruning and remaining memory growth documented (`D:315–321`). |
| Q6 | RESOLVED | H strengthened and a092 formula replacement required (`D:140–173`; `T:80–82`), subject to deferred-work qualification in X3. |
| Q7 | RESOLVED | F2 disposition now names B′ (`D:404`). |
| V1 | RESOLVED | Pre/post-claim epoch bracket prevents the identified post-claim snapshot race (`D:217–220`; claim commit `alert_claim.go:299–302`). |
| V2 | RESOLVED | Clear no longer proves zero backlog at that instant; row recheck preserves the independent-enqueue judgment (`D:252–263`; `replay.go:551`). |
| V3 | RESOLVED | Proposal/design/spec now agree on two epoch reads (`D:217–220`; `Δ:22–23`). |
| V4 | RESOLVED | Gate-only critical section permitted; nested external work prohibited (`Δ:30–33`; `flatten.go:643`). |
| V5 | RESOLVED | Same-ID carryover and listing-counter epoch comparison are explicit (`D:323–327`). |
| V6 | RESOLVED | Final listing cost included and task requires formula replacement (`D:148`; `T:81–82`). |
| W1 | RESOLVED | The original DELIVERED-without-human-confirmation discard is removed (`D:231–232,255–257`); replacement creates X2. |
| W2 | RESOLVED | Listing failures have a separate fence/reset path, rather than nonexistent row lookup (`D:333–335`). |
| W3 | PARTIAL | Unconditional fallback removed, but next-cycle invocation can bypass the promised acknowledgement check; X1. |
| W4 | RESOLVED | Escalation flag carried through retry and deferral (`D:225,234`; `T:36–37`). |
| W5 | PARTIAL | Stale unconditional-fallback text and lifecycle contradictions remain; X3/X5. |

| ID | Severity | Finding | Evidence | Suggested fix |
|---|---|---|---|---|
| X1 | **P1** | **Deferred replay has contradictory execution order.** D says to call `latchConfirmed` with a **new epoch**. That helper first conditionally blocks and only reads the ledger after a failed fence. If the row was acknowledged between cycles, the fresh epoch matches: it blocks and escalates without reading ACKNOWLEDGED. The prose promises the opposite. | `D:227–234,237–238,277`; acknowledgement changes state then clears at `notifier.go:864–875`. | Retain the last validated epoch in deferred state, or explicitly perform **epoch read → row lookup → conditional block** before every deferred application. Specify terminal removal. Add a test asserting **no transient Block or escalation**, not merely the eventual gate state. |
| X2 | **P1** | **ACKNOWLEDGED-only evidence cannot represent manual acknowledgement after delivery.** Counterexample: failed-attempt judgment waits; another sender delivers; operator calls Acknowledge and clears the gate; late judgment sees changed epoch plus DELIVERED and re-blocks/escalates. Even explicitly naming that ID does not turn DELIVERED into ACKNOWLEDGED. Thus D’s claimed empty-backlog acknowledgement remedy can itself be undone. | `D:231–232,271,283`; delivery `outbox.go:453–456`; acknowledgement updates **PENDING only** at `:494–503`; nonpending result is ignored by `notifier.go:864–865`, followed by Clear at `:875`. Canonical manual recovery: `openspec/specs/engine-safety/spec.md:183,954`. | Supply acknowledgement evidence covering delivered episodes and distinguish **clear before that episode was covered** from **manual clear after delivery**. Coordinate the acknowledgement/ledger surface and revise the non-goal if necessary. Simply accepting every DELIVERED row would reintroduce W1. |
| X3 | **P2** | **Lifecycle contract is incomplete.** Deferred entries are called a “small map,” yet every entry is processed before listing, with no explicit work budget, terminal deletion or merge rule. D8 also prunes ClaimSettled/absent rows although Δ says only Applied or epoch change breaks continuity. Release also returns Applied and must explicitly remain excluded. | `D:234–240,311–320`; `Δ:37–39`; `alert_claim.go:263–264,375–379`; a092 delta `:48` requires bounded cycle workload. | Specify deferred deletion and escalation-flag merging; budget processing without starving fresh deliveries. Enumerate counter pruning exceptions and exclude Release from reset. Include deferred overhead in D6 or explicitly assume none. |
| X4 | **P2** | **Protection latency claim is too broad.** B′ removes mutex-mediated remote waiting, but new settlement reads/escalation transactions still occupy the shared journal connection. Delaying an actual executor transaction can delay a protection-path journal operation; “exit residence unchanged” cannot be an unconditional assertion. | `D:288–292`; `T:31–33`; `journal.go:174`; `notifier.go:254–262`; `operating_mode.go:391–396`. | Separate “executor waiting without holding resources” from “executor holding the journal transaction.” Measure the latter against the accepted local-work budget; retain the no-network-wait invariant. |
| X5 | **P3** | Historical dispositions still describe rejected behavior; one branch-map summary misstates baseline behavior. | `D:443` still says unconditional block after three races, contradicting `:234`. DeliverOne FLM “State mutations” says B8 increments attempts, but `alertdelivery.go:205–212` only logs/releases. | Correct the disposition and B8 summary so implementation cannot follow stale instructions. |

D7’s **basic epoch mechanism is sound**: compare-and-block must be atomic under `g.mu`; neither publication nor journal work belongs inside it. The executor introduces no new Notifier lock-order cycle. Protection paths can still wait for its short gate critical section and shared journal work; the pre-existing synchronous remote-wait chain remains real.

**Incrementing on every clear request is correct and necessary.** Incrementing only when a latch exists loses case ㉣: acknowledgement before initial Block leaves no epoch evidence. Preserve `revision`’s existing behavior. Current production Clear callers for this reason are exactly `Notifier.Acknowledge` at `notifier.go:846,875`; mode/reconciliation projections delete different reasons (`modegate.go:37`; `symbolgate.go:184–189`). That proves the clear source, **not which episode the human acknowledged**—X2 is the missing distinction.

D8’s ordinary counter logic is adequately conservative: unrelated successes cannot erase failures; only Run-context cancellation is excluded, while transport timeout comes from Ntfy’s child context (`ntfy.go:95–100`). Listing-counter reset is defensible in the engine’s journal-backed acknowledgement path, but successful Count does not prove the exact listing operation recovered. Describe it as the chosen reset policy; the next three listing failures must still block.

D1’s additive transaction-local `Attempts` read is implementable. Read through `tx`, return the value only after successful commit, and test read-failure rollback. The two NotFound policies correctly differ: failed-send settlement blocks only; successful-send settlement failure blocks and escalates (`notifier.go:452–491,519–523,309–315`).

D6’s stated **idealized** values are correct: **4/6 s**, **34/46 s**, **214/316 s**, with **304 s** for the tenth row’s third timeout. They are not measured production guarantees; deferred processing and shared-resource costs need the qualification above.

D2 **limit 3 is adequately justified** by the existing default and like-for-like single-row behavior (`notifier.go:45,48`). D3 **counting missing publisher is adequately justified** by existing safety intent (`exitwiring.go:62–70`) and a092’s explicit requirement (`spec.md:76`). Its additional delay is disclosed; it is not synchronous latency parity.

VERDICT: REJECT — D7 still lacks a safe, unambiguous deferred replay protocol and acknowledgement evidence that preserves manual recovery after delivery.