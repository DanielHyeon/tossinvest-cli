# Function Logic Map: `Journal.ExitApplierBound`

- Source: `internal/journal/apply_hook.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| hook state | mutex-protected apply hook set | Journal | pure boolean snapshot |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy path (branchless) | read lock only | bound boolean | apply-hook tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RLock/RUnlock` | race-free inspection | no error | AST |

## State mutations and fallbacks

- No journal, broker, Position, toggle, or hook mutation.

## Safety conclusion

- Safe edit boundary: read-only test/diagnostic binding.
- High-risk impact: low; race protected.
