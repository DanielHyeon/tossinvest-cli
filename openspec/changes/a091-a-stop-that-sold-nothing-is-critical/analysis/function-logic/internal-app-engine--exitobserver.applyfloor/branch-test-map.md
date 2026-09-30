# Branch Test Map: `ExitObserver.applyFloor`

- Source: `internal/app/engine/exitloop.go`

> Test 열은 그 함수를 지나는 현존 · 통과 시험이다. 「아니오」 분기의 인용은 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> a091 의 새 RED 는 구현 로트의 변이 원장이 잰다 — 이 표는 편집 **전** base 의 사실이다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1618` `if o.opts.Floor == nil {` | 예 | `TestNoFloorSourceCapsNothing` | n/a | yes |
| B2 | `:1622` `if err != nil {` | 예 | `TestAFloorThatCannotBeComputedSellsNothing` | n/a | yes |
| B3 | `:1630` `if !applies {` | 예 | `TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable` | n/a | yes |
| B4 | `:1634` `if err != nil {` | 아니오 | `TestTheConfirmedFloorCapsTheLiquidation` | n/a | yes |
| B5 | `:1637` `if cmp >= 0 {` | 예 | `TestTheConfirmedFloorCapsTheLiquidation` | n/a | yes |
| B6 | `:1641` `if err != nil {` | 아니오 | `TestTheConfirmedFloorCapsTheLiquidation` | n/a | yes |
