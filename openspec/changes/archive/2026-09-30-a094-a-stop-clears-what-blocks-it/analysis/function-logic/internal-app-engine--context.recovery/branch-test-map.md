# Branch Test Map: `Context.Recovery`

- Source: `internal/app/engine/runtime_wiring.go`

> Test 열은 그 분기를 지나는 시험(현존 · 통과)이다. 「미진입」 분기의 인용은 그 함수를 도는 시험이며 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> RED 는 새 a094 시험에 한해 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:176` `if c == nil \|\| c.Journal == nil \|\| c.Resolver == nil \|\| c.Entry == nil {` | 아니오 | `TestA094TheRecoveryCarriesTheBootCatchUp` | n/a | yes |
| B2 | `:179` `if opts.Clock == nil {` | 예 | `TestA094TheRecoveryCarriesTheBootCatchUp` | n/a | yes |
| B3 | `:189` `if c.Notifier != nil {` | 예 | `TestA094TheRecoveryCarriesTheBootCatchUp` | n/a | yes |
| B4 | `:195` `if notifier != nil {` | 예 | `TestA094TheRecoveryCarriesTheBootCatchUp` | n/a | yes |
| B5 | `:201` `if err != nil {` | 아니오 | `TestA094TheRecoveryCarriesTheBootCatchUp` | n/a | yes |
