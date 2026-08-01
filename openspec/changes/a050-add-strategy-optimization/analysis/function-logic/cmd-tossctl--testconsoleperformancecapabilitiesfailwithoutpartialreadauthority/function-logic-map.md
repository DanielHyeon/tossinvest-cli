# Function Logic Map: `TestConsolePerformanceCapabilitiesFailWithoutPartialReadAuthority`

- Source: `cmd/tossctl/optimization_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| blocked path | file where directory required | test fixture | assertion failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | fixture creation and zero-capability error assertion | temp file only | test failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| capability constructor | verifies atomic failure | immediate failure | AST |

## State mutations and fallbacks

- Test-only filesystem fixture.

## Safety conclusion

- Safe edit boundary: test only; high-risk impact: no.
