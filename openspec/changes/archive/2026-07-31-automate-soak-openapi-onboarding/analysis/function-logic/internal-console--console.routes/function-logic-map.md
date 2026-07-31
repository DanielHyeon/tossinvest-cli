# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.remote` | nil, trusted network, or token mode | validated `RemoteAccess` | remote security wrapper refuses invalid peer/Host |
| route handlers | all routes pass `session0` except health/login | static route tests | test failure blocks gate |
| mutations | POST + origin + parsed CSRF | `mutating` | 403/405/400 before handler |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | remote token mode | registers login/logout | none | remote login route tests |
| B2 | any remote mode | wraps mux in remote security | returns wrapped handler | peer/Host/origin tests |
| default | local mode | none | returns mux | local route tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `session0` | application session/trusted-network gate | refuses unauthorized | CodeGraph + static tests |
| `mutating` | POST/origin/form/CSRF gate | refuses before side effect | CodeGraph + static tests |
| `remote.security` | peer/Host/security headers | fail closed | remote tests |

## State mutations and fallbacks

- This method only assembles handlers.
- New GET setup remains session-gated and refuses plaintext transport.
- New POST setup must be `session0(httpsOnly(mutating(handler, bodyLimit)))` so
  plaintext is rejected before form parsing while preserving the existing
  method, origin, bounded-body, and CSRF sequence.

## Safety conclusion

- Safe edit boundary: add exactly two Open API routes with the credential-only
  TLS wrapper and preserve every existing route wrapper.
- High-risk impact: yes, authentication/secret ingress. The unrelated panic
  finding at line 889 is outside this function and its existing recovery tests.
