# Function Logic Map: `TestPairedProductionRouteAuthorityLoadsExactThreeLanesIndependently`

- Source: `internal/strategyrouter/production_test.go` (21-41)
- Function: `TestPairedProductionRouteAuthorityLoadsExactThreeLanesIndependently` in package `strategyrouter`
- Signature: `TestPairedProductionRouteAuthorityLoadsExactThreeLanesIndependently(params=1, results=0)`
- File SHA-256: `6bcf8e475597ac2322f973b843dd0dc37e48f9e2ebbb306483e82bc9a9334dc6`
- Pinned revision: `base` — the AST and the SHA-256 above are `a8c3d067470fe9cd00523a7629ee93ee05de8e5c`'s file, because the checker requires this record at the frozen comparison base (the function moved or was renamed).
- AST evidence: `ast.json` — AST branches 5.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The frozen-base record of the route-authority load test, renamed in this worktree to `TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently` and moved from three candidates to four.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: none.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 23:2 | no coverage block maps to this position |
| B2 | if | 25:3 | no coverage block maps to this position |
| B3 | if | 28:3 | no coverage block maps to this position |
| B4 | if | 32:3 | no coverage block maps to this position |
| B5 | if | 37:3 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `newProductionRouteFixture` | 22:13 |
| `LoadProductionRouteAuthority` | 24:21 |
| `context.Background` | 24:50 |
| `t.Fatalf` | 26:4 |
| `authority.ManifestDigest` | 28:6 |
| `authority.OwnerDigest` | 28:42 |
| `t.Fatalf` | 29:4 |
| `authority.Request` | 31:14 |
| `len` | 32:6 |
| `len` | 32:72 |
| `t.Fatalf` | 33:4 |
| `Route` | 35:13 |
| `t.Fatalf` | 38:4 |

## State mutations and fallbacks

- AST assignments: 5. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
