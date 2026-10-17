# Branch Test Map: `deliverEachStrategyHandoff`

- Source: `internal/app/engine/strategy_market_handoff_delivery.go` (66-93)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | range 68 | 조정자 순서대로 하나씩 | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Scope|Handoff|Delivery'` · `TestA146ASkipContinuesToTheNextScopeWithoutAnError` · `TestA146ASkipIsNeverJoinedIntoTheCycleError` · `TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops` · `TestA137AMixedCycleStillReportsItsSkippedScopes` |
| B2 | if 70 | 다음 범위 | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Scope|Handoff|Delivery'` · `TestA146ASkipContinuesToTheNextScopeWithoutAnError` · `TestA146ASkipIsNeverJoinedIntoTheCycleError` · `TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops` · `TestA137AMixedCycleStillReportsItsSkippedScopes` |
| B3 | if 74 | 다음 범위(끝에 Join 으로 보고 — 무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Scope|Handoff|Delivery'` · `TestA146ASkipContinuesToTheNextScopeWithoutAnError` · `TestA146ASkipIsNeverJoinedIntoTheCycleError` · `TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops` · `TestA137AMixedCycleStillReportsItsSkippedScopes` |
| B4 | if 82 | Join(skipped) — 일일 손실 계좌 차단은 주기를 조용히 끝냄(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Scope|Handoff|Delivery'` · `TestA146ASkipContinuesToTheNextScopeWithoutAnError` · `TestA146ASkipIsNeverJoinedIntoTheCycleError` · `TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops` · `TestA137AMixedCycleStillReportsItsSkippedScopes` |
| B5 | if 87 | 멈추는 결함 그대로 / skipped 와 Join(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|Scope|Handoff|Delivery'` · `TestA146ASkipContinuesToTheNextScopeWithoutAnError` · `TestA146ASkipIsNeverJoinedIntoTheCycleError` · `TestOnlyATypedScopeRefusalIsSkippedEveryOtherFaultStops` · `TestA137AMixedCycleStillReportsItsSkippedScopes` |
