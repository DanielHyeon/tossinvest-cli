# Function Logic Map: `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`

- Source: `internal/strategyrouter/production_test.go` (74-104)
- Function: `TestProductionRouteAuthoritySelectsEverySignedSymbolScope` in package `strategyrouter`
- Signature: `TestProductionRouteAuthoritySelectsEverySignedSymbolScope(params=1, results=0)`
- File SHA-256: `21d69cb2a4f5da5ae8d5aded5ea32e5e8fe4aa4d4b5ae0f64038a2bde6cb3d9a`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 4.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Test function. Proves a second signed symbol scope in the same market is selectable, and now asserts the sealed `RouteSet` accepts it rather than the legacy single-winner `Route`.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`
- tagged package suite: `go test -count=1 -tags tossos_testseams -covermode=count ./internal/strategyrouter/`
- Measured entry: the function body was not instrumented — `go test` does not add coverage counters to `_test.go` files; it ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/`.

Exact AST return positions: none.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | range | 76:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B2 | range | 86:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B3 | if | 96:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B4 | if | 100:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `newProductionRouteFixture` | 75:13 |
| `fixture.body` | 83:11 |
| `append` | 90:17 |
| `fixture.write` | 91:3 |
| `LoadProductionRouteAuthority` | 95:21 |
| `context.Background` | 95:50 |
| `t.Fatalf` | 97:4 |
| `authority.Request` | 99:14 |
| `len` | 100:43 |
| `RouteSet` | 100:75 |
| `t.Fatalf` | 101:4 |

## State mutations and fallbacks

- AST assignments: 12. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Test-only; `go test` does not instrument `_test.go` files, so the measured counts below come from the production blocks this test drives, not from the test body.

a112 게이트 준비(2026-10-04): a127 82080177(전략 권한 적재기가 현재 원장을 읽음)이 같은 파일을 편집해 이 번들이 낡았다 — 분기 · return · 호출 구조 동일
