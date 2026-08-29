# Function Logic Map: `TestProductionRouteAuthorityFailureIsMarketLocal`

- Source: `internal/strategyrouter/production_test.go` (45-67)
- Function: `TestProductionRouteAuthorityFailureIsMarketLocal` in package `strategyrouter`
- Signature: `TestProductionRouteAuthorityFailureIsMarketLocal(params=1, results=0)`
- File SHA-256: `fccd226dcf67215fe7792bb850fbf5ffbddd93c399f72028be3dc6a55400bd38`
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
| B1 | if | 49:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B2 | if | 53:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B3 | if | 58:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B4 | if | 61:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B5 | if | 64:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `newProductionRouteFixture` | 46:13 |
| `string` | 48:34 |
| `make` | 48:41 |
| `LoadProductionRouteAuthority` | 49:15 |
| `context.Background` | 49:44 |
| `t.Fatal` | 50:3 |
| `LoadProductionRouteAuthority` | 52:13 |
| `context.Background` | 52:42 |
| `RouteSet` | 53:19 |
| `us.Request` | 53:28 |
| `t.Fatalf` | 54:3 |
| `os.Chmod` | 58:12 |
| `filepath.Join` | 58:21 |
| `ProductionRouteFileName` | 58:48 |
| `t.Fatal` | 59:3 |
| `LoadProductionRouteAuthority` | 61:15 |
| `context.Background` | 61:44 |
| `t.Fatal` | 62:3 |
| `LoadProductionRouteAuthority` | 64:15 |
| `context.Background` | 64:44 |
| `t.Fatalf` | 65:3 |

## State mutations and fallbacks

- AST assignments: 9. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Test-only; it asserts market-local failure, which is the fault-isolation property design.md requires.
