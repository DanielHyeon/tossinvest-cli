# Function Logic Map: `TestConsoleBrokerTypedMarketCalendarReusesResolutionAndKeepsExactAccountRef`

- Source: `cmd/tossctl/marketschedule_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| factory result | calendar-capable broker plus whitespace-padded exact account reference | test fixture | any build/delegation drift fails the test |
| adapter calls | two identical KR/date reads | test loop | errors fail immediately |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | two adapter calls | invokes shared calendar adapter twice | each must succeed | self |
| B2 | adapter call errors | none after error | fatal | self |
| B3 | factory build count is not one | none | fatal | self |
| B4 | cached resolve errors | none | fatal | self |
| B5 | cached account reference is not exact trimmed value | none | assertion failure | self |
| B6 | delegated country/date/count differ | none | assertion failure | self |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleBroker.TypedMarketCalendar` | exercise direct adapter without an account identity output | error fails test | AST + focused test |
| `consoleBroker.resolve` | inspect the cache after adapter discarded its returned reference | error fails test | exact-reference assertion |
| `stubScheduleCalendar.observation` | verify delegation and call count | locked snapshot | fixture evidence |

## State mutations and fallbacks

- Replaces the package-level factory for the test and restores it with `t.Cleanup`.
- Proves discarding the adapter's account-reference return does not erase the resolver cache used by order lineage.

## Safety conclusion

- Safe edit boundary: command adapter integration test.
- High-risk impact: medium because it guards cross-change account identity and shared live-read construction without issuing a mutation.
