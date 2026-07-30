# Function Logic Map: `finishConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function parameters and local service state | typed Go values; config paths remain profile-scoped | caller/config service | return the underlying error without broadening authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 403 | Only the guarded return/call in `finishConsole` | branch result or propagated error | `TestFinishConsole*` shutdown matrix |
| B2 | if at line 406 | Only the guarded return/call in `finishConsole` | branch result or propagated error | `TestFinishConsole*` shutdown matrix |
| B3 | if at line 410 | Only the guarded return/call in `finishConsole` | branch result or propagated error | `TestFinishConsole*` shutdown matrix |
| B4 | if at line 413 | Only the guarded return/call in `finishConsole` | branch result or propagated error | `TestFinishConsole*` shutdown matrix |
| B5 | if at line 415 | Only the guarded return/call in `finishConsole` | branch result or propagated error | `TestFinishConsole*` shutdown matrix |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Direct callees recorded in `ast.json` | Preserve the existing config/console/process seam | No implicit retry; errors are returned or rendered by the caller | CodeGraph evidence plus current AST |

## State mutations and fallbacks

- Mutation is limited to the function's declared config key, test fixture, or console lifecycle seam.
- Missing/error paths fail closed; none grants LIVE order capability or bypasses Guardian/interlock checks.

## Safety conclusion

- Safe edit boundary: `finishConsole` and its typed seam only.
- High-risk impact: yes; covered by explicit lifecycle/config regression tests.
