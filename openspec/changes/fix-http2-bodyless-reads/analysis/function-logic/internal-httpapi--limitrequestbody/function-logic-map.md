# Function Logic Map: `LimitRequestBody`

- Source: `internal/httpapi/server.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `next` | handler or nil | caller | nil becomes `http.NotFoundHandler` |
| `r.ContentLength` | `0` empty, `-1` unknown, `>0` declared | Go inbound server request | unknown/declared body remains bounded and route decides rejection |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `next == nil` | replace local handler | continue with 404 handler | existing nil-handler test path |
| B2 | request body is unknown or declared | replace `r.Body` with `MaxBytesReader` | continue | body limit boundary tests |
| B3 | request is known empty, including HTTP/2 EOF body object | no body wrapper | continue | TLS HTTP/2 bodyless GET/HEAD test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `http.MaxBytesReader` | enforce 256 KiB process ceiling | surfaces `MaxBytesError`; no retry | AST + boundary tests |
| `next.ServeHTTP` | execute exact router | single synchronous dispatch | AST + router tests |

## State mutations and fallbacks

- Only `r.Body` may be wrapped; no domain, journal, broker, capability or lifecycle state changes.
- Nil handler fallback remains fail-closed 404.

## Safety conclusion

- Safe edit boundary: body framing classification before router dispatch only.
- High-risk impact: yes; an incorrect classification can reject all HTTP/2 reads or weaken body limits.
