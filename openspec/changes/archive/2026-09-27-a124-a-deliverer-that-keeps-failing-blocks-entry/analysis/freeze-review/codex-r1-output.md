Static review only. No files changed, tests run, network calls, or engine execution.

Paths below are repository-relative. `change/` abbreviates `openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/`.

| ID | Severity | Finding | Evidence (file:line) | Suggested fix |
|---|---|---|---|---|
| F1 | P0 | D1 requires a return value that does not exist. `MarkAlertAttemptFailed` returns `SettleResult`, containing no attempts count. Using `alert.Attempts` reads the pre-cycle snapshot; adding one does not make it authoritative. Another sender can increment attempts, or acknowledgement/re-arm can reset the episode before acquisition. | `change/design.md:18`; `internal/journal/outbox.go:467`; `internal/journal/alert_claim.go:140`; `internal/app/engine/alertdelivery.go:150,168`; `internal/journal/outbox.go:337` | Specify a journal API returning the committed post-increment count for the matching PENDING/token episode. Add journal API changes and stale-snapshot/re-arm cases to tasks. A separate unguarded lookup is insufficient. |
| F2 | P0 | Settlement and latch/escalation are not atomic with acknowledgement. Even `SettleApplied` can become obsolete before `Block`. An operator can acknowledge everything, clear the gate, relax the mode, then receive a stale latch and fresh escalation. Idempotence does not prevent this. | `internal/journal/alert_claim.go:330–334`; `internal/obs/notifier.go:851–875`; `internal/journal/operating_mode.go:398–445`; `change/design.md:18–21` | Define a local settlement/acknowledgement coordination protocol, plus transactionally guarded escalation against the current alert episode. No shared lock across publishing. Add deterministic interleaving tests; revise the non-goal prohibiting acknowledgement-path changes if necessary. |
| F3 | P0 | B10 has no safety policy. If failed-attempt recording repeatedly errors, attempts never advance, so threshold-only logic can leave entries open indefinitely after a092 removes synchronous handling. The executor logs, releases, and continues. | `internal/app/engine/alertdelivery.go:221–226`; `internal/app/engine/alertdelivery.go:124–133`; `change/tasks.md:3` | Specify fail-closed handling for accounting failure, independent of an unavailable counter; attempt durable escalation, retain the memory block if persistence fails, and record sanitized diagnostics. Distinguish failure from normal lease loss/acknowledgement. |
| F4 | P1 | D6 is not a worst-case bound. It counts one preceding batch but omits other rows serviced between the target’s retries, arbitrary queue depth, leases, journal work, and release waits. Ten timeout rows already falsify it. | `change/design.md:55`; `internal/app/engine/alertdelivery.go:133,150–161,249–253`; `internal/obs/ntfy.go:95–100` | Replace with a conditional queue/service model and explicit assumptions. State that unrestricted backlog/contention/storage stalls have no finite universal bound. Measurements cannot establish an absolute worst case. |
| F5 | P1 | R2 fixes only one starvation case. Ten under-limit `HeldElsewhere` rows repeatedly consume the entire selection without advancing attempts. Sustained under-limit arrivals can indefinitely exclude exhausted rows; even without arrivals, the oldest ten exhausted failures monopolize their tier. Re-armed old IDs regain priority. | `change/design.md:43`; `internal/journal/outbox.go:518–522,334–342`; `internal/app/engine/alertdelivery.go:178–192`; `internal/journal/alert_claim.go:162–165` | Specify claimable selection/refill and bounded scanning. Define fairness within tiers. Either narrow the starvation promise or revise “exhausted rows only in spare slots,” which cannot guarantee their eventual service under continuous arrivals. |
| F6 | P1 | Deployment semantics are incomplete. Existing PENDING rows already over limit are demoted immediately. D1 checks only after another failed attempt, so they may never trigger durable escalation if starved, held, or successfully delivered next. Startup restores only the memory latch. | `change/design.md:18,43`; `internal/app/engine/gateway.go:153–167,269`; `internal/app/engine/alertdelivery.go:104–105` | Define reconciliation of existing exhausted PENDING episodes, independently of delivery priority, with the same acknowledgement race protection. Record expected first-start behavior. |
| F7 | P1 | “Toggle OFF means no Notifier” is not established and is false for the engine automation gate as wired here. `buildGateway` constructs the notifier unconditionally; the executor accepts a nil notifier/publisher. | `change/design.md:61`; `internal/app/engine/engine.go:504–516`; `internal/app/engine/gateway.go:323`; `internal/app/engine/auxiliary.go:156–173` | Name the exact toggle and prove its actual boundary. Add an OFF-preservation acceptance case; do not equate missing publisher with disabled feature. |
| F8 | P1 | D5 omits escalation-error behavior and sanitized gate/log content. Copying the existing synchronous gate-detail format would include raw errors. Mode-transition errors can also contain account references. | `change/design.md:49–51`; `internal/obs/notifier.go:563–571`; `internal/journal/operating_mode.go:394,449`; `internal/app/engine/auxiliary.go:206–211` | Block first with fixed/count-only detail; explicitly handle failed escalation and retry policy. Never copy raw transport/DB errors, account identifiers, payloads, or tokens into new gate details/log fields. |
| F9 | P3 | `Block` does not refresh an existing reason’s detail. D1’s description is inaccurate. | `change/design.md:19`; `internal/execgw/retry.go:526–532` | Describe it as insert-if-absent; repeated calls preserve the original detail. |

The race in F2 has two distinct forms:

- Acknowledgement during publish: failure recording returns `SettleAlreadySettled`. Neither latch nor escalation may follow.
- Acknowledgement after successful failure recording: executor receives `SettleApplied`, pauses; operator acknowledges and clears, possibly relaxes mode; executor resumes and blocks/escalates from stale evidence. Returning the updated count fixes F1, but does not fix this race.

`SettleLeaseLost` and `SettleAlreadySettled` must NOT trigger delivery-failure latch/escalation. `SettleNotFound` needs a separate fail-closed ledger-integrity response, not an invented attempts count. Errors and unknown outcomes also need explicit handling; the zero-value result has `SettleApplied`, so error checks must come first (`internal/journal/alert_claim.go:107–119,316–334`). Release outcomes matter too: the executor currently discards them (`internal/app/engine/alertdelivery.go:252`).

A second empty-backlog `Acknowledge` can technically clear a late memory latch (`internal/obs/notifier.go:854–875`). Therefore this is not necessarily an irrecoverable lock, but it violates the promised acknowledgement semantics and can leave durable mode escalation after the acknowledged episode ended. The sibling explicitly forbids resurrecting that cause (`openspec/changes/a092-an-alert-does-not-hold-the-stop/specs/engine-safety/spec.md:66`).

D2: choose three committed failed attempts per current episode, but correct the timing rationale.

Defaults are:

- `DefaultCriticalAttempts = 3`: `internal/obs/notifier.go:45`.
- `DefaultRetryDelay = 2s`: `internal/obs/notifier.go:48`.
- `DefaultPublishTimeout = 10s`: `internal/obs/alert_lease.go:17`.
- `alertDeliveryInterval = 2s`, batch 10: `internal/app/engine/alertdelivery.go:63,74`.

The executor runs immediately, then sleeps after the entire batch (`internal/app/engine/alertdelivery.go:122–134`). The following calculations exclude journal, scheduling, logging, and escalation costs, assume successful failure recording, and start at the first cycle/attempt:

| Situation, limit 3 | Time to threshold |
|---|---:|
| Synchronous sender, instant failures | 4s |
| Executor, one row, instant failures | 4s |
| Synchronous sender, three 10s timeouts | 34s |
| Executor, one row, three 10s timeouts | 34s |
| Executor, ten rows all timing out | First row 214s; last row 304s |

For the full batch, cycles start at 0, 102, and 204 seconds. Third failures occur at 214, 224, …, 304 seconds. D6 predicts only `100 + 3×2 + 3×10 = 136s`, so it fails even before storage overhead.

For an arrival during an otherwise idle executor’s sleep, add up to 2s: instant failures take approximately 4–6s; timeout failures 34–36s. Arrival during active batch processing can wait much longer.

For k older, initially zero-attempt rows, assuming all fail, no leases/re-arms/new arrivals, and limit 3:

- Instant failures: let `q = floor(k/10)`. Older full groups consume three cycles each. The new row reaches threshold at approximately `6q + 4` seconds from the initial cycle start.
- Ten-second failures: each older full group can consume `3×102 = 306s` before the next group starts. With `r = k mod 10`, its first attempt starts after approximately `306q + 10r` seconds. Its own retries then include intervening batch work.
- Concrete counterexample: k = 10 gives the new row its first failure at 316s and third failure at 520s, because exhausted older rows fill spare slots.

These are conditional service calculations, not universal bounds. Held leases, repeated accounting errors, or re-arms can prevent progress. The executor itself imposes no publish timeout; the concrete Ntfy implementation does (`internal/app/engine/alertdelivery.go:215`; `internal/obs/ntfy.go:95–100`). The existing synchronous budget also includes writes and release: its documented default is 54s, not 34s (`internal/obs/alert_lease.go:19–29,57–66`).

For any chosen limit L, isolated-row timing is:

- Instant failure: `2(L−1)` seconds.
- Ten-second timeout: `10L + 2(L−1) = 12L−2` seconds.
- Full executor batch, row position j: `(L−1)×102 + 10j` seconds.

Thus option ㄴ has no universally time-equivalent attempt count. Picking 17 from `34/2` gives 32s for instant failures but 202s for timeouts; 18 gives 34s and 214s. Option ㄷ remains unspecified until L is chosen; the same formulas apply. The synchronous path remains at three attempts unless explicitly changed.

“3 attempts ≈ 6s versus 34s” compares different failure durations and start points. Under matching conditions, isolated synchronous and executor timing is the same. Three is defensible as the existing retry-count contract, not as a deliberately faster equivalent timeout policy.

D3: count “no publisher configured” as a failed delivery opportunity.

This matches existing safety intent:

- `newNotifier` explicitly documents nil publisher → PENDING, gate latch, ENTRY_BLOCKED (`internal/app/engine/exitwiring.go:60–70`).
- Synchronous `deliver` breaks immediately on nil publisher, releases, then blocks; `notifyCritical` escalates (`internal/obs/notifier.go:429–431,540–573,223–228`).
- Executor currently releases without incrementing (`internal/app/engine/alertdelivery.go:205–212`).

Record a fixed sanitized cause through the same guarded failure-accounting path. Note the behavioral difference: synchronous nil-publisher handling blocks immediately, while the proposed three-cycle policy takes approximately 4–6s when isolated. It preserves the consequence, not the latency. The sibling already explicitly requires counting this case (`openspec/changes/a092-an-alert-does-not-hold-the-stop/specs/engine-safety/spec.md:76,108–110`).

No direct stop-loss/flatten edit is proposed. The entry check bypasses non-exposure-increasing mutations (`internal/execgw/gateway.go:855–859`), and this trigger maps to ENTRY_BLOCKED, never automatic HALT_ALL (`internal/journal/operating_mode.go:537–545`). However, “no exit delay” is not proved merely by using another goroutine: the journal has one connection (`internal/journal/journal.go:174`), and additional transactions can contend with exit work. Race fixes must never hold a shared lock or transaction across publishing. The existing synchronous exit wait remains until a092 lands.

VERDICT: REJECT — P0 defects leave the proposed judgement input unavailable, permit stale post-acknowledgement safety actions, and leave persistent failed-attempt accounting errors without a fail-closed policy.