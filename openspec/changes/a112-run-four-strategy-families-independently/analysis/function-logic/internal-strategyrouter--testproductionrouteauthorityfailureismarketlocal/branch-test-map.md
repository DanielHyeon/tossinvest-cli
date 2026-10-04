# Branch Test Map: `TestProductionRouteAuthorityFailureIsMarketLocal`

- Source: `internal/strategyrouter/production_test.go`; file SHA-256 `21d69cb2a4f5da5ae8d5aded5ea32e5e8fe4aa4d4b5ae0f64038a2bde6cb3d9a`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`
- tagged package suite: `go test -count=1 -tags tossos_testseams -covermode=count ./internal/strategyrouter/`

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 54:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B2 | if at 58:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B3 | if at 63:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B4 | if at 66:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |
| B5 | if at 69:2 | not instrumented — `go test` does not add coverage counters to `_test.go` files; the enclosing test ran and passed in `go test -count=1 ./internal/strategyrouter/` and `go test -count=1 -tags tossos_testseams ./internal/strategyrouter/` |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.
