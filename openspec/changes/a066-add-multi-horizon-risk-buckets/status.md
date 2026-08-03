# Status — a066-add-multi-horizon-risk-buckets

- Updated: 2026-08-03
- Overall: IN PROGRESS
- Current wave: Wave 1A pure risk core + independent-review hardening GREEN
- Runtime authority: dormant; no journal/Guardian/Gateway/engine integration

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

## Pending Wave 1B

- Journal migration, replay and authoritative transaction primitives.
- Atomic q_final GuardianDecision + owner + all-bucket HELD commit and conflict rollback.
- Guardian/Gateway/entry-loss-lock integration and zero exposure-raising broker request spies.
- KR/US concurrent runtime integration, cancel/expiry, restart/orphan/snapshot-drift and risk-reducing
  bypass tests.
- Full repository validation, independent implementation review and `make gate`.

No existing runtime function, live order path or operating toggle is activated by Wave 1A.
