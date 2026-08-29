# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.remote` | nil or validated remote session manager | `New` | remote login/logout routes exist only in remote mode |
| all mutating handlers | exact paths | static route allowlist + spec | session or CSRF failure rejects before handler |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | remote mode enabled | register login and guarded logout | local mode has neither route | remote route tests |
| B2 | optional update download seam available | register/serve optional capability | absent seam renders refusal | existing route tests |

The change adds exact `/settings/autostart` registration with
`session0(mutating(...))`; it adds no branch.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `session0` | authenticate every non-health route | rejects invalid session | AST + static tests |
| `mutating` | enforce POST + CSRF | rejects before handler | AST + static tests |
| `startExclusive` | serialize engine/verify starts | no concurrent start across update commit | existing tests |

## State mutations and fallbacks

- Registers handlers in a fresh `ServeMux`; no request is executed here.
- State-changing path inventory is a closed static-test allowlist.

## Safety conclusion

- Safe edit boundary: one exact route using the same two wrappers as gate/trading saves.
- High-risk impact: yes — ON may call the engine-start seam after both gates.
