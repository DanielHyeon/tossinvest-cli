# Function Logic Map: `consoleBroker.TypedMarketCalendar`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| shared broker | resolved client implementing the narrow typed calendar read | production official client | resolve/type mismatch returns error |
| account reference returned by `resolve` | exact trimmed broker identity | shared resolver cache | deliberately discarded; calendar provenance is market/date scoped |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | shared broker resolution fails | none | original error | `TestConsoleBrokerTypedMarketCalendarFailsClosed/resolver_error` |
| B2 | broker lacks typed calendar method | none | fail-closed capability error naming the missing read | `TestConsoleBrokerTypedMarketCalendarFailsClosed/broker_lacks_typed_calendar` |
| success | capability present | discards returned account reference and performs one typed official read | typed response/error | direct adapter and provenance tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleBroker.resolve` | reuse sole official console client without a second account lookup | error propagated; returned account reference is not calendar input | CodeGraph + AST |
| `TypedMarketCalendar` | narrow read capability | official error propagated | provenance fixture |

## State mutations and fallbacks

- No broker object or account reference crosses into `internal/console`; only this typed read result does.
- The resolver still retains the exact account reference for the order-evidence consumer.

## Safety conclusion

- Safe edit boundary: read-only official calendar capability.
- High-risk impact: medium due to live API provenance, with no mutation call.
