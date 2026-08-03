# Function Logic Map: `router.read`

- Source: `internal/httpapi/router.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| resource | exact registry member | `readResourceForPath` | default returns internal unknown-resource error |
| strategy runtime reader | optional narrow read interface | router options | nil returns explicit paired dormant data, error never fabricates current facts |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | engine/positions/orders/candidates/performance/settings/optimization | existing reader call | data/error | existing router tests |
| B2 | positions/orders/candidates nil items | normalize collection only | empty array | existing schema tests |
| B3 | strategy-runtime configured | optional reader call + shared validation | paired data or error | REST parity test |
| B4 | strategy-runtime unconfigured | none | paired dormant OFF projection | dormant health test |
| B5 | unknown internal resource | none | error | route guard test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| existing `Reader` methods | existing seven resources | unchanged error propagation | CodeGraph + AST |
| strategy runtime reader | new read-only resource | one call; no retry or mutation | CodeGraph + AST |
| shared validator | strict typed projection | invalid payload becomes unavailable error | AST |

## State mutations and fallbacks

- Reads and normalizes response data only. Existing approved mutation route map is not expanded.

## Safety conclusion

- Safe edit boundary: add one exact read case and no mutation route.
- High-risk impact: no; transport-only GET projection.
