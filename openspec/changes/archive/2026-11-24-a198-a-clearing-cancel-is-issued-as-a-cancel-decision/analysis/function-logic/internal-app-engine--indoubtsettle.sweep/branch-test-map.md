# Branch Test Map: `inDoubtSettle.sweep` (post-edit 번호)

시험 이름은 grep 근거(internal/app/engine/a193_in_doubt_settle_test.go · a198_in_doubt_cancel_settle_test.go · a198_in_doubt_cancel_close_test.go). RED 는 `analysis/tdd/red-4c.txt`, GREEN 은 4c.8 선택 시험 원문(`analysis/verify/4c-*.txt`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 원장 읽기 실패 | TestA193AnUnreadableLedgerAlertsAfterTenMinutes | n/a(기존) | PASS |
| B2 | 읽기 실패 + ctx 취소 | TestA193AShutdownMidResolutionIsAGracefulAuxiliaryStop | n/a(기존) | PASS |
| B3 | 순회 | TestA193OnlyInDoubtPlaceAttemptsAreTouched | n/a(기존) | PASS |
| B4 | 대상 거름 switch | TestA193OnlyInDoubtPlaceAttemptsAreTouched · TestA198TheSettlerLeavesFlattenAndAckedCancelsAlone | — | PASS |
| B5 | IN_DOUBT PLACE → 해소 대상 | TestA193AnInDoubtPlaceFoundInTheListIsConfirmedWhileTheEngineRuns | n/a(기존) | PASS |
| B6 | (새) IN_DOUBT CANCEL → 닫기 목록 | TestA198TheSettlerClosesAnEngineCancelWithoutBrokerOrAlert · TestA198AnInDoubtBuyCancelClosesOnItsClosingRecordAndTheStopFollows | FAIL | PASS |
| B7 | 1단 알림(PLACE 만) | TestA193FirstSightIsAlertedOnce · TestA198TheSettlerClosesAnEngineCancelWithoutBrokerOrAlert(취소 알림 0) | FAIL(닫기 단언) | PASS |
| B8 | (새) 취소 닫기 — 발주 해소보다 먼저 | TestA198TheCancelIsClosedBeforeThePlaceResolution | FAIL | PASS |
| B9 | (새) 닫기 오류(ctx) 반환 | **not-covered**(구현 리뷰 r3 P2-4, Manager 처분: 기록만) — closeCancel 은 ctx 취소일 때만 오류를 돌려주고, 원장 함수는 `BeginTx(ctx)` 라 ctx 취소 시 롤백되어 원자적(부분 쓰기 없음). 반환된 ctx 오류는 스윕을 끝낼 뿐(Run 이 정상 정지로 읽음) | n/a | not-covered |
| B10 | 2단 직렬 해소 | TestA193AttemptsAreResolvedOneAtATime | n/a(기존) | PASS |
| B11 | 해소 중 ctx 취소 | TestA193AShutdownMidResolutionIsAGracefulAuxiliaryStop | n/a(기존) | PASS |
