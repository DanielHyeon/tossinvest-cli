# Function Logic Map: `ApplyPositionCampaignFill`

- Source: `internal/journal/position_campaign.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fill scope | exact account/market/day/symbol/side/order | authoritative AppliedFill + persisted attempt/intent | no campaign match is no-op |
| cumulative watermark | canonical non-retreating decimal | immutable scoped order watermark | lower/duplicate no-op |
| per-order remaining | `max(0, that order requested_cap - that order cumulative)` | each immutable watermark, not aggregate leg residual | calculation/storage error rolls back fill tx |
| authoritative fill | never rejected for campaign ambiguity/cap/CLOSED | existing fill transaction + Position hook | preserve and latch campaign/reconcile |
| Position generation | first positive fill binds expected successor set-once | authoritative positions projection | mismatch latches reconcile |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | no authoritative match | none | nil | unrelated fill tests |
| B2 | multiple legacy-corrupt matches | evidence command/event + reconcile latch; no watermark guess | nil | ambiguous fill test |
| B3 | lower/duplicate observation | none | nil | retry/restart tests |
| B4 | valid delta | watermark+leg+campaign+event in fill tx | nil | partial/full tests |
| B5 | cap/terminal predecessor ambiguity | preserve delta, recalc current and successor per-order remaining, latch reconcile | nil | late/cap tests |
| B6 | CLOSED late delta | keep CLOSED, advance watermark, durable account reconcile | nil | CLOSED late-fill test |
| B7 | all zero-fill terminal | cancel legs, close campaign, release claim | nil | zero-fill terminal test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `enterReconcileScopeInTx` | durable safety latch | same tx; never broker wait | AST |
| `insertCampaignEvent` | complete projection digest checkpoint | error rolls whole fill tx only for storage corruption, not domain ambiguity | AST |

## State mutations and fallbacks

- Position quantity remains owned by the preceding Project hook.
- Current and successor order remaining quantities are independently cap-based; leg residual remains aggregate-only.
- Campaign ambiguity is converted to durable evidence and nil, so authoritative fill commits.

## Safety conclusion

- Safe edit boundary: tx-scoped lineage/watermark projection after Position.
- High-risk impact: yes; fill preservation and safety-path continuity are mandatory.
