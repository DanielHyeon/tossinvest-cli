# Function Logic Map: `ExitObserver.snapshotContext`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| managed position, typed quote, cycle fallback | persisted position generation/quantity and one authoritative observation source | journal/quote/`ObserveOnce` | identity construction errors refuse judgement |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | observation identity cannot be constructed | none | error | decimal/refusal tests |
| B2 | identity succeeds | none; returns a value-only context | stable context | observation identity tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `stableObservationID` | hashes authoritative `FetchedAt` or the cycle fallback | canonical decimal error propagates | CodeGraph + AST |

## State mutations and fallbacks

- Does not read the clock or mutate state; the already-captured cycle fallback is passed in.

## Safety conclusion

- Safe edit boundary: typed context construction only; evaluation and proposal authority remain downstream.
- High-risk impact: yes, because deduplication identity protects exit proposal reuse.
