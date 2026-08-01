# Function Logic Map: `TestConsoleBrokerTypedMarketCalendarFailsClosed`

- Source: `cmd/tossctl/marketschedule_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| resolver-error case | factory returns a stable sentinel | subtest fixture | sentinel must survive wrapping/propagation |
| missing-capability case | broker satisfies `verifylive.Broker` but not typed calendar reader | subtest fixture | adapter must reject before any calendar read |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | resolver result is not the sentinel error | none | fatal | resolver-error subtest |
| B2 | missing-capability call succeeds or returns a different error | none | fatal | broker-lacks-typed-calendar subtest |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.Run` | isolate both fail-closed contracts and cleanup | subtest fatal is scoped | Go test output |
| `consoleBroker.TypedMarketCalendar` | exercise the production adapter boundary | errors are the expected output | AST + focused test |

## State mutations and fallbacks

- Each subtest replaces and restores `verifyBrokerFactory`; neither fixture exposes a mutating action through the adapter.

## Safety conclusion

- Safe edit boundary: adapter failure-path tests.
- High-risk impact: medium because fabricated calendar provenance must remain impossible on resolution or capability failure.
