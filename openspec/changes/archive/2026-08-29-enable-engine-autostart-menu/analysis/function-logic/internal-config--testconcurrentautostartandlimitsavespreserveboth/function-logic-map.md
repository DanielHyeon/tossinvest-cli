# Function Logic Map: `TestConcurrentAutostartAndLimitSavesPreserveBoth`

- Source: `internal/config/operating_io_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function parameters and local service state | typed Go values; config paths remain profile-scoped | caller/config service | return the underlying error without broadening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range at line 200 | Only the guarded return/call in `TestConcurrentAutostartAndLimitSavesPreserveBoth` | branch result or propagated error | `TestConcurrentAutostartAndLimitSavesPreserveBoth` (self-verifying test) |
| B2 | if at line 201 | Only the guarded return/call in `TestConcurrentAutostartAndLimitSavesPreserveBoth` | branch result or propagated error | `TestConcurrentAutostartAndLimitSavesPreserveBoth` (self-verifying test) |
| B3 | if at line 207 | Only the guarded return/call in `TestConcurrentAutostartAndLimitSavesPreserveBoth` | branch result or propagated error | `TestConcurrentAutostartAndLimitSavesPreserveBoth` (self-verifying test) |
| B4 | if at line 210 | Only the guarded return/call in `TestConcurrentAutostartAndLimitSavesPreserveBoth` | branch result or propagated error | `TestConcurrentAutostartAndLimitSavesPreserveBoth` (self-verifying test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Direct callees recorded in `ast.json` | Preserve the existing config/console/process seam | No implicit retry; errors are returned or rendered by the caller | CodeGraph evidence plus current AST |

## State mutations and fallbacks

- Mutation is limited to the function's declared config key, test fixture, or console lifecycle seam.
- Missing/error paths fail closed; none grants LIVE order capability or bypasses Guardian/interlock checks.

## Safety conclusion

- Safe edit boundary: `TestConcurrentAutostartAndLimitSavesPreserveBoth` and its typed seam only.
- High-risk impact: no; test evidence only.
