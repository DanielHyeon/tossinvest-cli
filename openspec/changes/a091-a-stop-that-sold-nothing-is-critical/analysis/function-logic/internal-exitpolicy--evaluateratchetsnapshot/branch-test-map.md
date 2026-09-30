# Branch Test Map: `EvaluateRatchetSnapshot`

- Source: `internal/exitpolicy/snapshot.go`

> Test 열은 그 함수를 지나는 현존 · 통과 시험이다. 「아니오」 분기의 인용은 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> a091 의 새 RED 는 구현 로트의 변이 원장이 잰다 — 이 표는 편집 **전** base 의 사실이다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:178` `if in.Input.Config != nil {` | 예 | `TestRatchetOneSharePartialIsStateOnlyButBreachIsFull` | n/a | yes |
| B2 | `:182` `if err != nil {` | 아니오 | `TestRatchetOneSharePartialIsStateOnlyButBreachIsFull` | n/a | yes |
| B3 | `:186` `if err != nil {` | 아니오 | `TestRatchetOneSharePartialIsStateOnlyButBreachIsFull` | n/a | yes |
| B4 | `:190` `if err != nil {` | 아니오 | `TestRatchetOneSharePartialIsStateOnlyButBreachIsFull` | n/a | yes |
| B5 | `:194` `if err != nil {` | 아니오 | `TestRatchetOneSharePartialIsStateOnlyButBreachIsFull` | n/a | yes |
| B6 | `:203` `if action.Orderable() {` | 예 | `TestRatchetOneSharePartialIsStateOnlyButBreachIsFull` | n/a | yes |
| B7 | `:205` `if err != nil {` | 아니오 | `TestRatchetOneSharePartialIsStateOnlyButBreachIsFull` | n/a | yes |
| B8 | `:212` `if err != nil {` | 아니오 | `TestRatchetOneSharePartialIsStateOnlyButBreachIsFull` | n/a | yes |
| B9 | `:216` `if !previousLevel.Valid() {` | 예 | `TestRatchetOneSharePartialIsStateOnlyButBreachIsFull` | n/a | yes |
| B10 | `:218` `if err != nil {` | 아니오 | `TestRatchetOneSharePartialIsStateOnlyButBreachIsFull` | n/a | yes |
