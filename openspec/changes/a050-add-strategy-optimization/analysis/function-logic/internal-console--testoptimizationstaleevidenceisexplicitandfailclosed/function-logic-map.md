# Function Logic Map: `TestOptimizationStaleEvidenceIsExplicitAndFailClosed`

- Source: `internal/console/optimization_ui_contract_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| stale evidence view | stale status plus concrete missing reason | fake commander | UI must not present stale data as complete/zero |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | required stale-state marker loop | assertions only | exact missing marker | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| dashboard harness / string checks | render evidence banner and status strip | local/no retry | AST |

## State mutations and fallbacks

- Test-local evidence mutation only; no optimization apply.

## Safety conclusion

- Safe edit boundary: stale evidence presentation.
- High-risk impact: no.
