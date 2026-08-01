# Function Logic Map: `TestConsoleOptimizationCommanderUsesSeparatePrivateControlStore`

- Source: `cmd/tossctl/optimization_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| temporary profile | private directory | test fixture | assertion failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B7 | constructor, filesystem mode, evidence and registry assertions | temp DBs only | test failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| command constructor/read | integration verification | test fails immediately | AST |

## State mutations and fallbacks

- Test-only temporary databases; journal path must remain absent.

## Safety conclusion

- Safe edit boundary: test only; high-risk impact: no.
