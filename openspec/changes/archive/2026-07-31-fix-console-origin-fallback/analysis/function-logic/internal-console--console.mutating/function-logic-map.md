# Function Logic Map: `Console.mutating`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request method | POST only | browser request | 405 before parsing |
| remote runtime | nil for loopback mode; configured for remote mode | `New` options | remote mode evaluates origin |
| form body | parseable URL-encoded form | browser request | 400 before handler |
| CSRF token | constant-time equal to process-local token | rendered console form | 403 before handler |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | method is not POST | response only | 405 and return | existing mutating method tests |
| B2 | remote mode and origin predicate rejects | response only | 403 and return | origin evidence matrix |
| B3 | form parsing fails | request parse state only | 400 and return | existing malformed form tests |
| B4 | CSRF mismatch | response only | 403 and return | headerless canonical request with wrong CSRF |
| fallthrough | all gates pass | invokes supplied handler | handler result | headerless canonical request with valid CSRF |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `remoteRuntime.sameOriginForMutation` | mutation-only explicit-header/TLS+Host gate | false refuses before form parsing | CodeGraph + post-GREEN AST |
| `http.Request.ParseForm` | decode CSRF and handler fields | parse error returns 400 | Go AST |
| `tokenEqual` | constant-time CSRF comparison | false returns 403 | Go AST |
| wrapped `next` | perform the requested console mutation | called only after every gate | Go AST |

## State mutations and fallbacks

- Gate order is POST → origin → form parse → CSRF → handler and must not change.
- The only implemented edit is the origin predicate name. No handler, audit, engine,
  trading, config, or journal behavior changes.
- Remote login does not call this wrapper and therefore cannot inherit the
  mutation-only fallback.

## Safety conclusion

- Implemented edit boundary: replaced one call target without reordering gates.
- High-risk impact: yes, request-authentication boundary; no trading-path impact.
