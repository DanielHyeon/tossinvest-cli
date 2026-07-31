# Function Logic Map: `stableObservationID`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account/market/symbol/position generation/price/time | canonical decimal and normalized identifiers | quote + journal position | invalid decimal refuses snapshot |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid price | none | error | decimal tests |
| B2 | nonzero FetchedAt | hash authoritative timestamp; ignore cycle fallback | opaque ID | fetched-at test |
| B3 | zero FetchedAt | hash one cycle instant/sequence | opaque ID | fallback test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `CanonicalDecimal` | make numeric aliases identical | parse error propagates | AST |
| SHA-256 writer | length-prefix every field | no raw identifier output | AST |

## State mutations and fallbacks

- Only the digest is exposed; source account/position values remain preimage data.

## Safety conclusion

- Safe edit boundary: pure deterministic identity function.
- High-risk impact: yes — decision dedup lineage.
