# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| route registrations | every non-health route has session; every write has mutating/CSRF | operator-console + static route table | static tests fail closed |
| policy routes | GET read + POST preview/apply with opaque token | a044 spec | 403/405/412 without write |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | untrusted remote adds login/logout | mux registrations | continue | remote route tests |
| B2 | remote runtime wraps mux | security middleware | return wrapped or local mux | remote/local tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `session0` | possession authentication | reject before handler | static/runtime tests |
| `mutating` | POST + origin + bounded parse + CSRF | reject before commander | CSRF tests |
| `register*` | modular read routes | route extractor validates | static tests |

## State mutations and fallbacks

- Registers handlers only; mutations occur exclusively inside injected capability handlers.

## Safety conclusion

- Safe edit boundary: policy GET and token-only POST routes behind existing gates.
- High-risk impact: yes — policy lifecycle write; stale/forged/method negative tests required.
