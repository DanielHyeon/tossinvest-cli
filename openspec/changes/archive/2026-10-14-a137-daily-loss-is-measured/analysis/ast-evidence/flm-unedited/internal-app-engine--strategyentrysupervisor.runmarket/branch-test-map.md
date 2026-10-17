# Branch Test Map: `StrategyEntrySupervisor.runMarket`

편집 전 좌표. a137 은 이 함수를 편집하지 않는다 — 아래 새 시험은 「일일 손실 거절이 B11 로 오고 B12 · (966) 에 닿지 않음」을 고정한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B11 | 일일 손실 한도 거절 · 그 뒤 모드 투영 거절이 오류 없이 주기를 끝냄 | `TestA137ADailyLossStopEndsTheCycleWithoutAnError`(새, refreshOnly · effective 두 조립) | no | no |
| B12 | refreshOnly 레인이 일일 손실 거절을 삼키지 않음(계수 0) | `TestA137ARefreshOnlyLaneNeverSwallowsTheDailyLossStop`(새) · `TestARefreshOnlyWorkerCountsTheCycleErrorsItSwallows`(기존) | no | no |
| B13 · B14 | 중앙 무결성 분류 불변 | `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal`(새 표) · `TestARefreshOnlyWorkerCentralIntegrityErrorLeavesTheEngineRunning`(기존) | no | no |
| (966) | effective 레인이 일일 손실 거절로 잠기지 않음(`RecordStrategyLaneLatch` 호출 0) | `TestA137AnActivatedLaneIsNotLatchedByTheDailyLossStop`(새) | no | no |
| B1~B10 · B15~B17 | a137 무관 | 기존 감독자 시험 | — | — |
