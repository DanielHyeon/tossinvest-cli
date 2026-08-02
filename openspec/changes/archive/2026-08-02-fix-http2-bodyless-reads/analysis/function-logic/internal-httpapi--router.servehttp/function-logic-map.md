# Function Logic Map: `router.ServeHTTP`

- Source: `internal/httpapi/router.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request path/method | exact fixed `/api/v1` routes and allowed methods | router allowlists | stable 404/405 JSON |
| `request.ContentLength` | `0` empty, `-1` unknown, `>0` declared | Go inbound server request | unknown/declared read body is stable 400 |
| `request.URL.RawQuery` | empty only | fixed no-input API contract | stable 400 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | exact configured mutation path | response headers; dispatch or 405 | early return | mutation allowlist/method tests |
| B2 | stream path | delegate to `serveStream` | early return | stream contract tests |
| B3 | unknown read resource | write stable error | 404 | unknown resource tests |
| B4 | method is not GET/HEAD | set Allow; write stable error | 405 | method tests |
| B5 | unknown/declared request body | write stable error | 400 | HTTP/2 known/unknown body tests |
| B6 | non-empty query | write stable error | 400 | query rejection tests |
| B7 | reader returns error | write stable error | 503 | unavailable resource tests |
| B8 | valid bodyless fixed read | write JSON envelope | 200 | HTTP/1.1 and TLS HTTP/2 read tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `setResponseHeaders` | pin no-store/security headers | no error/retry | AST + header tests |
| `r.read` | invoke one fixed read seam | error becomes stable 503; no retry | AST + reader tests |
| `r.writeError` / `r.writeEnvelope` | stable contract output | single response write | AST + golden contract tests |

## State mutations and fallbacks

- No router-owned durable state mutation; request output only.
- Unknown resource/method/body/query and read errors all return before the reader executes.

## Safety conclusion

- Safe edit boundary: B5 request framing predicate only.
- High-risk impact: yes; this is the no-input read boundary, but it has no mutation/order authority.
