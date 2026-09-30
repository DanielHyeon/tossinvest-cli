# Branch Test Map: `Converger.ConvergeQuantities`

- Source: `internal/reconcile/converge.go`

> Test 열은 그 함수를 지나는 현존 · 통과 시험이다. 「아니오」 분기의 인용은 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> a091 의 새 RED 는 구현 로트의 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다 — 이 표는 편집 뒤 `3ec1efd2` 의 사실이다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:144` `if len(diff.Quantities) == 0 {` | 예 | `TestNoMismatchesIsANoOp` | n/a | yes |
| B2 | `:147` `if c == nil \|\| c.Journal == nil {` | 예 | `TestAQuantityMismatchConvergesToTheAccount` | n/a | yes |
| B3 | `:152` `if account == "" {` | 아니오 | `TestAQuantityMismatchConvergesToTheAccount` | n/a | yes |
| B4 | `:156` `if asOf == "" {` | 아니오 | `TestAQuantityMismatchConvergesToTheAccount` | n/a | yes |
| B5 | `:173` `if len(credited) == 0 {` | 예 | `TestConvergenceMakesTheBlockReleasable` | n/a | yes |
| B6 | `:178` `if c.Credit != nil {` | 예 | `TestConvergenceMakesTheBlockReleasable` | n/a | yes |
| B7 | `:194` `if err != nil {` | 아니오 | `TestAQuantityMismatchConvergesToTheAccount` | n/a | yes |
| B8 | `:197` `for _, mismatch := range diff.Quantities {` | 예 | `TestAQuantityMismatchConvergesToTheAccount` | n/a | yes |
| B9 | `:200` `if reason != "" {` | 예 | `TestASymbolWithNoLiveInstanceIsRefusedRatherThanFolded` | n/a | yes |
| B10 | `:206` `if err != nil {` | 아니오 | `TestAQuantityMismatchConvergesToTheAccount` | n/a | yes |
| B11 | `:228` `if errors.Is(err, journal.ErrAdjustmentStale) {` | 예 | `TestAStaleAdjustmentStopsThePass` | n/a | yes |
| B12 | `:233` `if err != nil {` | 예 | `TestAStaleAdjustmentStopsThePass` | n/a | yes |
| B13 | `:244` `if result.ClosedExitState {` | 예 | `TestConvergingAManagedPositionToZeroAlerts` | n/a | yes |
| B14 | `:251` `if c.Alert == nil {` | 아니오 | `TestConvergingAManagedPositionToZeroAlerts` | n/a | yes |
| B15 | `:254` `if err := c.Alert.ManagedPositionClosedExternally(ctx, ManagedCloseAlert{` | 예 | `TestConvergingAManagedPositionToZeroAlerts` | n/a | yes |
