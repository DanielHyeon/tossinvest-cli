# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Console.opts` | constructor-validated, optional capabilities remain nil when unwired | `console.New` and `Options` | nil optional seams render fail-closed/read-only state |
| HTTP request | route-specific method/authentication contract | `session0`, `readOnly`, `mutating` wrappers | authentication/method refusal; no account side effect |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | remote access exists but is not trusted-network mode | registers login/logout before session-protected application routes | same mux with explicit authentication entry points | remote security tests |
| B2 | remote access exists | wraps the complete mux with remote transport/origin security | secured handler | remote access tests |
| Success | local loopback mode | returns session-protected mux directly | handler | console route/static tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `session0` | authentication/session boundary for every application route | fail closed before handler | CodeGraph + AST |
| `readOnly` | states that a route owns no mutating capability | non-GET/HEAD rejected by handler/static contract | CodeGraph + AST |
| `mutating` | CSRF/origin/body limit boundary for existing write routes | bounded body and refusal before action | CodeGraph + AST |
| `register*` | package-local read-only route groups | each registrar is covered by static route enumeration | CodeGraph + AST |

## State mutations and fallbacks

- Route registration changes only the mux. Business state is neither read nor written here.
- a047 adds one GET/HEAD-only status card behind `session0`; its handler rejects every other method, receives display data only, and has no form/action.

## Safety conclusion

- Safe edit boundary: add the exact `/strategy-runtime` status route behind `session0`; keep method refusal in the handler and do not expose a writer/broker/journal capability.
- High-risk impact: no direct account mutation, but the page describes high-risk entry blockers and therefore must fail closed.
