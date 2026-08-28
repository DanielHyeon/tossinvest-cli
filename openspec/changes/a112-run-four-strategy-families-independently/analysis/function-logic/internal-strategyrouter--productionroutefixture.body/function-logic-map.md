# Function Logic Map: `productionRouteFixture.body`

- Source: `internal/strategyrouter/production_test.go` (247-272)
- Function: `productionRouteFixture.body` in package `strategyrouter`
- Signature: `productionRouteFixture.body(params=1, results=1)`
- File SHA-256: `4a6fe328016fbef89ac4b186f65b5561ef7ef89b9f379837a20f12911f2eca70`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 2.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Builds the signed manifest body for both markets. L3 added the breakout candidate row to each market's lane list.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 266:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 251:2 | no coverage block maps to this position |
| B2 | else | 258:9 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `string` | 268:56 |
| `Format` | 268:93 |
| `fixture.now.Add` | 268:93 |
| `string` | 269:48 |
| `string` | 269:101 |
| `string` | 270:65 |
| `Format` | 270:118 |
| `fixture.now.Add` | 270:118 |
| `Format` | 271:15 |
| `fixture.now.Add` | 271:15 |

## State mutations and fallbacks

- AST assignments: 5. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
