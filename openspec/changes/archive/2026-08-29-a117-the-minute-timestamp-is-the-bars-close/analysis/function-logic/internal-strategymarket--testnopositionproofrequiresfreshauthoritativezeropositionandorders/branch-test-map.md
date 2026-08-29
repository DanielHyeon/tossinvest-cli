# Branch Test Map: `TestNoPositionProofRequiresFreshAuthoritativeZeroPositionAndOrders`

- Source: `internal/strategymarket/bars_test.go`, SHA-256 `b94852735790b61add409d2cf155b54e1b7113bcfc389bc120d7236c722c8ed3`; branch IDs follow `ast.json` (regenerated 2026-08-18 after the edit).
- AST counts: branches 4, returns 0, defers 0, go statements 0. Source range `153:1-177:2`.
- Test-fixture bundle: this function is test-only; it has no production branch to hold.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | the position reading must be authoritative | `TestNoPositionProofRequiresFreshAuthoritativeZeroPositionAndOrders` (this function is the test) | not edited | green |
| B2 | the reading must be fresh | `TestNoPositionProofRequiresFreshAuthoritativeZeroPositionAndOrders` | not edited | green |
| B3 | each forged/zero variant is walked | `TestNoPositionProofRequiresFreshAuthoritativeZeroPositionAndOrders` | not edited | green |
| B4 | each variant must be refused | `TestNoPositionProofRequiresFreshAuthoritativeZeroPositionAndOrders` | not edited | green |

Verification: `go test ./... -count=1` green on 2026-08-18 (9,425 tests, 102 packages, exit 0).
