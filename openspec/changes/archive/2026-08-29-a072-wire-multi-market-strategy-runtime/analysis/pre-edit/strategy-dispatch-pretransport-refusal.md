# Strategy dispatch pre-transport refusal contract

**Status:** approved implementation refinement, 2026-08-04  
**Scope:** journal-only `CLAIMED + RESERVED -> REFUSED + RELEASED`; paired KR/US

## Decision

Gateway validation can fail after a lease was claimed but before a durable transport-start marker.
That provably-not-sent branch must atomically consume the exact lease and release its exact aggregate
plus five-bucket holds. It must not call `BeginStrategyDispatchSubmitting`, a broker, activation,
recovery or submission code.

The Gateway-provided refusal reason is record-only. A closed journal enum maps it to a canonical
stored code, but it does not satisfy or strengthen any owner, fence, lease, first-leg or reservation
check. Supplying a different valid reason therefore cannot make stale authority current.

## Requirements

- PR-FR-1: The API MUST require the exact current central owner epoch/fencing token and an exact
  `CLAIMED` lease revision in one SQL transaction.
- PR-FR-2: The request MUST repeat the complete immutable durable lease plan. The transaction MUST
  compare every plan field and reject cross-attempt, cross-market, cross-campaign, cross-leg,
  cross-decision, cross-reservation, router/lane, time or digest substitution.
- PR-FR-3: The transaction MUST prove the exact v26 first-leg binding, its operation/client-order
  identity and prospective generation, one exact aggregate reservation and exactly five distinct
  monetary buckets. A row may be `HELD`, or it may already be safely `RELEASED`; mixed safe-release
  states are normalizable and MUST NOT strand the lease.
- PR-FR-4: Success MUST atomically release every remaining HELD row while preserving metadata on
  already-released rows, prove the aggregate plus five buckets are all released, set lease
  `REFUSED + RELEASED`, increment revision once and append one transition using only the canonical
  record-only reason.
- PR-FR-5: Replay, stale revision, terminal/non-CLAIMED state, old owner, owner A→B→A, binding drift
  and cardinality mismatch MUST leave lease and holds unchanged.
- PR-FR-6: KR and US MUST use the same API and test matrix. One market's success is not completion
  without the peer.
- PR-FR-7: The API MUST contain no broker, Gateway, activation, recovery, submit, lease-mint or live
  operating capability.
- PR-FR-8: Missing/wrong-scope rows, wrong prospective generation, `FILLED` or partial monetary
  state, nonzero fill/overage, or any pre-existing risk-order mapping are integrity failures. They
  MUST roll back unchanged rather than be rewritten as a clean pre-transport refusal.

## Canonical refusal classes

- `GATEWAY_DECISION_REFUSED`
- `GATEWAY_PROTECTION_REFUSED`
- `GATEWAY_RESERVATION_REFUSED`
- `GATEWAY_ACCOUNT_BASE_FX_REFUSED`
- `GATEWAY_POLICY_REFUSED`

These values classify an already-proven refusal only; none is an authorization credential.

## Paired acceptance matrix

| Case | KR | US | Expected |
|---|---|---|---|
| exact current claim | KR/KRW binding | US/USD binding | one atomic `REFUSED + RELEASED`, revision +1, aggregate 1 + buckets 5 released |
| replay/concurrency | same request twice | same | exactly one winner; terminal lease never revives |
| owner ABA | owner A→B→A | same | old epoch/token fenced; holds unchanged |
| cross attempt/market | peer operation or US scope | peer operation or KR scope | fail before mutation |
| rollback | abort lease terminal update | same | claim and all six holds remain intact |
| reason substitution | another valid canonical reason | same | cannot bypass stale owner/revision/binding |
| safe partial release | aggregate or 1..5 buckets already RELEASED | same | preserve prior release metadata, normalize remaining HELD rows, terminalize once |
| integrity drift | missing/FILLED/partial/nonzero fill or order mapping | same | rollback unchanged; no normalization |

## Out of scope

- Gateway integration and mapping from execgw rejection details.
- `CLAIMED -> SUBMITTING`, broker transport and post-transport outcome classification.
- Restart recovery, lease issuance, activation and operating toggles.
- Schema changes.
