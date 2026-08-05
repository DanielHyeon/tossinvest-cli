# Branch Test Map: `Converger.ConvergeQuantities`

AST의 모든 분기를 1행씩 덮는다. ★ 표시가 이 change가 편집하는 분기다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (138) 불일치가 없으면 아무것도 하지 않는다 | `TestNoMismatchesIsANoOp` | n/a | 기존 |
| B2 | (141) journal 없이 수렴할 수 없다 | `TestConvergeRefusesWithoutAJournal` | n/a | 기존 |
| B3 | (146) 계좌가 지목되지 않으면 거부한다 | `TestConvergeRefusesWithoutAnAccount` | n/a | 기존 |
| B4 | (150) as-of 없는 비교는 수렴 자체를 거부한다 | `TestAConvergenceNeedsAnAsOf` | n/a | 기존 |
| B5 | (157) 투영 읽기 실패는 pass를 멈춘다 | `TestAStaleAdjustmentStopsThePass` | n/a | 기존 |
| B6 | (165) 불일치 심볼마다 조정 1건을 발행한다 | `TestAQuantityMismatchConvergesToTheAccount` | n/a | 기존 |
| B7 | (168) 단일 venue를 고를 수 없으면 거부하고 다른 심볼은 계속한다 | `TestASymbolWithNoLiveInstanceIsRefusedRatherThanFolded` | n/a | 기존 |
| B8 | (174) watermark 읽기 실패는 pass를 멈춘다 | `TestAStaleAdjustmentStopsThePass` | n/a | 기존 |
| B9 | (196) stale 조정은 pass를 멈추고 커밋된 것만 report에 남는다 | `TestAStaleAdjustmentStopsThePass` | n/a | 기존 |
| B10 | (201) 기타 조정 오류는 pass를 멈춘다 | `TestAStaleAdjustmentStopsThePass` | n/a | 기존 |
| B11 | (212) 관리 포지션이 0이 되면 exit state가 닫히고 계수된다 | `TestConvergingAManagedPositionToZeroAlerts` | n/a | 기존 |
| B12 | (219) Alert가 없으면 조용히 넘어간다 | `TestConvergingAnUnmanagedPositionToZeroDoesNotAlert` | n/a | 기존 |
| B13 | (222) 알림 실패는 모아서 반환하되 조정은 유지된다 | `TestConvergingAManagedPositionToZeroAlerts` | n/a | 기존 |
| B14 | (236) 수렴한 심볼이 있을 때만 credit한다 | `TestNoMismatchesIsANoOp` | n/a | 기존 |
| B15 | ★(239) credit이 diff의 as-of를 함께 전달하고, 재적용된 조정도 credit되며, Credit이 nil이면 수렴만 한다 | `TestTheCreditCarriesTheComparisonItWasComputedFrom`, `TestAReappliedAdjustmentIsStillCredited`, `TestConvergenceMakesTheBlockReleasable` | yes | yes |
