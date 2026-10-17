# Function Logic Map: `productionRouteFixture.write`

- Source: `internal/strategyrouter/production_test.go` (271-297)
- Function: `productionRouteFixture.write` in package `strategyrouter`
- Signature: `productionRouteFixture.write(params=3, results=0)`
- File SHA-256: `6bcf8e475597ac2322f973b843dd0dc37e48f9e2ebbb306483e82bc9a9334dc6`
- Pinned revision: `base` — the AST and the SHA-256 above are `a8c3d067470fe9cd00523a7629ee93ee05de8e5c`'s file, because the checker requires this record at the frozen comparison base (the function moved or was renamed).
- AST evidence: `ast.json` — AST branches 6.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Signs and writes the fixture manifest, pinned at the base revision because it moved without its body changing.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: **none available**. `go test` does not instrument `_test.go` files, so no coverage profile can speak for this function. Each row below is classified from the arm's own source text instead, and the run that exercised the function is named.
- Measured entry: no measured profile entered this function body.

Exact AST return positions: none.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 274:2 | no coverage block maps to this position |
| B2 | if | 279:2 | no coverage block maps to this position |
| B3 | if | 283:2 | no coverage block maps to this position |
| B4 | if | 284:3 | no coverage block maps to this position |
| B5 | if | 288:2 | no coverage block maps to this position |
| B6 | if | 291:2 | no coverage block maps to this position |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 272:2 |
| `json.Marshal` | 273:20 |
| `t.Fatal` | 275:3 |
| `base64.StdEncoding.EncodeToString` | 277:76 |
| `ed25519.Sign` | 277:110 |
| `json.Marshal` | 278:15 |
| `t.Fatal` | 280:3 |
| `filepath.Join` | 282:10 |
| `ProductionRouteFileName` | 282:37 |
| `os.Stat` | 283:15 |
| `os.Chmod` | 284:13 |
| `t.Fatal` | 285:4 |
| `os.WriteFile` | 288:12 |
| `t.Fatal` | 289:3 |
| `os.Chmod` | 291:12 |
| `t.Fatal` | 292:3 |
| `productionRouteDigest` | 295:26 |

## State mutations and fallbacks

- AST assignments: 11. Defers: 0. Goroutine statements: 0.
- A test function mutates only its own fixtures; it opens no journal, issues no order and touches no shared state.

## Safety conclusion

- Test-only. It cannot change production behaviour; its value is the assertion it makes, and a green run means only that no guard arm fired.
