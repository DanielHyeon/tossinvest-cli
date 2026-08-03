# Function Logic Map: `Client.BaseURL`

- Source: `internal/official/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| client configuration | sealed or construction-time base | `official.New` and `WithBaseURL` | returns a lock-consistent value |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy path | acquires read lock only | returns base URL | base and replay tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `configMu.RLock` | synchronize with constructor/replayed options | no timeout or I/O | AST |

## State mutations and fallbacks

- Read-only; post-construction options cannot alter the returned value.

## Safety conclusion

- Safe edit boundary: keep all configuration reads synchronized with option writes.
- High-risk impact: supports proof that endpoint origin is immutable.
