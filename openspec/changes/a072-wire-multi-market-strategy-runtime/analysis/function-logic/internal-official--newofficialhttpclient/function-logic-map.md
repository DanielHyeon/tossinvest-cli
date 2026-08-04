# Function Logic Map: `newOfficialHTTPClient`

- Source: `internal/official/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fixed transport policy | official default origin only | package constructor | no caller customization accepted |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy path | allocates private transport and client | returns non-nil client | authority-origin test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `net.Dialer.DialContext` | standard network dial with bounded timeout/keepalive | request context governs cancellation | AST |

## State mutations and fallbacks

- Creates a private transport rather than trusting mutable `http.DefaultTransport` state.

## Safety conclusion

- Safe edit boundary: preserve ordinary HTTP timeouts while retaining exact transport identity.
- High-risk impact: yes, this transport anchors official-origin authority.
