# Function Logic Map: `requestHasBody`

- Source: `internal/httpapi/server.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `request` | inbound server request or nil | `net/http` server | nil is treated as no body |
| `request.ContentLength` | `0` empty, `-1` unknown, `>0` declared | Go inbound request contract | unknown/declared returns true for fail-closed handling |
| `Content-Length` header | absent/zero empty; nonzero or malformed is suspicious | preserved inbound header | nonzero/malformed returns true even if field was normalized to zero |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | request is nil | none | false | direct helper test through middleware safety |
| B2 | content length is negative or positive | none | true | HTTP/2 unknown/known-length body tests |
| B3 | iterate preserved `Content-Length` header values | none | continue classification | direct helper absent/single-header tests |
| B4 | split comma-joined header values | none | continue classification | direct helper zero/nonzero/malformed tests |
| B5 | iterate every rune in the trimmed header part | none | continue digit validation | direct helper digit-only and signed-zero tests |
| B6 | rune is outside ASCII `0`-`9` | none | true | direct helper malformed and signed-zero tests |
| B7 | unsigned parse fails or parsed length is nonzero | none | true; otherwise continue and ultimately false | direct helper zero/positive/overflow tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | pure framing classification | no I/O, timeout, retry or error | AST |

## State mutations and fallbacks

- Pure function; no request read and no state mutation.
- Unknown length and suspicious preserved headers remain fail-closed and cannot bypass the route body rejection.

## Safety conclusion

- Safe edit boundary: inbound body framing classification only.
- High-risk impact: yes; shared by all read and SSE body gates.
