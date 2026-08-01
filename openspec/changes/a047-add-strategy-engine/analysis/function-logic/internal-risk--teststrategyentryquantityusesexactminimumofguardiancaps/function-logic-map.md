# Function Logic Map: `TestStrategyEntryQuantityUsesExactMinimumOfGuardianCaps`

- Source: `internal/risk/contract_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| table cases | risk/quantity/notional minimums | test fixture | test failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate all cap cases | test-only assertions | fail test | self |
| B2 | result/error differs | test failure | fatal | self |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `StrategyEntryQuantity` | subject under test | no retry | test source |

## State mutations and fallbacks

- Test-only; no production mutation.

## Safety conclusion

- Safe edit boundary: test table.
- High-risk impact: no.
