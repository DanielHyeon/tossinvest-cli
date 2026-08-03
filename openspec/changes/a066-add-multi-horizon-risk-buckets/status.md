# Status — a066-add-multi-horizon-risk-buckets

- Updated: 2026-08-04
- Overall: IN PROGRESS
- Current wave: Wave 1B additive journal admission and replay primitives GREEN; fill/runtime integration pending
- Runtime authority: dormant; no Guardian/Gateway/engine/broker/toggle integration

## Completed in Wave 1B checkpoint

- Additive schema 21→22 for immutable policy/snapshot provenance, final admission decisions,
  prospective/actual owners, five bucket reservations, order/fill/allocation records, events, state
  digests and conservative scope latches. The migration performs no legacy backfill.
- Journal-owned admission recalculation and one atomic transaction for `q_final`, the pre-existing
  HELD Guardian reservation reference, one owner and all five HELD bucket reservations.
- Database-enforced one-owner-per-account/market/symbol arbitration across concurrent journal
  processes, exact digest idempotence and stable mismatch on divergent transaction replay.
- Read/replay projection of owner and HELD/FILLED usage with a persisted state digest. Missing
  legacy state returns `ErrRiskBucketStateUnknown`; drift returns `ErrRiskBucketReplayMismatch`
  and is never silently repaired or deleted.
- Independent source-review hardening rejects immutable key/digest collisions, requires exact
  prospective identity for owner reuse, and gives same-owner scale-in a monotonic sequence plus a
  canonical aggregate replay digest independent of timestamp text ordering.
- Admission receipts contain exactly five unique reservation IDs; scale-in is restricted to the
  owner's exact existing five bucket keys and policy versions.
- Owner market/symbol must equal the corresponding bucket values, and snapshot references must
  exactly match the sealed policy/snapshot evidence consumed by admission; tampering produces zero
  writes and a stable snapshot mismatch.
- Idempotence hashes the canonical ordered full consumed bucket bindings, so equal caps cannot hide
  changed limit/FILLED/HELD values or different authority evidence on retry.

## Completed in Wave 1A

- Exact account-base-minor reservation using worst executable price, official frozen fresh FX,
  haircut, minimum fee and ceil.
- Bounded exact cap search and five-dimensional `q_final` intersection that cannot increase
  `q_candidate` or the existing Guardian cap.
- Typed fail-closed refusals for missing/stale/unknown policy, dimension and arithmetic evidence.
- Pure fill accounting for proportional HELD transfer, greater actual exposure, late actual-evidence
  completion, duplicate/retry idempotence, crash-pure error rollback and overage/unknown entry latch.
- Pure one-symbol owner acquisition, same-owner reuse, prospective-to-actual set-once binding and
  conservative idempotent clean release.
- Review hardening: bounded stored-minor addition/recompute with crash-pure overflow refusal;
  immutable policy/snapshot provenance bound to exact bucket identity and snapshot amounts; fresh
  release attestation bound to owner/lane/campaign/prospective/actual generations; exact actual-FX
  quote/base pair binding; deterministic duplicate-owner reconstruction refusal; bucket-local latch
  enforcement in `EntryBlocked`.

## Verification checkpoint

| Check | Result |
|---|---|
| Initial RED compile | PASS (missing implementation symbols observed) |
| `go test ./internal/riskbucket` | PASS |
| `go test -race ./internal/riskbucket` | PASS |
| `go vet ./internal/riskbucket` | PASS |
| Review-fix RED compile | PASS: missing provenance, attestation and currency-pair contracts observed |
| Repeated overflow/reconstruction/property tests, 25x | PASS |
| `FuzzReservationIsMonotone`, 3 seconds | PASS, 475,508 executions |
| `FuzzApplyFillRetryIsPure`, 3 seconds | PASS, 222,136 executions |
| Focused statement coverage | 77.6% |
| `git diff --check` | PASS |
| CodeGraph sync/status | PASS: 1,368 files / 23,745 nodes / 77,709 edges |
| CodeGraphContext advisory update | INCOMPLETE: stalled after DB load; terminated |
| Wave 1B focused journal tests | PASS |
| Wave 1B focused journal race | PASS |
| `go vet ./internal/journal` | PASS |
| Strict OpenSpec validation | PASS |
| Full `go test ./internal/journal` | INCOMPLETE: no-output timeout at 240 seconds; focused suites remain GREEN |

## Pending Wave 1B integration

- Authoritative fill/cancel/expiry and clean owner-release journal transactions, including
  replacement/predecessor-late fill and late actual-evidence completion.
- Atomic integration with the actual GuardianDecision writer (the new journal decision is a dormant
  sidecar and does not claim Guardian authority).
- Guardian/Gateway/entry-loss-lock integration and zero exposure-raising broker request spies.
- KR/US concurrent runtime integration, cancel/expiry, restart/orphan/snapshot-drift and risk-reducing
  bypass tests.
- Full repository validation, independent implementation review and `make gate`.

No existing runtime function, live order path or operating toggle is activated by Wave 1B.
