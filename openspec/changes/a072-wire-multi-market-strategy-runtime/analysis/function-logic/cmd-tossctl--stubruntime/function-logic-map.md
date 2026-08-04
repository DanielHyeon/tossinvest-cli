# Function Logic Map: `stubRuntime`

- Source: `cmd/tossctl/engine_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| test handle | non-nil `testing.T` | caller | test helper panics naturally if invalid |
| build callback | returns a controlled runtime/error from the supplied engine context | caller | result forwarded to command under test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | helper installed | temporarily replaces package factory with context-aware wrapper | callback result | command tests |
| B2 | cleanup | restores exact prior factory | none | `t.Cleanup` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.Helper` | attribute failure to caller | test-only | AST |
| `engineRuntimeFactory` wrapper | ignore clock/logger but preserve request-context signature and engine context | forwards callback exactly once | AST |
| `t.Cleanup` | prevent global seam leakage | always registered after replacement | AST |

## State mutations and fallbacks

- The request context is deliberately ignored only by this test seam; production `engineRuntime` consumes it for paired schedule reads.
- The previous package-global factory is restored even when the test fails.

## Safety conclusion

- Safe edit boundary: signature adaptation only; do not let the seam start real loops or read production artifacts.
- High-risk impact: no — test-only package seam.
