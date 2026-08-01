# Function Logic Map: `TestConsolePerformanceConstructorUsesReadOnlyOpenOnly`

- Source: `cmd/tossctl/optimization_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| production constructor source | contains `performance.OpenReadOnly`; contains no writer/collector entry point | `cmd/tossctl/optimization.go` | static guard fails on missing read-only seam or forbidden authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | source file cannot be read | test failure only | fatal | this test |
| B2 | read-only opener call is absent | test failure only | fatal | this test |
| B3 | iterate forbidden writer symbols | none | inspect all authority markers | this test |
| B4 | any forbidden symbol is present | test failure only | report error | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.ReadFile`, `strings.Contains` | bind composition source to a read-only constructor contract | test-only, no retry | source assertion |

## State mutations and fallbacks

- Reads source text only. It never opens a performance DB or invokes any trading operation.

## Safety conclusion

- Safe edit boundary: static production-wiring regression guard.
- High-risk impact: no direct side effect; prevents accidental acquisition of write/collector authority.
