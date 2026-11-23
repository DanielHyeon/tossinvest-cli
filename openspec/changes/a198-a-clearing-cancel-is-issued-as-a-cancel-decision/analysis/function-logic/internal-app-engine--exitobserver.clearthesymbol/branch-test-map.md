# Branch Test Map: `ExitObserver.clearTheSymbol` (pre-edit)

덮음 측정: `go test ./internal/app/engine -run 'A094|ExitClear|Clear|WorkingOrder|CancelPending' -count=1 -coverprofile`(HEAD `e1eaf720c`, rc 0, 35.3s) 의 블록 계수. **선택 실행 기준** — 이 선택 밖의 시험이 덮는지는 미측정. 시험 이름은 grep 근거.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 목록 읽기 실패 | TestA094AListReadFailureIsAnErrorAsBefore (덮음 1) | n/a(기존) | 기존 green |
| B2 | 미종결 읽기 실패 | 없음 (블록 1656-1658 계수 0) | n/a | 미덮음 |
| B3 | 같은 종목 미종결 | TestA094TheClearAndTheGatewayShareOneUnsettledJudgement · TestA094AnInFlightEngineCancelIsNotCountedButAnInDoubtOneIs (덮음) | n/a | 기존 green · a198 §4-6 예정 |
| B4 | 순회 | 다수 (덮음) | n/a | 기존 green |
| B5 | 매도 주문 | TestA094ACancelledSellIsNotClearedUntilItsClosingRecord (덮음) | n/a | 기존 green |
| B6 | ConfirmedCancelOf 오류 | 없음 (1671-1674 계수 0) | n/a | 미덮음 |
| B7 | 접수된 엔진 취소 대기 | TestA094ACancelledSellIsNotClearedUntilItsClosingRecord · TestA094ACancelledSellStillHoldsTheStopAfterTheProposalIsReleased (덮음 — 시험이 원장에 직접 쓴 CANCEL 행) | n/a | 기존 green · a198 §4-3 예정(게이트웨이 행) |
| B8 | 발의 없는 매도 | TestA094AnOwnSideSellDoesNotWithholdAStopWithoutAProposal (덮음) | n/a | 기존 green |
| B9 | 결정 발행 실패 | 없음 (1696-1698 계수 0) | a198 §4-5 예정 | 미덮음 |
| B10 | 수량·가격 해석 실패 | 없음 (1702-1704 계수 0) | n/a | 미덮음 |
| B11 | 취소 오류·비확정 | TestAnUncancellableEntryWithholdsTheLiquidationAndAlertsPastTheBound · TestA094ConsecutiveClearFailuresRaiseAnEarlierAlert 등 `cancelFails` (덮음 — 가짜 제출자) | a198 §4-1 예정(진짜 게이트웨이 거절) | 기존 green(가짜) |
| B12 | 접수된 매도 취소 | a094 취소 매도 시험들 (덮음 — 가짜 제출자 + a094RecordCancelOn) | n/a | 기존 green · a198 §4-3 예정 |
| B13 | 미치움 반환 | (덮음) | n/a | 기존 green |
| B14 | 발의 있음 | (덮음) | n/a | 기존 green |
| B15 | 발의 intent 없음 | 없음 (1730-1733 계수 0) | n/a | 미덮음 |
| B16 | 해제 판정 오류 | 없음 (1735-1737 계수 0) | n/a | 미덮음 |
| B17 | 해제 안 됨 | TestA094ALiveProposalAttemptIsNotReleased (덮음) | n/a | 기존 green |
