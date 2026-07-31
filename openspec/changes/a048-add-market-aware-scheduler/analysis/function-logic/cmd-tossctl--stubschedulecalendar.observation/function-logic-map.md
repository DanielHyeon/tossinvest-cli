# Function Logic Map: `stubScheduleCalendar.observation`

- Source: `cmd/tossctl/marketschedule_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| recorded fixture state | country, date, and non-negative call count | `stubScheduleCalendar.TypedMarketCalendar` | returned as one locked snapshot |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| happy path | every observation | locks fixture during read | latest country/date/call count | direct adapter and provenance tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sync.Mutex.Lock/Unlock` | prevent torn or raced assertions | unlock is deferred | AST + focused tests |

## State mutations and fallbacks

- Read-only snapshot; it does not reset the counter or mutate fixture values.

## Safety conclusion

- Safe edit boundary: test-only fixture accessor.
- High-risk impact: no; it removes unsynchronized test reads.
