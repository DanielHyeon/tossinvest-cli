You are an independent ADVERSARIAL senior engineer doing a proposal-freeze review of an OpenSpec change in a Go
repository that runs a real-money automated trading engine. You are read-only: do NOT edit, create, or delete any file,
do NOT run git commands that write, do NOT run network calls, do NOT run the engine or any command that could place an
order. Reading files, grep/rg, and `go vet`-free static reading are fine. Go tests are NOT needed; do not run them.

Repository root: the current directory. Base commit: 4798d3992c95f8dcd0fdf7faf812203910bbf9e3 (working tree Go code is
identical to that commit).

Change under review (read ALL of it first):
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/proposal.md
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/design.md
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/tasks.md
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/specs/engine-safety/spec.md
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/analysis/function-logic/*/function-logic-map.md
  (AST-derived branch maps; branch IDs like B8/B9 refer to these)

Code the change will touch or relies on (verify every claim of the proposal against the code, with file:line):
- internal/app/engine/alertdelivery.go (alertDeliverer.Run / cycle / deliverOne)
- internal/app/engine/auxiliary.go (Context.AlertDeliverer, blockEntryOnDeliveryStop)
- internal/app/engine/gateway.go (restoreAlertEntryLatch, buildGateway)
- internal/app/engine/exitwiring.go (newNotifier)
- internal/journal/outbox.go (PendingAlerts, MarkAlertAttemptFailed, re-arm path in ClaimAlertForDelivery)
- internal/journal/alert_claim.go (ClaimAlertByID, settleUnderClaim, SettleResult)
- internal/journal/operating_mode.go (EscalateOperatingMode, TransitionOperatingMode)
- internal/execgw/retry.go (EntryGate.Block / Clear)
- internal/obs/notifier.go (claimAndDeliver, deliver, notifyCritical, escalate, Acknowledge), internal/obs/alert_lease.go
- openspec/specs/engine-safety/spec.md (the canonical requirement 「등급화된 알림」 and its scenarios)
- openspec/changes/a092-an-alert-does-not-hold-the-stop/specs/engine-safety/spec.md (the sibling change that will remove
  the synchronous deliver from the exit loop)

Safety invariants that override everything: no weakening/delaying of stop-loss or emergency flatten; toggle OFF must
equal prior behaviour; only conservative direction changes to entry blocking; operating-mode relaxation needs a human;
no secrets/account data in logs or gate details.

What to produce (plain text, in English or Korean, concise, evidence with file:line):
1. Findings table: id | severity (P0 = unsafe or unimplementable as written, P1 = must fix before freeze, P2 = should
   record, P3 = editorial) | finding | evidence (file:line) | suggested fix. Attack in particular:
   - Is the proposed judgement input (the row's `attempts` after MarkAlertAttemptFailed) actually obtainable from the
     current API as design D1 says? What value would an implementation really read, and is it stale?
   - Races between the executor (which takes no mutex) and Notifier.Acknowledge / the synchronous sender: can the new
     latch or escalation fire AFTER an operator has acknowledged the backlog, leaving the gate latched with nothing to
     acknowledge, or re-escalating a mode a human just relaxed? Which SettleOutcome values must NOT latch?
   - What happens when MarkAlertAttemptFailed itself errors (FLM deliverOne B10)?
   - Does the design's worst-case latch time (D6) hold? Compute it for: one failing row with instant failures; one with
     10s publish timeouts; a full batch of 10 timing-out rows; k under-limit rows older than a new row.
   - Does R2's ordering (under-limit rows first) create any new starvation or livelock, e.g. rows HeldElsewhere, rows
     whose attempts were already raised by the synchronous path, re-armed rows?
   - Deployment: what does the first cycle after deploying this do to existing PENDING rows with attempts >= limit?
   - Does anything here touch or delay the stop-loss / exit path?
2. Open decision D2 (the limit value): answer with evidence. Note in particular DefaultCriticalAttempts, DefaultRetryDelay,
   DefaultPublishTimeout, alertDeliveryInterval, and compute the elapsed time to reach the limit under each option for
   both the synchronous path and the executor. Is the proposal's "3 attempts ≈ 6 s vs 34 s" comparison correct?
3. Open decision D3 (count "no publisher configured" as a failed attempt): answer with evidence, including what the
   synchronous path and newNotifier's documentation already say about a nil publisher.
4. A final verdict line: `VERDICT: PASS` (freeze may proceed with the recorded P1 fixes applied to design/tasks) or
   `VERDICT: REJECT` (P0 present), with one sentence why.
