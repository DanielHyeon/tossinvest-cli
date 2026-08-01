# Function Logic Map: `dedupeKey`

- Source: `internal/console/orders.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| broker order identity | opaque non-empty id plus account/market/time | broker payload + market clock | empty id is never deduplicated; invalid scope uses tagged raw fallback |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | broker id has zero bytes | none | no dedupe key | missing-id overlap fixture |
| B2 | full evidence key is valid | none | canonical account/market/local-day + exact id | exact/market/day/whitespace table |
| B3 | market alone is canonicalizable for fallback | none | tagged fallback keeps raw invalid timestamp | invalid timestamp table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `evidenceKey` | reuse evidence identity when fully valid | pure; false selects fallback | current AST + scoped collision table |
| `clock.ParseMarket` | canonicalize known market even when time is invalid | pure; error preserves raw market | current AST + invalid identity table |

## State mutations and fallbacks

- Pure function. Fallback is explicitly tagged and retains raw time, so it cannot collide with a valid identity or another malformed timestamp.

## Safety conclusion

- Safe edit boundary: OPEN/CLOSED presentation dedupe only.
- High-risk impact: no; it cannot mutate or submit an order.
