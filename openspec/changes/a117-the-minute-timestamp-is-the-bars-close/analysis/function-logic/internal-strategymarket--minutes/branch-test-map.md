# Branch Test Map: `minutes`

- Source: `internal/strategymarket/bars_test.go`, SHA-256 `ffeec4a89d78fe4855e5ce30829c88c97f959a21d242602cf3b9b1221f81558f`; branch IDs follow `ast.json` (regenerated 2026-08-18 after the edit).
- AST counts: branches 1, returns 1, defers 0, go statements 0. Source range `16:1-23:2`.
- Test-fixture bundle: this function is test-only; it has no production branch to hold.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | five consecutive minute labels for the `09:00` bucket | every test in `bars_test.go` that calls `minutes()` | the pre-shift fixture made `TestAggregateClosedKRXFiveMinutePreservesExactDecimals` and three table cases fail with `outside_regular_session: 2026-07-31T09:00:00+09:00` | green |

Verification: `go test ./... -count=1` green on 2026-08-18 (9,425 tests, 102 packages, exit 0).
