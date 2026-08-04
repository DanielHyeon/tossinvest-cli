# Function Logic Map: `Journal.prepareQFinalCampaignFirstLeg`

- Source: `internal/journal/strategy_first_leg_atomic.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request | exact q_final, strategy, campaign, router and optional weekly binding | caller plus sealed lane lineage | typed mismatch, zero writes |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B8 | risk intent, q_final, time, lineage, owner or scope invalid | none | typed refusal | existing substitution matrix |
| B9 | weekly lane has no exact active durable reservation, or non-weekly lane supplies one | none | weekly binding refusal | paired weekly first-leg test |
| success | all cross-family facts match | returns sealed prepared value | nil | paired KR/US admission |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| q_final/risk helpers | recompute authoritative quantity and bucket binding | pure refusal | CodeGraph + AST |
| weekly reservation validator | bind active journal key to weekly lane/campaign/market/ordinal | fail closed | CodeGraph + AST |

## State mutations and fallbacks

- Pure preparation only; no database mutation. Weekly binding is incorporated into request and record digests.

## Safety conclusion

- Safe edit boundary: validate and digest the optional weekly binding without moving transaction ownership.
- High-risk impact: yes/no
