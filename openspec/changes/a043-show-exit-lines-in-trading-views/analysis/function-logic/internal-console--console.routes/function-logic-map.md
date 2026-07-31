# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| console options | optional remote/auth/read/write seams fixed at construction | `New` / `Options` | absent capability renders unavailable or route refuses through existing wrappers |
| `/positions` request | authenticated; a043 requires GET/HEAD only | operator-console spec | non-read method must return 405 before handler |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | remote exists and is not trusted-network mode | register login/logout | mux registration only | existing remote route tests |
| B2 | remote security wrapper is configured | return the secured mux | mux construction only | existing remote route tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `session0` | require console session before route behavior | denies unauthenticated requests | current route tests |
| `readOnly` | admit GET/HEAD and reject POST/other methods with 405 | no retry or side effect | existing `/orders` precedent + planned `/positions` POST test |
| `mutating` | protect existing explicit write routes | unchanged by a043 | static route tests |

## State mutations and fallbacks

- `routes` mutates only a newly allocated `http.ServeMux` during construction.
- a043 narrows `/positions` from ANY-after-auth to GET/HEAD; it does not add a route or capability.
- The `/positions` binding is an unconditional call, not an AST branch: `session0(readOnly(handlePositions))` is checked by the static route suite and POST test.

## Safety conclusion

- Safe edit boundary: wrap the existing `/positions` handler with the established method gate, leaving every mutating route untouched.
- High-risk impact: no. This is a capability reduction on a read-only UI path.
