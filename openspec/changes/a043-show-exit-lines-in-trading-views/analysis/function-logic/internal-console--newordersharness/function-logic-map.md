# Function Logic Map: `newOrdersHarness`

- Source: `internal/console/orders_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| test reader/tweaks | deterministic fixture | console tests | default only missing account fixture |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | fixture account absent | mutate test fixture | supply canonical account | origin/evidence tests |
| B2 | option tweaks | test-only options | apply each | harness tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newHarness` | build authenticated test console | test failure propagates | AST + package tests |

## State mutations and fallbacks

- Test-only fixture normalization; production behavior unchanged.

## Safety conclusion

- Safe edit boundary: tests only.
- High-risk impact: no.
