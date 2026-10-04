# Function Logic Map: `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot`

- Source: `internal/strategyrouter/production_test.go` (106-145)
- Function: `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot` in package `strategyrouter`
- Signature: `TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot(params=1, results=0)`
- File SHA-256: `21d69cb2a4f5da5ae8d5aded5ea32e5e8fe4aa4d4b5ae0f64038a2bde6cb3d9a`
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
| B1 | range | 108:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B2 | range | 120:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B3 | if | 129:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B4 | if | 132:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B5 | range | 135:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B6 | if | 137:4 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B7 | if | 141:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `newProductionRouteFixture` | 107:13 |
| `fixture.body` | 116:11 |
| `append` | 124:17 |
| `fixture.write` | 125:3 |
| `LoadProductionRouteAuthorityBatch` | 126:17 |
| `context.Background` | 126:51 |
| `t.Fatalf` | 130:4 |
| `batch.Len` | 132:6 |
| `batch.ManifestDigest` | 132:26 |
| `t.Fatalf` | 133:4 |
| `batch.For` | 136:21 |
| `RouteSet` | 137:14 |
| `authority.Request` | 137:23 |
| `authority.Request` | 137:67 |
| `t.Fatalf` | 138:5 |
| `batch.For` | 141:15 |
| `t.Fatalf` | 142:4 |

## State mutations and fallbacks

- AST assignments: 11. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Test-only; it exercises the batch path that the engine route loader uses in production.

a112 게이트 준비(2026-10-04): a127 82080177(전략 권한 적재기가 현재 원장을 읽음)이 같은 파일을 편집해 이 번들이 낡았다 — 분기 · return · 호출 구조 동일
