# Function Logic Map: `router.serveStream`

- Source: `internal/httpapi/router.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| method/path | `/api/v1/stream`, GET/HEAD | fixed router | stable 405 for other methods |
| `request.ContentLength` | `0` empty, `-1` unknown, `>0` declared | Go inbound server request | stable 400 for unknown/declared body |
| stream handler | handler or nil | daemon assembly | stable 503 if unavailable |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | method is not GET/HEAD | set Allow; write error | 405 | stream method tests |
| B2 | unknown/declared body | write error | 400 | HTTP/2 body tests |
| B3 | non-empty query | write error | 400 | stream query tests |
| B4 | stream handler is nil | write error | 503 | unavailable stream test |
| B5 | valid bodyless request | delegate to stream | handler result | HTTP/1.1 and TLS HTTP/2 stream tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.writeError` | stable failure envelope | one response; no retry | AST + contract tests |
| `r.stream.ServeHTTP` | run bounded SSE handler | stream owns heartbeat/queue/close | AST + SSE tests |

## State mutations and fallbacks

- No durable state mutation; only response headers/body and stream subscription lifecycle.
- All invalid framing returns before opening a stream client.

## Safety conclusion

- Safe edit boundary: B2 framing predicate only.
- High-risk impact: yes; wrong acceptance can consume SSE capacity, wrong rejection blocks mobile clients.
