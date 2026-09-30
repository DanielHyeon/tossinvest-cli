# Branch Test Map: `ExitObserver.judge`

- Source: `internal/app/engine/exitloop.go`

> Test 열은 그 분기를 지나는 시험(현존 · 통과)이다. 「미진입」 분기의 인용은 그 함수를 도는 시험이며 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> RED 는 새 a094 시험에 한해 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:900` `if !o.quoteUsable(quote) {` | 아니오 | `TestA094TheParkCheckChangesNoOutcome` | n/a | yes |
| B2 | `:906` `if m.identityErr != nil {` | 예 | `TestARungTableSwappedUnderALivePositionIsRefused` | n/a | yes |
| B3 | `:910` `if m.reJudge {` | 예 | `TestCrossingTheFirstTakeProfitKeepsThePositionUnderJudgement` | n/a | yes |
| B4 | `:920` `if err := o.opts.Journal.StampExitSnapshotQuarantineSelector(ctx,` | — | `TestCrossingTheFirstTakeProfitKeepsThePositionUnderJudgement` | n/a | yes |
| B5 | `:926` `if err != nil {` | 아니오 | `TestA094TheParkCheckChangesNoOutcome` | n/a | yes |
| B6 | `:931` `switch m.state.PolicyKind {` | 예 | `TestA094AParkedStopIsNamedOnceAndStaysSuppressed` | n/a | yes |
| B7 | `:932` `case journal.ExitPolicyLadder:` | 예 | `TestA094AnUnacceptedStopIsReleasedAndProposedAgain` | n/a | yes |
| B8 | `:934` `default:` | 예 | `TestA094AParkedStopIsNamedOnceAndStaysSuppressed` | n/a | yes |
