# Function Logic Map: `inputFor`

- Source: `internal/strategyflow/flow_test.go` (264-285)
- Function: `inputFor` in package `strategyflow`
- Signature: `inputFor(params=1, results=1)`
- File SHA-256: `59776edda49cc64112b0a744fb25fdfefb39d484df7cd87ea8cf6171f25b656b`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 10.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The shared fixture that produces a `LaneInput` for a given descriptor. L3 added the two breakout arms, which is why its branch count is high: one arm per lane plus the error guards.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 267:3, 269:3, 271:3, 273:3, 275:3, 277:3, 279:3, 281:3, 283:3.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | switch | 265:2 | no coverage block maps to this position |
| B2 | case | 266:2 | no coverage block maps to this position |
| B3 | case | 268:2 | no coverage block maps to this position |
| B4 | case | 270:2 | no coverage block maps to this position |
| B5 | case | 272:2 | no coverage block maps to this position |
| B6 | case | 274:2 | no coverage block maps to this position |
| B7 | case | 276:2 | no coverage block maps to this position |
| B8 | case | 278:2 | no coverage block maps to this position |
| B9 | case | 280:2 | no coverage block maps to this position |
| B10 | case | 282:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `ContinuationKR` | 267:10 |
| `ContinuationUS` | 269:10 |
| `ReversalKR` | 271:10 |
| `ReversalUS` | 273:10 |
| `WeeklyKR` | 275:10 |
| `WeeklyUS` | 277:10 |
| `BreakoutKR` | 279:10 |
| `BreakoutUS` | 281:10 |

## State mutations and fallbacks

- AST assignments: 0. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
