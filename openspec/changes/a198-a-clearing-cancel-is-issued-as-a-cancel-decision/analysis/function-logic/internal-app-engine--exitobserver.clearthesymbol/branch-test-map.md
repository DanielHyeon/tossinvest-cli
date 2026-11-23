# Branch Test Map: `ExitObserver.clearTheSymbol` (post-edit)

RED/GREEN 은 `analysis/tdd/red.txt`(HEAD production 코드)와 4.2 의 선택 시험 원문(review.md)이 근거. 기존 시험 덮음은 pre-edit 판의 선택 실행 coverprofile(블록 계수) 기준 — post-edit 커버리지는 재측정하지 않음.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 목록 읽기 실패 | TestA094AListReadFailureIsAnErrorAsBefore | n/a(기존) | 기존 green(4.2 선택 실행) |
| B2 | 미종결 읽기 실패 | 없음 | n/a | 미덮음 |
| B3 | 같은 종목 미종결(IN_DOUBT·ACKED 취소) | TestA198AnInDoubtCancelHoldsTheStopAndBothAlertsFire · TestA198AnAckedCancelHoldsTheStopAndTheDelayAlertStillFires | InDoubt: FAIL(취소 0) · Acked: HEAD 에서도 PASS(기존 규칙 특성 시험) | PASS |
| B4 | 순회 | 다수 | — | PASS |
| B5 | 매도 주문 | TestA198AnArmedSellIsCancelledOnceAndTheStopWaitsForItsClosingRecord | FAIL(취소 0) | PASS |
| B6 | ConfirmedCancelOf 오류 | 없음 | n/a | 미덮음 |
| B7 | 게이트웨이가 쓴 확정 취소 대기 | 같은 ArmedSell 시험(둘째 주기: 취소 여전히 1) | FAIL | PASS |
| B8 | 발의 없는 매도 | TestA094AnOwnSideSellDoesNotWithholdAStopWithoutAProposal | n/a | 기존 green |
| B9 | 수량·가격 해석 실패 | 없음 | n/a | 미덮음 |
| B10 | 취소 결정 발행 실패 | TestA198AnUnrecordableCancelDecisionIsLoggedOnce(원장 트리거로 위험 감소 결정 기록 거절) | FAIL(줄 0) | PASS — 줄 1, 반복에도 1 |
| B11 | 취소 거절·모호 | TestA198ARefusedCancelIsLoggedOnceWithoutAccountQuantityPriceOrDetail · InDoubt 시험 | FAIL(줄 0 / 취소 0) | PASS — 400 줄 1, 반복 1, 500 으로 바뀌면 2 |
| B11 거짓 | 매수 취소 접수 → 같은 주기 손절 | TestA198AWorkingEngineBuyIsCancelledAndTheStopFollowsInTheSameCycle · TestA198TheProductionAssemblyCancelsTheWorkingBuyAndPlacesTheStop | FAIL(취소 0) | PASS |
| B12 | 접수된 매도 취소 | ArmedSell | FAIL | PASS |
| B13 | 미치움 반환 | 위 실패 시험들 | — | PASS |
| B14 | 발의 있음 | ArmedSell | — | PASS |
| B15 | 발의 intent 없음 | 없음 | n/a | 미덮음 |
| B16 | 해제 판정 오류 | 없음 | n/a | 미덮음 |
| B17 | 해제 안 됨 | TestA094ALiveProposalAttemptIsNotReleased | n/a | 기존 green |
