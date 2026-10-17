# Branch Test Map: `Journal.ResolveExitProposal`

- Source: `internal/journal/apply_hook.go`

> Test 열은 그 분기를 지나는 시험(현존 · 통과)이다. 「미진입」 분기의 인용은 그 함수를 도는 시험이며 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> RED 는 새 a094 시험에 한해 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:832` `if strings.TrimSpace(expectedIntentID) == "" {` | 예 | `TestA094ALateReleaseDoesNotClearAnotherProposal` | n/a | yes |
| B2 | `:836` `if err != nil {` | 아니오 | `TestResolvingNothingIsNotAnError` | n/a | yes |
| B3 | `:840` `if err != nil {` | 아니오 | `TestResolvingNothingIsNotAnError` | n/a | yes |
| B4 | `:845` `if err != nil \|\| !released {` | 예 | `TestResolvingNothingIsNotAnError` · `TestA094ALateReleaseDoesNotClearAnotherProposal` | n/a | yes |
| B5 | `:848` `if err := tx.Commit(); err != nil {` | 예 | `TestARefusalReArmsTheLevel` · `TestACancelledRungIsProposableAgain` | n/a | yes |
