# Function Logic Map: `ApplyPositionCampaignFill`

- Source: `internal/journal/position_campaign.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fill scope | exact account/market/day/symbol/side/order | authoritative AppliedFill + persisted attempt/intent | no campaign match is no-op |
| cumulative watermark | canonical non-retreating decimal | immutable scoped order watermark | lower/duplicate no-op |
| per-order remaining | `min(max(0, cap - cumulative), leg residual)` for a live successor; `max(0, cap - cumulative)` otherwise | `positioncampaign.StoredOrderRemaining` (single definition; the domain ledger and offline reconstruction call the same function) | calculation/storage error rolls back fill tx |
| authoritative fill | never rejected for campaign ambiguity/cap/CLOSED | existing fill transaction + Position hook | preserve and latch campaign/reconcile |
| Position generation | first positive fill binds expected successor set-once | authoritative positions projection | mismatch latches reconcile |
| leg state | derived, never computed here | `positioncampaign.LegStateAfterFill` (single definition; offline reconstruction calls the same function) | table refuses the fact: keep the prior state and latch RECONCILE |
| `entry_blocked` | monotone: a fill never clears it | stored column OR the transition, via `positioncampaign.LatchEntryBlocked` | recomputing it from campaign state alone erases the stop fail-closed latch |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | no authoritative match | none | nil | unrelated fill tests |
| B2 | multiple legacy-corrupt matches | evidence command/event + reconcile latch; no watermark guess | nil | ambiguous fill test |
| B3 | lower/duplicate observation | none | nil | retry/restart tests |
| B4 | valid delta | watermark+leg+campaign+event in fill tx | nil | partial/full tests |
| B5 | cap/terminal predecessor ambiguity | preserve delta, recalc live successor remaining against the new leg residual, latch reconcile | nil | late/cap tests |
| B6 | CLOSED late delta | keep CLOSED, advance watermark, durable account reconcile | nil | CLOSED late-fill test |
| B7 | all zero-fill terminal | cancel legs, close campaign, release claim | nil | zero-fill terminal test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `enterReconcileScopeInTx` | durable safety latch | same tx; never broker wait | AST |
| `insertCampaignEvent` | complete projection digest checkpoint | error rolls whole fill tx only for storage corruption, not domain ambiguity | AST |

## State mutations and fallbacks

- Position quantity remains owned by the preceding Project hook.
- A live successor's remaining is bounded by the leg residual as well as its own cap, so a predecessor's
  late fill actually reduces it (design D5). The cap-only form made that recalculation a provable no-op.
- `entry_blocked` is latched, not recomputed: the stop fail-closed latch is not encoded in campaign state,
  so deriving the column from state alone erased it on any ordinary fill.
- Campaign ambiguity is converted to durable evidence and nil, so authoritative fill commits.

## Safety conclusion

- Safe edit boundary: tx-scoped lineage/watermark projection after Position.
- High-risk impact: yes; fill preservation and safety-path continuity are mandatory.
