# Function Logic Map: `allSixStrategyFirstLegResults`

- Source: `internal/app/engine/strategy_first_leg_admission_test.go` (193-211)
- Function: `allSixStrategyFirstLegResults` in package `engine`
- Signature: `allSixStrategyFirstLegResults(params=1, results=1)`
- File SHA-256: `367a5e296f03b904d72489ab4dd505f8d0ce93cc74b5a80cd29971ac3434f60f`
- Pinned revision: `base` — the AST and the SHA-256 above are `a8c3d067470fe9cd00523a7629ee93ee05de8e5c`'s file, because the checker requires this record at the frozen comparison base (the function moved or was renamed).
- AST evidence: `ast.json` — AST branches 3.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The frozen-base record of the fixture that builds one accepted first-leg result per descriptor. Renamed here to `allPairedStrategyFirstLegResults` because it never depended on the count being six — it iterates whatever the frozen matrix holds, and only its name claimed otherwise.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 210:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 197:2 | no coverage block maps to this position |
| B2 | if | 200:3 | no coverage block maps to this position |
| B3 | if | 205:3 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 194:2 |
| `strategyflow.Descriptors` | 195:17 |
| `make` | 196:9 |
| `len` | 196:47 |
| `string` | 198:13 |
| `strategyflow.AcceptedResultForJournalTest` | 203:18 |
| `t.Fatal` | 206:4 |
| `append` | 208:9 |

## State mutations and fallbacks

- AST assignments: 7. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
