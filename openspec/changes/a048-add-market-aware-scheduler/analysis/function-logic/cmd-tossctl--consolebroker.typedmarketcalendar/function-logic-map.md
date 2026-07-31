# Function Logic Map: `consoleBroker.TypedMarketCalendar`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| shared broker | resolved client implementing the narrow typed calendar read | production official client | resolve/type mismatch returns error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | shared broker resolution fails | none | error | console broker tests |
| B2 | broker lacks typed calendar method | none | fail-closed capability error | seam contract |
| success | capability present | one official read | typed response/error | provenance test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleBroker.resolve` | reuse sole official console client | error propagated | CodeGraph + AST |
| `TypedMarketCalendar` | narrow read capability | official error propagated | provenance fixture |

## State mutations and fallbacks

- No broker object crosses into `internal/console`; only this typed read result does.

## Safety conclusion

- Safe edit boundary: read-only official calendar capability.
- High-risk impact: medium due to live API provenance, with no mutation call.
