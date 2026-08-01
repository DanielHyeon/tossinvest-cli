# Function Logic Map: `Console.mutating`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request method/origin | POST; remote mode also exact same origin | route wrapper and remote config | 405/403 before handler |
| content type | parsed media type exactly `application/x-www-form-urlencoded` | server-rendered form contract | 400 before parsing any polluted multipart values |
| body | optional positive route-specific byte cap | route registration | 413 on max-byte parse error |
| CSRF | exact constant-time match to console token | authenticated console process | 403 before handler |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | method is not POST | response only | 405 with Allow | state-changing route method tests |
| B2 | remote mutation origin mismatch | response only | 403 | remote-origin tests |
| B3 | content type malformed or not URL-encoded | response only | 400 before `ParseForm`/handler | multipart pollution test |
| B4 | positive body limit supplied | wraps request body | continue with bounded reader | bounded body tests |
| B5 | form parse fails | response only | 413 or 400 | oversized/malformed form tests |
| B6 | parse error is max-bytes error | response only | 413 | request body bound test |
| B7 | CSRF mismatch | response only | 403 | CSRF tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `mime.ParseMediaType` | rejects alternate parsers before values are read | one parse/no fallback | multipart regression test |
| `http.MaxBytesReader`, `r.ParseForm` | bounds and parses only URL-encoded body | route cap; parse errors refuse | body-limit tests |
| `tokenEqual` | compares CSRF token | constant-time comparison | CSRF tests |
| wrapped handler | receives only accepted request | exactly once after every gate | zero-call assertions on refusal |

## State mutations and fallbacks

- Refusal branches write only an HTTP response. The wrapped mutating capability is unreachable until method, origin, media type, size, parse, and CSRF checks all pass.
- No content-type fallback to multipart or query-string pollution is allowed.

## Safety conclusion

- Safe edit boundary: shared console mutation authentication/parser gate.
- High-risk impact: yes; every console state mutation depends on this fail-closed wrapper.
