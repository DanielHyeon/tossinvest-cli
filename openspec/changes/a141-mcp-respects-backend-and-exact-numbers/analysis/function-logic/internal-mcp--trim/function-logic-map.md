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
| B1 | proportional keep >= n (330) | keep becomes n-1 | continue | cap regression; no individual branch-hit claim |
| B2 | keep < 0 (333) | keep becomes zero | continue | unchanged clamp; no focused branch-hit claim |
| B3 | n < len(s), so a prior marker exists (336) | inspect final marker | continue | repeated and pre-existing marker cases |
| B4 | omittedCount valid AND prior count fits after new drops (337) | recover prior count; otherwise use zero | continue | exact/repeated/malformed marker tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `keepable` | Exclude an existing marker from real-item count | no error return | call before B1 |
| `omittedCount` | Recover count safely from preserved JSON number | malformed/overflowing value becomes safe non-panicking fallback | B3/B4 |
| `fmt.Sprintf` | Render compatible marker note | no network/retry | AST return |

## State mutations and fallbacks

- Replaces a selected array with kept prefix plus a new omission marker.
- Its numeric marker logic must remain compatible with decoder exact-number representation.

## Safety conclusion

- Safe edit boundary: marker-count parsing only; preserve proportional trimming and marker shape.
- High-risk impact: yes. Incorrect cumulative omission count makes a partial financial list look complete or reports a false loss amount.
