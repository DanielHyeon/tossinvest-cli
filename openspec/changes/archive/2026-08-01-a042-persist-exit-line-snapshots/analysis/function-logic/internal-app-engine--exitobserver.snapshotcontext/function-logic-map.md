# Function Logic Map: `ExitObserver.snapshotContext`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| managed position, quote, fallback cycle | stable observation identity and nonnegative generation | a041 identity contract | typed error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | stable ID success/failure | none | context/error | observation identity tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `stableObservationID` | canonical opaque identity | pure | CodeGraph + AST |

## State mutations and fallbacks

- No planned body edit; observation source/time is derived beside the record handoff.

## Safety conclusion

- Safe edit boundary: unchanged.
- High-risk impact: yes.
