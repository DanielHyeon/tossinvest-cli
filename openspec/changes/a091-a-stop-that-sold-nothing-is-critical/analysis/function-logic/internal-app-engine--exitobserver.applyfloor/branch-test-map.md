# Branch Test Map: `ExitObserver.applyFloor`

- Source: `internal/app/engine/exitloop.go`

> Test 열은 그 함수를 지나는 현존 · 통과 시험이다. 「아니오」 분기의 인용은 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> a091 의 새 RED 는 구현 로트의 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다 — 이 표는 편집 뒤 `3ec1efd2` 의 사실이다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1625` `if o.opts.Floor == nil {` | 예 | `TestNoFloorSourceCapsNothing` | n/a | yes |
| B2 | `:1629` `if err != nil {` | 예 | `TestA091AFloorThatCannotBeComputedIsTheSameReport` · `TestA091OnlyACancellationOnlyFailureIsSuppressed` | n/a | yes |
| B3 | `:1638` `if !applies {` | 예 | `TestA091TheOutcomeIsUnchanged` | n/a | yes |
| B4 | `:1642` `if err != nil {` | 아니오 | `TestTheConfirmedFloorCapsTheLiquidation` | n/a | yes |
| B5 | `:1645` `if cmp >= 0 {` | 예 | `TestA091APartialCapIsUnchanged` | n/a | yes |
| B6 | `:1649` `if err != nil {` | 아니오 | `TestTheConfirmedFloorCapsTheLiquidation` | n/a | yes |
| B7 | `:1652` `if isZeroQuantity(floor.Quantity) {` | 예 | `TestA091AProtectiveZeroIsACriticalRow` · `TestA091AZeroHoldingIsNotAFailedStop` | n/a | yes |
