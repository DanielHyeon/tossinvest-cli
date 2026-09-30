# Branch Test Map: `ExitObserver.release`

- Source: `internal/app/engine/exitloop.go`

> Test 열은 그 분기를 지나는 시험(현존 · 통과)이다. 「미진입」 분기의 인용은 그 함수를 도는 시험이며 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> RED 는 새 a094 시험에 한해 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1487` `if _, _, err := o.opts.Journal.ReleaseUnacceptedExitProposal(ctx, m.position.ID, intentID, how); err != nil {` | 예 | `TestA094AnUnacceptedStopIsReleasedAndProposedAgain` · `TestARefusedProposalReleasesTheLevelAndAlerts` | n/a | yes |
