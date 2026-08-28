# Function Logic Map: `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`

- Source: `internal/strategyrouter/production_test.go` (68-98)
- Function: `TestProductionRouteAuthoritySelectsEverySignedSymbolScope` in package `strategyrouter`
- Signature: `TestProductionRouteAuthoritySelectsEverySignedSymbolScope(params=1, results=0)`
- File SHA-256: `4a6fe328016fbef89ac4b186f65b5561ef7ef89b9f379837a20f12911f2eca70`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 4.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Pins that every signed symbol scope loads and routes without refusal; its candidate count moved from three to four.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: none.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 70:2 | no coverage block maps to this position |
| B2 | range | 80:3 | no coverage block maps to this position |
| B3 | if | 90:3 | no coverage block maps to this position |
| B4 | if | 94:3 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `newProductionRouteFixture` | 69:13 |
| `fixture.body` | 77:11 |
| `append` | 84:17 |
| `fixture.write` | 85:3 |
| `LoadProductionRouteAuthority` | 89:21 |
| `context.Background` | 89:50 |
| `t.Fatalf` | 91:4 |
| `authority.Request` | 93:14 |
| `len` | 94:43 |
| `Route` | 94:75 |
| `t.Fatalf` | 95:4 |

## State mutations and fallbacks

- AST assignments: 12. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
