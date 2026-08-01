# Function Logic Map: `TestConsolePerformanceCapabilitiesOpenOneProfileDatabaseForBothReadSeams`

- Source: `cmd/tossctl/optimization_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| temporary profile and fixed clock | isolated DB | test fixture | assertion failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B7 | open, two seam presence, evidence result, close and post-close refusal | derived temp DB only | test failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| capability constructor/read/close | verifies shared source and lifecycle | immediate failure | AST |

## State mutations and fallbacks

- Test-only derived DB; no journal collection.

## Safety conclusion

- Safe edit boundary: test only; high-risk impact: no.
