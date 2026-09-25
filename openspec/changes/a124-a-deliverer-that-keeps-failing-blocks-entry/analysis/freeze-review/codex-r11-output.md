**REJECT: one new P1 remains—priority selection introduces backlog-sized work on the connection shared with protection.** D7’s former mutex and timestamp approaches are gone; their original blockers should not be carried forward unchanged.

Read-only review. No tests, engine commands, network calls, or writes performed. “RESOLVED” below means addressed by the proposal, not implementation-verified.

References to `design.md`, `tasks.md`, and delta `spec.md` mean files under `openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/`.

**Earlier findings**

| ID | Status | Reason and evidence |
|---|---|---|
| F1 | RESOLVED | Transaction-local attempts read explicitly fills the missing API field; `design.md:36`, `internal/journal/alert_claim.go:330`. |
| F2 | PARTIAL | Epoch fencing addresses delayed application, but literal acknowledgement precedence has explicit conservative exceptions; `design.md:292`, delta `spec.md:36`; see AB2. |
| F3 | RESOLVED | Per-row write/claim-error counting and listing-error counting cover failures that cannot increment attempts; `design.md:322`, `:338`, `:356`. |
| F4 | RESOLVED | Formula includes intervening batch service; 214/316-second examples are correct under H; `design.md:140`, `:155`. |
| F5 | PARTIAL | Exhausted-row starvation is addressed; held-row and exhausted-tier starvation remain explicitly accepted limitations; `design.md:108`, `:116`. |
| F6 | RESOLVED | Deployment escalation is now conditional on selection and another failure; startup itself only restores the latch; `design.md:411`, `internal/app/engine/gateway.go:153`. |
| F7 | RESOLVED | OFF explanation matches startup rejection, not absent Notifier; `internal/app/engine/gateway.go:323`, `cmd/tossctl/engine.go:220`. |
| F8 | RESOLVED | Fixed details, allowlisted new logs, and escalation-error fallback specified; `design.md:378`, `:383`, `:215`. |
| F9 | RESOLVED | First-detail-wins semantics correctly stated; `internal/execgw/retry.go:529`, `design.md:74`. |
| F10 | RESOLVED | Error-first handling prevents zero-valued `SettleApplied` interpretation; `design.md:45`, `tasks.md:37`, `internal/journal/alert_claim.go:109`. |
| F11 | RESOLVED | Like-for-like isolated-row comparison is 4/34 seconds; `design.md:82`, `:87`. |
| F12 | RESOLVED | Publisher-absent write contention is explicitly included in protection measurements; `tasks.md:31`. |
| F13 | RESOLVED | Immediate synchronous latch versus approximately 4–6-second executor delay is disclosed; `design.md:102`, `internal/obs/notifier.go:429`. |
| N1 | RESOLVED | Executor no longer holds `n.mu` around settlement, escalation, or logging; `design.md:125`, `:301`. |
| N2 | RESOLVED | Error-hidden acknowledgement has an explicit conservative policy rather than an unreliable state/timestamp inference; `design.md:264`, delta `spec.md:37`. |
| N3 | RESOLVED | Successful publish followed by settlement failure immediately latches/escalates and retains the lease; `design.md:60`, `internal/obs/notifier.go:465`. |
| N4 | RESOLVED | Counter retention, pruning, release exclusion, and episode carryover are specified; `design.md:320`, `:335`. |
| N5 | RESOLVED | General bound withdrawn; finite examples require H; `design.md:131`, `:140`, `:162`. |
| N6 | RESOLVED | Failed-attempt `NotFound` means latch only: synchronous `lost=true` suppresses escalation; `internal/obs/notifier.go:519`, `:309`. |
| N7 | PARTIAL | Main policies now appear in spec/tasks, but scenario exceptions and listing-cost coverage remain incomplete; AB1/AB2. |
| N8 | RESOLVED | Deployment description no longer promises an unconditional first-cycle mode record; `design.md:411`. |
| R1 | RESOLVED | Journal waits are outside executor-held gate/Notifier locks; `design.md:253`, `:310`. |
| R2 | RESOLVED | New logs prohibit account/raw errors, including escalation failures; `design.md:383`, `tasks.md:59`; raw logger behavior confirmed at `internal/obs/log.go:162`. |
| R3 | RESOLVED | Other rows and zero-write settlement results cannot reset the failing row’s counter; `design.md:323`, `tasks.md:51`. |
| R4 | RESOLVED | Global pending-count inference removed; clear epochs and explicit error treatment replace it; `design.md:264`, `:338`. |
| R5 | RESOLVED | Lock contract consistently limits executor critical sections to gate state operations; delta `spec.md:40`, `tasks.md:77`. |
| R6 | RESOLVED | Suppression is explicitly lease-lifetime-limited; expiration/republication tests required; `design.md:71`, `tasks.md:61`. |
| R7 | RESOLVED | Unsupported 54-second mutex bound withdrawn; settlement CAS checks token/state, not expiry; `design.md:166`, `internal/journal/outbox.go:472`. |
| R8 | RESOLVED | Per-row counting restores the default minimum of two inter-cycle waits; `design.md:359`, `internal/app/engine/alertdelivery.go:133`. |
| Q1 | RESOLVED | Executor never acquires `n.mu`; existing synchronous `n.mu → g.mu` hazard is separately recorded; `design.md:301`, `:419`. |
| Q2 | RESOLVED | Epoch increments at `Clear`, not acknowledgement start; `design.md:230`, `internal/obs/notifier.go:875`. |
| Q3 | RESOLVED | Failed/partial acknowledgements do not call `Clear`, so cannot reset epochs/counters; `internal/obs/notifier.go:864`, `:870`, `design.md:343`. |
| Q4 | RESOLVED | Delivery/re-arm after established evidence does not erase the manual-clear obligation; `design.md:281`, `internal/journal/outbox.go:337`. |
| Q5 | RESOLVED | Complete-list pruning specified; truncated-list retention and potential map growth disclosed; `design.md:325`, `:332`. |
| Q6 | RESOLVED | H, escalation/listing costs, and replacement of a092’s double-counting formula specified; `design.md:140`, `:170`. |
| Q7 | RESOLVED | F2 disposition now identifies B′ rather than obsolete C; `design.md:430`. |
| V1 | RESOLVED | Obsolete claim-relative snapshot removed; error outcomes deliberately remain conservative even when acknowledgement is hidden; `design.md:264`, `:284`. |
| V2 | RESOLVED | Count–clear race accurately assigned to existing acknowledgement behavior; delayed escalation is retained; `design.md:285`, `internal/obs/notifier.go:870`, `internal/execgw/replay.go:551`. |
| V3 | RESOLVED | Documents consistently place D1 epoch capture after settlement, after release on failed-publish paths; `design.md:256`, `tasks.md:77`. |
| V4 | RESOLVED | Gate mutex explicitly permitted for state-only operations; delta `spec.md:40`; flatten uses that mutex through `Block` at `internal/flatten/flatten.go:643`. |
| V5 | RESOLVED | Episode carryover accepted conservatively; counter transition priority now explicit; `design.md:335`, `:338`. |
| V6 | RESOLVED | Final listing cost included and a092 formula replacement tasked; `design.md:148`, `tasks.md:97`. |
| W1 | RESOLVED | Rejection argument now holds for the specified interleaving: late clear removes only the alert latch; escalation survives, with latch fallback on write failure; `design.md:215`, `:286`. |
| W2 | RESOLVED | Listing failures use their own epoch-aware counter without row-state lookup; `design.md:356`, `tasks.md:54`. |
| W3 | RESOLVED | Arbitrary retry-exhaustion relatching removed; unconditional fallback is now specifically failed durable escalation; `design.md:215`, `:298`. |
| W4 | RESOLVED | Action table preserves latch-only outcomes through application; `design.md:209`, `tasks.md:39`. |
| W5 | RESOLVED | Old fence policy replaced across design/spec/tasks; `design.md:298`, delta `spec.md:25`, `tasks.md:77`. |
| X1 | RESOLVED | Deferred-judgement map removed; no fresh-epoch replay of old deferred work; `design.md:298`. |
| X2 | RESOLVED | Delivered-row acknowledgement-state ambiguity removed by temporal fencing, subject to the documented conservative capture gap; `design.md:287`, `:292`. |
| X3 | RESOLVED | Deferred map removed; pruning and release-`Applied` exclusion explicit; `design.md:323`, delta `spec.md:50`. |
| X4 | RESOLVED | Absolute unchanged-dwell claim withdrawn; shared-connection costs acknowledged; `design.md:301`. |
| X5 | RESOLVED | V2 disposition and B8 map now match baseline: missing publisher logs/releases without increment; `design.md:466`, `internal/app/engine/alertdelivery.go:205`. |
| Y1 | RESOLVED | Clear suppresses only latch application, not escalation; `design.md:208`, delta `spec.md:25`; acknowledgement never changes modes at `internal/obs/notifier.go:840`. |
| Y2 | RESOLVED | `acknowledged_at` ordering removed; timestamp is indeed sampled before the write at `internal/journal/outbox.go:485`; `design.md:264`. |
| Y3 | RESOLVED | All counter-pruning exceptions now appear in delta; delta `spec.md:45`. |
| Y4 | PARTIAL | Transaction contention measurement is improved, but excludes the new full-backlog selection work; `tasks.md:33`; AB1. |
| Y5 | RESOLVED | Failed-publish release precedes gate wait; retained-lease path’s expiration risk disclosed; `design.md:257`, `:269`. |
| Y6 | RESOLVED | Proposal baseline matches missing-publisher code; temporal explanation matches revised D7; `internal/app/engine/alertdelivery.go:205`, `design.md:285`. |
| Z1 | RESOLVED | Threshold decision precedes epoch reset, preserving escalation; `design.md:342`, `tasks.md:54`. |
| Z2 | RESOLVED | Failed escalation always falls back to unconditional latch; `design.md:215`, delta `spec.md:28`. |
| Z3 | RESOLVED | D1/action table/spec consistently distinguish latch-only outcomes; `design.md:57`, `:209`, delta `spec.md:26`. |
| Z4 | RESOLVED | Transaction-read and commit-failure atomicity tests required for all three callers; `design.md:39`, `tasks.md:37`. |
| Z5 | RESOLVED | Self-expanding allowance replaced with fixed-budget measurement and failing positive control; `tasks.md:35`. |
| AA1 | RESOLVED | Fallback no longer depends on the earlier conditional-block return value; `design.md:215`, `tasks.md:50`. |
| AA2 | RESOLVED | Threshold-first transition table and three timing expectations agree; `design.md:338`, `tasks.md:54`. |
| AA3 | RESOLVED | Fixed 250-ms allowance references the actual existing constant; `tasks.md:35`, `internal/app/engine/a098_the_backlog_does_not_delay_protection_test.go:65`. |
| AA4 | RESOLVED | New delivery-only query preserves existing `PendingAlerts` callers/signature; `design.md:107`, `tasks.md:74`. |

**New findings**

| ID | Severity | Finding | Evidence | Suggested fix |
|---|---|---|---|---|
| AB1 | **P1** | **Batch size no longer bounds selection work.** `ORDER BY (attempts >= ?), id LIMIT ?` requires evaluating the pending population; existing `(state,id)` index cannot supply that ordering. Large backlogs therefore introduce backlog-sized connection occupancy before any publish. Protection shares that connection. Existing transport-stall tests and proposed settlement/escalation injections can pass while this regression remains. This is a source-derived cost finding, not a measured latency claim. | `design.md:108`; index `internal/journal/outbox.go:67`; old indexed selection `:518`; shared connection `internal/journal/journal.go:174`; incomplete measurement scope `tasks.md:33`. | Specify a selection/index strategy that avoids full-backlog occupancy, or an enforceable bounded read budget. Add query-plan evidence and large-backlog, concurrent protection tests covering **selection itself**, retaining the fixed dwell allowance. |
| AB2 | **P2** | **E is conservative approximation, not unconditional equivalence or literal acknowledgement precedence.** Spec scenarios promise no late relatch, while the normative text permits capture-gap relatching and mandates fallback relatching after escalation failure. The design acknowledges these exceptions; scenario wording and explicit capture-gap tests lag behind. | `design.md:215`, `:292`; delta `spec.md:36` versus `:79`, `:87`, `:96`; `tasks.md:43`. | Qualify those scenarios with the permitted exceptions. Add `settlement → clear → epoch read → application`, expecting conservative relatch, alongside the existing `epoch read → clear → application` no-relatch test. Describe E as equivalence with enumerated conservative exceptions. |

**Focused safety assessment**

- **D7 locking:** B′ removes the executor-induced `n.mu → journal/g.mu` wait chain. Gate operations contain only map work; publish, settlement, escalation, and logging remain outside. Existing synchronous locking remains hazardous when strategy dispatch holds `g.mu` over transport: `internal/obs/notifier.go:254`, `:484`; `internal/execgw/strategy_entry_gate_authority.go:60`. The new executor does not add that chain. Protection still shares journal connection waits—AB1 matters independently of mutex correctness.

- **E and interleavings:** Two orderings are insufficient. The relevant boundaries are evidence, epoch capture, conditional block, clear, and escalation completion. Capture after evidence prevents discarding a judgement because of a pre-evidence clear. A clear between evidence and capture can instead cause excess blocking, as explicitly allowed. Timestamp comparison is absent from fourth printing and must stay absent. Re-arm before settlement is rejected by token/state CAS; re-arm after committed evidence does not invalidate the earlier obligation (`internal/journal/alert_claim.go:357`; `internal/journal/outbox.go:337`).

- **W1:** The revised rejection is defensible **for delayed application**, because escalation remains even when an old clear suppresses the alert latch. It does not establish that count–clear is safe: `Acknowledge` counts at `internal/obs/notifier.go:870`, clears at `:875`, and independent enqueue does not participate in `n.mu`. That existing race remains a prerequisite for a092 to address.

- **D8:** Threshold-first handling closes Z1. Other-row successes, lease loss, and successful release cannot defeat the counter. Same-row applied failure resets the error counter but advances durable attempts, eventually reaching D1. Partial acknowledgements cannot reset it. Full-clear resets, episode carryover, truncated-list retention, and shutdown exclusion are explicit. Tasks 2.10 cover these policies; tasks 4.1 cover several destructive mutations. No additional fail-open counter interleaving was established.

- **D1:** Reading attempts through `tx.QueryRowContext` before commit is implementable and preserves the same-transaction CAS evidence. Using `j.db` would wait for the connection owned by that transaction. Failed-attempt `NotFound → latch only` parity is correct; successful-publish `NotFound → latch + escalation` is separately correct (`internal/obs/notifier.go:452`, `:519`, `:309`).

- **D6:** Under H, `C=102s`, first threshold is `2C+10=214s`; last batch row is `304s`; maximum stated queue allowance gives `316s`. These are calculations, not measurements or universal bounds. Including `I_list` makes the algebra honest but does not solve AB1’s protection interference.

- **Coverage:** Tasks 2.9/2.10 address late-clear escalation loss, failed-escalation fallback, re-arm, and threshold/reset ordering. Task 4.1 addresses premature epoch capture, timestamp/state-based suppression, and removal of clear increments. Missing explicit coverage: AB1 selection occupancy and AB2’s permitted pre-capture-clear exception.

**D2 and D3**

Both are adequately justified. Three preserves the existing default (`internal/obs/notifier.go:45`) without inventing a misleading time-equivalent retry count. Missing publisher is genuinely undeliverable and must count (`internal/app/engine/exitwiring.go:62`; a092 delta `spec.md:76`). The approximately 4–6-second versus immediate-latch difference is disclosed; it is not time parity.

VERDICT: REJECT — AB1 leaves a new backlog-dependent protection delay outside the proposed selection design and acceptance tests.