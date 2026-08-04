# Function Logic Map: `SetChainForTest`

- Source: `internal/execgw/export_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| test chain function | non-nil deterministic test evaluator | test only | restore closure rebinds the prior chain |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy path | temporarily swaps the test chain | returns restore closure | existing execgw tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | direct test-global assignment | test process only | AST |

## State mutations and fallbacks

- Mutates only the package test seam and provides exact restoration.

## Safety conclusion

- Safe edit boundary: this symbol is compiled only into tests.
- High-risk impact: no production impact.
