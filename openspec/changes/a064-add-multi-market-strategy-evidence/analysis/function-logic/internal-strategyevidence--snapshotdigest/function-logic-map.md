# Function Logic Map: `snapshotDigest`

- Source: `internal/strategyevidence/store.go`
- AST evidence: `ast.json` (pre-edit source hash captured before the provenance-binding fix)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `query` | normalized KR/US market, symbol, issuer/mapping and dual cutoffs | `SealSnapshot` | invalid query is rejected before this function |
| `items` | immutable envelopes selected for the exact query | evidence.db as-of selection | every full immutable Header and payload digest must affect the result |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | each normalized query field is added to the digest preimage | hash state only | none | deterministic snapshot tests |
| B2 | items are cloned and ordered by immutable EvidenceID | cloned slice only | none | deterministic snapshot tests |
| B3 | every field of each immutable Header plus payload digest is length-prefixed into the hash | hash state only | final lowercase SHA-256 | header-tamper replay rejection tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `cloneEnvelopes`, `sort.Slice` | deterministic order without caller mutation | in-memory, bounded by snapshot item count | AST + store tests |
| header canonical preimage helper | encode every immutable Header field unambiguously | no fallback or omitted provenance | tamper RED/GREEN tests |
| `sha256`, `hex.EncodeToString` | canonical snapshot identity | deterministic, no I/O | snapshot replay tests |

## State mutations and fallbacks

- Only local hash state and a cloned slice are mutated.
- No database, source, journal, broker, Guardian or toggle call occurs.
- Header provenance cannot fall back to EvidenceID/payload-only identity.

## Safety conclusion

- High-risk integrity function: the snapshot ID gates historical evidence replay.
- The fix must bind the complete normalized Header and retain deterministic ordering and legacy test semantics.
