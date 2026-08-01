# Function Logic Map: `TestRunConsoleWiresAndClosesPerformanceWithoutJournalCollection`

- Source: `cmd/tossctl/optimization_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| production source | current console.go | repository | assertion failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | read source, required wiring strings, forbidden authority strings | none | test failure | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.ReadFile`, string checks | static least-authority guard | immediate failure | AST |

## State mutations and fallbacks

- Read-only static test.

## Safety conclusion

- Safe edit boundary: test only; high-risk impact: guards production authority.
