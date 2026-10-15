# Function Logic Map: `trim`

- Source: `internal/mcp/server.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| slice | decoded array, optionally ending in omission marker | `shrinkToCap` | returns a new shortened array |
| existing marker | numeric omission count or malformed value | last element | malformed/overflowing values must not panic or corrupt count |
| size/excess | positive encoded sizes from caller | `shrinkToCap` | keeps at least one element removal when possible |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | Existing omission placeholder is present | attempts cumulative count recovery | no error return available | exact-number decoder changes marker dynamic type |
| B2 | Marker is json.Number, compatible integral float64, or int | validates before adding prior count | malformed/overflow values fail closed | exact, repeated, malformed regressions |
| B3 | proportional keep would retain all elements | force one removal | continue | Existing cap test |
| B4 | calculated keep is negative | clamp to zero | continue | Oversized-item regression |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `keepable` | Exclude an existing marker from real-item count | no error return | AST B1 |
| numeric marker parser | Recover count safely from preserved JSON number | malformed/overflowing value becomes safe non-panicking fallback | AST B2 edit target |
| `fmt.Sprintf` | Render compatible marker note | no network/retry | AST return |

## State mutations and fallbacks

- Replaces a selected array with kept prefix plus a new omission marker.
- Its numeric marker logic must remain compatible with decoder exact-number representation.

## Safety conclusion

- Safe edit boundary: marker-count parsing only; preserve proportional trimming and marker shape.
- High-risk impact: yes. Incorrect cumulative omission count makes a partial financial list look complete or reports a false loss amount.
