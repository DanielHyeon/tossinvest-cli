# Status — a066-add-multi-horizon-risk-buckets

- Updated: 2026-08-04
- Overall: IN PROGRESS
- Current wave: Wave 1C single-decision authoritative journal fill accounting GREEN; multi-decision owner/runtime integration pending
- Runtime authority: dormant; no Guardian/Gateway/engine/broker/toggle integration

## Completed in Wave 1B checkpoint

- Released schema 21→22 remains byte-for-byte immutable for policy/snapshot provenance, final
  admission decisions, owners, reservations and its original order/fill evidence. Wave 1C adds a
  separate atomic 22→23 transition: released order/fill/allocation tables are retained under `_v22`
  legacy names and new scoped tables are created without auto-promoting legacy rows.
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

## Completed in Wave 1C checkpoint

- A confirmed official BUY order can be registered against its exact risk decision, owner and five
  reservations. Broker order IDs are scoped by account and market, so equal KR/US IDs cannot cross
  release or actual-evidence boundaries.
- `Journal.RecordFill` applies the a066 sidecar inside the same `BEGIN IMMEDIATE` transaction as its
  authoritative fill snapshot, Position and existing campaign/exit hooks. A later hook failure rolls
  back both projections; semantic a066 drift preserves the fill and Position and latches new entry.
- Partial fills transfer proportional HELD in every bucket. Monotonic actual evidence completes
  usage as `filled=max(transfer,actual)`; missing actual price/fee/FX latches
  `UNKNOWN_ACTUAL_RISK` without rejecting the fill.
- Replacement ownership is handed to the successor reservation. A replaced predecessor cannot be
  released independently; predecessor-late or post-cancel/expiry fills remain authoritative and
  latch overage when released HELD is insufficient.
- Duplicate/retry, restart replay, orphan mapping, snapshot drift and crash atomicity have focused
  coverage. All monetary sums use bounded exact 256-bit integers rather than SQLite integer casts.
- SELL/risk-reducing fills bypass entry accounting. Single-decision owners are supported; owners with
  multiple final decisions fail closed at order registration until aggregate order binding lands.
- Order replay seals are canonical ordered DTOs, never JSON encodings of struct-keyed maps. The
  reservation-policy seal and quote/base currencies are derived from the five persisted authority
  policies; caller strings cannot become authority. Order quantity must equal the one exact confirmed
  intent quantity. US coverage pins USD/KRW while KR remains isolated.
- Ambiguous or corrupt sidecar reconstruction is classified separately from database transport
  failure. It preserves the broker fill and Position, latches every applicable reservation and owner,
  and records `FILL_UNACCOUNTED`; genuine storage failures still roll back the outer transaction.
- v22→v23 preservation, failed-migration rollback and older-v22-build refusal are step-pinned. The
  original v22 SQL matches commit `4aee6853`; a v22 binary sees v23 as `ErrSchemaTooNew`.
- Actual-evidence completion and cancel/expiry release remain package-private test seams with zero
  non-test callers. They cannot become production authority until official sealed fill evidence and
  journal-derived confirmed cancellation/expiry, broker-zero and clean lifecycle adapters land.

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
| Wave 1C focused journal tests | PASS (including canonical replay, replacement/release and ambiguous non-drop regressions) |
| Wave 1C focused journal race | PASS (54.071s) |
| Wave 1C existing fill/apply-hook regressions | PASS |
| Wave 1C journal+riskbucket vet | PASS |
| Wave 1C strict OpenSpec validation / diff check | PASS |
| v22 SQL versus released `4aee6853` | PASS: exact diff |
| v22→v23 row preservation / no auto-promotion | PASS |
| broken v23 rename+version atomic rollback | PASS |
| v22 build opening v23 | PASS: `ErrSchemaTooNew` |
| v22→v23 pre-migration backup | PASS: self-contained v22 copy with exact legacy rows |
| v23 hardened focused journal race | PASS (34.960s) |

## Pending integration

- Multi-decision aggregate order registration and authoritative prospective-to-actual owner binding,
  plus clean owner release after reconciliation/protection/sell evidence is complete.
- Atomic integration with the actual GuardianDecision writer (the new journal decision is a dormant
  sidecar and does not claim Guardian authority).
- Guardian/Gateway/entry-loss-lock integration and zero exposure-raising broker request spies.
- KR/US concurrent runtime integration and independent lane-operation tests.
- Full repository validation, independent implementation review and `make gate`.

No existing runtime function, live order path or operating toggle is activated by Wave 1B.
