# Function Logic Map: `Console.mutating`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request method | POST only | HTTP request | 405 before handler |
| remote origin | exact configured HTTPS origin | `remoteRuntime` | 403 before parse |
| form/CSRF | parseable body + process token | request + console | 400/403 before handler |
| optional body limit | absent or positive bytes | route registration | 413 before parse |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | method is not POST | response only | 405 | existing mutation gate tests |
| B2 | remote origin mismatch | response only | 403 | origin tests |
| B3 | form parse fails | response only | 400 | malformed form test |
| B4 | CSRF mismatch | response only | 403 | CSRF tests |
| new limit | route body exceeds declared limit | response only | 413 | Open API oversize test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sameOriginForMutation` | remote origin proof | fail closed | CodeGraph + remote tests |
| `ParseForm` | decode after limit | one attempt | AST |
| `tokenEqual` | constant-time CSRF check | false refuses | current HEAD |
| `next` | execute authorized mutation | exactly once | gate tests |

## State mutations and fallbacks

- The middleware writes only refusal responses.
- Optional route limit will install `MaxBytesReader` before `ParseForm`; existing
  routes pass no limit and preserve behavior.

## Safety conclusion

- Safe edit boundary: variadic optional limit with zero-value compatibility.
- High-risk impact: yes, shared mutation/auth middleware; full console and
  remote test/race suites are mandatory.
