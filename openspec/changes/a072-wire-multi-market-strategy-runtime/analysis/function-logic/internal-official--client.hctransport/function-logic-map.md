# Function Logic Map: `Client.hcTransport`

- Source: `internal/official/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| HTTP client transport | non-nil concrete `*http.Transport` | constructor-owned client | returns nil,false for nil or nonstandard transport without interface comparison panic |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | HTTP client is nil | none | nil,false | authority refusal tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| type assertion to `*http.Transport` | restrict authority to comparable constructor transport | pure, no I/O | AST |

## State mutations and fallbacks

- Avoids comparing arbitrary `http.RoundTripper` interface values, whose dynamic type may be uncomparable.

## Safety conclusion

- Safe edit boundary: nonstandard or missing transport must return false, never panic or mint authority.
- High-risk impact: yes, supports fail-closed origin verification.
