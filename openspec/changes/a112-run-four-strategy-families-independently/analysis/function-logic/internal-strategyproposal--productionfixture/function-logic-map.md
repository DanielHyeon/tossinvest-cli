# Function Logic Map: `productionFixture`

- Source: `internal/strategyproposal/production_test.go` (100-193)
- Function: `productionFixture` in package `strategyproposal`
- Signature: `productionFixture(params=3, results=3)`
- File SHA-256: `28dc27e289908691099ee4c43139f1f8b0796bf186409849fd810c07209af0ee`
- Pinned revision: `base` — the AST and the SHA-256 above are `a8c3d067470fe9cd00523a7629ee93ee05de8e5c`'s file, because the checker requires this record at the frozen comparison base (the function moved or was renamed).
- AST evidence: `ast.json` — AST branches 15.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

The manifest/evidence fixture for the production proposal tests, pinned at the base revision because it moved without its body changing. It is behind `tossos_testseams`, so CI never runs it (decision 50).

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: 192:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 107:2 | no coverage block maps to this position |
| B2 | if | 118:2 | no coverage block maps to this position |
| B3 | if | 123:2 | no coverage block maps to this position |
| B4 | if | 126:2 | no coverage block maps to this position |
| B5 | if | 131:2 | no coverage block maps to this position |
| B6 | if | 134:2 | no coverage block maps to this position |
| B7 | if | 137:2 | no coverage block maps to this position |
| B8 | if | 143:2 | no coverage block maps to this position |
| B9 | if | 147:2 | no coverage block maps to this position |
| B10 | if | 153:2 | no coverage block maps to this position |
| B11 | if | 169:2 | no coverage block maps to this position |
| B12 | if | 174:2 | no coverage block maps to this position |
| B13 | if | 178:2 | no coverage block maps to this position |
| B14 | if | 181:2 | no coverage block maps to this position |
| B15 | if | 185:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 101:2 |
| `t.TempDir` | 102:9 |
| `string` | 112:62 |
| `string` | 113:124 |
| `now.Add` | 114:82 |
| `now.Add` | 114:128 |
| `now.Add` | 115:15 |
| `strategyevidence.NewEnvelope` | 117:19 |
| `?` | 117:56 |
| `t.Fatal` | 119:3 |
| `filepath.Join` | 121:18 |
| `strategyevidence.Open` | 122:16 |
| `context.Background` | 122:38 |
| `marketclock.NewFake` | 122:112 |
| `t.Fatal` | 124:3 |
| `store.Append` | 126:15 |
| `context.Background` | 126:28 |
| `t.Fatal` | 127:3 |
| `store.SealSnapshot` | 129:19 |
| `context.Background` | 129:38 |
| `t.Fatal` | 132:3 |
| `store.Close` | 134:12 |
| `t.Fatal` | 135:3 |
| `os.Chmod` | 137:12 |
| `t.Fatal` | 138:3 |
| `strategy.ApprovedSnapshotForTest` | 141:14 |
| `string` | 141:47 |
| `strategyrouter.NewOwnerKey` | 142:14 |
| `t.Fatal` | 144:3 |
| `strategyrouter.ProductionRouteAuthorityForTest` | 146:16 |
| `string` | 146:162 |
| `t.Fatal` | 148:3 |
| `route.Request` | 150:57 |
| `ed25519.GenerateKey` | 152:26 |
| `t.Fatal` | 154:3 |
| `string` | 159:40 |
| `Format` | 159:93 |
| `now.Add` | 159:93 |
| `Format` | 160:15 |
| `now.Add` | 160:15 |

(29 further call sites omitted; `ast.json` carries all 69.)

## State mutations and fallbacks

- AST assignments: 30. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
