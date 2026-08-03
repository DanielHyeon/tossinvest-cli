# Function Logic Map: `New`

- Source: `internal/official/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| credentials/cache/options | Options may configure ordinary client; any base/HTTP override forfeits authority | `official.New` | returns initialized client; auth failures occur on request |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range over options | applies each option before token manager construction | client returned | existing client tests + authority-origin tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newTokenManager` | bind credentials, endpoint and transport | no request until later call | CodeGraph + current source |

## State mutations and fallbacks

- Initializes default endpoint/transport as authority-eligible; override options can only clear eligibility.

## Safety conclusion

- Safe edit boundary: preserve option ordering and shared transport/token manager.
- High-risk impact: yes — establishes official-origin capability.
