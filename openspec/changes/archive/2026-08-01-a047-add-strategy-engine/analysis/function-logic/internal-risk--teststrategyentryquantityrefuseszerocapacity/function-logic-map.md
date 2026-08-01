# Function Logic Map: `TestStrategyEntryQuantityRefusesZeroCapacity`

- Source: `internal/risk/contract_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| zero-capacity cases | notional floor zero or stop width zero | test fixture | test failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate zero cases | test-only assertions | fail test | self |
| B2 | non-sentinel result | test failure | fatal | self |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `StrategyEntryQuantity` | subject under test | no retry | test source |

## State mutations and fallbacks

- Test-only; no production mutation.

## Safety conclusion

- Safe edit boundary: zero-capacity test.
- High-risk impact: no.
