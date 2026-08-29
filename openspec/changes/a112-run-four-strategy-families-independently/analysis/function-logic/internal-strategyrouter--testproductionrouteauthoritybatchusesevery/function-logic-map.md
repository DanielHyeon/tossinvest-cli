# Function Logic Map: `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`

- Source: `internal/strategyrouter/production_test.go` (101-140)
- Function: `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot` in package `strategyrouter`
- Signature: `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot(params=1, results=0)`
- File SHA-256: `fccd226dcf67215fe7792bb850fbf5ffbddd93c399f72028be3dc6a55400bd38`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 7.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Test function. Proves one batch read reconstructs every signed scope, now asserted through `RouteSet`.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`
- tagged package suite: `go test -count=1 -tags tossos_testseams -covermode=count ./internal/strategyrouter/`
- Measured entry: the function body was not instrumented — `go test` does not add coverage counters to `_test.go` files; it ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/`.

Exact AST return positions: none.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 103:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B2 | range | 115:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B3 | if | 124:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B4 | if | 127:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B5 | range | 130:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B6 | if | 132:4 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B7 | if | 136:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `newProductionRouteFixture` | 102:13 |
| `fixture.body` | 111:11 |
| `append` | 119:17 |
| `fixture.write` | 120:3 |
| `LoadProductionRouteAuthorityBatch` | 121:17 |
| `context.Background` | 121:51 |
| `t.Fatalf` | 125:4 |
| `batch.Len` | 127:6 |
| `batch.ManifestDigest` | 127:26 |
| `t.Fatalf` | 128:4 |
| `batch.For` | 131:21 |
| `RouteSet` | 132:14 |
| `authority.Request` | 132:23 |
| `authority.Request` | 132:67 |
| `t.Fatalf` | 133:5 |
| `batch.For` | 136:15 |
| `t.Fatalf` | 137:4 |

## State mutations and fallbacks

- AST assignments: 11. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Test-only; it exercises the batch path that the engine route loader uses in production.
