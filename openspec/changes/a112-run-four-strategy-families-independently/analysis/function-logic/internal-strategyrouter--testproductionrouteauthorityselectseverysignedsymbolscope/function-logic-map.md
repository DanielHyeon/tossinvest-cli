# Function Logic Map: `TestProductionRouteAuthoritySelectsEverySignedSymbolScope`

- Source: `internal/strategyrouter/production_test.go` (69-99)
- Function: `TestProductionRouteAuthoritySelectsEverySignedSymbolScope` in package `strategyrouter`
- Signature: `TestProductionRouteAuthoritySelectsEverySignedSymbolScope(params=1, results=0)`
- File SHA-256: `fccd226dcf67215fe7792bb850fbf5ffbddd93c399f72028be3dc6a55400bd38`
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
| B1 | range | 71:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B2 | range | 81:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B3 | if | 91:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B4 | if | 95:3 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `newProductionRouteFixture` | 70:13 |
| `fixture.body` | 78:11 |
| `append` | 85:17 |
| `fixture.write` | 86:3 |
| `LoadProductionRouteAuthority` | 90:21 |
| `context.Background` | 90:50 |
| `t.Fatalf` | 92:4 |
| `authority.Request` | 94:14 |
| `len` | 95:43 |
| `RouteSet` | 95:75 |
| `t.Fatalf` | 96:4 |

## State mutations and fallbacks

- AST assignments: 12. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Test-only; `go test` does not instrument `_test.go` files, so the measured counts below come from the production blocks this test drives, not from the test body.
