# Function Logic Map: `ExitObserver.observe`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| broker quotes | symbol, positive last, optional fetched timestamp | `domain.Quote` | invalid last omitted; empty answer errors |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | retrier/query fails | no state mutation | error | existing query tests |
| B2 | last price non-positive | none | omit quote | existing invalid-price test |
| B3 | at least one valid quote | preserve price plus `FetchedAt` | return map | identity tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Retrier.Query` | stamp and retry the price read | configured retry contract | CodeGraph + AST |
| `Prices.Prices` | fetch one quote batch | source error propagated | CodeGraph + AST |

## State mutations and fallbacks

- Preserve the source timestamp instead of collapsing the observation to a decimal string.

## Safety conclusion

- Safe edit boundary: return a typed observed quote without changing query frequency.
- High-risk impact: yes — price observations trigger exits.
