# Function Logic Map: `stubScheduleCalendar.TypedMarketCalendar`

- Source: `cmd/tossctl/marketschedule_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| country/date | exact values supplied by adapter tests | test call site | recorded verbatim for assertions |
| response/error | configured fixture values | test setup | returned without transformation |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| happy path | every invocation | locks, records latest request, increments call count | configured response and error | adapter and screen reuse tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sync.Mutex.Lock/Unlock` | make concurrent screen observations race-safe | unlock is deferred | AST + race-safe fixture |

## State mutations and fallbacks

- The mutex protects `country`, `date`, and `calls`; the immutable fixture response/error are returned while the same lock is held.

## Safety conclusion

- Safe edit boundary: test-only calendar fixture state.
- High-risk impact: no; it adds synchronization so the production concurrency assertion can run under `-race`.
