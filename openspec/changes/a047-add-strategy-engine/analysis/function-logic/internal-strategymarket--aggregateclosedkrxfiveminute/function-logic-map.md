# Function Logic Map: `AggregateClosedKRXFiveMinute`

- Source: `internal/strategymarket/bars.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| caller raw slice | any value | deprecated pre-a047 API | always source refusal; cannot mint a bar |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | every invocation | none | `RefusalSource` requiring opaque official page | raw API refusal test / replacement API tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | legacy function is a fail-closed compatibility tombstone | no retry/timeout | AST |

## State mutations and fallbacks

- No external mutation/fallback. Raw candle slices cannot directly mint a verified bar after this change.

## Safety conclusion

- Safe edit boundary: the legacy raw-slice path is unable to return a valid proof.
- High-risk impact: yes — verification moved to the new opaque-page sealer and adapter integration tests.
