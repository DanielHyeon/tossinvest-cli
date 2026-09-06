# Function Logic Map: `snapshotDigest`

- Source: `internal/strategyevidence/store.go`
- AST evidence: `ast.json` (re-extracted at the re-baselined comparison base; the function body is
  byte-identical to a064's implementation commit `23794f86`, only the enclosing file moved)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `query` | normalized KR/US market, symbol, issuer/mapping and dual cutoffs | `SealSnapshot` | invalid query is rejected before this function |
| `items` | immutable envelopes selected for the exact query | evidence.db as-of selection | every full immutable Header and payload digest must affect the result |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | each normalized query field is added to the digest preimage | hash state only | none | `TestSnapshotDigestBindsEveryQueryField` |
| B2 | items are cloned and ordered by immutable EvidenceID | cloned slice only | none | `TestSnapshotDigestOrdersItemsAscendingWithoutMutatingCaller` |
| B3 | every field of each immutable Header plus payload digest is length-prefixed into the hash | hash state only | final lowercase SHA-256 | `TestSnapshotDigestBindsEveryHeaderFieldAndPayloadDigest` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `cloneEnvelopes`, `sort.Slice` | deterministic order without caller mutation | in-memory, bounded by snapshot item count | AST + `TestSnapshotDigestOrdersItemsAscendingWithoutMutatingCaller` |
| `writeSnapshotDigestField` | length-prefix every value so two adjacent fields cannot be confused for one | pure; no fallback or omitted provenance | AST + `TestSnapshotDigestMatchesFrozenGoldenVector` |
| `sha256`, `hex.EncodeToString` | canonical snapshot identity | deterministic, no I/O | AST + `TestSnapshotDigestMatchesFrozenGoldenVector` |

## State mutations and fallbacks

- Only local hash state and a cloned slice are mutated.
- No database, source, journal, broker, Guardian or toggle call occurs.
- Header provenance cannot fall back to EvidenceID/payload-only identity.

## Safety conclusion

- High-risk integrity function: the snapshot ID gates historical evidence replay.
- The fix must bind the complete normalized Header and retain deterministic ordering and legacy test semantics.
- Corrected 2026-09-07. The rows above previously cited the deterministic snapshot tests and
  `TestDormantSnapshotReadRejectsTamperedHeaderScopeAndCutoffs`, none of which can reach these branches:
  the first assert digest *stability* only, and the second is refused earlier by `snapshotItemMatchesQuery`.
  Measured: reducing the item preimage to `EvidenceID + PayloadDigest` left the package green.
- This function is one half of a pair. `snapshotItemMatchesQuery` judges the same scope rule in
  `Snapshot.Valid` and `Replay`, and it is bound separately by
  `TestSnapshotItemScopeIsCheckedIndependentlyOfTheDigest`, which supplies a *correct* digest so only the
  scope check can refuse. Binding one half and not the other is what left both deletable for a month.
