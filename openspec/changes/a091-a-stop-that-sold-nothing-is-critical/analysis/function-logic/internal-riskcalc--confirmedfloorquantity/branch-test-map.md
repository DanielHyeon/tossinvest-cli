# Branch Test Map: `ConfirmedFloorQuantity`

- Source: `internal/riskcalc/confirmed_floor.go`

> Test 열은 그 함수를 지나는 현존 · 통과 시험이다. 「아니오」 분기의 인용은 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> a091 의 새 RED 는 구현 로트의 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다 — 이 표는 편집 뒤 `3ec1efd2` 의 사실이다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:130` `if in.Now.IsZero() {` | 예 | `TestTheFloorIsTheFormula` | n/a | yes |
| B2 | `:142` `if err != nil {` | 예 | `TestAnUnknownLocalQuantityIsRefused` | n/a | yes |
| B3 | `:147` `if err != nil {` | 예 | `TestMalformedQuantitiesAreRefused` | n/a | yes |
| B4 | `:150` `if !ok {` | 예 | `TestAnAbsentSnapshotIsZero` | n/a | yes |
| B5 | `:154` `if err != nil {` | 예 | `TestMalformedQuantitiesAreRefused` | n/a | yes |
| B6 | `:157` `if !ok {` | 예 | `TestAStaleSnapshotIsZero` | n/a | yes |
| B7 | `:165` `if err != nil {` | 아니오 | `TestTheFloorIsTheFormula` | n/a | yes |
| B8 | `:169` `if base == sellable && sellable != holdings {` | 예 | `TestTheSellableQuantityOnlyLowers` | n/a | yes |
| B9 | `:174` `if err != nil {` | 아니오 | `TestTheFloorIsTheFormula` | n/a | yes |
| B10 | `:181` `if err != nil {` | 아니오 | `TestTheFloorIsTheFormula` | n/a | yes |
| B11 | `:184` `if floor != base {` | 예 | `TestALocalQuantityCanNeverRaiseTheFloor` | n/a | yes |
