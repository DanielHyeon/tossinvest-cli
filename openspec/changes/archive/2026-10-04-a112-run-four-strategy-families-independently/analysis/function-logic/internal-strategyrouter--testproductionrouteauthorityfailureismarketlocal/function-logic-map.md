# Function Logic Map: `TestProductionRouteAuthorityFailureIsMarketLocal`

- Source: `internal/strategyrouter/production_test.go` (50-72)
- Function: `TestProductionRouteAuthorityFailureIsMarketLocal` in package `strategyrouter`
- Signature: `TestProductionRouteAuthorityFailureIsMarketLocal(params=1, results=0)`
- File SHA-256: `21d69cb2a4f5da5ae8d5aded5ea32e5e8fe4aa4d4b5ae0f64038a2bde6cb3d9a`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 5.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Test function. Proves a corrupt or wrong-mode KR manifest does not gate the US market, and now asserts that through `RouteSet` rather than `Route`.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`
- tagged package suite: `go test -count=1 -tags tossos_testseams -covermode=count ./internal/strategyrouter/`
- Measured entry: the function body was not instrumented — `go test` does not add coverage counters to `_test.go` files; it ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/`.

Exact AST return positions: none.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 54:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B2 | if | 58:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B3 | if | 63:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B4 | if | 66:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B5 | if | 69:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `newProductionRouteFixture` | 51:13 |
| `string` | 53:34 |
| `make` | 53:41 |
| `LoadProductionRouteAuthority` | 54:15 |
| `context.Background` | 54:44 |
| `t.Fatal` | 55:3 |
| `LoadProductionRouteAuthority` | 57:13 |
| `context.Background` | 57:42 |
| `RouteSet` | 58:19 |
| `us.Request` | 58:28 |
| `t.Fatalf` | 59:3 |
| `os.Chmod` | 63:12 |
| `filepath.Join` | 63:21 |
| `ProductionRouteFileName` | 63:48 |
| `t.Fatal` | 64:3 |
| `LoadProductionRouteAuthority` | 66:15 |
| `context.Background` | 66:44 |
| `t.Fatal` | 67:3 |
| `LoadProductionRouteAuthority` | 69:15 |
| `context.Background` | 69:44 |
| `t.Fatalf` | 70:3 |

## State mutations and fallbacks

- AST assignments: 9. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Test-only; it asserts market-local failure, which is the fault-isolation property design.md requires.

a112 게이트 준비(2026-10-04): a127 82080177(전략 권한 적재기가 현재 원장을 읽음)이 같은 파일을 편집해 이 번들이 낡았다 — 분기 · return · 호출 구조 동일
