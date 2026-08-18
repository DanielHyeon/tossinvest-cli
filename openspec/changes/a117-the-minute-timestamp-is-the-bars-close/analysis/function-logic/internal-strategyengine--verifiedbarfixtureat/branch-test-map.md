# Branch Test Map: `verifiedBarFixtureAt`

- Source: `internal/strategyengine/lane_test.go`, SHA-256 `6b224a50c8db46a7c66b6fe0d0ef120bb103045b54a715f6a13a64f9b0c8fdf2`; branch IDs follow `ast.json` (regenerated 2026-08-18 after the edit).
- AST counts: branches 3, returns 1, defers 0, go statements 0. Source range `76:1-96:2`.
- Test-fixture bundle: this function is test-only; it has no production branch to hold.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | five minute labels for the bucket opening at `start` | `TestParkerSessionUsesInjectedEvaluationTimeAndInclusiveCutoff` and every other consumer of the fixture | the pre-shift labels made every consuming test fail with `incomplete_bucket: not aligned to KRX five-minute boundary` | green |
| B2 | caller mutates a row | `TestParkerIlliquidAndFakeBreakoutGatesPreserveFrozenPrecedence` and `TestVersionedMarketInputRejectsMalformedDecimalsWithoutCoercion` | not edited | green |
| B3 | seal refuses | not-applicable: reached only when a mutator makes the page invalid, and those cases assert the refusal at the lane instead | not edited | not-applicable |

Verification: `go test ./... -count=1` green on 2026-08-18 (9,425 tests, 102 packages, exit 0).
