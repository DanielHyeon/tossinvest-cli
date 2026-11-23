# Branch Test Map: `ExitObserver.clearTheSymbol` (post-edit 2판 — 보강 4b.7 뒤, 2판 번호)

RED 원문: `analysis/tdd/red.txt`(base production) · `analysis/tdd/red-4b.txt`(1판 production). GREEN: 4b.8 선택 시험 원문(`analysis/verify/4b-*.txt`). 기존 시험 덮음은 pre-edit 판의 선택 실행 coverprofile 기준 — post-edit 커버리지 재측정 없음. 번호 대응은 function-logic-map.md 「재번호 2판」.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 목록 읽기 실패 | TestA094AListReadFailureIsAnErrorAsBefore | n/a(기존) | 4b.8 |
| B2 | 미종결 읽기 실패 | 없음 | n/a | 미덮음 |
| B3 | 같은 종목 미종결(IN_DOUBT·ACKED 취소, already-modified) | TestA198AnInDoubtCancelHoldsTheStopAndBothAlertsFire · TestA198AnAckedCancelHoldsTheStopAndTheDelayAlertStillFires · TestA198AnAlreadyModifiedCancelStaysInDoubt | InDoubt: FAIL(red.txt) | PASS |
| B4 | 순회 | 다수 | — | PASS |
| B5 | ConfirmedCancelOf 오류(매수·매도) | 없음(읽기 오류 주입 수단 없음 — review.md 미실행) | n/a | 미덮음 |
| B6 | 엔진 취소 접수 확정 있음 → 재취소 0 (매수 D8 · 매도) | TestA198AnAcceptedBuyCancelIsNotSentAgainAndTheStopIsResubmitted · TestA198OnlyTheRefusedBuyIsCancelledAgain · TestA198AnArmedSellIsCancelledOnceAndTheStopWaitsForItsClosingRecord | 매수: FAIL(red-4b.txt 재취소) | PASS |
| B7 | (B6 안) 매도 → 종결 증거 대기 | TestA198AnArmedSellIsCancelledOnceAndTheStopWaitsForItsClosingRecord · TestA094ACancelledSellIsNotClearedUntilItsClosingRecord | n/a | PASS |
| B8 | 매도 && !withPending | TestA094AnOwnSideSellDoesNotWithholdAStopWithoutAProposal | n/a | 4b.8 |
| B9 | 수량·가격 해석 실패 → 실패 줄 1 | TestA198AnUnreadableWorkingOrderIsLoggedOnce(quantity · price) | FAIL(red-4b.txt 줄 0) | PASS |
| B10 | 취소 결정 발행 실패 | TestA198AnUnrecordableCancelDecisionIsLoggedOnce · TestA198OnlyTheCancelDecisionFailingStillWithholdsTheStop | 줄: FAIL(red.txt) | PASS |
| B11 | 취소 거절·모호 → 실패 줄(같은 결과 1, 바뀌면 2, 쓰기 패닉 뒤 재기록) · 409 세 code 는 확정 거절 | TestA198ARefusedCancelIsLoggedOnceWithoutAccountQuantityPriceOrDetail · TestA198AFailureLineWhoseWritePanickedIsWrittenNextCycle · TestA198AnAlreadyFilledBuyIsADefinitiveRefusalAndTheStopFollowsItsClosingRecord · TestA198AnAlreadyCanceledAfterAlreadyProcessingIsDefinitive | FAIL(red.txt · red-4b.txt) | PASS |
| B12 | 접수된 매도 취소 | TestA198AnArmedSellIsCancelledOnceAndTheStopWaitsForItsClosingRecord | FAIL(red.txt) | PASS |
| B13 | 미치움 반환 | 위 실패 시험들 | — | PASS |
| B14 | 발의 있음 | ArmedSell 시험 | — | PASS |
| B15 | 발의 intent 없음 | 없음 | n/a | 미덮음 |
| B16 | 해제 판정 오류 | 없음 | n/a | 미덮음 |
| B17 | 해제 안 됨 | TestA094ALiveProposalAttemptIsNotReleased | n/a | 4b.8 |
