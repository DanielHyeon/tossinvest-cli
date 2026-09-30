# Branch Test Map: `EvaluateLadderSnapshot`

- Source: `internal/exitpolicy/snapshot.go`

> Test 열은 그 함수를 지나는 현존 · 통과 시험이다. 「아니오」 분기의 인용은 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> a091 의 새 RED 는 구현 로트의 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다 — 이 표는 편집 뒤 `3ec1efd2` 의 사실이다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:98` `if err != nil {` | 예 | `TestOneShareIntermediateLadderTargetIsStateOnly` | n/a | yes |
| B2 | `:102` `if err != nil {` | 아니오 | `TestOneShareIntermediateLadderTargetIsStateOnly` | n/a | yes |
| B3 | `:111` `if strings.TrimSpace(eval.State.PolicyID) == "" {` | 아니오 | `TestOneShareIntermediateLadderTargetIsStateOnly` | n/a | yes |
| B4 | `:114` `if strings.TrimSpace(eval.State.PolicyVersion) == "" {` | 예 | `TestOneShareIntermediateLadderTargetIsStateOnly` | n/a | yes |
| B5 | `:117` `if strings.TrimSpace(eval.State.PolicyDigest) == "" {` | 예 | `TestOneShareIntermediateLadderTargetIsStateOnly` | n/a | yes |
| B6 | `:121` `if err != nil {` | 예 | `TestOneShareIntermediateLadderTargetIsStateOnly` | n/a | yes |
| B7 | `:126` `if err != nil {` | 아니오 | `TestOneShareIntermediateLadderTargetIsStateOnly` | n/a | yes |
| B8 | `:132` `if !transition.Proposal.Zero() {` | 예 | `TestOneShareIntermediateLadderTargetIsStateOnly` | n/a | yes |
| B9 | `:138` `if action.Orderable() {` | 예 | `TestOneShareIntermediateLadderTargetIsStateOnly` | n/a | yes |
| B10 | `:140` `if err != nil {` | 아니오 | `TestOneShareIntermediateLadderTargetIsStateOnly` | n/a | yes |
| B11 | `:147` `if err != nil {` | 아니오 | `TestOneShareIntermediateLadderTargetIsStateOnly` | n/a | yes |
