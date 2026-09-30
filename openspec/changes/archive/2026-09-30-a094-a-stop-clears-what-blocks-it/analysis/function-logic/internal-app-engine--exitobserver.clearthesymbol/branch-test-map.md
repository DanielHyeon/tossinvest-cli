# Branch Test Map: `ExitObserver.clearTheSymbol`

- Source: `internal/app/engine/exitloop.go`

> Test 열은 그 분기를 지나는 시험(현존 · 통과)이다. 「미진입」 분기의 인용은 그 함수를 도는 시험이며 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> RED 는 새 a094 시험에 한해 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1515` `if err != nil {` | 예 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B2 | `:1519` `if err != nil {` | 아니오 | `TestA094AnotherIntentInFlightStopsTheArmReleaseLoop` | n/a | yes |
| B3 | `:1522` `if len(unsettled) > 0 {` | 예 | `TestA094AnotherIntentInFlightStopsTheArmReleaseLoop` · `TestA094AnInFlightEngineCancelIsNotCountedButAnInDoubtOneIs` | n/a | yes |
| B4 | `:1528` `for _, order := range live {` | 예 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` | n/a | yes |
| B5 | `:1530` `if !buy {` | 예 | `TestABreachDisplacesAnOutstandingTakeProfit` | n/a | yes |
| B6 | `:1534` `if err != nil {` | 아니오 | `TestA094ACancelledSellIsNotClearedUntilItsClosingRecord` | n/a | yes |
| B7 | `:1539` `if waiting {` | 예 | `TestA094ACancelledSellIsNotClearedUntilItsClosingRecord` | n/a | yes |
| B8 | `:1543` `if !withPending {` | 예 | `TestA094ACancelledSellIsNotClearedUntilItsClosingRecord` | n/a | yes |
| B9 | `:1559` `if err != nil {` | 아니오 | `TestA094ConsecutiveClearFailuresRaiseAnEarlierAlert` | n/a | yes |
| B10 | `:1565` `if qerr != nil \|\| perr != nil {` | 아니오 | `TestA094AnEmptyPriceIsNotAClearFailure` | n/a | yes |
| B11 | `:1580` `if err != nil \|\| out.State != journal.StateConfirmed {` | 예 | `TestAnUncancellableEntryWithholdsTheLiquidationAndAlertsPastTheBound` | n/a | yes |
| B12 | `:1584` `if !buy {` | 예 | `TestA094ACancelledSellIsNotClearedUntilItsClosingRecord` · `TestABreachDisplacesAnOutstandingTakeProfit` | n/a | yes |
| B13 | `:1589` `if !res.cleared {` | 예 | `TestA094AClosingRecordThatNeverComesHoldsAndAlerts` | n/a | yes |
| B14 | `:1592` `if withPending && m.state.Pending() {` | 예 | `TestA094AParkedTakeProfitIsNotClearedAndIsNamed` | n/a | yes |
| B15 | `:1593` `if strings.TrimSpace(m.state.PendingIntentID) == "" {` | 아니오 | `TestA094AParkedTakeProfitIsNotClearedAndIsNamed` | n/a | yes |
| B16 | `:1598` `if err != nil {` | 아니오 | `TestA094AParkedTakeProfitIsNotClearedAndIsNamed` | n/a | yes |
| B17 | `:1601` `if !released {` | 예 | `TestA094AParkedTakeProfitIsNotClearedAndIsNamed` | n/a | yes |
