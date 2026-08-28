# Function Logic Map: `TestAcceptedProjectionCoversKRUSContinuationReversalAndWeeklyTogether`

- Source: `internal/strategyflow/canonical_projection_test.go` (23-54)
- Function: `TestAcceptedProjectionCoversKRUSContinuationReversalAndWeeklyTogether` in package `strategyflow`
- Signature: `TestAcceptedProjectionCoversKRUSContinuationReversalAndWeeklyTogether(params=1, results=0)`
- File SHA-256: `ac31a33808c126bebe48a706d67aa69783547b85d2aa7d24ce184371a84d40b4`
- Pinned revision: `base` — the AST and the SHA-256 above are `a8c3d067470fe9cd00523a7629ee93ee05de8e5c`'s file, because the checker requires this record at the frozen comparison base (the function moved or was renamed).
- AST evidence: `ast.json` — AST branches 7.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The frozen-base record of the projection coverage test. In this worktree the function was renamed to `TestAcceptedProjectionCoversAllFourFamiliesInBothMarketsTogether` and widened to the 8-set, so the checker requires this record at the base revision.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: none.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 25:2 | no coverage block maps to this position |
| B2 | range | 29:2 | no coverage block maps to this position |
| B3 | if | 32:3 | no coverage block maps to this position |
| B4 | if | 36:3 | no coverage block maps to this position |
| B5 | if | 40:3 | no coverage block maps to this position |
| B6 | if | 44:3 | no coverage block maps to this position |
| B7 | if | 51:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `Descriptors` | 24:17 |
| `len` | 25:5 |
| `t.Fatalf` | 26:3 |
| `len` | 26:47 |
| `acceptedProjectionFixture` | 30:13 |
| `ProjectAccepted` | 31:22 |
| `t.Fatalf` | 33:4 |
| `ProjectAccepted` | 35:18 |
| `replay.Payload` | 36:20 |
| `projection.Payload` | 36:40 |
| `replay.PayloadDigest` | 36:64 |
| `projection.PayloadDigest` | 36:90 |
| `t.Fatalf` | 37:4 |
| `VerifyAcceptedProjection` | 39:20 |
| `projection.Payload` | 39:45 |
| `t.Fatalf` | 41:4 |
| `verified.Lineage` | 43:21 |
| `verified.ExecutionTerms` | 43:41 |
| `lineage.Valid` | 44:7 |
| `terms.Valid` | 44:27 |
| `terms.LineageIdentity` | 46:4 |
| `terms.Quantity` | 46:51 |
| `t.Fatalf` | 47:4 |
| `t.Fatalf` | 52:3 |

## State mutations and fallbacks

- AST assignments: 8. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
