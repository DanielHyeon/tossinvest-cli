# Wave 1D pre-edit evidence — owner-wide multi-decision fill accounting

- Date: 2026-08-04
- Scope: `internal/journal`, minimum `internal/riskbucket`, and a066 evidence only.
- Excluded: Gateway, engine, broker, protection, console, runtime toggles and schema migration.

## Existing function evidence

- `CommitRiskBucketAdmission` currently verifies the owner state digest and then refuses every
  same-owner scale-in while any ACTIVE/REPLACED a066 order exists. This prevents the requested
  owner-wide sequence of decision 1/order 1 followed by decision 2/order 2.
- `RegisterRiskBucketOrder` validates one exact confirmed BUY attempt, including decision,
  account, market, symbol and intent quantity, but then refuses when the owner has more than one
  final decision.
- `loadRiskBucketFillTransition` reconstructs reservations, orders, fills and allocations only for
  `target.decisionID`. `persistRiskBucketFillTransition` consequently writes aggregate values back
  to one decision row; using it unchanged for multiple decisions would overwrite rather than
  aggregate HELD/FILLED usage.
- `riskbucket.ApplyFill` keys `FillState.Orders` by broker `OrderID`. The journal requires an
  internal `order_key` identity so separate decisions cannot alias reconstruction state.

## Function Logic Map

### `CommitRiskBucketAdmission` owner-reuse branch

| Branch | Current | Wave 1D |
|---|---|---|
| state digest mismatch | refuse, zero writes | unchanged |
| active registered order | refuse scale-in | permit only after all existing owner/bucket identity checks |
| key/policy drift | refuse, zero writes | unchanged |
| exact owner/key/policy | append decision + five reservations atomically | unchanged |

### `RegisterRiskBucketOrder`

| Branch | Current | Wave 1D |
|---|---|---|
| confirmed decision/scope/quantity mismatch | refuse | unchanged |
| owner has multiple decisions | refuse | accept the exact target decision only |
| broker order ID already belongs to another owner decision | not explicit | typed replay mismatch, zero writes |
| target decision reservations/policy/currency mismatch | refuse | unchanged |
| replacement predecessor not exact ACTIVE same-decision order | refuse | unchanged |

### `loadRiskBucketFillTransition` / persistence

| Stage | Current | Wave 1D |
|---|---|---|
| buckets | target decision only | aggregate exact bounded HELD/FILLED/overage across every owner decision; conservative minimum limit |
| orders | target decision only, keyed by broker ID | every owner order, keyed by immutable `order_key` |
| fills/allocations | target decision only | reconstruct by order_key across the owner, retaining decision-specific reservation IDs |
| write monetary delta | replace target row with projected usage | apply only aggregate deltas to target order's mapped reservation |
| latch | target decision rows | propagate latch state to every matching owner bucket row |
| semantic ambiguity | preserve fill in known-order paths | preserve fill and latch every applicable owner/reservation path |

## Branch Test Map

| Branch | Required RED/GREEN evidence |
|---|---|
| active-order scale-in | second decision and five reservations commit under the exact owner |
| two decision-specific orders | both register only against their exact confirmed intent quantity/decision/scope |
| partial fills on both orders | aggregate usage is exact while each fill allocates only to its decision's five reservation IDs |
| late actual completion | `filled=max(transfer,actual)` updates aggregate usage; UNKNOWN clears only after every owner fill is resolved |
| duplicate/restart | no duplicate allocation and owner-wide digest/reconstruction is stable after reopen |
| cross-decision broker-ID collision | second registration fails with zero writes; a corrupted/ambiguous registered fill is never dropped and latches the full owner scope |
| semantic mapping drift | authoritative fill/Position commits; no partial risk movement; all owner reservations latch |
| KR/US isolation | owner aggregation never crosses account/market/symbol/prospective-generation scope |

## Safety invariant

The aggregate model may block later exposure but cannot reject or roll back an authoritative broker
fill or Position for semantic risk-state ambiguity. Database transport/commit failure remains the only
risk-sidecar reason to fail the enclosing transaction. Actual-evidence completion and order release
remain package-private test seams.

## RED → GREEN result

- RED: the exact second admission with an existing ACTIVE order returned
  `ErrRiskBucketSnapshotMismatch` with `scale-in while risk order accounting is active`.
- GREEN: two owner decisions and two exact confirmed orders now reconstruct and settle by
  `decision_id`/`order_key`; partial fills produce aggregate HELD 70 and FILLED 30 with five
  allocations per fill and zero cross-decision allocations, including after restart.
- GREEN: late actual evidence completes monotonically and clears UNKNOWN only after both owner fills
  are known. Broker-ID collision and semantic ambiguity regressions preserve zero false allocation,
  retain authoritative fill/Position, and latch all matching owner reservation rows.
