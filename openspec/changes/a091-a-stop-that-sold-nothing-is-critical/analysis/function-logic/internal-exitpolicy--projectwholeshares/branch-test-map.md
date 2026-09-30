# Branch Test Map: `ProjectWholeShares`

- Source: `internal/exitpolicy/snapshot.go`

> Test 열은 그 함수를 지나는 현존 · 통과 시험이다. 「아니오」 분기의 인용은 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> a091 의 새 RED 는 구현 로트의 변이 원장이 잰다 — 이 표는 편집 **전** base 의 사실이다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:75` `if !ok {` | 아니오 | `TestOneShareFinalAndBreachProjectExactlyOne` | n/a | yes |
| B2 | `:79` `if r := strings.TrimSpace(ratio); r != "" {` | 예 | `TestOneShareFinalAndBreachProjectExactlyOne` | n/a | yes |
| B3 | `:81` `if !ok {` | 아니오 | `TestOneShareFinalAndBreachProjectExactlyOne` | n/a | yes |
| B4 | `:85` `if share.Cmp(one) > 0 {` | 예 | `TestOneShareFinalAndBreachProjectExactlyOne` | n/a | yes |
| B5 | `:90` `if units.Sign() < 0 {` | 아니오 | `TestOneShareFinalAndBreachProjectExactlyOne` | n/a | yes |
